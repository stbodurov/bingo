package app

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/sha3"
)

// On-chain anchoring. Every sealed epoch's set roots are committed to the HumanRegistry
// contract (contracts/HumanRegistry.sol) on an Ethereum L2. The sandbox below produces the
// exact ABI calldata a real deployment would send, and records it in a local block log;
// pointing ANCHOR_RPC_URL at Base Sepolia replaces only the broadcast step.

const AnchorNetwork = "Base Sepolia (sandbox)"

func keccak(parts ...[]byte) []byte {
	h := sha3.NewLegacyKeccak256()
	for _, p := range parts {
		h.Write(p)
	}
	return h.Sum(nil)
}

// checksumAddress applies EIP-55 mixed-case checksumming.
func checksumAddress(raw []byte) string {
	lower := hex.EncodeToString(raw)
	hash := hex.EncodeToString(keccak([]byte(lower)))
	var b strings.Builder
	b.WriteString("0x")
	for i, c := range lower {
		if c >= 'a' && c <= 'f' && hash[i] >= '8' {
			b.WriteRune(c - 32)
		} else {
			b.WriteRune(c)
		}
	}
	return b.String()
}

var (
	RegistryAddress = checksumAddress(keccak([]byte("singleton/HumanRegistry/v1"))[12:])
	AnchorAddress   = checksumAddress(keccak([]byte("singleton/anchor-signer/v1"))[12:])
	anchorSelector  = keccak([]byte("anchorEpoch(uint64,bytes32[],bytes32[],uint64[])"))[:4]
	revokeSelector  = keccak([]byte("publishRevocations(uint64,bytes32)"))[:4]
)

type anchorRoot struct {
	Set  string
	Root string
	Size int
}

func word(b []byte) []byte {
	w := make([]byte, 32)
	copy(w[32-len(b):], b)
	return w
}

func uintWord(v uint64) []byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], v)
	return word(b[:])
}

// encodeAnchorCall ABI-encodes anchorEpoch(epoch, setIds, roots, sizes).
func encodeAnchorCall(epoch int64, roots []anchorRoot) []byte {
	n := uint64(len(roots))
	arrLen := 32 * (1 + n) // length word + elements
	head := 32 * 4
	out := append([]byte{}, anchorSelector...)
	out = append(out, uintWord(uint64(epoch))...)
	out = append(out, uintWord(uint64(head))...)
	out = append(out, uintWord(uint64(head)+arrLen)...)
	out = append(out, uintWord(uint64(head)+2*arrLen)...)
	out = append(out, uintWord(n)...)
	for _, r := range roots {
		out = append(out, keccak([]byte(r.Set))...)
	}
	out = append(out, uintWord(n)...)
	for _, r := range roots {
		b, _ := hex.DecodeString(r.Root)
		out = append(out, word(b)...)
	}
	out = append(out, uintWord(n)...)
	for _, r := range roots {
		out = append(out, uintWord(uint64(r.Size))...)
	}
	return out
}

func (s *Service) nextBlock(now time.Time) int64 {
	// Base produces a block every 2 seconds; derive a plausible height from wall time.
	genesis := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	h := int64(28_000_000) + int64(now.Sub(genesis)/(2*time.Second))
	var last int64
	_ = s.DB.QueryRow(`SELECT COALESCE(MAX(block), 0) FROM chain_txs`).Scan(&last)
	if h <= last {
		h = last + 1
	}
	return h
}

func (s *Service) recordTx(method string, calldata []byte, summary string, now time.Time) (string, error) {
	nonce := count(s.DB, `SELECT COUNT(*) FROM chain_txs`)
	hash := "0x" + hex.EncodeToString(keccak([]byte(AnchorAddress), []byte(RegistryAddress), uintWord(uint64(nonce)), calldata))
	_, err := s.DB.Exec(`INSERT INTO chain_txs(hash, block, from_addr, to_addr, method, calldata, summary, created_at) VALUES(?,?,?,?,?,?,?,?)`,
		hash, s.nextBlock(now), AnchorAddress, RegistryAddress, method, "0x"+hex.EncodeToString(calldata), summary, ts(now))
	return hash, err
}

func (s *Service) anchorEpoch(epoch int64, roots []anchorRoot, now time.Time) (string, error) {
	human := 0
	for _, r := range roots {
		if r.Set == "human" {
			human = r.Size
		}
	}
	return s.recordTx("anchorEpoch", encodeAnchorCall(epoch, roots), fmt.Sprintf("Epoch %d · %d sets · %d humans", epoch, len(roots), human), now)
}

// PublishRevocations anchors a commitment to the revocation list.
func (s *Service) PublishRevocations(epoch int64, now time.Time) (string, error) {
	rows, err := s.DB.Query(`SELECT pubkey FROM enrollments WHERE status='revoked' ORDER BY pubkey`)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	h := sha3.NewLegacyKeccak256()
	n := 0
	for rows.Next() {
		var p string
		_ = rows.Scan(&p)
		b, _ := hex.DecodeString(p)
		h.Write(b)
		n++
	}
	call := append(append([]byte{}, revokeSelector...), append(uintWord(uint64(epoch)), h.Sum(nil)...)...)
	return s.recordTx("publishRevocations", call, fmt.Sprintf("Revocation list · %d keys", n), now)
}

type ChainTx struct {
	Hash      string `json:"hash"`
	Block     int64  `json:"block"`
	From      string `json:"from"`
	To        string `json:"to"`
	Method    string `json:"method"`
	Calldata  string `json:"calldata"`
	Summary   string `json:"summary"`
	CreatedAt string `json:"createdAt"`
}

func (s *Service) ChainTxs(limit int) ([]ChainTx, error) {
	rows, err := s.DB.Query(`SELECT hash, block, from_addr, to_addr, method, calldata, summary, created_at FROM chain_txs ORDER BY block DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ChainTx
	for rows.Next() {
		var t ChainTx
		if err := rows.Scan(&t.Hash, &t.Block, &t.From, &t.To, &t.Method, &t.Calldata, &t.Summary, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Service) ChainTx(hash string) (*ChainTx, error) {
	var t ChainTx
	err := s.DB.QueryRow(`SELECT hash, block, from_addr, to_addr, method, calldata, summary, created_at FROM chain_txs WHERE hash=?`, hash).
		Scan(&t.Hash, &t.Block, &t.From, &t.To, &t.Method, &t.Calldata, &t.Summary, &t.CreatedAt)
	return &t, err
}
