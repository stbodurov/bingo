package web

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/stbodurov/bingo/singleton/internal/app"
)

const consoleCookie = "sgl_console"

type consoleCtx struct {
	User *app.User
	Org  *app.Org
	CSRF string
	Apps []app.App
}

func (s *Server) consoleRoutes() {
	m := s.mux
	m.HandleFunc("GET /console/login", s.consoleLoginPage)
	m.HandleFunc("POST /console/login", s.limit("console-login", 20, 15*time.Minute, s.consoleLogin))
	m.HandleFunc("POST /console/signup", s.limit("console-login", 20, 15*time.Minute, s.consoleSignup))
	m.HandleFunc("POST /console/demo", s.limit("console-login", 40, 15*time.Minute, s.consoleDemo))
	m.HandleFunc("POST /console/logout", s.consoleAuth(s.consoleLogout))
	m.HandleFunc("GET /console", s.consoleAuth(s.consoleOverview))
	m.HandleFunc("GET /console/apps", s.consoleAuth(s.consoleApps))
	m.HandleFunc("POST /console/apps", s.consoleAuth(s.consoleCreateApp))
	m.HandleFunc("GET /console/apps/{id}", s.consoleAuth(s.consoleApp))
	m.HandleFunc("POST /console/apps/{id}", s.consoleAuth(s.consoleUpdateApp))
	m.HandleFunc("POST /console/apps/{id}/rotate", s.consoleAuth(s.consoleRotate))
	m.HandleFunc("POST /console/apps/{id}/test", s.consoleAuth(s.consoleTestSession))
	m.HandleFunc("GET /console/sessions", s.consoleAuth(s.consoleSessions))
	m.HandleFunc("GET /console/sessions/{id}", s.consoleAuth(s.consoleSession))
	m.HandleFunc("GET /console/humans", s.consoleAuth(s.consoleHumans))
	m.HandleFunc("GET /console/agents", s.consoleAuth(s.consoleAgents))
	m.HandleFunc("POST /console/agents/{id}/revoke", s.consoleAuth(s.consoleRevokeAgent))
	m.HandleFunc("GET /console/webhooks", s.consoleAuth(s.consoleWebhooks))
	m.HandleFunc("GET /console/billing", s.consoleAuth(s.consoleBilling))
	m.HandleFunc("POST /console/billing/plan", s.consoleAuth(s.consolePlan))
	m.HandleFunc("GET /console/compliance", s.consoleAuth(s.consoleCompliance))
	m.HandleFunc("GET /console/compliance.csv", s.consoleAuth(s.consoleComplianceCSV))
}

// consoleAuth loads the console session and enforces CSRF on every POST.
func (s *Server) consoleAuth(h func(http.ResponseWriter, *http.Request, *consoleCtx)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(consoleCookie)
		if err != nil {
			http.Redirect(w, r, "/console/login?next="+url.QueryEscape(r.URL.RequestURI()), http.StatusSeeOther)
			return
		}
		u, o, csrf, err := s.svc.ConsoleUser(c.Value)
		if err != nil {
			http.SetCookie(w, &http.Cookie{Name: consoleCookie, Value: "", Path: "/console", MaxAge: -1})
			http.Redirect(w, r, "/console/login", http.StatusSeeOther)
			return
		}
		if r.Method == http.MethodPost {
			_ = r.ParseForm()
			given := r.PostFormValue("_csrf")
			if given == "" {
				given = r.Header.Get("X-CSRF-Token")
			}
			if subtle.ConstantTimeCompare([]byte(given), []byte(csrf)) != 1 {
				http.Error(w, "CSRF check failed — reload the page and try again", http.StatusForbidden)
				return
			}
		}
		apps, _ := s.svc.Apps(o.ID)
		h(w, r, &consoleCtx{User: u, Org: o, CSRF: csrf, Apps: apps})
	}
}

func (s *Server) renderConsole(w http.ResponseWriter, r *http.Request, c *consoleCtx, name, active string, data pageData) {
	if data == nil {
		data = pageData{}
	}
	data["C"] = c
	data["Active"] = active
	if f, err := r.Cookie("sgl_flash"); err == nil {
		if b, err := base64.RawURLEncoding.DecodeString(f.Value); err == nil {
			var flash map[string]string
			if json.Unmarshal(b, &flash) == nil {
				data["Flash"] = flash
			}
		}
		http.SetCookie(w, &http.Cookie{Name: "sgl_flash", Value: "", Path: "/console", MaxAge: -1})
	}
	s.render(w, r, name, data)
}

