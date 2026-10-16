package app

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/stbodurov/bingo/singleton/internal/proof"
	"golang.org/x/crypto/bcrypt"
)

// Demo data: a sandbox network of synthetic holders (so anonymity buckets are realistic),
// a demo relying-party org with three apps, and 30 days of verification history. All of
// it is labelled synthetic in the UI and excluded from nothing but billing realism.

const (
	DemoEmail    = "demo@singleton.dev"
	DemoPassword = "singleton-demo"
)

type DemoApp struct {
	Key, Name, Scope, Category, Required, Color string
}

var DemoApps = []DemoApp{
	{"driftwood", "Driftwood", "app:driftwood-demo", "social", "human,age_over_16", "#7cf2c4"},
	{"frontrow", "Front Row", "app:frontrow-demo", "ticketing", "human_doc", "#ffb86b"},
	{"commons", "Commons", "app:commons-demo", "civic", "human", "#9db4ff"},
}

func (s *Service) Seeded() bool { return s.DB.Meta("seeded") == "1" }

// Seed populates the sandbox. holders controls the synthetic network size.
func (s *Service) Seed(holders int) error {
	if s.Seeded() {
		return nil
	}
	r := rand.New(rand.NewPCG(2026, 10))
	now := time.Now().UTC()
	if err := s.seedHolders(r, holders, now); err != nil {
		return err
	}
	// Epoch history: the registry grows over the past week.
	for _, back := range []time.Duration{144 * time.Hour, 96 * time.Hour, 48 * time.Hour, 24 * time.Hour, 3 * time.Hour} {
		if _, err := s.SealEpoch(now.Add(-back)); err != nil {
			return err
		}
	}
	if _, err := s.PublishRevocations(5, now.Add(-2*time.Hour)); err != nil {
		return err
	}
	if _, err := s.SealEpoch(now); err != nil {
		return err
	}
	if err := s.seedOrg(r, now); err != nil {
		return err
	}
	return s.DB.SetMeta("seeded", "1")
}

