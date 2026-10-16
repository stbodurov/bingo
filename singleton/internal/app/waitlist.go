package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/bits"
	"net/mail"
	"strings"
	"sync"
	"time"
)

// Singleton's waitlist practises what it preaches: no CAPTCHA and no third-party bot
// scoring. The browser solves a small hashcash proof-of-work in a Web Worker; the server
// verifies a stateless HMAC-signed challenge. Two tracks: people (want a Pass) and
// platforms (want to integrate). Platforms are prioritised by urgency and fit; people
// move up five places per friend who joins.

const (
	PowBits          = 18
	powTTL           = 10 * time.Minute
	SpotsPerReferral = 5
)

type PowChallenge struct {
	Challenge string `json:"challenge"`
	Bits      int    `json:"bits"`
	Sig       string `json:"sig"`
}

func (s *Service) NewPowChallenge() PowChallenge {
	var raw [24]byte
	copy(raw[:16], []byte(randomHex(8)))
	binary.BigEndian.PutUint64(raw[16:], uint64(time.Now().Unix()))
	c := base64.RawURLEncoding.EncodeToString(raw[:])
	return PowChallenge{Challenge: c, Bits: PowBits, Sig: s.powSig(c)}
}

func (s *Service) powSig(c string) string {
	m := hmac.New(sha256.New, s.powKey)
	m.Write([]byte(c))
	return hex.EncodeToString(m.Sum(nil))
}

var usedPow sync.Map

func leadingZeroBits(b []byte) int {
	n := 0
	for _, x := range b {
		if x == 0 {
			n += 8
			continue
		}
		return n + bits.LeadingZeros8(x)
	}
	return n
}

// VerifyPow checks signature, freshness, single use and work.
func (s *Service) VerifyPow(challenge, sig, nonce string) (int, error) {
	if !hmac.Equal([]byte(sig), []byte(s.powSig(challenge))) {
		return 0, errors.New("invalid challenge — reload the page")
	}
	raw, err := base64.RawURLEncoding.DecodeString(challenge)
	if err != nil || len(raw) != 24 {
		return 0, errors.New("malformed challenge")
	}
	issued := time.Unix(int64(binary.BigEndian.Uint64(raw[16:])), 0)
	if time.Since(issued) > powTTL {
		return 0, errors.New("challenge expired — reload the page")
	}
	if _, loaded := usedPow.LoadOrStore(challenge, issued); loaded {
		return 0, errors.New("challenge already used")
	}
	sum := sha256.Sum256([]byte(challenge + ":" + nonce))
	z := leadingZeroBits(sum[:])
	if z < PowBits {
		return z, fmt.Errorf("proof-of-work too weak (%d bits)", z)
	}
	return z, nil
}

// SweepPow forgets expired challenges.
func SweepPow() {
	usedPow.Range(func(k, v any) bool {
		if time.Since(v.(time.Time)) > powTTL {
			usedPow.Delete(k)
		}
		return true
	})
}

type WaitlistInput struct {
	Email    string `json:"email"`
	Track    string `json:"track"` // "person" | "platform"
	Company  string `json:"company"`
	Role     string `json:"role"`
	UseCase  string `json:"useCase"`
	Volume   string `json:"volume"`
	Region   string `json:"region"`
	Ref      string `json:"ref"`
	Website  string `json:"website"` // honeypot
	Pow      struct {
		Challenge string `json:"challenge"`
		Sig       string `json:"sig"`
		Nonce     string `json:"nonce"`
	} `json:"pow"`
}

type WaitlistStatus struct {
	Code      string `json:"code"`
	Track     string `json:"track"`
	Position  int    `json:"position"`
	Total     int    `json:"total"`
	Referrals int    `json:"referrals"`
	ShareURL  string `json:"shareUrl"`
	Existing  bool   `json:"existing"`
	Tier      string `json:"tier"`
	PowBits   int    `json:"powBits"`
}

var volumes = map[string]int{"<10k": 1, "10k-100k": 3, "100k-1m": 6, ">1m": 10}
var urgentRegions = map[string]int{"au": 4, "gb": 4, "eu": 2, "us": 2}
var hotUseCases = map[string]int{"dating": 2, "social": 2, "ticketing": 2, "marketplace": 1, "gaming": 1, "ai_agents": 2, "civic": 1, "fintech": 1}

func platformPriority(in WaitlistInput) int {
	return volumes[in.Volume] + urgentRegions[in.Region] + hotUseCases[in.UseCase]
}

