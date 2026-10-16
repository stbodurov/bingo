package proof

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/bits"
	"sort"
)

// The human set for an epoch is committed with a binary Merkle tree over sorted leaves
// leaf = SHA256(0x00 || pubkey), node = SHA256(0x01 || left || right); a lone node at the
// end of a level is promoted unchanged. Roots are anchored on-chain so anyone can check
// that the set we serve to verifiers is the set we committed to.

func leafHash(pub []byte) []byte {
	h := sha256.Sum256(append([]byte{0x00}, pub...))
	return h[:]
}

func nodeHash(l, r []byte) []byte {
	buf := make([]byte, 0, 65)
	buf = append(buf, 0x01)
	buf = append(buf, l...)
	buf = append(buf, r...)
	h := sha256.Sum256(buf)
	return h[:]
}

var emptyRoot = func() []byte { h := sha256.Sum256([]byte("singleton/empty-set/v1")); return h[:] }()

// MerkleTree holds every level so inclusion proofs are O(log n).
type MerkleTree struct {
	levels [][][]byte // levels[0] = sorted leaves
}

// BuildMerkle builds a tree from raw 32-byte public keys (order-independent).
func BuildMerkle(pubs [][]byte) *MerkleTree {
	leaves := make([][]byte, len(pubs))
	for i, p := range pubs {
		leaves[i] = leafHash(p)
	}
	sort.Slice(leaves, func(i, j int) bool { return bytes.Compare(leaves[i], leaves[j]) < 0 })
	t := &MerkleTree{levels: [][][]byte{leaves}}
	for cur := leaves; len(cur) > 1; {
		next := make([][]byte, 0, (len(cur)+1)/2)
		for i := 0; i < len(cur); i += 2 {
			if i+1 < len(cur) {
				next = append(next, nodeHash(cur[i], cur[i+1]))
			} else {
				next = append(next, cur[i])
			}
		}
		t.levels = append(t.levels, next)
		cur = next
	}
	return t
}

func (t *MerkleTree) Root() []byte {
	top := t.levels[len(t.levels)-1]
	if len(top) == 0 {
		return emptyRoot
	}
	return top[0]
}

func (t *MerkleTree) RootHex() string { return hex.EncodeToString(t.Root()) }

// ProofStep is one sibling on the path to the root.
type ProofStep struct {
	Sibling string `json:"sibling"`
	Left    bool   `json:"left"` // sibling is on the left
}

var ErrNotInTree = errors.New("proof: leaf not in tree")

// Prove returns the inclusion path for pub.
func (t *MerkleTree) Prove(pub []byte) ([]ProofStep, error) {
	leaf := leafHash(pub)
	idx := sort.Search(len(t.levels[0]), func(i int) bool { return bytes.Compare(t.levels[0][i], leaf) >= 0 })
	if idx >= len(t.levels[0]) || !bytes.Equal(t.levels[0][idx], leaf) {
		return nil, ErrNotInTree
	}
	var path []ProofStep
	for lvl := 0; lvl < len(t.levels)-1; lvl++ {
		nodes := t.levels[lvl]
		if idx%2 == 1 {
			path = append(path, ProofStep{Sibling: hex.EncodeToString(nodes[idx-1]), Left: true})
		} else if idx+1 < len(nodes) {
			path = append(path, ProofStep{Sibling: hex.EncodeToString(nodes[idx+1]), Left: false})
		}
		idx /= 2
	}
	return path, nil
}

// VerifyInclusion recomputes the root from pub and its path.
func VerifyInclusion(pub []byte, path []ProofStep, rootHex string) bool {
	cur := leafHash(pub)
	for _, s := range path {
		sib, err := hex.DecodeString(s.Sibling)
		if err != nil || len(sib) != 32 {
			return false
		}
		if s.Left {
			cur = nodeHash(sib, cur)
		} else {
			cur = nodeHash(cur, sib)
		}
	}
	return hex.EncodeToString(cur) == rootHex
}

// ---------------------------------------------------------------- anonymity buckets

// BucketTarget is the default ring size target. Rings are O(n) to sign and verify,
// so v1 signs over a hash-prefix bucket of the set instead of the whole set.
const BucketTarget = 64

// bucketKey is the stable position of a key in the prefix tree.
func bucketKey(pub []byte) []byte {
	h := sha256.Sum256(append([]byte("singleton/bucket/v1"), pub...))
	return h[:]
}

func prefixMatch(a, b []byte, nbits int) bool {
	full := nbits / 8
	if !bytes.Equal(a[:full], b[:full]) {
		return false
	}
	rem := nbits % 8
	if rem == 0 {
		return true
	}
	mask := byte(0xFF << (8 - rem))
	return a[full]&mask == b[full]&mask
}

// MaxBucketBits is the deepest prefix a set of size n supports for the target size.
func MaxBucketBits(n, target int) int {
	if target < 2 {
		target = 2
	}
	if n < 2*target {
		return 0
	}
	return bits.Len(uint(n/target)) - 1
}

// Bucket returns the canonical anonymity bucket containing member: the deepest prefix
// (≤ MaxBucketBits) whose bucket still has at least max(2, target/2) members. Prover and
// verifier run the same rule, so a prover cannot shrink their ring to deanonymise
// themselves or a verifier be tricked into accepting a tiny ring.
// Members are returned sorted by encoding, which fixes the ring order.
func Bucket(set [][]byte, member []byte, target int) (nbits int, ring [][]byte) {
	mk := bucketKey(member)
	keys := make([][]byte, len(set))
	for i, p := range set {
		keys[i] = bucketKey(p)
	}
	minSize := target / 2
	if minSize < 2 {
		minSize = 2
	}
	for b := MaxBucketBits(len(set), target); b >= 0; b-- {
		ring = ring[:0]
		for i, p := range set {
			if prefixMatch(keys[i], mk, b) {
				ring = append(ring, p)
			}
		}
		if len(ring) >= minSize || b == 0 {
			sort.Slice(ring, func(i, j int) bool { return bytes.Compare(ring[i], ring[j]) < 0 })
			return b, ring
		}
	}
	return 0, ring
}