func (s *Service) seedHolders(r *rand.Rand, n int, now time.Time) error {
	type att struct {
		id, region string
		w          float64
	}
	paths := []att{{"att_mdl", "us", 0.35}, {"att_passport", "", 0.30}, {"att_eudi", "eu", 0.15}, {"att_inperson", "", 0.12}, {"att_vouch", "", 0.08}}
	pick := func() att {
		x := r.Float64()
		for _, p := range paths {
			if x < p.w {
				return p
			}
			x -= p.w
		}
		return paths[0]
	}
	regions := []string{"us", "gb", "au", "eu", ""}
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`INSERT INTO enrollments(id, pubkey, attester_id, assurance, region, age13_at, age16_at, age18_at, age21_at, status, revoked_reason, revoked_at, synthetic, created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,1,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for i := 0; i < n; i++ {
		k := proof.NewKeyPair()
		a := pick()
		assurance := "IAL2"
		region := a.region
		if region == "" {
			region = regions[r.IntN(len(regions))]
		}
		var ages [4]any
		if a.id == "att_vouch" {
			assurance, region = "IAL1", ""
		} else {
			// age distribution of an online population
			var years float64
			switch x := r.Float64(); {
			case x < 0.08:
				years = 13 + r.Float64()*5
			case x < 0.15:
				years = 18 + r.Float64()*3
			case x < 0.45:
				years = 21 + r.Float64()*14
			case x < 0.80:
				years = 35 + r.Float64()*20
			default:
				years = 55 + r.Float64()*30
			}
			birth := now.AddDate(0, 0, -int(years*365.25))
			for j, th := range thresholdDates(birth.Format("2006-01-02"), now) {
				idx := map[int]int{13: 0, 16: 1, 18: 2, 21: 3}[j]
				if th == "past" {
					ages[idx] = "1900-01-01"
				} else {
					ages[idx] = th
				}
			}
		}
		created := now.Add(-time.Duration(math.Pow(r.Float64(), 0.6) * float64(160*24*time.Hour)))
		status, reason, revokedAt := "active", any(nil), any(nil)
		if r.Float64() < 0.006 {
			status, reason, revokedAt = "revoked", "attester fraud review", ts(created.Add(48*time.Hour))
		}
		if _, err := stmt.Exec(newID("enr"), k.PublicHex(), a.id, assurance, region, ages[0], ages[1], ages[2], ages[3], status, reason, revokedAt, ts(created)); err != nil {
			return err
		}
	}
	// a few rejected attempts for attester stats
	for i := 0; i < n/40; i++ {
		outcome, why := "duplicate", "document or person already enrolled"
		if i%3 == 0 {
			outcome, why = "rejected", "issuer signature is invalid"
		}
		if _, err := tx.Exec(`INSERT INTO enrollment_attempts(id, attester_id, outcome, reason, created_at) VALUES(?,?,?,?,?)`, newID("att"), paths[i%4].id, outcome, why, ts(now.Add(-time.Duration(r.IntN(30*24))*time.Hour))); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Service) seedOrg(r *rand.Rand, now time.Time) error {
	hash, _ := bcrypt.GenerateFromPassword([]byte(DemoPassword), bcrypt.DefaultCost)
	orgID := "org_demo"
	if _, err := s.DB.Exec(`INSERT INTO orgs(id, name, plan, demo, created_at) VALUES(?,?,?,?,?)`, orgID, "Demo Platforms Inc.", "growth", 1, ts(now.AddDate(0, -3, 0))); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`INSERT INTO users(id, org_id, email, name, password_hash, role, created_at) VALUES(?,?,?,?,?,?,?)`, "usr_demo", orgID, DemoEmail, "Dana Demo", string(hash), "owner", ts(now.AddDate(0, -3, 0))); err != nil {
		return err
	}
	for _, d := range DemoApps {
		appID := "app_" + d.Key
		webhook := ""
		if d.Key == "driftwood" {
			webhook = s.Cfg.BaseURL + "/demo/webhook-sink"
		}
		_, err := s.DB.Exec(`INSERT INTO apps(`+appCols+`) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, appID, orgID, d.Name, d.Scope, d.Category, "live",
			s.Cfg.BaseURL, d.Required, "flag", webhook, "whsec_"+randomHex(16), d.Color, ts(now.AddDate(0, -2, 0)))
		if err != nil {
			return err
		}
		pk, err := s.issueKey(appID, "publishable", "live")
		if err != nil {
			return err
		}
		sk, err := s.issueKey(appID, "secret", "live")
		if err != nil {
			return err
		}
		_ = s.DB.SetMeta("demo_pk_"+d.Key, pk)
		_ = s.DB.SetMeta("demo_sk_"+d.Key, sk)
		if err := s.seedSessions(r, appID, d, now); err != nil {
			return err
		}
	}
	return s.seedDelegations(r, now)
}

