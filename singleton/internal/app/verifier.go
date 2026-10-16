package app

import (
	"crypto/ed25519"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/stbodurov/bingo/singleton/internal/proof"
)

// VerifyPresentation checks a presentation against the registry's epoch snapshots.
func (s *Service) VerifyPresentation(p *proof.Presentation, scope, challenge string, required []string) (*proof.Verified, error) {
	cur, err := s.CurrentEpoch()
	if err != nil {
		return nil, err
	}
	if p.Epoch > cur.ID || p.Epoch < cur.ID-s.Cfg.EpochGrace {
		return nil, fmt.Errorf("presentation uses epoch %d; current is %d — refresh your pass", p.Epoch, cur.ID)
	}
	return proof.VerifyPresentation(p, scope, challenge, required, s, s.Cfg.MinRing)
}

// ---------------------------------------------------------------- sessions

type SessionRequest struct {
	Required  []string `json:"required"`
	Action    string   `json:"action"`
	Reference string   `json:"reference"`
	ReturnURL string   `json:"returnUrl"`
}

type VerifySession struct {
	ID          string   `json:"id"`
	AppID       string   `json:"appId"`
	AppName     string   `json:"appName,omitempty"`
	Challenge   string   `json:"challenge"`
	Scope       string   `json:"scope"`
	Required    []string `json:"required"`
	Action      string   `json:"action,omitempty"`
	Reference   string   `json:"reference,omitempty"`
	Origin      string   `json:"origin,omitempty"`
	ReturnURL   string   `json:"returnUrl,omitempty"`
	Status      string   `json:"status"`
	Nullifier   string   `json:"nullifier,omitempty"`
	Predicates  []string `json:"predicates,omitempty"`
	RingSizes   []int    `json:"ringSizes,omitempty"`
	Error       string   `json:"error,omitempty"`
	DuplicateOf string   `json:"duplicateOf,omitempty"`
	Returning   bool     `json:"returning"`
	DurationMS  int      `json:"durationMs,omitempty"`
	CreatedAt   string   `json:"createdAt"`
	CompletedAt string   `json:"completedAt,omitempty"`
	ExpiresAt   string   `json:"expiresAt"`
	URL         string   `json:"url,omitempty"`
	Token       string   `json:"token,omitempty"`
}

var actionRe = func(a string) bool {
	if len(a) > 64 {
		return false
	}
	for _, r := range a {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == ':' || r == '.') {
			return false
		}
	}
	return true
}

func normalizeRequired(req []string, fallback string) ([]string, error) {
	if len(req) == 0 {
		req = strings.Split(fallback, ",")
	}
	seen := map[string]bool{}
	var out []string
	for _, r := range req {
		r = strings.TrimSpace(r)
		if r == "" || seen[r] {
			continue
		}
		if _, ok := SetByName(r); !ok {
			return nil, fmt.Errorf("unknown predicate %q", r)
		}
		seen[r] = true
		out = append(out, r)
	}
	if len(out) == 0 {
		out = []string{"human"}
	}
	if len(out) > 4 {
		return nil, errors.New("ask for at most 4 predicates per session")
	}
	return out, nil
}

func (s *Service) CreateSession(app *App, req SessionRequest, origin string) (*VerifySession, error) {
	required, err := normalizeRequired(req.Required, app.Required)
	if err != nil {
		return nil, err
	}
	action := strings.ToLower(strings.TrimSpace(req.Action))
	if action != "" && !actionRe(action) {
		return nil, errors.New("action may contain a-z, 0-9, - _ : . (max 64)")
	}
	scope := app.Scope
	if action != "" {
		scope += "#" + action
	}
	if len(req.Reference) > 128 || len(req.ReturnURL) > 512 {
		return nil, errors.New("reference or returnUrl too long")
	}
	if req.ReturnURL != "" && !app.AllowsURL(req.ReturnURL) {
		return nil, errors.New("returnUrl is not on this app's allowed origins")
	}
	now := time.Now().UTC()
	vs := &VerifySession{
		ID: newID("vs"), AppID: app.ID, AppName: app.Name, Challenge: randomHex(24), Scope: scope, Required: required,
		Action: action, Reference: req.Reference, Origin: origin, ReturnURL: req.ReturnURL, Status: "pending",
		CreatedAt: ts(now), ExpiresAt: ts(now.Add(s.Cfg.SessionTTL)),
	}
	reqJSON, _ := json.Marshal(required)
	_, err = s.DB.Exec(`INSERT INTO verify_sessions(id, app_id, challenge, scope, required, action, reference, origin, return_url, created_at, expires_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?)`, vs.ID, app.ID, vs.Challenge, scope, string(reqJSON), action, req.Reference, origin, req.ReturnURL, vs.CreatedAt, vs.ExpiresAt)
	if err != nil {
		return nil, err
	}
	vs.URL = s.Cfg.BaseURL + "/verify/" + vs.ID
	return vs, nil
}