func setFlash(w http.ResponseWriter, kv map[string]string) {
	b, _ := json.Marshal(kv)
	http.SetCookie(w, &http.Cookie{Name: "sgl_flash", Value: base64.RawURLEncoding.EncodeToString(b), Path: "/console", MaxAge: 60, HttpOnly: true, SameSite: http.SameSiteLaxMode})
}

func (s *Server) startConsole(w http.ResponseWriter, r *http.Request, userID string) {
	tok, _, err := s.svc.StartConsoleSession(userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: consoleCookie, Value: tok, Path: "/console", MaxAge: 14 * 86400, HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"})
	next := r.PostFormValue("next")
	if !strings.HasPrefix(next, "/console") {
		next = "/console"
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}

func (s *Server) consoleLoginPage(w http.ResponseWriter, r *http.Request) {
	s.render(w, r, "console_login", pageData{"Next": r.URL.Query().Get("next"), "Mode": r.URL.Query().Get("mode"), "Error": r.URL.Query().Get("error"), "DemoEmail": app.DemoEmail})
}

func loginError(w http.ResponseWriter, r *http.Request, mode, msg string) {
	http.Redirect(w, r, "/console/login?mode="+mode+"&error="+url.QueryEscape(msg), http.StatusSeeOther)
}

func (s *Server) consoleLogin(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	u, err := s.svc.Login(r.PostFormValue("email"), r.PostFormValue("password"))
	if err != nil {
		loginError(w, r, "login", err.Error())
		return
	}
	s.startConsole(w, r, u.ID)
}

func (s *Server) consoleSignup(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	u, err := s.svc.Signup(r.PostFormValue("email"), r.PostFormValue("password"), r.PostFormValue("name"), r.PostFormValue("company"))
	if err != nil {
		loginError(w, r, "signup", err.Error())
		return
	}
	s.startConsole(w, r, u.ID)
}

func (s *Server) consoleDemo(w http.ResponseWriter, r *http.Request) {
	u, err := s.svc.Login(app.DemoEmail, app.DemoPassword)
	if err != nil {
		loginError(w, r, "login", "demo workspace not seeded — start the server without -no-seed")
		return
	}
	s.startConsole(w, r, u.ID)
}

func (s *Server) consoleLogout(w http.ResponseWriter, r *http.Request, c *consoleCtx) {
	if ck, err := r.Cookie(consoleCookie); err == nil {
		s.svc.EndConsoleSession(ck.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: consoleCookie, Value: "", Path: "/console", MaxAge: -1})
	http.Redirect(w, r, "/console/login", http.StatusSeeOther)
}

func appIDs(c *consoleCtx) []string {
	ids := make([]string, len(c.Apps))
	for i, a := range c.Apps {
		ids[i] = a.ID
	}
	return ids
}

func appNames(c *consoleCtx) map[string]app.App {
	m := map[string]app.App{}
	for _, a := range c.Apps {
		m[a.ID] = a
	}
	return m
}

func (s *Server) consoleOverview(w http.ResponseWriter, r *http.Request, c *consoleCtx) {
	days := 30
	if r.URL.Query().Get("range") == "7" {
		days = 7
	}
	st, err := s.svc.Stats(appIDs(c), days)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	recent, _ := s.svc.Sessions(appIDs(c), 8, "")
	inv, _ := s.svc.Invoice(c.Org)
	type pc struct {
		Name  string
		Label string
		N     int
	}
	var preds []pc
	for k, v := range st.Predicates {
		d, _ := app.SetByName(k)
		preds = append(preds, pc{k, d.Label, v})
	}
	sort.Slice(preds, func(i, j int) bool { return preds[i].N > preds[j].N })
	s.renderConsole(w, r, c, "console_overview", "overview", pageData{"Stats": st, "Recent": recent, "Apps": appNames(c), "Invoice": inv, "Preds": preds, "Range": days})
}

func (s *Server) consoleApps(w http.ResponseWriter, r *http.Request, c *consoleCtx) {
	st, _ := s.svc.Stats(appIDs(c), 30)
	s.renderConsole(w, r, c, "console_apps", "apps", pageData{"Stats": st, "Categories": app.Categories, "Sets": app.Sets})
}

func (s *Server) consoleCreateApp(w http.ResponseWriter, r *http.Request, c *consoleCtx) {
	required := strings.Join(r.PostForm["required"], ",")
	keys, err := s.svc.CreateApp(c.Org.ID, r.PostFormValue("name"), r.PostFormValue("category"), r.PostFormValue("origins"), required)
	if err != nil {
		setFlash(w, map[string]string{"error": err.Error()})
		http.Redirect(w, r, "/console/apps", http.StatusSeeOther)
		return
	}
	setFlash(w, map[string]string{"ok": "App created. Copy your secret key now — it will not be shown again.", "secret": keys.Secret, "publishable": keys.Publishable})
	http.Redirect(w, r, "/console/apps/"+keys.App.ID, http.StatusSeeOther)
}

func (s *Server) consoleApp(w http.ResponseWriter, r *http.Request, c *consoleCtx) {
	a, err := s.svc.OrgApp(c.Org.ID, r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	keys, _ := s.svc.Keys(a.ID)
	st, _ := s.svc.Stats([]string{a.ID}, 30)
	req := map[string]bool{}
	for _, x := range a.RequiredList() {
		req[x] = true
	}
	pk := ""
	if c.Org.Demo {
		pk = s.svc.DB.Meta("demo_pk_" + strings.TrimPrefix(a.ID, "app_"))
	}
	s.renderConsole(w, r, c, "console_app", "apps", pageData{"App": a, "Keys": keys, "Stats": st, "Sets": app.Sets, "Req": req, "DemoPK": pk})
}

func (s *Server) consoleUpdateApp(w http.ResponseWriter, r *http.Request, c *consoleCtx) {
	upd := map[string]string{
		"name": r.PostFormValue("name"), "origins": r.PostFormValue("origins"), "required": strings.Join(r.PostForm["required"], ","),
		"duplicatePolicy": r.PostFormValue("duplicatePolicy"), "mode": r.PostFormValue("mode"), "webhookUrl": r.PostFormValue("webhookUrl"),
	}
	if _, err := s.svc.UpdateApp(c.Org.ID, r.PathValue("id"), upd); err != nil {
		setFlash(w, map[string]string{"error": err.Error()})
	} else {
		setFlash(w, map[string]string{"ok": "Settings saved."})
	}
	http.Redirect(w, r, "/console/apps/"+r.PathValue("id"), http.StatusSeeOther)
}

func (s *Server) consoleRotate(w http.ResponseWriter, r *http.Request, c *consoleCtx) {
	sk, err := s.svc.RotateSecret(c.Org.ID, r.PathValue("id"))
	if err != nil {
		setFlash(w, map[string]string{"error": err.Error()})
	} else {
		setFlash(w, map[string]string{"ok": "Secret key rotated. The previous key stopped working immediately.", "secret": sk})
	}
	http.Redirect(w, r, "/console/apps/"+r.PathValue("id"), http.StatusSeeOther)
}

// consoleTestSession creates a session and sends the developer through the hosted flow.
func (s *Server) consoleTestSession(w http.ResponseWriter, r *http.Request, c *consoleCtx) {
	a, err := s.svc.OrgApp(c.Org.ID, r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	vs, err := s.svc.CreateSession(a, app.SessionRequest{Reference: "console-test", ReturnURL: "/console/sessions/"}, s.svc.Cfg.BaseURL)
	if err != nil {
		setFlash(w, map[string]string{"error": err.Error()})
		http.Redirect(w, r, "/console/apps/"+a.ID, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/verify/"+vs.ID, http.StatusSeeOther)
}

func (s *Server) consoleSessions(w http.ResponseWriter, r *http.Request, c *consoleCtx) {
	status := r.URL.Query().Get("status")
	ids := appIDs(c)
	if a := r.URL.Query().Get("app"); a != "" {
		ids = []string{}
		for _, x := range c.Apps {
			if x.ID == a {
				ids = append(ids, a)
			}
		}
	}
	list, err := s.svc.Sessions(ids, 200, status)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.renderConsole(w, r, c, "console_sessions", "sessions", pageData{"Sessions": list, "Status": status, "AppFilter": r.URL.Query().Get("app"), "Apps": appNames(c)})
}

func (s *Server) consoleSession(w http.ResponseWriter, r *http.Request, c *consoleCtx) {
	vs, err := s.svc.Session(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	a, ok := appNames(c)[vs.AppID]
	if !ok {
		http.NotFound(w, r)
		return
	}
	var claims map[string]any
	if vs.Status == "verified" || vs.Status == "duplicate" {
		claims, _ = s.svc.VerifyResultToken(s.svc.ResultToken(vs))
	}
	pretty, _ := json.MarshalIndent(claims, "", "  ")
	s.renderConsole(w, r, c, "console_session", "sessions", pageData{"S": vs, "App": a, "Claims": string(pretty)})
}

func (s *Server) consoleHumans(w http.ResponseWriter, r *http.Request, c *consoleCtx) {
	bs, _ := s.svc.Bindings(appIDs(c), 150)
	dups, _ := s.svc.Sessions(appIDs(c), 60, "duplicate")
	st, _ := s.svc.Stats(appIDs(c), 30)
	s.renderConsole(w, r, c, "console_humans", "humans", pageData{"Bindings": bs, "Dups": dups, "Apps": appNames(c), "Stats": st})
}

func (s *Server) consoleAgents(w http.ResponseWriter, r *http.Request, c *consoleCtx) {
	ds, _ := s.svc.Delegations(appIDs(c), 150)
	active := 0
	for _, d := range ds {
		if d.RevokedAt == "" && d.NotAfter > app.Now() {
			active++
		}
	}
	s.renderConsole(w, r, c, "console_agents", "agents", pageData{"Delegations": ds, "Apps": appNames(c), "Active": active})
}

func (s *Server) consoleRevokeAgent(w http.ResponseWriter, r *http.Request, c *consoleCtx) {
	if err := s.svc.RevokeDelegation(appIDs(c), r.PathValue("id")); err != nil {
		setFlash(w, map[string]string{"error": err.Error()})
	} else {
		setFlash(w, map[string]string{"ok": "Delegation revoked — the agent's requests will now fail verification."})
	}
	http.Redirect(w, r, "/console/agents", http.StatusSeeOther)
}

func (s *Server) consoleWebhooks(w http.ResponseWriter, r *http.Request, c *consoleCtx) {
	ds, _ := s.svc.Deliveries(appIDs(c), 100)
	s.renderConsole(w, r, c, "console_webhooks", "webhooks", pageData{"Deliveries": ds, "Apps": appNames(c), "Example": app.SignWebhook("whsec_example", 1760000000, []byte(`{"type":"session.verified"}`))})
}

func (s *Server) consoleBilling(w http.ResponseWriter, r *http.Request, c *consoleCtx) {
	inv, err := s.svc.Invoice(c.Org)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	order := []string{"developer", "growth", "scale", "enterprise"}
	var plans []app.Plan
	for _, k := range order {
		plans = append(plans, app.Plans[k])
	}
	s.renderConsole(w, r, c, "console_billing", "billing", pageData{"Inv": inv, "Plans": plans})
}

func (s *Server) consolePlan(w http.ResponseWriter, r *http.Request, c *consoleCtx) {
	if err := s.svc.SetPlan(c.Org.ID, r.PostFormValue("plan")); err != nil {
		setFlash(w, map[string]string{"error": err.Error()})
	} else {
		setFlash(w, map[string]string{"ok": "Plan updated."})
	}
	http.Redirect(w, r, "/console/billing", http.StatusSeeOther)
}

type complianceRow struct {
	Regime, Obligation, Predicate string
	Checks, Passed                int
	Evidence                      string
}

func (s *Server) complianceRows(c *consoleCtx) []complianceRow {
	st, _ := s.svc.Stats(appIDs(c), 30)
	row := func(regime, obligation, pred, evidence string) complianceRow {
		return complianceRow{regime, obligation, pred, st.Requested[pred], st.Predicates[pred], evidence}
	}
	return []complianceRow{
		row("Australia — Online Safety Amendment (Social Media Minimum Age) Act 2024", "Take reasonable steps to prevent under-16s holding accounts; offer a non-ID alternative", "age_over_16", "Cryptographic age_over_16 proofs; in-person and vouching paths available; no ID retained"),
		row("United Kingdom — Online Safety Act 2023 (Part 5)", "Highly effective age assurance for adult content", "age_over_18", "Issuer-signed mDL/passport/EUDI attestation (IAL2); no personal data stored by the platform"),
		row("United States — state age-verification laws (e.g. TX HB 1181)", "Verify users are 18+ before granting access", "age_over_18", "Age predicate proofs with per-session challenge; replay-protected"),
		row("Platform integrity — one account per human", "Prevent ban evasion and sybil accounts", "human", "Scoped key images (nullifiers) with duplicate detection"),
	}
}

func (s *Server) consoleCompliance(w http.ResponseWriter, r *http.Request, c *consoleCtx) {
	s.renderConsole(w, r, c, "console_compliance", "compliance", pageData{"Rows": s.complianceRows(c), "Generated": time.Now().UTC().Format("2 Jan 2006 15:04 MST")})
}

func (s *Server) consoleComplianceCSV(w http.ResponseWriter, r *http.Request, c *consoleCtx) {
	var rows [][]string
	for _, x := range s.complianceRows(c) {
		rows = append(rows, []string{x.Regime, x.Obligation, x.Predicate, fmt.Sprint(x.Checks), fmt.Sprint(x.Passed), x.Evidence})
	}
	writeCSV(w, "singleton-compliance-evidence.csv", []string{"regime", "obligation", "predicate", "checks_30d", "passed_30d", "evidence"}, rows)
}
