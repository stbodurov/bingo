package proof

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

// A Presentation answers a verifier's request ("is this a unique human over 18?") with one
// ring signature per requested predicate set, all under the verifier's scope. Every proof
// must carry the same key image, which becomes the user's pseudonym (nullifier) at that app.

const (
	PresentationVersion = 1
	MaxRing             = 2048
)

type PredicateProof struct {
	Set        string `json:"set"`        // e.g. "human", "age_over_18"
	BucketBits int    `json:"bucketBits"` // depth of the anonymity bucket in the prefix tree
	Prefix     string `json:"prefix"`     // hex, bits beyond BucketBits are zero
	Signature  string `json:"signature"`  // base64url bLSAG
}

type Presentation struct {
	Version   int              `json:"v"`
	Scope     string           `json:"scope"`
	Challenge string           `json:"challenge"`
	Epoch     int64            `json:"epoch"`
	Proofs    []PredicateProof `json:"proofs"`
}

// SetSource gives verifiers the published members of a predicate set at an epoch.
type SetSource interface {
	Members(epoch int64, set string) ([][]byte, error)
}

// Result of a successful verification.
type Verified struct {
	Nullifier  string   `json:"nullifier"`
	KeyImage   string   `json:"keyImage"`
	Predicates []string `json:"predicates"`
	RingSizes  []int    `json:"ringSizes"`
	Epoch      int64    `json:"epoch"`
}

func presentationMessage(scope, challenge string, epoch int64, set string, nbits int, prefix string) []byte {
	var b bytes.Buffer
	b.WriteString("singleton/presentation/v1\x00")
	b.WriteString(scope)
	b.WriteByte(0)
	b.WriteString(challenge)
	b.WriteByte(0)
	var e [8]byte
	binary.BigEndian.PutUint64(e[:], uint64(epoch))
	b.Write(e[:])
	b.WriteString(set)
	b.WriteByte(0)
	binary.BigEndian.PutUint64(e[:], uint64(nbits))
	b.Write(e[:])
	b.WriteString(prefix)
	return b.Bytes()
}

func maskedPrefix(key []byte, nbits int) string {
	n := (nbits + 7) / 8
	p := append([]byte(nil), key[:n]...)
	if rem := nbits % 8; rem != 0 {
		p[n-1] &= byte(0xFF << (8 - rem))
	}
	return hex.EncodeToString(p)
}

// BucketByPrefix returns the sorted ring for a published prefix.
func BucketByPrefix(set [][]byte, prefixHex string, nbits int) ([][]byte, error) {
	prefix, err := hex.DecodeString(prefixHex)
	if err != nil || nbits < 0 || nbits > 8*len(prefix) || len(prefix) != (nbits+7)/8 {
		return nil, errors.New("proof: malformed bucket prefix")
	}
	if maskedPrefix(append(prefix, make([]byte, 32)...), nbits) != prefixHex {
		return nil, errors.New("proof: prefix has bits beyond bucket depth")
	}
	padded := append(append([]byte(nil), prefix...), make([]byte, 32)...)
	var ring [][]byte
	for _, p := range set {
		if prefixMatch(bucketKey(p), padded, nbits) {
			ring = append(ring, p)
		}
	}
	sort.Slice(ring, func(i, j int) bool { return bytes.Compare(ring[i], ring[j]) < 0 })
	return ring, nil
}

func decodeRing(raw [][]byte) ([]*Element, error) {
	ring := make([]*Element, len(raw))
	for i, r := range raw {
		e, err := DecodeElement(r)
		if err != nil {
			return nil, fmt.Errorf("proof: bad ring member %d", i)
		}
		ring[i] = e
	}
	return ring, nil
}

