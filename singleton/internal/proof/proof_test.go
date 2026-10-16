package proof

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

func ring(n int) ([]*KeyPair, []*Element) {
	keys := make([]*KeyPair, n)
	pubs := make([]*Element, n)
	for i := range keys {
		keys[i] = NewKeyPair()
		pubs[i] = keys[i].Public
	}
	return keys, pubs
}

func TestSignVerify(t *testing.T) {
	keys, pubs := ring(8)
	msg := []byte("hello")
	sig, err := Sign(msg, "app:driftwood", pubs, keys[5])
	if err != nil {
		t.Fatal(err)
	}
	if err := Verify(msg, "app:driftwood", pubs, sig); err != nil {
		t.Fatalf("valid signature rejected: %v", err)
	}
	// Round-trip encoding.
	parsed, err := ParseSignature(sig.String(), 64)
	if err != nil {
		t.Fatal(err)
	}
	if err := Verify(msg, "app:driftwood", pubs, parsed); err != nil {
		t.Fatalf("decoded signature rejected: %v", err)
	}
}

func TestTamperingFails(t *testing.T) {
	keys, pubs := ring(6)
	sig, _ := Sign([]byte("m"), "s", pubs, keys[2])
	if Verify([]byte("m2"), "s", pubs, sig) == nil {
		t.Fatal("message tampering not detected")
	}
	if Verify([]byte("m"), "other-scope", pubs, sig) == nil {
		t.Fatal("scope tampering not detected")
	}
	swapped := append([]*Element{pubs[1], pubs[0]}, pubs[2:]...)
	if Verify([]byte("m"), "s", swapped, sig) == nil {
		t.Fatal("ring reordering not detected")
	}
	_, other := ring(6)
	if Verify([]byte("m"), "s", other, sig) == nil {
		t.Fatal("foreign ring accepted")
	}
	// Forged key image (claims a different pseudonym) must fail.
	forged := *sig
	forged.KeyImage = NewKeyPair().KeyImage("s")
	if Verify([]byte("m"), "s", pubs, &forged) == nil {
		t.Fatal("forged key image accepted")
	}
}

func TestNonMemberCannotSign(t *testing.T) {
	_, pubs := ring(5)
	outsider := NewKeyPair()
	if _, err := Sign([]byte("m"), "s", pubs, outsider); !errors.Is(err, ErrSignerNotInRing) {
		t.Fatalf("expected ErrSignerNotInRing, got %v", err)
	}
}

func TestLinkabilityWithinScopeUnlinkableAcross(t *testing.T) {
	keys, pubs := ring(10)
	a, _ := Sign([]byte("signup-1"), "app:a", pubs, keys[3])
	b, _ := Sign([]byte("signup-2"), "app:a", pubs, keys[3])
	c, _ := Sign([]byte("signup-3"), "app:b", pubs, keys[3])
	d, _ := Sign([]byte("signup-4"), "app:a", pubs, keys[4])
	if a.KeyImage.Equal(b.KeyImage) != 1 {
		t.Fatal("same holder + same scope must link")
	}
	if a.KeyImage.Equal(c.KeyImage) == 1 {
		t.Fatal("same holder in different scopes must not link")
	}
	if a.KeyImage.Equal(d.KeyImage) == 1 {
		t.Fatal("different holders must not link")
	}
	if Nullifier(a.KeyImage) != Nullifier(keys[3].KeyImage("app:a")) {
		t.Fatal("nullifier must be derivable by the holder")
	}
}

func TestSeedDeterminism(t *testing.T) {
	k := NewKeyPair()
	k2 := KeyPairFromSeed(k.Seed)
	if k.Public.Equal(k2.Public) != 1 || Nullifier(k.KeyImage("x")) != Nullifier(k2.KeyImage("x")) {
		t.Fatal("seed must restore keys and pseudonyms")
	}
}

func TestMerkle(t *testing.T) {
	var pubs [][]byte
	for i := 0; i < 37; i++ {
		pubs = append(pubs, NewKeyPair().Public.Encode(nil))
	}
	tree := BuildMerkle(pubs)
	root := tree.RootHex()
	for _, p := range pubs {
		path, err := tree.Prove(p)
		if err != nil {
			t.Fatal(err)
		}
		if !VerifyInclusion(p, path, root) {
			t.Fatal("inclusion proof failed")
		}
	}
	if _, err := tree.Prove(NewKeyPair().Public.Encode(nil)); err == nil {
		t.Fatal("non-member proved")
	}
	// order independence
	rev := make([][]byte, len(pubs))
	for i := range pubs {
		rev[i] = pubs[len(pubs)-1-i]
	}
	if BuildMerkle(rev).RootHex() != root {
		t.Fatal("root must not depend on input order")
	}
	if BuildMerkle(nil).RootHex() == "" {
		t.Fatal("empty root must be defined")
	}
}