func (s *Service) seedSessions(r *rand.Rand, appID string, d DemoApp, now time.Time) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	ins, err := tx.Prepare(`INSERT INTO verify_sessions(id, app_id, challenge, scope, required, action, reference, origin, status, nullifier, predicates, ring_sizes, error, duplicate_of, duration_ms, synthetic, created_at, completed_at, expires_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,1,?,?,?)`)
	if err != nil {
		return err
	}
	defer ins.Close()
	bind, err := tx.Prepare(`INSERT OR IGNORE INTO bindings(scope, nullifier, app_id, reference, first_seen, last_seen, uses) VALUES(?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer bind.Close()
	base := map[string]float64{"driftwood": 140, "frontrow": 70, "commons": 35}[d.Key]
	reqs := strings.Split(d.Required, ",")
	reqJSON, _ := json.Marshal(reqs)
	type b struct{ nullifier, ref, scope string }
	var known []b
	for day := 29; day >= 0; day-- {
		// growth + weekly seasonality
		trend := 0.55 + 0.45*float64(29-day)/29
		dow := now.AddDate(0, 0, -day).Weekday()
		season := 1.0
		if dow == time.Saturday || dow == time.Sunday {
			season = 1.25
		}
		if d.Key == "frontrow" && (29-day)%7 == 3 {
			season = 3.2 // ticket drop day
		}
		n := int(base * trend * season * (0.85 + r.Float64()*0.3))
		for i := 0; i < n; i++ {
			created := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -day).Add(time.Duration(r.IntN(86400)) * time.Second)
			if created.After(now) {
				created = now.Add(-time.Duration(r.IntN(3600)) * time.Second)
			}
			scope := d.Scope
			action := ""
			if d.Key == "frontrow" {
				action = fmt.Sprintf("drop:week-%d", (29-day)/7)
				scope += "#" + action
			}
			if d.Key == "commons" {
				action = fmt.Sprintf("poll:%d", (29-day)/10+1)
				scope += "#" + action
			}
			dur := int(math.Exp(r.NormFloat64()*0.5+2.1) * 1000) // median ~8s
			ref := fmt.Sprintf("user_%06d", r.IntN(999999))
			status := "verified"
			var nullifier, preds, rings, errS, dupOf any
			x := r.Float64()
			switch {
			case x < 0.10:
				status = "expired"
			case x < 0.13:
				status = "failed"
				errS = []string{"presentation uses an expired epoch", "challenge mismatch (replay?)", "holder declined"}[r.IntN(3)]
			case x < 0.18 && len(known) > 20:
				status = "duplicate"
				k := known[r.IntN(len(known))]
				if k.scope != scope {
					status = "verified"
				} else {
					nullifier, dupOf = k.nullifier, k.ref
				}
			}
			if status == "verified" || status == "duplicate" {
				if nullifier == nil {
					nb := make([]byte, 32)
					for j := range nb {
						nb[j] = byte(r.IntN(256))
					}
					nh := hex.EncodeToString(nb)
					nullifier = nh
					known = append(known, b{nh, ref, scope})
					if _, err := bind.Exec(scope, nh, appID, ref, ts(created), ts(created), 1); err != nil {
						return err
					}
				}
				preds = string(reqJSON)
				rs := make([]int, len(reqs))
				for j := range rs {
					rs[j] = 60 + r.IntN(70)
				}
				rb, _ := json.Marshal(rs)
				rings = string(rb)
			}
			var completed any
			if status != "expired" {
				completed = ts(created.Add(time.Duration(dur) * time.Millisecond))
			}
			if _, err := ins.Exec(newID("vs"), appID, randomHex(24), scope, string(reqJSON), action, ref, s.Cfg.BaseURL, status, nullifier, preds, rings, errS, dupOf, dur,
				ts(created), completed, ts(created.Add(s.Cfg.SessionTTL))); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (s *Service) seedDelegations(r *rand.Rand, now time.Time) error {
	labels := []struct {
		label string
		perms []string
	}{
		{"Scheduling assistant", []string{"events:create", "messages:send"}},
		{"Posting copilot", []string{"posts:draft", "posts:publish"}},
		{"Shopping agent", []string{"tickets:purchase"}},
		{"Research agent", []string{"read:public"}},
	}
	for i := 0; i < 38; i++ {
		l := labels[r.IntN(len(labels))]
		app := "app_driftwood"
		if l.label == "Shopping agent" {
			app = "app_frontrow"
		}
		var nul string
		_ = s.DB.QueryRow(`SELECT nullifier FROM bindings WHERE app_id=? LIMIT 1 OFFSET ?`, app, r.IntN(200)).Scan(&nul)
		if nul == "" {
			continue
		}
		perms, _ := json.Marshal(l.perms)
		agent := make([]byte, 32)
		for j := range agent {
			agent[j] = byte(r.IntN(256))
		}
		created := now.Add(-time.Duration(r.IntN(25*24)) * time.Hour)
		var revoked any
		if r.Float64() < 0.12 {
			revoked = ts(created.Add(72 * time.Hour))
		}
		if _, err := s.DB.Exec(`INSERT INTO delegations(id, app_id, nullifier, agent, label, permissions, not_after, created_at, revoked_at) VALUES(?,?,?,?,?,?,?,?,?)`,
			newID("dlg"), app, nul, hex.EncodeToString(agent), l.label, string(perms), ts(created.AddDate(0, 3, 0)), ts(created), revoked); err != nil {
			return err
		}
	}
	return nil
}
