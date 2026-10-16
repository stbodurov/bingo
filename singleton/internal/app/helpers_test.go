package app

import (
	"crypto/sha256"
	"math/rand/v2"
	"strconv"
)

func randFor(seed uint64) *rand.Rand { return rand.New(rand.NewPCG(seed, seed+1)) }

func solvePow(challenge string, bits int) string {
	for i := 0; ; i++ {
		n := strconv.Itoa(i)
		sum := sha256.Sum256([]byte(challenge + ":" + n))
		if leadingZeroBits(sum[:]) >= bits {
			return n
		}
	}
}
