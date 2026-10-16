package app

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ---------------------------------------------------------------- orgs, users, console sessions

type Org struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Plan      string `json:"plan"`
	Demo      bool   `json:"demo"`
	CreatedAt string `json:"createdAt"`
}

type User struct {
	ID        string `json:"id"`
	OrgID     string `json:"orgId"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
}

var ErrAuth = errors.New("email or password is incorrect")

func hashToken(t string) string {
	sum := sha256.Sum256([]byte(t))
	return hex.EncodeToString(sum[:])
}

func (s *Service) Signup(email, password, name, orgName string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, errors.New("enter a valid email")
	}
	if len(password) < 10 {
		return nil, errors.New("password must be at least 10 characters")
	}
	if strings.TrimSpace(name) == "" {
		name = strings.Split(email, "@")[0]
	}
	if strings.TrimSpace(orgName) == "" {
		orgName = name + "'s team"
	}
	if count(s.DB, `SELECT COUNT(*) FROM users WHERE email=?`, email) > 0 {
		return nil, errors.New("an account with that email already exists")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	orgID := newID("org")
	if _, err := s.DB.Exec(`INSERT INTO orgs(id, name, created_at) VALUES(?,?,?)`, orgID, orgName, Now()); err != nil {
		return nil, err
	}
	u := &User{ID: newID("usr"), OrgID: orgID, Email: email, Name: name, Role: "owner", CreatedAt: Now()}
	if _, err := s.DB.Exec(`INSERT INTO users(id, org_id, email, name, password_hash, role, created_at) VALUES(?,?,?,?,?,?,?)`, u.ID, u.OrgID, u.Email, u.Name, string(hash), u.Role, u.CreatedAt); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) Login(email, password string) (*User, error) {
	var u User
	var hash string
	err := s.DB.QueryRow(`SELECT id, org_id, email, name, role, created_at, password_hash FROM users WHERE email=?`, strings.ToLower(strings.TrimSpace(email))).
		Scan(&u.ID, &u.OrgID, &u.Email, &u.Name, &u.Role, &u.CreatedAt, &hash)
	if err != nil {
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$CwTycUXWue0Thq9StjUM0uJ8.y1Yy8GhB1Ue1QiL/2Q1qQ0zqZf2."), []byte(password))
		return nil, ErrAuth
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return nil, ErrAuth
	}
	return &u, nil
}

// StartConsoleSession returns (session token, csrf token).
func (s *Service) StartConsoleSession(userID string) (string, string, error) {
	tok := randomHex(32)
	csrf := randomHex(16)
	_, err := s.DB.Exec(`INSERT INTO console_sessions(token_hash, user_id, csrf, expires_at) VALUES(?,?,?,?)`, hashToken(tok), userID, csrf, ts(time.Now().Add(14*24*time.Hour)))
	return tok, csrf, err
}

func (s *Service) EndConsoleSession(tok string) {
	_, _ = s.DB.Exec(`DELETE FROM console_sessions WHERE token_hash=?`, hashToken(tok))
}

func (s *Service) ConsoleUser(tok string) (*User, *Org, string, error) {
	var u User
	var o Org
	var csrf, exp string
	var demo int
	err := s.DB.QueryRow(`SELECT u.id, u.org_id, u.email, u.name, u.role, u.created_at, o.id, o.name, o.plan, o.demo, o.created_at, c.csrf, c.expires_at
		FROM console_sessions c JOIN users u ON u.id=c.user_id JOIN orgs o ON o.id=u.org_id WHERE c.token_hash=?`, hashToken(tok)).
		Scan(&u.ID, &u.OrgID, &u.Email, &u.Name, &u.Role, &u.CreatedAt, &o.ID, &o.Name, &o.Plan, &demo, &o.CreatedAt, &csrf, &exp)
	if err != nil {
		return nil, nil, "", err
	}
	if parseTS(exp).Before(time.Now()) {
		s.EndConsoleSession(tok)
		return nil, nil, "", errors.New("session expired")
	}
	o.Demo = demo == 1
	return &u, &o, csrf, nil
}

func (s *Service) SetPlan(orgID, plan string) error {
	if _, ok := Plans[plan]; !ok {
		return errors.New("unknown plan")
	}
	_, err := s.DB.Exec(`UPDATE orgs SET plan=? WHERE id=?`, plan, orgID)
	return err
}

// ---------------------------------------------------------------- apps & keys

type App struct {
	ID              string `json:"id"`
	OrgID           string `json:"orgId"`
	Name            string `json:"name"`
	Scope           string `json:"scope"`
	Category        string `json:"category"`
	Mode            string `json:"mode"`
	Origins         string `json:"origins"`
	Required        string `json:"required"`
	DuplicatePolicy string `json:"duplicatePolicy"`
	WebhookURL      string `json:"webhookUrl"`
	WebhookSecret   string `json:"-"`
	Color           string `json:"color"`
	CreatedAt       string `json:"createdAt"`
}

var Categories = []string{"social", "dating", "marketplace", "ticketing", "gaming", "community", "fintech", "ai_agents", "civic"}

func (a *App) OriginList() []string {
	var out []string
	for _, o := range strings.FieldsFunc(a.Origins, func(r rune) bool { return r == ',' || r == '\n' || r == ' ' }) {
		if o = strings.TrimRight(strings.TrimSpace(o), "/"); o != "" {
			out = append(out, o)
		}
	}
	return out
}

// AllowsOrigin is exact scheme://host[:port] matching; "*" is never accepted in live mode.
func (a *App) AllowsOrigin(origin string) bool {
	origin = strings.TrimRight(origin, "/")
	for _, o := range a.OriginList() {
		if o == origin || (o == "*" && a.Mode == "test") {
			return true
		}
	}
	return false
}

func (a *App) AllowsURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return strings.HasPrefix(raw, "/") && !strings.HasPrefix(raw, "//")
	}
	return a.AllowsOrigin(u.Scheme + "://" + u.Host)
}

func (a *App) RequiredList() []string { return strings.Split(a.Required, ",") }

const appCols = `id, org_id, name, scope, category, mode, origins, required, duplicate_policy, webhook_url, webhook_secret, color, created_at`

func scanApp(row interface{ Scan(...any) error }) (*App, error) {
	var a App
	err := row.Scan(&a.ID, &a.OrgID, &a.Name, &a.Scope, &a.Category, &a.Mode, &a.Origins, &a.Required, &a.DuplicatePolicy, &a.WebhookURL, &a.WebhookSecret, &a.Color, &a.CreatedAt)
	return &a, err
}

func (s *Service) App(id string) (*App, error) {
	return scanApp(s.DB.QueryRow(`SELECT `+appCols+` FROM apps WHERE id=?`, id))
}

func (s *Service) OrgApp(orgID, id string) (*App, error) {
	a, err := scanApp(s.DB.QueryRow(`SELECT `+appCols+` FROM apps WHERE id=? AND org_id=?`, id, orgID))
	if err != nil {
		return nil, errors.New("app not found")
	}
	return a, nil
}

func (s *Service) Apps(orgID string) ([]App, error) {
	rows, err := s.DB.Query(`SELECT `+appCols+` FROM apps WHERE org_id=? ORDER BY created_at`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []App
	for rows.Next() {
		a, err := scanApp(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, rows.Err()
}

func (s *Service) AppIDs(orgID string) []string {
	apps, _ := s.Apps(orgID)
	ids := make([]string, len(apps))
	for i, a := range apps {
		ids[i] = a.ID
	}
	return ids
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

type NewAppKeys struct {
	App         *App   `json:"app"`
	Publishable string `json:"publishableKey"`
	Secret      string `json:"secretKey"`
}

func (s *Service) CreateApp(orgID, name, category, origins, required string) (*NewAppKeys, error) {
	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 60 {
		return nil, errors.New("app name must be 2–60 characters")
	}
	validCat := false
	for _, c := range Categories {
		if c == category {
			validCat = true
		}
	}
	if !validCat {
		category = "social"
	}
	reqList, err := normalizeRequired(strings.Split(required, ","), "human")
	if err != nil {
		return nil, err
	}
	slug := strings.Trim(slugRe.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if slug == "" {
		slug = "app"
	}
	a := &App{ID: newID("app"), OrgID: orgID, Name: name, Scope: "app:" + slug + "-" + randomHex(2), Category: category, Mode: "test",
		Origins: origins, Required: strings.Join(reqList, ","), DuplicatePolicy: "flag", WebhookSecret: "whsec_" + randomHex(16),
		Color: palette[int(time.Now().UnixNano()%int64(len(palette)))], CreatedAt: Now()}
	_, err = s.DB.Exec(`INSERT INTO apps(`+appCols+`) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, a.ID, a.OrgID, a.Name, a.Scope, a.Category, a.Mode, a.Origins, a.Required, a.DuplicatePolicy, a.WebhookURL, a.WebhookSecret, a.Color, a.CreatedAt)
	if err != nil {
		return nil, err
	}
	pk, err := s.issueKey(a.ID, "publishable", "test")
	if err != nil {
		return nil, err
	}
	sk, err := s.issueKey(a.ID, "secret", "test")
	if err != nil {
		return nil, err
	}
	return &NewAppKeys{App: a, Publishable: pk, Secret: sk}, nil
}