func TestBucketsAreCanonicalAndLargeEnough(t *testing.T) {
	var set [][]byte
	keys := map[string]*KeyPair{}
	for i := 0; i < 600; i++ {
		k := NewKeyPair()
		set = append(set, k.Public.Encode(nil))
		keys[string(k.Public.Encode(nil))] = k
	}
	for _, m := range set[:50] {
		nbits, ring := Bucket(set, m, BucketTarget)
		if len(ring) < BucketTarget/2 {
			t.Fatalf("bucket too small: %d", len(ring))
		}
		found := false
		for _, r := range ring {
			if bytes.Equal(r, m) {
				found = true
			}
		}
		if !found {
			t.Fatal("member missing from own bucket")
		}
		byPrefix, err := BucketByPrefix(set, maskedPrefix(bucketKey(m), nbits), nbits)
		if err != nil {
			t.Fatal(err)
		}
		if len(byPrefix) != len(ring) {
			t.Fatal("verifier must reconstruct the same ring")
		}
	}
}

type memSets map[string][][]byte

func (m memSets) Members(epoch int64, set string) ([][]byte, error) {
	if s, ok := m[fmt.Sprintf("%d/%s", epoch, set)]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("no set %s", set)
}

func TestPresentationEndToEnd(t *testing.T) {
	src := memSets{}
	var human, adults [][]byte
	var me, minor *KeyPair
	for i := 0; i < 300; i++ {
		k := NewKeyPair()
		human = append(human, k.Public.Encode(nil))
		if i%2 == 0 {
			adults = append(adults, k.Public.Encode(nil))
		}
		if i == 42 {
			me = k
		}
		if i == 43 {
			minor = k
		}
	}
	src["7/human"] = human
	src["7/age_over_18"] = adults
	sets := map[string][][]byte{"human": human, "age_over_18": adults}

	p, err := Prove(me, "app:front-row", "nonce-123", 7, sets, []string{"human", "age_over_18"})
	if err != nil {
		t.Fatal(err)
	}
	v, err := VerifyPresentation(p, "app:front-row", "nonce-123", []string{"human", "age_over_18"}, src, 8)
	if err != nil {
		t.Fatalf("verification failed: %v", err)
	}
	if v.Nullifier != Nullifier(me.KeyImage("app:front-row")) {
		t.Fatal("unexpected nullifier")
	}
	if _, err := VerifyPresentation(p, "app:front-row", "nonce-999", nil, src, 8); err == nil {
		t.Fatal("replayed presentation accepted")
	}
	if _, err := VerifyPresentation(p, "app:other", "nonce-123", nil, src, 8); err == nil {
		t.Fatal("cross-app presentation accepted")
	}
	if _, err := VerifyPresentation(p, "app:front-row", "nonce-123", []string{"age_over_21"}, src, 8); err == nil {
		t.Fatal("missing predicate accepted")
	}
	// A minor is human but not in the adult set, so cannot prove age_over_18.
	if _, err := Prove(minor, "app:front-row", "n", 7, sets, []string{"age_over_18"}); err == nil {
		t.Fatal("minor proved age_over_18")
	}
	if _, err := Prove(minor, "app:front-row", "n", 7, sets, []string{"human"}); err != nil {
		t.Fatalf("minor should still prove humanity: %v", err)
	}
	// Splicing a proof from another holder must be rejected.
	other, _ := Prove(minor, "app:front-row", "nonce-123", 7, sets, []string{"human"})
	spliced := *p
	spliced.Proofs = []PredicateProof{other.Proofs[0], p.Proofs[1]}
	if _, err := VerifyPresentation(&spliced, "app:front-row", "nonce-123", nil, src, 8); err == nil {
		t.Fatal("spliced presentation from two holders accepted")
	}
	outsider := NewKeyPair()
	if _, err := Prove(outsider, "app:front-row", "n", 7, sets, []string{"human"}); err == nil {
		t.Fatal("non-member produced a presentation")
	}
}

func TestShamir(t *testing.T) {
	k := NewKeyPair()
	shares, err := Split(k.Seed, 5, 3)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Combine([]Share{shares[4], shares[0], shares[2]})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, k.Seed) {
		t.Fatal("3-of-5 reconstruction failed")
	}
	wrong, _ := Combine([]Share{shares[0], shares[1]})
	if bytes.Equal(wrong, k.Seed) {
		t.Fatal("2 shares must not reveal a 3-of-5 secret")
	}
	parsed, err := ParseShare(shares[1].String())
	if err != nil || parsed.X != shares[1].X || !bytes.Equal(parsed.Y, shares[1].Y) {
		t.Fatal("share encoding round-trip failed")
	}
}

func TestDelegationChallengeStable(t *testing.T) {
	d := Delegation{Version: 1, Scope: "app:a", Agent: "ab", Label: "Travel agent", Permissions: []string{"book"}, IssuedAt: 1, NotAfter: 2}
	if d.Challenge() != d.Challenge() || len(d.Challenge()) != 64 {
		t.Fatal("delegation challenge must be deterministic sha256 hex")
	}
	d2 := d
	d2.NotAfter = 3
	if d.Challenge() == d2.Challenge() {
		t.Fatal("challenge must bind every field")
	}
}

func BenchmarkVerifyRing64(b *testing.B) {
	keys, pubs := ring(64)
	sig, _ := Sign([]byte("m"), "s", pubs, keys[7])
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Verify([]byte("m"), "s", pubs, sig)
	}
}