func scanSession(row interface{ Scan(...any) error }) (*VerifySession, error) {
	var vs VerifySession
	var req string
	var nullifier, preds, rings, errS, dup, completed *string
	var dur *int
	var returning int
	err := row.Scan(&vs.ID, &vs.AppID, &vs.AppName, &vs.Challenge, &vs.Scope, &req, &vs.Action, &vs.Reference, &vs.Origin, &vs.ReturnURL,
		&vs.Status, &nullifier, &preds, &rings, &errS, &dup, &returning, &dur, &vs.CreatedAt, &completed, &vs.ExpiresAt)
	vs.Returning = returning == 1
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(req), &vs.Required)
	if nullifier != nil {
		vs.Nullifier = *nullifier
	}
	if preds != nil {
		_ = json.Unmarshal([]byte(*preds), &vs.Predicates)
	}
	if rings != nil {
		_ = json.Unmarshal([]byte(*rings), &vs.RingSizes)
	}
	if errS != nil {
		vs.Error = *errS
	}
	if dup != nil {
		vs.DuplicateOf = *dup
	}
	if dur != nil {
		vs.DurationMS = *dur
	}
	if completed != nil {
		vs.CompletedAt = *completed
	}
	return &vs, nil
}

const sessionCols = `v.id, v.app_id, a.name, v.challenge, v.scope, v.required, v.action, v.reference, v.origin, v.return_url, v.status, v.nullifier, v.predicates, v.ring_sizes, v.error, v.duplicate_of, v.is_returning, v.duration_ms, v.created_at, v.completed_at, v.expires_at`

func (s *Service) Session(id string) (*VerifySession, error) {
	vs, err := scanSession(s.DB.QueryRow(`SELECT `+sessionCols+` FROM verify_sessions v JOIN apps a ON a.id = v.app_id WHERE v.id = ?`, id))
	if err != nil {
		return nil, err
	}
	if vs.Status == "pending" && parseTS(vs.ExpiresAt).Before(time.Now()) {
		vs.Status = "expired"
		_, _ = s.DB.Exec(`UPDATE verify_sessions SET status='expired' WHERE id=? AND status='pending'`, id)
	}
	vs.URL = s.Cfg.BaseURL + "/verify/" + vs.ID
	return vs, nil
}

func (s *Service) Sessions(appIDs []string, limit int, status string) ([]VerifySession, error) {
	if len(appIDs) == 0 {
		return nil, nil
	}
	q := `SELECT ` + sessionCols + ` FROM verify_sessions v JOIN apps a ON a.id = v.app_id WHERE v.app_id IN (` + placeholders(len(appIDs)) + `)`
	args := make([]any, 0, len(appIDs)+2)
	for _, id := range appIDs {
		args = append(args, id)
	}
	if status != "" {
		q += ` AND v.status = ?`
		args = append(args, status)
	}
	q += ` ORDER BY v.created_at DESC LIMIT ?`
	args = append(args, limit)
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []VerifySession
	for rows.Next() {
		vs, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *vs)
	}
	return out, rows.Err()
}

func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