var palette = []string{"#d4ff4f", "#7cf2c4", "#9db4ff", "#ffb86b", "#ff8fb1", "#c8a6ff"}

func (s *Service) UpdateApp(orgID, id string, upd map[string]string) (*App, error) {
	a, err := s.OrgApp(orgID, id)
	if err != nil {
		return nil, err
	}
	if v, ok := upd["name"]; ok && strings.TrimSpace(v) != "" {
		a.Name = strings.TrimSpace(v)
	}
	if v, ok := upd["origins"]; ok {
		a.Origins = v
	}
	if v, ok := upd["required"]; ok {
		list, err := normalizeRequired(strings.Split(v, ","), "human")
		if err != nil {
			return nil, err
		}
		a.Required = strings.Join(list, ",")
	}
	if v, ok := upd["duplicatePolicy"]; ok && (v == "flag" || v == "block") {
		a.DuplicatePolicy = v
	}
	if v, ok := upd["mode"]; ok && (v == "test" || v == "live") {
		if v == "live" {
			for _, o := range a.OriginList() {
				if o == "*" {
					return nil, errors.New("remove the * origin before going live")
				}
			}
		}
		a.Mode = v
	}
	if v, ok := upd["webhookUrl"]; ok {
		if v != "" {
			if u, err := url.Parse(v); err != nil || (u.Scheme != "https" && u.Scheme != "http") {
				return nil, errors.New("webhook URL must be http(s)")
			}
		}
		a.WebhookURL = v
	}
	_, err = s.DB.Exec(`UPDATE apps SET name=?, origins=?, required=?, duplicate_policy=?, mode=?, webhook_url=? WHERE id=?`, a.Name, a.Origins, a.Required, a.DuplicatePolicy, a.Mode, a.WebhookURL, a.ID)
	return a, err
}

