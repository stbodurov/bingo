package proof

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"github.com/gtank/ristretto255"
)

// KeyPair is a Singleton holder key. The secret never leaves the holder's device.
type KeyPair struct {
	Seed   []byte // 32 bytes; what social recovery protects
	Secret *Scalar
	Public *Element
}

// NewKeyPair creates a fresh holder key from 32 bytes of OS entropy.
func NewKeyPair() *KeyPair {
	seed := make([]byte, 32)
	if _, err := rand.Read(seed); err != nil {
		panic(err)
	}
	return KeyPairFromSeed(seed)
}

// KeyPairFromSeed deterministically derives the key so a recovered seed restores every
// per-app pseudonym (key images depend only on the secret and the scope).
func KeyPairFromSeed(seed []byte) *KeyPair {
	x := HashToScalar("singleton/holder-key/v1", seed)
	return &KeyPair{Seed: append([]byte(nil), seed...), Secret: x, Public: ristretto255.NewElement().ScalarBaseMult(x)}
}

// PublicHex is the 64-char hex encoding published in the human set.
func (k *KeyPair) PublicHex() string { return ElementHex(k.Public) }

// Fingerprint is a short, human-comparable identifier for a public key.
func Fingerprint(pub *Element) string {
	sum := sha256.Sum256(pub.Encode(nil))
	h := hex.EncodeToString(sum[:])
	return h[0:4] + "-" + h[4:8] + "-" + h[8:12]
}

// ScopeBase is H(scope): the per-app base point for key images. Distinct scopes give
// independent bases, so pseudonyms in different apps are unlinkable under DDH.
func ScopeBase(scope string) *Element {
	return HashToElement("singleton/scope/v1", []byte(scope))
}

// KeyImage I = x·H(scope) — the holder's stable pseudonym inside one scope.
func (k *KeyPair) KeyImage(scope string) *Element {
	return ristretto255.NewElement().ScalarMult(k.Secret, ScopeBase(scope))
}

// Nullifier is the hex pseudonym apps store (a hash of the key image, 32 bytes).
func Nullifier(keyImage *Element) string {
	sum := sha256.Sum256(append([]byte("singleton/nullifier/v1"), keyImage.Encode(nil)...))
	return hex.EncodeToString(sum[:])
}
