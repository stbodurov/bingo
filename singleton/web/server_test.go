package web

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stbodurov/bingo/singleton/internal/app"
)

func testServer(t *testing.T) (*httptest.Server, *app.Service) {
	t.Helper()
	t.Setenv("RATE_LIMIT", "off")
	db, err := app.OpenDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	cfg := app.DefaultConfig()
	svc, err := app.NewService(db, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Seed(300); err != nil {
		t.Fatal(err)
	}
	srv, err := New(svc, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(func() { ts.Close(); db.Close() })
	return ts, svc
}

func TestPagesRender(t *testing.T) {
	ts, _ := testServer(t)
	for _, p := range []string{"/", "/docs", "/transparency", "/pass", "/demo/driftwood", "/demo/frontrow", "/demo/commons", "/attester", "/admin", "/console/login", "/.well-known/jwks.json", "/sdk/singleton.js", "/api/epoch", "/api/attesters", "/api/predicates"} {
		res, err := http.Get(ts.URL + p)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode != 200 {
			t.Fatalf("%s → %d: %s", p, res.StatusCode, body)
		}
		if strings.Contains(string(body), "ZgotmplZ") {
			t.Fatalf("%s contains an unsafe template substitution", p)
		}
	}
	// Security headers are set on every response.
	res, _ := http.Get(ts.URL + "/")
	if !strings.Contains(res.Header.Get("Content-Security-Policy"), "script-src 'self'") || res.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("missing security headers")
	}
}

func TestV1RequiresKeysAndOrigins(t *testing.T) {
	ts, svc := testServer(t)
	post := func(key, origin, body string) (int, map[string]any) {
		req, _ := http.NewRequest("POST", ts.URL+"/v1/sessions", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if key != "" {
			req.Header.Set("Authorization", "Bearer "+key)
		}
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var out map[string]any
		_ = json.NewDecoder(res.Body).Decode(&out)
		return res.StatusCode, out
	}
	if code, _ := post("", "", `{}`); code != 401 {
		t.Fatalf("missing key should be 401, got %d", code)
	}
	sk := svc.DB.Meta("demo_sk_driftwood")
	pk := svc.DB.Meta("demo_pk_driftwood")
	code, out := post(sk, "", `{"required":["human"],"reference":"u1"}`)
	if code != 201 || out["url"] == nil {
		t.Fatalf("secret key session failed: %d %v", code, out)
	}
	if code, _ := post(pk, "https://evil.example", `{}`); code != 403 {
		t.Fatalf("publishable key from unknown origin should be 403, got %d", code)
	}
	if code, _ := post(pk, svc.Cfg.BaseURL, `{"required":["age_over_18"]}`); code != 201 {
		t.Fatalf("publishable key from allowed origin should work, got %d", code)
	}
	if code, out := post(sk, "", `{"required":["not_a_predicate"]}`); code != 400 {
		t.Fatalf("unknown predicate should be 400, got %d %v", code, out)
	}
	// GET with a publishable key is refused (results need a secret key).
	req, _ := http.NewRequest("GET", ts.URL+"/v1/sessions/"+out["id"].(string), nil)
	req.Header.Set("Authorization", "Bearer "+pk)
	req.Header.Set("Origin", svc.Cfg.BaseURL)
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != 403 {
		t.Fatalf("publishable key must not read sessions, got %d", res.StatusCode)
	}
}

func TestConsoleRequiresLoginAndCSRF(t *testing.T) {
	ts, _ := testServer(t)
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	res, _ := client.Get(ts.URL + "/console")
	if res.StatusCode != http.StatusSeeOther || !strings.HasPrefix(res.Header.Get("Location"), "/console/login") {
		t.Fatalf("console should redirect to login, got %d %s", res.StatusCode, res.Header.Get("Location"))
	}
	res, _ = client.Post(ts.URL+"/console/demo", "application/x-www-form-urlencoded", nil)
	var cookie *http.Cookie
	for _, c := range res.Cookies() {
		if c.Name == consoleCookie {
			cookie = c
		}
	}
	if cookie == nil {
		t.Fatal("demo login did not set a session cookie")
	}
	for _, p := range []string{"/console", "/console/apps", "/console/sessions", "/console/humans", "/console/agents", "/console/webhooks", "/console/billing", "/console/compliance", "/console/apps/app_driftwood"} {
		req, _ := http.NewRequest("GET", ts.URL+p, nil)
		req.AddCookie(cookie)
		r, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(r.Body)
		r.Body.Close()
		if r.StatusCode != 200 {
			t.Fatalf("%s → %d", p, r.StatusCode)
		}
		if strings.Contains(string(b), "ZgotmplZ") {
			t.Fatalf("%s contains an unsafe template substitution", p)
		}
	}
	// POST without CSRF token is rejected.
	req, _ := http.NewRequest("POST", ts.URL+"/console/apps", strings.NewReader("name=Evil"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)
	r, _ := client.Do(req)
	if r.StatusCode != http.StatusForbidden {
		t.Fatalf("POST without CSRF should be 403, got %d", r.StatusCode)
	}
}
