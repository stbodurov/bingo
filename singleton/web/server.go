// Package web is Singleton's HTTP surface: the landing page, the Singleton Pass wallet,
// the hosted consent screen, the developer API (/v1), the relying-party console, the
// transparency explorer, the in-person attester console and three demo apps.
package web

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/stbodurov/bingo/singleton/internal/app"
)

//go:embed templates static
var assets embed.FS

type Server struct {
	svc     *app.Service
	log     *log.Logger
	pages   map[string]*template.Template
	static  fs.FS
	mux     *http.ServeMux
	limiter *limiter
	wasmMu  sync.Mutex
	wasm    []byte
	wasmGz  []byte
}

func New(svc *app.Service, logger *log.Logger) (*Server, error) {
	s := &Server{svc: svc, log: logger, mux: http.NewServeMux(), limiter: newLimiter()}
	st, _ := fs.Sub(assets, "static")
	s.static = st
	if err := s.parseTemplates(); err != nil {
		return nil, err
	}
	s.routes()
	return s, nil
}

func (s *Server) parseTemplates() error {
	s.pages = map[string]*template.Template{}
	layouts := map[string]string{"site": "templates/layout_site.html", "console": "templates/layout_console.html", "app": "templates/layout_app.html"}
	entries, err := fs.Glob(assets, "templates/*.html")
	if err != nil {
		return err
	}
	for _, path := range entries {
		name := strings.TrimSuffix(filepath.Base(path), ".html")
		if strings.HasPrefix(name, "layout_") || strings.HasPrefix(name, "partial_") {
			continue
		}
		layout := "site"
		switch {
		case strings.HasPrefix(name, "console_"):
			layout = "console"
		case strings.HasPrefix(name, "pass") || strings.HasPrefix(name, "verify") || strings.HasPrefix(name, "attester") || strings.HasPrefix(name, "demo_") || name == "admin":
			layout = "app"
		}
		t := template.New(name).Funcs(funcs())
		files := []string{layouts[layout], path}
		partials, _ := fs.Glob(assets, "templates/partial_*.html")
		files = append(files, partials...)
		if _, err := t.ParseFS(assets, files...); err != nil {
			return fmt.Errorf("template %s: %w", name, err)
		}
		s.pages[name] = t
	}
	return nil
}

func (s *Server) Handler() http.Handler { return s.middleware(s.mux) }

// ---------------------------------------------------------------- middleware

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(c int) { w.status = c; w.ResponseWriter.WriteHeader(c) }

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), interest-cohort=()")
		if strings.HasPrefix(r.URL.Path, "/verify/") {
			// The hosted consent page is opened as a popup by other origins and must be able
			// to postMessage its result back to its opener.
			h.Set("Cross-Origin-Opener-Policy", "unsafe-none")
		} else {
			h.Set("Cross-Origin-Opener-Policy", "same-origin-allow-popups")
		}
		// 'wasm-unsafe-eval' lets the wallet instantiate WebAssembly; no inline scripts anywhere.
		h.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'wasm-unsafe-eval'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		if !strings.HasPrefix(r.URL.Path, "/demo/") {
			h.Set("X-Frame-Options", "DENY")
		}
		sw := &statusWriter{ResponseWriter: w, status: 200}
		defer func() {
			if rec := recover(); rec != nil {
				s.log.Printf("panic %s %s: %v", r.Method, r.URL.Path, rec)
				http.Error(sw, "internal error", 500)
			}
			if s.log != nil && (strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/v1/") || sw.status >= 400) {
				s.log.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.status, time.Since(start).Round(time.Millisecond))
			}
		}()
		next.ServeHTTP(sw, r)
	})
}

// limiter is a per-IP fixed-window rate limiter.
type limiter struct {
	mu   sync.Mutex
	hits map[string]*window
}

type window struct {
	start time.Time
	n     int
}

func newLimiter() *limiter { return &limiter{hits: map[string]*window{}} }

