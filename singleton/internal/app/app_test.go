package app

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/stbodurov/bingo/singleton/internal/proof"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	db, err := OpenDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	cfg := DefaultConfig()
	cfg.MinRing = 8
	s, err := NewService(db, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func enroll(t *testing.T, s *Service, birth, doc, issuer string) *proof.KeyPair {
	t.Helper()
	k := proof.NewKeyPair()
	c, err := s.IssueSandboxCredential(issuer, Claims{GivenName: "Ada", FamilyName: "Lovelace" + doc, BirthDate: birth, DocumentNumber: doc})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.EnrollWithCredential(k.PublicHex(), c); err != nil {
		t.Fatal(err)
	}
	return k
}

// present fetches the holder's buckets exactly like the browser wallet does.
func present(t *testing.T, s *Service, k *proof.KeyPair, scope, challenge string, sets []string) *proof.Presentation {
	t.Helper()
	ep, err := s.CurrentEpoch()
	if err != nil {
		t.Fatal(err)
	}
	buckets := map[string]proof.BucketData{}
	for _, set := range sets {
		b, err := s.BucketFor(ep.ID, set, k.PublicHex())
		if err != nil {
			t.Fatal(err)
		}
		bd := proof.BucketData{Bits: b.Bits, Prefix: b.Prefix}
		for _, m := range b.Members {
			raw, _ := hex.DecodeString(m)
			bd.Members = append(bd.Members, raw)
		}
		buckets[set] = bd
	}
	p, err := proof.ProveBuckets(k, scope, challenge, ep.ID, buckets, sets, s.Cfg.MinRing)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestEndToEndVerification(t *testing.T) {
	s := newTestService(t)
	if err := s.seedHolders(randFor(1), 400, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	adult := enroll(t, s, "1990-04-02", "D1234567", "sbx-dmv-ca")
	teen := enroll(t, s, time.Now().AddDate(-15, 0, 0).Format("2006-01-02"), "P7654321", "sbx-csca-au")
	if _, err := s.SealEpoch(time.Now().UTC().Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	// Duplicate enrollment with the same document is refused.
	c, _ := s.IssueSandboxCredential("sbx-dmv-ca", Claims{GivenName: "Ada", FamilyName: "LovelaceD1234567", BirthDate: "1990-04-02", DocumentNumber: "D1234567"})
	if _, err := s.EnrollWithCredential(proof.NewKeyPair().PublicHex(), c); err != ErrDuplicateHuman {
		t.Fatalf("expected duplicate refusal, got %v", err)
	}
	// Tampered credential is refused.
	c.Claims.BirthDate = "1980-01-01"
	if _, err := s.EnrollWithCredential(proof.NewKeyPair().PublicHex(), c); err == nil || !strings.Contains(err.Error(), "signature") {
		t.Fatalf("expected signature failure, got %v", err)
	}

	keys, err := s.Signup("owner@example.com", "correct-horse-battery", "Owner", "Example")
	if err != nil {
		t.Fatal(err)
	}
	ak, err := s.CreateApp(keys.OrgID, "Example Social", "social", "http://localhost:4201", "human,age_over_16")
	if err != nil {
		t.Fatal(err)
	}
	app := ak.App

	vs, err := s.CreateSession(app, SessionRequest{Reference: "alice"}, "http://localhost:4201")
	if err != nil {
		t.Fatal(err)
	}
	done, err := s.CompleteSession(vs.ID, present(t, s, adult, vs.Scope, vs.Challenge, vs.Required))
	if err != nil {
		t.Fatalf("adult verification failed: %v", err)
	}
	if done.Status != "verified" || done.Nullifier == "" {
		t.Fatalf("unexpected result %+v", done)
	}
	claims, err := s.VerifyResultToken(done.Token)
	if err != nil || claims["sub"] != done.Nullifier {
		t.Fatalf("token invalid: %v %v", err, claims)
	}

	// Same human tries a second account → duplicate, pointing at the first account.
	vs2, _ := s.CreateSession(app, SessionRequest{Reference: "alice-alt"}, "")
	dup, err := s.CompleteSession(vs2.ID, present(t, s, adult, vs2.Scope, vs2.Challenge, vs2.Required))
	if err != nil {
		t.Fatal(err)
	}
	if dup.Status != "duplicate" || dup.DuplicateOf != "alice" || dup.Nullifier != done.Nullifier {
		t.Fatalf("duplicate not detected: %+v", dup)
	}

	// A 15-year-old passes 'human' + 'age_over_13' elsewhere, but cannot satisfy age_over_16.
	vs3, _ := s.CreateSession(app, SessionRequest{Reference: "teen"}, "")
	if _, err := s.BucketFor(mustEpoch(t, s), "age_over_16", teen.PublicHex()); err == nil {
		t.Fatal("teen should not be in age_over_16")
	}
	_ = vs3

	// Different app → different pseudonym for the same human.
	ak2, _ := s.CreateApp(keys.OrgID, "Other App", "dating", "", "human")
	vs4, _ := s.CreateSession(ak2.App, SessionRequest{}, "")
	other, err := s.CompleteSession(vs4.ID, present(t, s, adult, vs4.Scope, vs4.Challenge, vs4.Required))
	if err != nil {
		t.Fatal(err)
	}
	if other.Nullifier == done.Nullifier {
		t.Fatal("pseudonyms must differ across apps")
	}

	// Replaying a used presentation fails.
	if _, err := s.CompleteSession(vs.ID, present(t, s, adult, vs.Scope, vs.Challenge, vs.Required)); err == nil {
		t.Fatal("completed session accepted a second presentation")
	}

	// Agent delegation.
	agentPub, agentPriv, _ := ed25519.GenerateKey(nil)
	d := proof.Delegation{Version: 1, Scope: app.Scope, Agent: hex.EncodeToString(agentPub), Label: "Scheduling assistant", Permissions: []string{"events:create"}, IssuedAt: time.Now().Unix(), NotAfter: time.Now().Add(48 * time.Hour).Unix()}
	rec, err := s.CreateDelegation(app, d, present(t, s, adult, app.Scope, d.Challenge(), []string{"human"}))
	if err != nil {
		t.Fatal(err)
	}
	if rec.Nullifier != done.Nullifier {
		t.Fatal("delegation must carry the holder's app pseudonym")
	}
	msg := "POST /events {\"title\":\"standup\"}"
	sig := base64.RawURLEncoding.EncodeToString(ed25519.Sign(agentPriv, []byte(msg)))
	if _, err := s.VerifyAgentRequest(app.ID, rec.ID, msg, sig); err != nil {
		t.Fatalf("agent request rejected: %v", err)
	}
	if _, err := s.VerifyAgentRequest(app.ID, rec.ID, msg+"tampered", sig); err == nil {
		t.Fatal("tampered agent request accepted")
	}

	// Self-revocation removes the key from the next epoch.
	if err := s.SelfRevoke(adult.PublicHex(), adult.SignSchnorr([]byte("singleton/revoke/v1:"+adult.PublicHex())).String()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SealEpoch(time.Now().UTC().Add(2 * time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.BucketFor(mustEpoch(t, s), "human", adult.PublicHex()); err == nil {
		t.Fatal("revoked key still in set")
	}

	st, err := s.Stats([]string{app.ID, ak2.App.ID}, 30)
	if err != nil {
		t.Fatal(err)
	}
	if st.Verified != 2 || st.Duplicates != 1 || st.UniqueHumans != 2 {
		t.Fatalf("unexpected stats %+v", st)
	}
	txs, _ := s.ChainTxs(10)
	if len(txs) < 2 || !strings.HasPrefix(txs[0].Calldata, "0x"+hex.EncodeToString(anchorSelector)) {
		t.Fatalf("epochs must be anchored with anchorEpoch calldata: %+v", txs)
	}
}

func TestVouching(t *testing.T) {
	s := newTestService(t)
	if err := s.seedHolders(randFor(2), 200, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	var vouchers []*proof.KeyPair
	for i := 0; i < 3; i++ {
		vouchers = append(vouchers, enroll(t, s, "1985-01-0"+string(rune('1'+i)), "DOC00"+string(rune('1'+i))+"77", "sbx-csca-gb"))
	}
	if _, err := s.SealEpoch(time.Now().UTC().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	newcomer := proof.NewKeyPair()
	var enrolled *Enrollment
	for i, v := range vouchers {
		n, e, err := s.AddVouch(newcomer.PublicHex(), present(t, s, v, VouchScope, "vouch:"+newcomer.PublicHex(), []string{"human_doc"}))
		if err != nil {
			t.Fatal(err)
		}
		if n != i+1 {
			t.Fatalf("vouch count %d", n)
		}
		enrolled = e
	}
	if enrolled == nil || enrolled.Assurance != "IAL1" {
		t.Fatalf("expected IAL1 enrollment after 3 vouches, got %+v", enrolled)
	}
	// The same voucher cannot vouch twice for the same person.
	other := proof.NewKeyPair()
	if _, _, err := s.AddVouch(other.PublicHex(), present(t, s, vouchers[0], VouchScope, "vouch:"+other.PublicHex(), []string{"human_doc"})); err != nil {
		t.Fatal(err)
	}
}

func TestWaitlistPow(t *testing.T) {
	s := newTestService(t)
	ch := s.NewPowChallenge()
	nonce := solvePow(ch.Challenge, ch.Bits)
	in := WaitlistInput{Email: "ceo@platform.com", Track: "platform", Volume: ">1m", Region: "au", UseCase: "social"}
	in.Pow.Challenge, in.Pow.Sig, in.Pow.Nonce = ch.Challenge, ch.Sig, nonce
	st, err := s.JoinWaitlist(in, "ip")
	if err != nil {
		t.Fatal(err)
	}
	if st.Position != 1 || !strings.HasPrefix(st.Tier, "Design partner") {
		t.Fatalf("unexpected platform status %+v", st)
	}
	// Reusing the same challenge fails.
	in.Email = "other@platform.com"
	if _, err := s.JoinWaitlist(in, "ip"); err == nil {
		t.Fatal("reused proof-of-work accepted")
	}
	// Bad nonce fails.
	ch2 := s.NewPowChallenge()
	in.Pow.Challenge, in.Pow.Sig, in.Pow.Nonce = ch2.Challenge, ch2.Sig, "0"
	if _, err := s.JoinWaitlist(in, "ip"); err == nil {
		t.Fatal("weak proof-of-work accepted")
	}
}

func mustEpoch(t *testing.T, s *Service) int64 {
	ep, err := s.CurrentEpoch()
	if err != nil {
		t.Fatal(err)
	}
	return ep.ID
}