// CompleteSession verifies the holder's presentation, applies one-human-one-account
// bindings and returns the session with a signed result token.
func (s *Service) CompleteSession(id string, p *proof.Presentation) (*VerifySession, error) {
	vs, err := s.Session(id)
	if err != nil {
		return nil, errors.New("unknown session")
	}
	if vs.Status != "pending" {
		return nil, fmt.Errorf("session is %s", vs.Status)
	}
	app, err := s.App(vs.AppID)
	if err != nil {
		return nil, err
	}
	started := parseTS(vs.CreatedAt)
	v, verr := s.VerifyPresentation(p, vs.Scope, vs.Challenge, vs.Required)
	now := time.Now().UTC()
	dur := int(now.Sub(started) / time.Millisecond)
	if verr != nil {
		_, _ = s.DB.Exec(`UPDATE verify_sessions SET status='failed', error=?, completed_at=?, duration_ms=? WHERE id=?`, verr.Error(), ts(now), dur, id)
		s.emit(app, "session.failed", map[string]any{"session": id, "error": verr.Error()})
		return nil, verr
	}
	status := "verified"
	returning := false
	var dupOf string
	var existingRef string
	var uses int
	err = s.DB.QueryRow(`SELECT reference, uses FROM bindings WHERE scope=? AND nullifier=?`, vs.Scope, v.Nullifier).Scan(&existingRef, &uses)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		_, err = s.DB.Exec(`INSERT INTO bindings(scope, nullifier, app_id, reference, first_seen, last_seen) VALUES(?,?,?,?,?,?)`, vs.Scope, v.Nullifier, app.ID, vs.Reference, ts(now), ts(now))
	case err == nil:
		returning = true
		if vs.Reference != "" && existingRef != "" && existingRef != vs.Reference {
			status = "duplicate"
			dupOf = existingRef
		}
		_, err = s.DB.Exec(`UPDATE bindings SET last_seen=?, uses=uses+1 WHERE scope=? AND nullifier=?`, ts(now), vs.Scope, v.Nullifier)
	}
	if err != nil {
		return nil, err
	}
	preds, _ := json.Marshal(v.Predicates)
	rings, _ := json.Marshal(v.RingSizes)
	_, err = s.DB.Exec(`UPDATE verify_sessions SET status=?, nullifier=?, predicates=?, ring_sizes=?, duplicate_of=?, is_returning=?, completed_at=?, duration_ms=? WHERE id=?`,
		status, v.Nullifier, string(preds), string(rings), nullable(dupOf), boolInt(returning), ts(now), dur, id)
	if err != nil {
		return nil, err
	}
	out, err := s.Session(id)
	if err != nil {
		return nil, err
	}
	out.Token = s.ResultToken(out)
	s.emit(app, "session."+status, map[string]any{"session": id, "nullifier": out.Nullifier, "predicates": out.Predicates, "reference": out.Reference, "duplicateOf": dupOf, "returning": returning, "scope": out.Scope})
	return out, nil
}

// ---------------------------------------------------------------- result tokens (JWT, EdDSA)

func (s *Service) KeyID() string {
	sum := sha256.Sum256(s.TokenPublicKey())
	return hex.EncodeToString(sum[:8])
}

func b64(v any) string {
	b, _ := json.Marshal(v)
	return base64.RawURLEncoding.EncodeToString(b)
}

// ResultToken is a compact JWS relying parties can verify offline with our JWKS.
func (s *Service) ResultToken(vs *VerifySession) string {
	now := time.Now().UTC().Unix()
	header := b64(map[string]string{"alg": "EdDSA", "typ": "JWT", "kid": s.KeyID()})
	payload := b64(map[string]any{
		"iss": s.Cfg.BaseURL, "aud": vs.AppID, "sub": vs.Nullifier, "sid": vs.ID, "scope": vs.Scope,
		"predicates": vs.Predicates, "status": vs.Status, "duplicate": vs.Status == "duplicate", "duplicate_of": vs.DuplicateOf,
		"returning": vs.Returning, "reference": vs.Reference, "iat": now, "exp": now + 600,
	})
	sig := ed25519.Sign(s.tokenKey, []byte(header+"."+payload))
	return header + "." + payload + "." + base64.RawURLEncoding.EncodeToString(sig)
}