func (s *Service) issueKey(appID, kind, mode string) (string, error) {
	prefix := map[string]string{"publishable": "pk_", "secret": "sk_"}[kind] + mode + "_"
	raw := prefix + randomHex(20)
	_, err := s.DB.Exec(`INSERT INTO api_keys(id, app_id, kind, prefix, key_hash, created_at) VALUES(?,?,?,?,?,?)`, newID("key"), appID, kind, raw[:len(prefix)+6], hashToken(raw), Now())
	return raw, err
}

// RotateSecret revokes existing secret keys and issues a new one.
func (s *Service) RotateSecret(orgID, appID string) (string, error) {
	a, err := s.OrgApp(orgID, appID)
	if err != nil {
		return "", err
	}
	if _, err := s.DB.Exec(`UPDATE api_keys SET revoked_at=? WHERE app_id=? AND kind='secret' AND revoked_at IS NULL`, Now(), a.ID); err != nil {
		return "", err
	}
	return s.issueKey(a.ID, "secret", a.Mode)
}

type APIKey struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Prefix     string `json:"prefix"`
	CreatedAt  string `json:"createdAt"`
	LastUsedAt string `json:"lastUsedAt"`
	RevokedAt  string `json:"revokedAt"`
}

func (s *Service) Keys(appID string) ([]APIKey, error) {
	rows, err := s.DB.Query(`SELECT id, kind, prefix, created_at, COALESCE(last_used_at,''), COALESCE(revoked_at,'') FROM api_keys WHERE app_id=? ORDER BY created_at DESC`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []APIKey
	for rows.Next() {
		var k APIKey
		if err := rows.Scan(&k.ID, &k.Kind, &k.Prefix, &k.CreatedAt, &k.LastUsedAt, &k.RevokedAt); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

// AuthenticateKey resolves an API key to its app.
func (s *Service) AuthenticateKey(raw string) (*App, string, error) {
	var appID, kind string
	err := s.DB.QueryRow(`SELECT app_id, kind FROM api_keys WHERE key_hash=? AND revoked_at IS NULL`, hashToken(raw)).Scan(&appID, &kind)
	if err != nil {
		return nil, "", errors.New("invalid API key")
	}
	_, _ = s.DB.Exec(`UPDATE api_keys SET last_used_at=? WHERE key_hash=?`, Now(), hashToken(raw))
	a, err := s.App(appID)
	return a, kind, err
}

// ---------------------------------------------------------------- analytics

type DayStat struct {
	Day       string `json:"day"`
	Verified  int    `json:"verified"`
	Duplicate int    `json:"duplicate"`
	Failed    int    `json:"failed"`
	Abandoned int    `json:"abandoned"`
}

type Stats struct {
	Sessions      int            `json:"sessions"`
	Verified      int            `json:"verified"`
	Duplicates    int            `json:"duplicates"`
	Failed        int            `json:"failed"`
	Abandoned     int            `json:"abandoned"`
	UniqueHumans  int            `json:"uniqueHumans"`
	Conversion    float64        `json:"conversion"`
	MedianSeconds float64        `json:"medianSeconds"`
	Days          []DayStat      `json:"days"`
	Predicates    map[string]int `json:"predicates"`
	Requested     map[string]int `json:"requested"`
	Delegations   int            `json:"delegations"`
	ByApp         map[string]int `json:"byApp"`
}

func (s *Service) Stats(appIDs []string, days int) (*Stats, error) {
	st := &Stats{Predicates: map[string]int{}, Requested: map[string]int{}, ByApp: map[string]int{}}
	if len(appIDs) == 0 {
		return st, nil
	}
	since := time.Now().UTC().AddDate(0, 0, -days+1).Format("2006-01-02")
	args := []any{}
	for _, id := range appIDs {
		args = append(args, id)
	}
	ph := placeholders(len(appIDs))
	rows, err := s.DB.Query(`SELECT substr(created_at,1,10) d, status, app_id, COALESCE(predicates,''), COALESCE(duration_ms,0), expires_at, required FROM verify_sessions
		WHERE app_id IN (`+ph+`) AND created_at >= ?`, append(args, since)...)
	if err != nil {
		return nil, err
	}
	byDay := map[string]*DayStat{}
	var durations []int
	now := time.Now().UTC()
	for rows.Next() {
		var d, status, appID, preds, exp, required string
		var dur int
		if err := rows.Scan(&d, &status, &appID, &preds, &dur, &exp, &required); err != nil {
			rows.Close()
			return nil, err
		}
		if status == "pending" && parseTS(exp).Before(now) {
			status = "expired"
		}
		ds := byDay[d]
		if ds == nil {
			ds = &DayStat{Day: d}
			byDay[d] = ds
		}
		st.Sessions++
		st.ByApp[appID]++
		for _, p := range strings.Split(strings.Trim(required, `[]`), ",") {
			if p = strings.Trim(p, `"`); p != "" {
				st.Requested[p]++
			}
		}
		switch status {
		case "verified":
			st.Verified++
			ds.Verified++
		case "duplicate":
			st.Duplicates++
			ds.Duplicate++
		case "failed":
			st.Failed++
			ds.Failed++
		case "expired":
			st.Abandoned++
			ds.Abandoned++
		}
		if status == "verified" || status == "duplicate" {
			durations = append(durations, dur)
			for _, p := range strings.Split(strings.Trim(preds, `[]`), ",") {
				if p = strings.Trim(p, `"`); p != "" {
					st.Predicates[p]++
				}
			}
		}
	}
	rows.Close()
	for i := days - 1; i >= 0; i-- {
		d := now.AddDate(0, 0, -i).Format("2006-01-02")
		if ds, ok := byDay[d]; ok {
			st.Days = append(st.Days, *ds)
		} else {
			st.Days = append(st.Days, DayStat{Day: d})
		}
	}
	if st.Sessions > 0 {
		st.Conversion = float64(st.Verified+st.Duplicates) / float64(st.Sessions)
	}
	if len(durations) > 0 {
		sort.Ints(durations)
		st.MedianSeconds = float64(durations[len(durations)/2]) / 1000
	}
	st.UniqueHumans = count(s.DB, `SELECT COUNT(*) FROM bindings WHERE app_id IN (`+ph+`)`, args...)
	st.Delegations = count(s.DB, `SELECT COUNT(*) FROM delegations WHERE app_id IN (`+ph+`)`, args...)
	return st, nil
}

// ---------------------------------------------------------------- billing

type Plan struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	MonthlyFee  float64 `json:"monthlyFee"`
	Included    int     `json:"included"`
	PerVerify   float64 `json:"perVerify"`
	PerPremium  float64 `json:"perPremium"` // per age/region predicate
	Description string  `json:"description"`
}

var Plans = map[string]Plan{
	"developer":  {"developer", "Developer", 0, 1000, 0, 0, "1,000 live verifications a month, unlimited test mode."},
	"growth":     {"growth", "Growth", 0, 1000, 0.05, 0.02, "Pay as you go: $0.05 per verified human, +$0.02 per age or region predicate. First 1,000 free."},
	"scale":      {"scale", "Scale", 2000, 80000, 0.025, 0.01, "$2,000/mo commit with 80,000 verifications, compliance evidence packs, 99.95% SLA."},
	"enterprise": {"enterprise", "Enterprise", 0, 0, 0, 0, "Annual contract, on-prem verifier, dedicated attester routing, custom DPA."},
}

type Invoice struct {
	Plan       Plan    `json:"plan"`
	Period     string  `json:"period"`
	Billable   int     `json:"billable"`
	Premium    int     `json:"premium"`
	TestMode   int     `json:"testMode"`
	Overage    int     `json:"overage"`
	Platform   float64 `json:"platform"`
	UsageCost  float64 `json:"usageCost"`
	Total      float64 `json:"total"`
	OverLimit  bool    `json:"overLimit"`
}

func (s *Service) Invoice(org *Org) (*Invoice, error) {
	plan := Plans[org.Plan]
	period := time.Now().UTC().Format("2006-01")
	inv := &Invoice{Plan: plan, Period: period}
	rows, err := s.DB.Query(`SELECT a.mode, COALESCE(v.predicates,'') FROM verify_sessions v JOIN apps a ON a.id=v.app_id
		WHERE a.org_id=? AND v.status IN ('verified','duplicate') AND substr(v.created_at,1,7)=?`, org.ID, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var mode, preds string
		_ = rows.Scan(&mode, &preds)
		if mode != "live" {
			inv.TestMode++
			continue
		}
		inv.Billable++
		for _, p := range strings.Split(strings.Trim(preds, `[]`), ",") {
			p = strings.Trim(p, `"`)
			if strings.HasPrefix(p, "age_") || strings.HasPrefix(p, "region_") {
				inv.Premium++
			}
		}
	}
	inv.Platform = plan.MonthlyFee
	if inv.Billable > plan.Included {
		inv.Overage = inv.Billable - plan.Included
	}
	switch plan.ID {
	case "developer":
		inv.OverLimit = inv.Overage > 0
	case "growth", "scale":
		share := 0.0
		if inv.Billable > 0 {
			share = float64(inv.Overage) / float64(inv.Billable)
		}
		inv.UsageCost = float64(inv.Overage)*plan.PerVerify + float64(inv.Premium)*share*plan.PerPremium
	}
	inv.Total = inv.Platform + inv.UsageCost
	return inv, nil
}

// ---------------------------------------------------------------- bindings view

type Binding struct {
	Scope     string `json:"scope"`
	Nullifier string `json:"nullifier"`
	AppID     string `json:"appId"`
	Reference string `json:"reference"`
	FirstSeen string `json:"firstSeen"`
	LastSeen  string `json:"lastSeen"`
	Uses      int    `json:"uses"`
}

func (s *Service) Bindings(appIDs []string, limit int) ([]Binding, error) {
	if len(appIDs) == 0 {
		return nil, nil
	}
	args := []any{}
	for _, id := range appIDs {
		args = append(args, id)
	}
	args = append(args, limit)
	rows, err := s.DB.Query(`SELECT scope, nullifier, app_id, reference, first_seen, last_seen, uses FROM bindings WHERE app_id IN (`+placeholders(len(appIDs))+`) ORDER BY last_seen DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Binding
	for rows.Next() {
		var b Binding
		if err := rows.Scan(&b.Scope, &b.Nullifier, &b.AppID, &b.Reference, &b.FirstSeen, &b.LastSeen, &b.Uses); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// Org lookups for seeding and admin.
func (s *Service) OrgByID(id string) (*Org, error) {
	var o Org
	var demo int
	err := s.DB.QueryRow(`SELECT id, name, plan, demo, created_at FROM orgs WHERE id=?`, id).Scan(&o.ID, &o.Name, &o.Plan, &demo, &o.CreatedAt)
	o.Demo = demo == 1
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("org %s not found", id)
	}
	return &o, err
}