// Prove builds a presentation for the requested sets. sets maps set name → members.
func Prove(key *KeyPair, scope, challenge string, epoch int64, sets map[string][][]byte, order []string) (*Presentation, error) {
	if len(order) == 0 {
		return nil, errors.New("proof: no predicates requested")
	}
	me := key.Public.Encode(nil)
	p := &Presentation{Version: PresentationVersion, Scope: scope, Challenge: challenge, Epoch: epoch}
	for _, name := range order {
		members, ok := sets[name]
		if !ok {
			return nil, fmt.Errorf("proof: set %q unavailable", name)
		}
		nbits, rawRing := Bucket(members, me, BucketTarget)
		found := false
		for _, r := range rawRing {
			if bytes.Equal(r, me) {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("proof: you are not in the %q set for epoch %d", name, epoch)
		}
		ring, err := decodeRing(rawRing)
		if err != nil {
			return nil, err
		}
		prefix := maskedPrefix(bucketKey(me), nbits)
		sig, err := Sign(presentationMessage(scope, challenge, epoch, name, nbits, prefix), scope, ring, key)
		if err != nil {
			return nil, err
		}
		p.Proofs = append(p.Proofs, PredicateProof{Set: name, BucketBits: nbits, Prefix: prefix, Signature: sig.String()})
	}
	return p, nil
}

// VerifyPresentation checks every predicate proof against the published sets and requires
// that all proofs share one key image. It returns the app-scoped nullifier.
func VerifyPresentation(p *Presentation, wantScope, wantChallenge string, required []string, src SetSource, minRing int) (*Verified, error) {
	if p == nil || p.Version != PresentationVersion {
		return nil, errors.New("proof: unsupported presentation version")
	}
	if p.Scope != wantScope {
		return nil, errors.New("proof: presentation is for a different app")
	}
	if p.Challenge != wantChallenge {
		return nil, errors.New("proof: challenge mismatch (replay?)")
	}
	if minRing < 2 {
		minRing = 2
	}
	got := map[string]bool{}
	var image *Element
	out := &Verified{Epoch: p.Epoch}
	for _, pp := range p.Proofs {
		if got[pp.Set] {
			return nil, fmt.Errorf("proof: duplicate proof for %q", pp.Set)
		}
		members, err := src.Members(p.Epoch, pp.Set)
		if err != nil {
			return nil, err
		}
		rawRing, err := BucketByPrefix(members, pp.Prefix, pp.BucketBits)
		if err != nil {
			return nil, err
		}
		if len(rawRing) < minRing || len(rawRing) > MaxRing {
			return nil, fmt.Errorf("proof: anonymity bucket for %q has %d members (need %d–%d)", pp.Set, len(rawRing), minRing, MaxRing)
		}
		ring, err := decodeRing(rawRing)
		if err != nil {
			return nil, err
		}
		sig, err := ParseSignature(pp.Signature, MaxRing)
		if err != nil {
			return nil, err
		}
		if err := Verify(presentationMessage(p.Scope, p.Challenge, p.Epoch, pp.Set, pp.BucketBits, pp.Prefix), p.Scope, ring, sig); err != nil {
			return nil, fmt.Errorf("proof: %q proof invalid: %w", pp.Set, err)
		}
		if image == nil {
			image = sig.KeyImage
		} else if image.Equal(sig.KeyImage) != 1 {
			return nil, errors.New("proof: predicate proofs come from different holders")
		}
		got[pp.Set] = true
		out.Predicates = append(out.Predicates, pp.Set)
		out.RingSizes = append(out.RingSizes, len(rawRing))
	}
	for _, r := range required {
		if !got[r] {
			return nil, fmt.Errorf("proof: missing required predicate %q", r)
		}
	}
	if image == nil {
		return nil, errors.New("proof: empty presentation")
	}
	out.KeyImage = ElementHex(image)
	out.Nullifier = Nullifier(image)
	return out, nil
}

// ---------------------------------------------------------------- agent delegation

// Delegation lets a verified human deputise an AI agent inside one app's scope. The app
// learns "agent K acts for pseudonym N until T with these permissions" — never who N is.
type Delegation struct {
	Version     int      `json:"v"`
	Scope       string   `json:"scope"`
	Agent       string   `json:"agent"` // agent's Ed25519 public key, hex
	Label       string   `json:"label"`
	Permissions []string `json:"permissions"`
	IssuedAt    int64    `json:"issuedAt"`
	NotAfter    int64    `json:"notAfter"`
}

// Challenge binds a presentation to this exact delegation statement.
func (d Delegation) Challenge() string {
	b, _ := json.Marshal(d)
	sum := sha256.Sum256(append([]byte("singleton/delegation/v1\x00"), b...))
	return hex.EncodeToString(sum[:])
}

// PrefixFor returns the masked bucket prefix (hex) for a member at depth nbits.
func PrefixFor(member []byte, nbits int) string { return maskedPrefix(bucketKey(member), nbits) }

// BucketData is a pre-fetched anonymity bucket, as a wallet receives it from the registry.
type BucketData struct {
	Bits    int      `json:"bits"`
	Prefix  string   `json:"prefix"`
	Members [][]byte `json:"-"`
}

// ProveBuckets is Prove for wallets that fetch only their own bucket per set (instead of
// downloading whole sets). The wallet checks that the bucket matches its own prefix and
// contains its key before signing, so a malicious server cannot shrink the ring unnoticed
// below minRing.
func ProveBuckets(key *KeyPair, scope, challenge string, epoch int64, buckets map[string]BucketData, order []string, minRing int) (*Presentation, error) {
	me := key.Public.Encode(nil)
	p := &Presentation{Version: PresentationVersion, Scope: scope, Challenge: challenge, Epoch: epoch}
	for _, name := range order {
		b, ok := buckets[name]
		if !ok {
			return nil, fmt.Errorf("proof: bucket for %q missing", name)
		}
		if b.Prefix != maskedPrefix(bucketKey(me), b.Bits) {
			return nil, fmt.Errorf("proof: bucket for %q is not yours", name)
		}
		if len(b.Members) < minRing {
			return nil, fmt.Errorf("proof: bucket for %q is too small (%d) to protect your anonymity", name, len(b.Members))
		}
		sorted := append([][]byte(nil), b.Members...)
		sort.Slice(sorted, func(i, j int) bool { return bytes.Compare(sorted[i], sorted[j]) < 0 })
		ring, err := decodeRing(sorted)
		if err != nil {
			return nil, err
		}
		sig, err := Sign(presentationMessage(scope, challenge, epoch, name, b.Bits, b.Prefix), scope, ring, key)
		if err != nil {
			return nil, err
		}
		p.Proofs = append(p.Proofs, PredicateProof{Set: name, BucketBits: b.Bits, Prefix: b.Prefix, Signature: sig.String()})
	}
	return p, nil
}