// VerifyResultToken validates a token (used by our SDK examples and tests).
func (s *Service) VerifyResultToken(tok string) (map[string]any, error) {
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		return nil, errors.New("malformed token")
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !ed25519.Verify(s.TokenPublicKey(), []byte(parts[0]+"."+parts[1]), sig) {
		return nil, errors.New("bad signature")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims map[string]any
	if err := json.Unmarshal(raw, &claims); err != nil {
		return nil, err
	}
	if exp, ok := claims["exp"].(float64); !ok || int64(exp) < time.Now().Unix() {
		return nil, errors.New("token expired")
	}
	return claims, nil
}

// JWKS publishes the token signing key (RFC 8037 OKP).
func (s *Service) JWKS() map[string]any {
	return map[string]any{"keys": []map[string]string{{
		"kty": "OKP", "crv": "Ed25519", "use": "sig", "alg": "EdDSA", "kid": s.KeyID(),
		"x": base64.RawURLEncoding.EncodeToString(s.TokenPublicKey()),
	}}}
}

// ---------------------------------------------------------------- agent delegation

type DelegationRecord struct {
	ID          string   `json:"id"`
	AppID       string   `json:"appId"`
	Nullifier   string   `json:"nullifier"`
	Agent       string   `json:"agent"`
	Label       string   `json:"label"`
	Permissions []string `json:"permissions"`
	NotAfter    string   `json:"notAfter"`
	CreatedAt   string   `json:"createdAt"`
	RevokedAt   string   `json:"revokedAt,omitempty"`
}

func (s *Service) CreateDelegation(app *App, d proof.Delegation, p *proof.Presentation) (*DelegationRecord, error) {
	if d.Scope != app.Scope {
		return nil, errors.New("delegation scope must equal the app scope")
	}
	agent, err := hex.DecodeString(d.Agent)
	if err != nil || len(agent) != ed25519.PublicKeySize {
		return nil, errors.New("agent must be a 32-byte Ed25519 public key (hex)")
	}
	if d.NotAfter <= time.Now().Unix() || d.NotAfter > time.Now().Add(366*24*time.Hour).Unix() {
		return nil, errors.New("notAfter must be in the future and within a year")
	}
	if strings.TrimSpace(d.Label) == "" || len(d.Label) > 80 || len(d.Permissions) > 16 {
		return nil, errors.New("label required (≤80 chars), ≤16 permissions")
	}
	v, err := s.VerifyPresentation(p, app.Scope, d.Challenge(), []string{"human"})
	if err != nil {
		return nil, err
	}
	rec := &DelegationRecord{ID: newID("dlg"), AppID: app.ID, Nullifier: v.Nullifier, Agent: d.Agent, Label: d.Label, Permissions: d.Permissions,
		NotAfter: ts(time.Unix(d.NotAfter, 0)), CreatedAt: Now()}
	perms, _ := json.Marshal(d.Permissions)
	_, err = s.DB.Exec(`INSERT INTO delegations(id, app_id, nullifier, agent, label, permissions, not_after, created_at) VALUES(?,?,?,?,?,?,?,?)`,
		rec.ID, rec.AppID, rec.Nullifier, rec.Agent, rec.Label, string(perms), rec.NotAfter, rec.CreatedAt)
	if err != nil {
		return nil, err
	}
	s.emit(app, "delegation.created", map[string]any{"delegation": rec.ID, "nullifier": rec.Nullifier, "agent": rec.Agent, "permissions": d.Permissions})
	return rec, nil
}

func (s *Service) Delegations(appIDs []string, limit int) ([]DelegationRecord, error) {
	if len(appIDs) == 0 {
		return nil, nil
	}
	args := []any{}
	for _, id := range appIDs {
		args = append(args, id)
	}
	args = append(args, limit)
	rows, err := s.DB.Query(`SELECT id, app_id, nullifier, agent, label, permissions, not_after, created_at, COALESCE(revoked_at,'') FROM delegations WHERE app_id IN (`+placeholders(len(appIDs))+`) ORDER BY created_at DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DelegationRecord
	for rows.Next() {
		var d DelegationRecord
		var perms string
		if err := rows.Scan(&d.ID, &d.AppID, &d.Nullifier, &d.Agent, &d.Label, &perms, &d.NotAfter, &d.CreatedAt, &d.RevokedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(perms), &d.Permissions)
		out = append(out, d)
	}
	return out, rows.Err()
}

// VerifyAgentRequest checks that msg was signed by a currently delegated agent.
func (s *Service) VerifyAgentRequest(appID, delegationID, msg, sigB64 string) (*DelegationRecord, error) {
	var d DelegationRecord
	var perms, revoked string
	err := s.DB.QueryRow(`SELECT id, app_id, nullifier, agent, label, permissions, not_after, created_at, COALESCE(revoked_at,'') FROM delegations WHERE id=? AND app_id=?`, delegationID, appID).
		Scan(&d.ID, &d.AppID, &d.Nullifier, &d.Agent, &d.Label, &perms, &d.NotAfter, &d.CreatedAt, &revoked)
	if err != nil {
		return nil, errors.New("unknown delegation")
	}
	_ = json.Unmarshal([]byte(perms), &d.Permissions)
	if revoked != "" {
		return nil, errors.New("delegation revoked")
	}
	if parseTS(d.NotAfter).Before(time.Now()) {
		return nil, errors.New("delegation expired")
	}
	pub, _ := hex.DecodeString(d.Agent)
	sig, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil || !ed25519.Verify(pub, []byte(msg), sig) {
		return nil, errors.New("agent signature invalid")
	}
	return &d, nil
}

func (s *Service) RevokeDelegation(appIDs []string, id string) error {
	args := []any{Now(), id}
	for _, a := range appIDs {
		args = append(args, a)
	}
	res, err := s.DB.Exec(`UPDATE delegations SET revoked_at=? WHERE id=? AND revoked_at IS NULL AND app_id IN (`+placeholders(len(appIDs))+`)`, args...)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("delegation not found")
	}
	return nil
}
