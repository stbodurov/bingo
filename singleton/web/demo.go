package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/stbodurov/bingo/singleton/internal/app"
)

// Demo relying parties built on the public SDK exactly as a customer would build them,
// plus the in-person attester console and the waitlist admin.

type demoDef struct {
	Key, Name, Tagline, Action string
}

var demos = map[string]demoDef{
	"driftwood": {"driftwood", "Driftwood", "A social network where every account is a real, unique person aged 16+.", ""},
	"frontrow":  {"frontrow", "Front Row", "Fair ticket drops: one ticket per human, no bots, no scalpers.", "drop:fall-tour-2026"},
	"commons":   {"commons", "Commons", "Community polls with one vote per person — and nobody learns who voted.", "poll:park-redesign"},
}

func (s *Server) demoRoutes() {
	m := s.mux
	m.HandleFunc("GET /demo", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/demo/driftwood", http.StatusSeeOther) })
	m.HandleFunc("GET /demo/apps.json", s.demoApps)
	m.HandleFunc("GET /demo/{app}", s.demoPage)
	m.HandleFunc("POST /demo/{app}/api/agent-verify", s.limit("demo", 120, time.Minute, s.demoAgentVerify))
	m.HandleFunc("POST /demo/{app}/api/claim", s.limit("demo", 60, time.Minute, s.demoClaim))
	m.HandleFunc("POST /demo/webhook-sink", s.demoWebhookSink)
	m.HandleFunc("GET /attester", s.attesterPage)
	m.HandleFunc("POST /attester/login", s.limit("attester", 20, 15*time.Minute, s.attesterLogin))
	m.HandleFunc("POST /attester/decide", s.attesterDecide)
	m.HandleFunc("GET /admin", s.adminPage)
	m.HandleFunc("GET /admin/waitlist.csv", s.adminCSV)
}

// demoApps lists the demo relying parties with their publishable keys (demo only).
func (s *Server) demoApps(w http.ResponseWriter, r *http.Request) {
	var out []map[string]string
	for _, d := range app.DemoApps {
		out = append(out, map[string]string{"key": d.Key, "name": d.Name, "scope": d.Scope, "pk": s.svc.DB.Meta("demo_pk_" + d.Key)})
	}
	writeJSON(w, 200, out)
}

// demoAgentVerify plays the demo app's backend: it verifies an agent request with the
// app's secret key, exactly like POST /v1/agents/verify from a customer server.
func (s *Server) demoAgentVerify(w http.ResponseWriter, r *http.Request) {
	if _, ok := demos[r.PathValue("app")]; !ok {
		writeErr(w, 404, "unknown demo")
		return
	}
	var in struct {
		DelegationID string `json:"delegationId"`
		Message      string `json:"message"`
		Signature    string `json:"signature"`
	}
	if err := decode(r, &in); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	d, err := s.svc.VerifyAgentRequest("app_"+r.PathValue("app"), in.DelegationID, in.Message, in.Signature)
	if err != nil {
		writeJSON(w, 200, map[string]any{"valid": false, "error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"valid": true, "nullifier": d.Nullifier, "label": d.Label, "permissions": d.Permissions})
}

type demoRecord struct {
	Key, Nullifier, Value, CreatedAt string
}

func (s *Server) demoRecords(appKey string, limit int) []demoRecord {
	rows, err := s.svc.DB.Query(`SELECT rkey, nullifier, value, created_at FROM demo_records WHERE app=? ORDER BY created_at DESC LIMIT ?`, appKey, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []demoRecord
	for rows.Next() {
		var d demoRecord
		_ = rows.Scan(&d.Key, &d.Nullifier, &d.Value, &d.CreatedAt)
		out = append(out, d)
	}
	return out
}

func (s *Server) demoPage(w http.ResponseWriter, r *http.Request) {
	d, ok := demos[r.PathValue("app")]
	if !ok {
		http.NotFound(w, r)
		return
	}
	pk := s.svc.DB.Meta("demo_pk_" + d.Key)
	tally := map[string]int{}
	for _, rec := range s.demoRecords(d.Key, 10000) {
		if d.Key == "commons" {
			tally[rec.Value]++
		}
	}
	s.render(w, r, "demo_"+d.Key, pageData{"D": d, "PK": pk, "Records": s.demoRecords(d.Key, 12), "Tally": tally, "Demos": demos})
}

func (s *Server) demoClaim(w http.ResponseWriter, r *http.Request) {
	d, ok := demos[r.PathValue("app")]
	if !ok {
		writeErr(w, 404, "unknown demo")
		return
	}
	var in struct {
		Token string `json:"token"`
		Value string `json:"value"`
	}
	if err := decode(r, &in); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	// The demo backend verifies the signed result exactly as a customer backend would.
	claims, err := s.svc.VerifyResultToken(in.Token)
	if err != nil {
		writeErr(w, 401, "result token invalid: "+err.Error())
		return
	}
	if claims["aud"] != "app_"+d.Key {
		writeErr(w, 401, "token was issued for a different app")
		return
	}
	nul, _ := claims["sub"].(string)
	in.Value = strings.TrimSpace(in.Value)
	if len(in.Value) > 40 {
		in.Value = in.Value[:40]
	}
	key := d.Key
	if d.Action != "" {
		key = d.Action
	}
	var existing string
	_ = s.svc.DB.QueryRow(`SELECT value FROM demo_records WHERE app=? AND rkey=? AND nullifier=?`, d.Key, key, nul).Scan(&existing)
	if existing != "" {
		msg := map[string]string{
			"driftwood": "This person already has a Driftwood account: @" + existing + ". One human, one account.",
			"frontrow":  "This person already claimed a ticket for this drop (seat " + existing + "). One per human.",
			"commons":   "This person already voted in this poll. One vote per human.",
		}[d.Key]
		writeJSON(w, 409, map[string]any{"error": msg, "nullifier": nul, "existing": existing})
		return
	}
	switch d.Key {
	case "driftwood":
		if !validHandle(in.Value) {
			writeErr(w, 400, "pick a handle: 3–20 letters, numbers or underscores")
			return
		}
		if count := countRows(s, `SELECT COUNT(*) FROM demo_records WHERE app='driftwood' AND value=?`, in.Value); count > 0 {
			writeErr(w, 409, "@"+in.Value+" is taken")
			return
		}
	case "frontrow":
		in.Value = fmt.Sprintf("%s-%d", string(rune('A'+countRows(s, `SELECT COUNT(*) FROM demo_records WHERE app='frontrow'`, nil)%12)), 1+countRows(s, `SELECT COUNT(*) FROM demo_records WHERE app='frontrow'`, nil)%30)
	case "commons":
		if in.Value != "meadow" && in.Value != "playground" && in.Value != "courts" {
			writeErr(w, 400, "choose one of the options")
			return
		}
	}
	if _, err := s.svc.DB.Exec(`INSERT INTO demo_records(app, rkey, nullifier, value, created_at) VALUES(?,?,?,?,?)`, d.Key, key, nul, in.Value, app.Now()); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "value": in.Value, "nullifier": nul, "predicates": claims["predicates"]})
}

func countRows(s *Server, q string, arg any) int {
	var n int
	if arg == nil {
		_ = s.svc.DB.QueryRow(q).Scan(&n)
	} else {
		_ = s.svc.DB.QueryRow(q, arg).Scan(&n)
	}
	return n
}

func validHandle(h string) bool {
	if len(h) < 3 || len(h) > 20 {
		return false
	}
	for _, c := range h {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_') {
			return false
		}
	}
	return true
}

// demoWebhookSink verifies Singleton-Signature like a customer endpoint would.
func (s *Server) demoWebhookSink(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	a, err := s.svc.App("app_driftwood")
	if err != nil {
		http.Error(w, "no demo app", 404)
		return
	}
	sig := r.Header.Get("Singleton-Signature")
	var t int64
	var v1 string
	for _, part := range strings.Split(sig, ",") {
		if strings.HasPrefix(part, "t=") {
			t, _ = strconv.ParseInt(part[2:], 10, 64)
		}
		if strings.HasPrefix(part, "v1=") {
			v1 = part[3:]
		}
	}
	expected := app.SignWebhook(a.WebhookSecret, t, body)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(fmt.Sprintf("t=%d,v1=%s", t, v1))) != 1 || time.Since(time.Unix(t, 0)) > 5*time.Minute {
		http.Error(w, "bad signature", 400)
		return
	}
	w.WriteHeader(204)
}

// ---------------------------------------------------------------- in-person attester console

func (s *Server) attesterToken() string {
	if t := os.Getenv("ATTESTER_TOKEN"); t != "" {
		return t
	}
	return "dev-attester"
}

func (s *Server) attesterCSRF(session string) string {
	m := hmac.New(sha256.New, []byte(s.attesterToken()))
	m.Write([]byte("csrf:" + session))
	return hex.EncodeToString(m.Sum(nil))[:32]
}

func (s *Server) attesterSession(r *http.Request) (string, bool) {
	c, err := r.Cookie("sgl_attester")
	if err != nil {
		return "", false
	}
	m := hmac.New(sha256.New, []byte(s.attesterToken()))
	m.Write([]byte("session"))
	return c.Value, subtle.ConstantTimeCompare([]byte(c.Value), []byte(hex.EncodeToString(m.Sum(nil)))) == 1
}

func (s *Server) attesterPage(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.attesterSession(r)
	if !ok {
		s.render(w, r, "attester", pageData{"LoggedIn": false, "Error": r.URL.Query().Get("error")})
		return
	}
	pending, _ := s.svc.PendingInPerson()
	s.render(w, r, "attester", pageData{"LoggedIn": true, "Pending": pending, "CSRF": s.attesterCSRF(sess), "OK": r.URL.Query().Get("ok"), "Error": r.URL.Query().Get("error"), "Code": r.URL.Query().Get("code")})
}

func (s *Server) attesterLogin(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	tok := r.PostFormValue("token")
	if r.PostFormValue("demo") == "1" {
		tok = s.attesterToken()
	}
	if subtle.ConstantTimeCompare([]byte(tok), []byte(s.attesterToken())) != 1 {
		http.Redirect(w, r, "/attester?error=Wrong+token", http.StatusSeeOther)
		return
	}
	m := hmac.New(sha256.New, []byte(s.attesterToken()))
	m.Write([]byte("session"))
	http.SetCookie(w, &http.Cookie{Name: "sgl_attester", Value: hex.EncodeToString(m.Sum(nil)), Path: "/attester", MaxAge: 8 * 3600, HttpOnly: true, SameSite: http.SameSiteStrictMode})
	http.Redirect(w, r, "/attester", http.StatusSeeOther)
}

func (s *Server) attesterDecide(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.attesterSession(r)
	_ = r.ParseForm()
	if !ok || subtle.ConstantTimeCompare([]byte(r.PostFormValue("_csrf")), []byte(s.attesterCSRF(sess))) != 1 {
		http.Error(w, "not authorised", http.StatusForbidden)
		return
	}
	code := strings.ToUpper(strings.TrimSpace(r.PostFormValue("code")))
	if r.PostFormValue("decision") == "reject" {
		_ = s.svc.RejectInPerson(code)
		http.Redirect(w, r, "/attester?ok=Rejected+"+code, http.StatusSeeOther)
		return
	}
	cl := app.Claims{
		GivenName: r.PostFormValue("given_name"), FamilyName: r.PostFormValue("family_name"), BirthDate: r.PostFormValue("birth_date"),
		DocumentNumber: r.PostFormValue("document_number"), IssuingCountry: strings.ToUpper(r.PostFormValue("issuing_country")),
	}
	if _, err := s.svc.ApproveInPerson(code, cl, r.PostFormValue("doc_type")); err != nil {
		http.Redirect(w, r, "/attester?code="+code+"&error="+urlq(err.Error()), http.StatusSeeOther)
		return
	}
	s.svc.MaybeSeal(time.Now().UTC(), true)
	http.Redirect(w, r, "/attester?ok=Approved+"+code+"+%E2%80%94+the+ID+details+were+hashed+and+discarded", http.StatusSeeOther)
}

func urlq(s string) string { return strings.ReplaceAll(strings.ReplaceAll(s, " ", "+"), "&", "and") }

// ---------------------------------------------------------------- waitlist admin

func (s *Server) adminOK(r *http.Request) bool {
	tok := r.URL.Query().Get("token")
	return tok != "" && subtle.ConstantTimeCompare([]byte(tok), []byte(s.svc.Cfg.AdminToken)) == 1
}

func (s *Server) adminPage(w http.ResponseWriter, r *http.Request) {
	if !s.adminOK(r) {
		s.render(w, r, "admin", pageData{"Locked": true})
		return
	}
	rows, _ := s.svc.WaitlistRows()
	people, platforms := s.svc.WaitlistCounts()
	s.render(w, r, "admin", pageData{"Rows": rows, "People": people, "Platforms": platforms, "Token": r.URL.Query().Get("token")})
}

func (s *Server) adminCSV(w http.ResponseWriter, r *http.Request) {
	if !s.adminOK(r) {
		http.Error(w, "admin token required", http.StatusUnauthorized)
		return
	}
	rows, _ := s.svc.WaitlistRows()
	var out [][]string
	for _, x := range rows {
		out = append(out, []string{x.Email, x.Track, x.Company, x.Role, x.UseCase, x.Volume, x.Region, x.Code, strconv.Itoa(x.Referrals), strconv.Itoa(x.Priority), strconv.Itoa(x.PowBits), x.CreatedAt})
	}
	writeCSV(w, "singleton-waitlist.csv", []string{"email", "track", "company", "role", "use_case", "volume", "region", "code", "referrals", "priority", "pow_bits", "created_at"}, out)
}
