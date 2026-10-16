package proof

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// Social recovery: the 32-byte holder seed is split with Shamir's scheme over GF(2^8)
// (AES polynomial x^8+x^4+x^3+x+1), byte by byte. Any `threshold` guardians can restore
// the seed — and therefore every per-app pseudonym — while fewer learn nothing.

var gfExp, gfLog [256]byte

func init() {
	x := byte(1)
	for i := 0; i < 255; i++ {
		gfExp[i] = x
		gfLog[x] = byte(i)
		// multiply by generator 3
		x = x ^ xtime(x)
	}
	gfExp[255] = gfExp[0]
}

func xtime(b byte) byte {
	if b&0x80 != 0 {
		return (b << 1) ^ 0x1b
	}
	return b << 1
}

func gfMul(a, b byte) byte {
	if a == 0 || b == 0 {
		return 0
	}
	return gfExp[(int(gfLog[a])+int(gfLog[b]))%255]
}

func gfDiv(a, b byte) byte {
	if b == 0 {
		panic("gf division by zero")
	}
	if a == 0 {
		return 0
	}
	return gfExp[(int(gfLog[a])-int(gfLog[b])+255)%255]
}

// Share is one guardian's piece: x in 1..255 and the evaluated bytes.
type Share struct {
	X byte
	Y []byte
}

// String encodes a share as "sgl1-<x>-<hex>" for printing on a recovery card.
func (s Share) String() string { return fmt.Sprintf("sgl1-%02x-%s", s.X, hex.EncodeToString(s.Y)) }

func ParseShare(s string) (Share, error) {
	parts := strings.Split(strings.TrimSpace(s), "-")
	if len(parts) != 3 || parts[0] != "sgl1" {
		return Share{}, errors.New("proof: share must look like sgl1-xx-<hex>")
	}
	xb, err := hex.DecodeString(parts[1])
	if err != nil || len(xb) != 1 || xb[0] == 0 {
		return Share{}, errors.New("proof: bad share index")
	}
	y, err := hex.DecodeString(parts[2])
	if err != nil || len(y) == 0 {
		return Share{}, errors.New("proof: bad share body")
	}
	return Share{X: xb[0], Y: y}, nil
}

// Split divides secret into n shares, any threshold of which reconstruct it.
func Split(secret []byte, n, threshold int) ([]Share, error) {
	if threshold < 2 || threshold > n || n > 255 {
		return nil, errors.New("proof: need 2 ≤ threshold ≤ n ≤ 255")
	}
	shares := make([]Share, n)
	for i := range shares {
		shares[i] = Share{X: byte(i + 1), Y: make([]byte, len(secret))}
	}
	coeffs := make([]byte, threshold)
	for bi, sb := range secret {
		coeffs[0] = sb
		if _, err := rand.Read(coeffs[1:]); err != nil {
			return nil, err
		}
		for i := range shares {
			x := shares[i].X
			// Horner evaluation
			var y byte
			for d := threshold - 1; d >= 0; d-- {
				y = gfMul(y, x) ^ coeffs[d]
			}
			shares[i].Y[bi] = y
		}
	}
	return shares, nil
}

// Combine reconstructs the secret from ≥ threshold distinct shares (Lagrange at x=0).
func Combine(shares []Share) ([]byte, error) {
	if len(shares) < 2 {
		return nil, errors.New("proof: need at least two shares")
	}
	size := len(shares[0].Y)
	seen := map[byte]bool{}
	for _, s := range shares {
		if len(s.Y) != size {
			return nil, errors.New("proof: shares have different lengths")
		}
		if seen[s.X] {
			return nil, errors.New("proof: duplicate share")
		}
		seen[s.X] = true
	}
	out := make([]byte, size)
	for bi := 0; bi < size; bi++ {
		var acc byte
		for i, si := range shares {
			num, den := byte(1), byte(1)
			for j, sj := range shares {
				if i == j {
					continue
				}
				num = gfMul(num, sj.X)
				den = gfMul(den, si.X^sj.X)
			}
			acc ^= gfMul(si.Y[bi], gfDiv(num, den))
		}
		out[bi] = acc
	}
	return out, nil
}