func (l *limiter) allow(key string, max int, per time.Duration) bool {
	if os.Getenv("RATE_LIMIT") == "off" {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	w := l.hits[key]
	if w == nil || now.Sub(w.start) > per {
		if len(l.hits) > 50000 {
			l.hits = map[string]*window{}
		}
		w = &window{start: now}
		l.hits[key] = w
	}
	w.n++
	return w.n <= max
}

func clientIP(r *http.Request) string {
	if f := r.Header.Get("X-Forwarded-For"); f != "" {
		return strings.TrimSpace(strings.Split(f, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func ipHash(r *http.Request) string {
	sum := sha256.Sum256([]byte("singleton/ip/" + clientIP(r)))
	return hex.EncodeToString(sum[:8])
}

func (s *Server) limit(name string, max int, per time.Duration, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.limiter.allow(name+"|"+clientIP(r), max, per) {
			w.Header().Set("Retry-After", fmt.Sprint(int(per.Seconds())))
			writeErr(w, http.StatusTooManyRequests, "too many requests — slow down")
			return
		}
		h(w, r)
	}
}

// ---------------------------------------------------------------- responses

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decode(r *http.Request, v any) error {
	ct := r.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		return errors.New("Content-Type must be application/json")
	}
	dec := json.NewDecoder(io.LimitReader(r.Body, 2<<20))
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("malformed JSON: %v", err)
	}
	return nil
}

type pageData map[string]any

func (s *Server) render(w http.ResponseWriter, r *http.Request, name string, data pageData) {
	t, ok := s.pages[name]
	if !ok {
		http.Error(w, "template not found: "+name, 500)
		return
	}
	if data == nil {
		data = pageData{}
	}
	data["Path"] = r.URL.Path
	data["BaseURL"] = s.svc.Cfg.BaseURL
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout", data); err != nil {
		s.log.Printf("render %s: %v", name, err)
		http.Error(w, "render error", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(buf.Bytes())
}

// ---------------------------------------------------------------- static + wasm

func (s *Server) serveStatic(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(r.URL.Path, "/static/")
	if p == "wasm/wallet.wasm" {
		s.serveWasm(w, r)
		return
	}
	if strings.HasSuffix(p, ".js") || strings.HasSuffix(p, ".css") {
		w.Header().Set("Cache-Control", "public, max-age=300")
	}
	http.ServeFileFS(w, r, s.static, p)
}

// serveWasm serves the wallet core. If it was not built ahead of time (make wasm), it is
// compiled once on first request with the local Go toolchain so `go run` just works.
func (s *Server) serveWasm(w http.ResponseWriter, r *http.Request) {
	s.wasmMu.Lock()
	if s.wasm == nil {
		if b, err := fs.ReadFile(s.static, "wasm/wallet.wasm"); err == nil {
			s.wasm = b
		} else if b, err := os.ReadFile("web/static/wasm/wallet.wasm"); err == nil {
			s.wasm = b
		} else {
			out := filepath.Join(os.TempDir(), "singleton-wallet.wasm")
			cmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", out, "./cmd/wallet-wasm")
			cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
			if msg, err := cmd.CombinedOutput(); err != nil {
				s.wasmMu.Unlock()
				s.log.Printf("wasm build failed: %v\n%s", err, msg)
				http.Error(w, "wallet core not built — run `make wasm`", 503)
				return
			}
			s.wasm, _ = os.ReadFile(out)
		}
	}
	if s.wasmGz == nil {
		var gz bytes.Buffer
		zw, _ := gzip.NewWriterLevel(&gz, gzip.BestCompression)
		_, _ = zw.Write(s.wasm)
		_ = zw.Close()
		s.wasmGz = gz.Bytes()
	}
	b, gz := s.wasm, s.wasmGz
	s.wasmMu.Unlock()
	w.Header().Set("Content-Type", "application/wasm")
	w.Header().Set("Cache-Control", "public, max-age=600")
	w.Header().Set("Vary", "Accept-Encoding")
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Length", fmt.Sprint(len(gz)))
		_, _ = w.Write(gz)
		return
	}
	http.ServeContent(w, r, "wallet.wasm", time.Time{}, bytes.NewReader(b))
}
