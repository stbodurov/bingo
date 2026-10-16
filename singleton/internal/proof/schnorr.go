package proof

import (
	"encoding/base64"

	"github.com/gtank/ristretto255"
)

// Schnorr signatures over ristretto255 let a holder prove control of *their own* key, e.g.
// to revoke it or to poll an in-person enrollment. Unlike ring signatures they reveal the key,
// so they are only used for actions that already name the key.

type SchnorrSig struct {
	R *Element
	S *Scalar
}

func schnorrChallenge(pub, R *Element, msg []byte) *Scalar {
	return HashToScalar("singleton/schnorr/v1", pub.Encode(nil), R.Encode(nil), msg)
}

func (k *KeyPair) SignSchnorr(msg []byte) *SchnorrSig {
	r := RandomScalar()
	R := ristretto255.NewElement().ScalarBaseMult(r)
	c := schnorrChallenge(k.Public, R, msg)
	s := ristretto255.NewScalar().Add(r, ristretto255.NewScalar().Multiply(c, k.Secret))
	return &SchnorrSig{R: R, S: s}
}

func VerifySchnorr(pub *Element, msg []byte, sig *SchnorrSig) bool {
	if sig == nil || sig.R == nil || sig.S == nil {
		return false
	}
	c := schnorrChallenge(pub, sig.R, msg)
	// s·G == R + c·P
	lhs := ristretto255.NewElement().ScalarBaseMult(sig.S)
	rhs := ristretto255.NewElement().Add(sig.R, ristretto255.NewElement().ScalarMult(c, pub))
	return lhs.Equal(rhs) == 1
}

func (s *SchnorrSig) String() string {
	return base64.RawURLEncoding.EncodeToString(append(s.R.Encode(nil), s.S.Encode(nil)...))
}

func ParseSchnorr(v string) (*SchnorrSig, error) {
	b, err := base64.RawURLEncoding.DecodeString(v)
	if err != nil || len(b) != 64 {
		return nil, ErrDecode
	}
	R, err := DecodeElement(b[:32])
	if err != nil {
		return nil, err
	}
	S, err := DecodeScalar(b[32:])
	if err != nil {
		return nil, err
	}
	return &SchnorrSig{R: R, S: S}, nil
}