func (s *Service) JoinWaitlist(in WaitlistInput, ipHash string) (*WaitlistStatus, error) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if _, err := mail.ParseAddress(in.Email); err != nil || len(in.Email) > 254 {
		return nil, errors.New("enter a valid email")
	}
	if in.Track != "platform" {
		in.Track = "person"
	}
	if in.Website != "" {
		return &WaitlistStatus{Code: "BOT", Track: in.Track, Position: 1}, nil
	}
	var existing string
	if err := s.DB.QueryRow(`SELECT code FROM waitlist WHERE email=?`, in.Email).Scan(&existing); err == nil {
		st, err := s.WaitlistStatus(existing)
		if st != nil {
			st.Existing = true
		}
		return st, err
	}
	z, err := s.VerifyPow(in.Pow.Challenge, in.Pow.Sig, in.Pow.Nonce)
	if err != nil {
		return nil, err
	}
	for _, f := range []*string{&in.Company, &in.Role, &in.UseCase, &in.Volume, &in.Region} {
		if len(*f) > 120 {
			*f = (*f)[:120]
		}
	}
	prio := 0
	if in.Track == "platform" {
		prio = platformPriority(in)
	}
	code := strings.ToUpper(randomHex(4))
	var referrer string
	if in.Ref != "" {
		_ = s.DB.QueryRow(`SELECT code FROM waitlist WHERE code=?`, strings.ToUpper(in.Ref)).Scan(&referrer)
	}
	_, err = s.DB.Exec(`INSERT INTO waitlist(id, email, track, company, role, use_case, volume, region, code, referred_by, priority, pow_bits, ip_hash, created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, newID("wl"), in.Email, in.Track, in.Company, in.Role, in.UseCase, in.Volume, in.Region, code, nullable(referrer), prio, z, ipHash, Now())
	if err != nil {
		return nil, err
	}
	if referrer != "" {
		_, _ = s.DB.Exec(`UPDATE waitlist SET referrals = referrals + 1 WHERE code=?`, referrer)
	}
	return s.WaitlistStatus(code)
}

func (s *Service) WaitlistStatus(code string) (*WaitlistStatus, error) {
	var st WaitlistStatus
	var created string
	var prio int
	err := s.DB.QueryRow(`SELECT code, track, referrals, created_at, priority, pow_bits FROM waitlist WHERE code=?`, strings.ToUpper(code)).Scan(&st.Code, &st.Track, &st.Referrals, &created, &prio, &st.PowBits)
	if err != nil {
		return nil, errors.New("waitlist spot not found")
	}
	st.Total = count(s.DB, `SELECT COUNT(*) FROM waitlist WHERE track=?`, st.Track)
	if st.Track == "platform" {
		st.Position = 1 + count(s.DB, `SELECT COUNT(*) FROM waitlist WHERE track='platform' AND (priority > ? OR (priority = ? AND created_at < ?))`, prio, prio, created)
		switch {
		case prio >= 12:
			st.Tier = "Design partner — we'll reach out this week"
		case prio >= 7:
			st.Tier = "Priority integration"
		default:
			st.Tier = "Standard access"
		}
	} else {
		// rank by signup order minus referral boost
		var seq int
		_ = s.DB.QueryRow(`SELECT COUNT(*) FROM waitlist WHERE track='person' AND created_at <= ?`, created).Scan(&seq)
		score := seq - SpotsPerReferral*st.Referrals
		st.Position = 1 + count(s.DB, `SELECT COUNT(*) FROM (SELECT (SELECT COUNT(*) FROM waitlist w2 WHERE w2.track='person' AND w2.created_at <= w.created_at) - ?*w.referrals AS sc, w.created_at AS c, w.code AS cd FROM waitlist w WHERE w.track='person') t WHERE t.sc < ? OR (t.sc = ? AND t.c < ?)`, SpotsPerReferral, score, score, created)
		switch {
		case st.Referrals >= 10:
			st.Tier = "Founding Human — lifetime Pass Plus"
		case st.Referrals >= 3:
			st.Tier = "Early access"
		default:
			st.Tier = "In line"
		}
	}
	st.ShareURL = s.Cfg.BaseURL + "/?ref=" + st.Code
	return &st, nil
}

type WaitlistRow struct {
	Email     string `json:"email"`
	Track     string `json:"track"`
	Company   string `json:"company"`
	Role      string `json:"role"`
	UseCase   string `json:"useCase"`
	Volume    string `json:"volume"`
	Region    string `json:"region"`
	Code      string `json:"code"`
	Referrals int    `json:"referrals"`
	Priority  int    `json:"priority"`
	PowBits   int    `json:"powBits"`
	CreatedAt string `json:"createdAt"`
}

func (s *Service) WaitlistRows() ([]WaitlistRow, error) {
	rows, err := s.DB.Query(`SELECT email, track, company, role, use_case, volume, region, code, referrals, priority, pow_bits, created_at FROM waitlist ORDER BY track, priority DESC, created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []WaitlistRow
	for rows.Next() {
		var r WaitlistRow
		if err := rows.Scan(&r.Email, &r.Track, &r.Company, &r.Role, &r.UseCase, &r.Volume, &r.Region, &r.Code, &r.Referrals, &r.Priority, &r.PowBits, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Service) WaitlistCounts() (people, platforms int) {
	return count(s.DB, `SELECT COUNT(*) FROM waitlist WHERE track='person'`), count(s.DB, `SELECT COUNT(*) FROM waitlist WHERE track='platform'`)
}
