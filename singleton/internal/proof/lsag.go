package proof

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"errors"

	"github.com/gtank/ristretto255"
)

// Signature is a scoped linkable ring signature (bLSAG, Back 2015 / Noether 2016) where
// the key image uses a *scope* base point instead of H(P). It proves, in zero knowledge,
// knowledge of x such that P_j = x·G for some ring member j AND I = x·H(scope).
//
//   - Same holder, same scope  → identical key image (sybil detection).
//   - Same holder, other scope → independent key image (no cross-app tracking).
//   - Verifier learns nothing about j beyond membership in the ring.
type Signature struct {
	C0       *Scalar
	R        []*Scalar
	KeyImage *Element
}

var (
	ErrRingTooSmall   = errors.New("proof: ring must have at least 2 members")
	ErrSignerNotInRing = errors.New("proof: signer public key not in ring")
	ErrBadSignature   = errors.New("proof: signature does not verify")
)

const lsagDomain = "singleton/lsag/v1"

func ringDigest(ring []*Element) []byte {
	h := sha512.New()
	var n [8]byte
	binary.BigEndian.PutUint64(n[:], uint64(len(ring)))
	h.Write(n[:])
	for _, p := range ring {
		h.Write(p.Encode(nil))
	}
	return h.Sum(nil)
}

func challenge(msg []byte, scope string, ringHash []byte, keyImage, L, R *Element) *Scalar {
	return HashToScalar(lsagDomain, msg, []byte(scope), ringHash, keyImage.Encode(nil), L.Encode(nil), R.Encode(nil))
}

// Sign produces a signature on msg for the given scope using the holder key, which must be
// in ring. The ring order is part of the statement; verifiers must use the same order.
func Sign(msg []byte, scope string, ring []*Element, key *KeyPair) (*Signature, error) {
	n := len(ring)
	if n < 2 {
		return nil, ErrRingTooSmall
	}
	pi := -1
	for i, p := range ring {
		if p.Equal(key.Public) == 1 {
			pi = i
			break
		}
	}
	if pi < 0 {
		return nil, ErrSignerNotInRing
	}
	hs := ScopeBase(scope)
	img := ristretto255.NewElement().ScalarMult(key.Secret, hs)
	rh := ringDigest(ring)

	c := make([]*Scalar, n)
	r := make([]*Scalar, n)
	alpha := RandomScalar()
	L := ristretto255.NewElement().ScalarBaseMult(alpha)
	R := ristretto255.NewElement().ScalarMult(alpha, hs)
	c[(pi+1)%n] = challenge(msg, scope, rh, img, L, R)

	for k := 1; k < n; k++ {
		i := (pi + k) % n
		r[i] = RandomScalar()
		// L_i = r_i·G + c_i·P_i ; R_i = r_i·H + c_i·I
		Li := ristretto255.NewElement().VarTimeDoubleScalarBaseMult(c[i], ring[i], r[i])
		Ri := ristretto255.NewElement().VarTimeMultiScalarMult([]*Scalar{r[i], c[i]}, []*Element{hs, img})
		c[(i+1)%n] = challenge(msg, scope, rh, img, Li, Ri)
	}
	// r_π = α − c_π·x
	r[pi] = ristretto255.NewScalar().Subtract(alpha, ristretto255.NewScalar().Multiply(c[pi], key.Secret))
	return &Signature{C0: c[0], R: r, KeyImage: img}, nil
}

// Verify checks sig over msg/scope/ring. It does not check ring membership against the
// published human set — callers must build the ring from the set (see Bucket).
func Verify(msg []byte, scope string, ring []*Element, sig *Signature) error {
	n := len(ring)
	if n < 2 {
		return ErrRingTooSmall
	}
	if sig == nil || sig.C0 == nil || sig.KeyImage == nil || len(sig.R) != n {
		return ErrBadSignature
	}
	if sig.KeyImage.Equal(identity()) == 1 {
		return ErrBadSignature
	}
	hs := ScopeBase(scope)
	rh := ringDigest(ring)
	c := ristretto255.NewScalar().Set(sig.C0)
	for i := 0; i < n; i++ {
		if sig.R[i] == nil {
			return ErrBadSignature
		}
		Li := ristretto255.NewElement().VarTimeDoubleScalarBaseMult(c, ring[i], sig.R[i])
		Ri := ristretto255.NewElement().VarTimeMultiScalarMult([]*Scalar{sig.R[i], c}, []*Element{hs, sig.KeyImage})
		c = challenge(msg, scope, rh, sig.KeyImage, Li, Ri)
	}
	if c.Equal(sig.C0) != 1 {
		return ErrBadSignature
	}
	return nil
}

// Encode serialises as: keyImage(32) || c0(32) || n(4, BE) || r_0..r_{n-1}(32 each).
func (s *Signature) Encode() []byte {
	out := make([]byte, 0, 68+32*len(s.R))
	out = append(out, s.KeyImage.Encode(nil)...)
	out = append(out, s.C0.Encode(nil)...)
	var n [4]byte
	binary.BigEndian.PutUint32(n[:], uint32(len(s.R)))
	out = append(out, n[:]...)
	for _, r := range s.R {
		out = append(out, r.Encode(nil)...)
	}
	return out
}

func (s *Signature) String() string { return base64.RawURLEncoding.EncodeToString(s.Encode()) }

// DecodeSignature parses Encode output; maxRing guards against oversized inputs.
func DecodeSignature(b []byte, maxRing int) (*Signature, error) {
	if len(b) < 68 {
		return nil, ErrDecode
	}
	img, err := DecodeElement(b[0:32])
	if err != nil {
		return nil, err
	}
	c0, err := DecodeScalar(b[32:64])
	if err != nil {
		return nil, err
	}
	n := int(binary.BigEndian.Uint32(b[64:68]))
	if n < 2 || n > maxRing || len(b) != 68+32*n {
		return nil, ErrDecode
	}
	rs := make([]*Scalar, n)
	for i := 0; i < n; i++ {
		off := 68 + 32*i
		if rs[i], err = DecodeScalar(b[off : off+32]); err != nil {
			return nil, err
		}
	}
	return &Signature{C0: c0, R: rs, KeyImage: img}, nil
}

func ParseSignature(s string, maxRing int) (*Signature, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, ErrDecode
	}
	return DecodeSignature(b, maxRing)
}
