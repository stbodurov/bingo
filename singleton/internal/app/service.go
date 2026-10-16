package app

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"
)

// Config tunes a Singleton deployment.
type Config struct {
	BaseURL       string
	EpochInterval time.Duration // how often a new epoch is sealed when the registry changed
	MinRing       int           // smallest anonymity bucket a verifier accepts
	EpochGrace    int64         // how many past epochs a presentation may reference
	SessionTTL    time.Duration
	AdminToken    string
}

func DefaultConfig() Config {
	return Config{
		BaseURL:       "http://localhost:4201",
		EpochInterval: 90 * time.Second,
		MinRing:       32,
		EpochGrace:    2,
		SessionTTL:    15 * time.Minute,
		AdminToken:    "dev-admin",
	}
}

// Service bundles every Singleton subsystem around one database.
type Service struct {
	DB  *DB
	Cfg Config
	Log *log.Logger

	pepper    []byte             // HMAC key for de-duplication hashes (HSM-held in production)
	tokenKey  ed25519.PrivateKey // signs verification result tokens (JWKS published)
	powKey    []byte             // signs waitlist proof-of-work challenges
	issuers   map[string]ed25519.PublicKey
	issuerSKs map[string]ed25519.PrivateKey // sandbox issuers only

	sealMu   sync.Mutex
	setCache sync.Map // "epoch/set" -> *cachedSet
}

func NewService(db *DB, cfg Config, logger *log.Logger) (*Service, error) {
	s := &Service{DB: db, Cfg: cfg, Log: logger, issuers: map[string]ed25519.PublicKey{}, issuerSKs: map[string]ed25519.PrivateKey{}}
	s.pepper = s.secret("pepper", 32)
	s.powKey = s.secret("pow_key", 32)
	seed := s.secret("token_key_seed", ed25519.SeedSize)
	s.tokenKey = ed25519.NewKeyFromSeed(seed)
	for _, iss := range SandboxIssuers {
		sk := ed25519.NewKeyFromSeed(s.secret("issuer_"+iss.ID, ed25519.SeedSize))
		s.issuerSKs[iss.ID] = sk
		s.issuers[iss.ID] = sk.Public().(ed25519.PublicKey)
	}
	if err := s.ensureAttesters(); err != nil {
		return nil, err
	}
	return s, nil
}

// secret returns a persistent random secret, creating it on first boot.
func (s *Service) secret(name string, n int) []byte {
	if v := s.DB.Meta("secret_" + name); v != "" {
		b, err := base64.StdEncoding.DecodeString(v)
		if err == nil && len(b) == n {
			return b
		}
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	_ = s.DB.SetMeta("secret_"+name, base64.StdEncoding.EncodeToString(b))
	return b
}

// keyedHash is HMAC-SHA256(pepper, domain || value): stable for de-duplication, useless
// to anyone who steals the database without the pepper.
func (s *Service) keyedHash(domain, value string) string {
	m := hmac.New(sha256.New, s.pepper)
	m.Write([]byte(domain))
	m.Write([]byte{0})
	m.Write([]byte(value))
	return hex.EncodeToString(m.Sum(nil))
}

func (s *Service) TokenPublicKey() ed25519.PublicKey { return s.tokenKey.Public().(ed25519.PublicKey) }

func newID(prefix string) string {
	b := make([]byte, 10)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, base64.RawURLEncoding.EncodeToString(b))
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Service) logf(format string, args ...any) {
	if s.Log != nil {
		s.Log.Printf(format, args...)
	}
}
