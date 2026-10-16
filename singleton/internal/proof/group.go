// Package proof is Singleton's cryptographic core: ristretto255 keys, scoped linkable
// ring signatures (one stable, unlinkable pseudonym per app), Merkle commitments for
// the public human set, anonymity buckets, Shamir social recovery and agent delegation.
//
// The same package compiles to the server verifier and, via GOOS=js GOARCH=wasm, to
// the in-browser Singleton Pass wallet — one implementation, two runtimes.
package proof

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/gtank/ristretto255"
)

// Element and Scalar are aliases so callers never import ristretto255 directly.
type (
	Element = ristretto255.Element
	Scalar  = ristretto255.Scalar
)

var ErrDecode = errors.New("proof: invalid encoding")

// hashParts writes length-prefixed parts into SHA-512 so boundaries are unambiguous.
func hashParts(domain string, parts ...[]byte) []byte {
	h := sha512.New()
	var lenBuf [8]byte
	binary.BigEndian.PutUint64(lenBuf[:], uint64(len(domain)))
	h.Write(lenBuf[:])
	h.Write([]byte(domain))
	for _, p := range parts {
		binary.BigEndian.PutUint64(lenBuf[:], uint64(len(p)))
		h.Write(lenBuf[:])
		h.Write(p)
	}
	return h.Sum(nil)
}

// HashToScalar maps arbitrary input to a uniformly distributed scalar.
func HashToScalar(domain string, parts ...[]byte) *Scalar {
	return ristretto255.NewScalar().FromUniformBytes(hashParts(domain, parts...))
}

// HashToElement maps input to a group element with unknown discrete log (Elligator 2).
func HashToElement(domain string, parts ...[]byte) *Element {
	return ristretto255.NewElement().FromUniformBytes(hashParts(domain, parts...))
}

// RandomScalar returns a uniformly random non-zero scalar.
func RandomScalar() *Scalar {
	var buf [64]byte
	for {
		if _, err := rand.Read(buf[:]); err != nil {
			panic(fmt.Sprintf("proof: entropy failure: %v", err))
		}
		s := ristretto255.NewScalar().FromUniformBytes(buf[:])
		if s.Equal(ristretto255.NewScalar()) == 0 {
			return s
		}
	}
}

// ScalarFromUint64 encodes a small integer as a scalar (used for Shamir x-coordinates).
func ScalarFromUint64(v uint64) *Scalar {
	var b [32]byte
	binary.LittleEndian.PutUint64(b[:8], v)
	s, err := ristretto255.NewScalar().SetCanonicalBytes(b[:])
	if err != nil {
		panic(err)
	}
	return s
}

func DecodeElement(b []byte) (*Element, error) {
	e := ristretto255.NewElement()
	if err := e.Decode(b); err != nil {
		return nil, ErrDecode
	}
	return e, nil
}

func DecodeScalar(b []byte) (*Scalar, error) {
	s := ristretto255.NewScalar()
	if err := s.Decode(b); err != nil {
		return nil, ErrDecode
	}
	return s, nil
}

func DecodeElementHex(h string) (*Element, error) {
	b, err := hex.DecodeString(h)
	if err != nil || len(b) != 32 {
		return nil, ErrDecode
	}
	return DecodeElement(b)
}

func ElementHex(e *Element) string { return hex.EncodeToString(e.Encode(nil)) }
func ScalarHex(s *Scalar) string   { return hex.EncodeToString(s.Encode(nil)) }

// Base returns the ristretto255 generator G.
func Base() *Element { return ristretto255.NewGeneratorElement() }

func identity() *Element { return ristretto255.NewIdentityElement() }
