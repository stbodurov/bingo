package web

import (
	"net/http"
	"strings"

	"github.com/stbodurov/bingo/singleton/internal/app"
	"github.com/stbodurov/bingo/singleton/internal/proof"
)

// The developer API. Secret keys (sk_) are for servers; publishable keys (pk_) may be used
// from browsers, but only from the app's allowed origins and only to start sessions or
// register delegations — results are always signed tokens the RP backend verifies.

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

func (s *Server) authKey(w http.ResponseWriter, r *http.Request, allowPublishable bool) (*app.App, string, bool) {
	raw := bearer(r)
	if raw == "" {
		writeErr(w, 401, "missing Authorization: Bearer <api key>")
		return nil, "", false
	}
	a, kind, err := s.svc.AuthenticateKey(raw)
	if err != nil {
		writeErr(w, 401, err.Error())
		return nil, "", false
	}
	if kind == "publishable" {
		if !allowPublishable {
			writeErr(w, 403, "this endpoint needs a secret key (sk_…) from your server")
			return nil, "", false
		}
		origin := r.Header.Get("Origin")
		if origin == "" || !a.AllowsOrigin(origin) {
			writeErr(w, 403, "origin "+origin+" is not allowed for this app — add it in the console")
			return nil, "", false
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
	}
	return a, kind, true
}

func (s *Server) v1Preflight(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
	}
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	w.Header().Set("Access-Control-Max-Age", "600")
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) v1CreateSession(w http.ResponseWriter, r *http.Request) {
	a, _, ok := s.authKey(w, r, true)
	if !ok {
		return
	}
	var req app.SessionRequest
	if err := decode(r, &req); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	vs, err := s.svc.CreateSession(a, req, r.Header.Get("Origin"))
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, vs)
}

func (s *Server) v1GetSession(w http.ResponseWriter, r *http.Request) {
	a, _, ok := s.authKey(w, r, false)
	if !ok {
		return
	}
	vs, err := s.svc.Session(r.PathValue("id"))
	if err != nil || vs.AppID != a.ID {
		writeErr(w, 404, "session not found")
		return
	}
	if vs.Status == "verified" || vs.Status == "duplicate" {
		vs.Token = s.svc.ResultToken(vs)
	}
	writeJSON(w, 200, vs)
}

// v1VerifyPresentation is for RPs that issue their own challenges (e.g. native apps).
func (s *Server) v1VerifyPresentation(w http.ResponseWriter, r *http.Request) {
	a, _, ok := s.authKey(w, r, false)
	if !ok {
		return
	}
	var in struct {
		Presentation proof.Presentation `json:"presentation"`
		Challenge    string             `json:"challenge"`
		Required     []string           `json:"required"`
		Action       string             `json:"action"`
	}
	if err := decode(r, &in); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	if len(in.Challenge) < 16 {
		writeErr(w, 400, "challenge must be at least 16 random characters you generated")
		return
	}
	scope := a.Scope
	if in.Action != "" {
		scope += "#" + strings.ToLower(in.Action)
	}
	v, err := s.svc.VerifyPresentation(&in.Presentation, scope, in.Challenge, in.Required)
	if err != nil {
		writeJSON(w, 200, map[string]any{"valid": false, "error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"valid": true, "nullifier": v.Nullifier, "predicates": v.Predicates, "ringSizes": v.RingSizes, "epoch": v.Epoch, "scope": scope})
}

func (s *Server) v1CreateDelegation(w http.ResponseWriter, r *http.Request) {
	a, _, ok := s.authKey(w, r, true)
	if !ok {
		return
	}
	var in struct {
		Delegation   proof.Delegation   `json:"delegation"`
		Presentation proof.Presentation `json:"presentation"`
	}
	if err := decode(r, &in); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	rec, err := s.svc.CreateDelegation(a, in.Delegation, &in.Presentation)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, rec)
}

func (s *Server) v1VerifyAgent(w http.ResponseWriter, r *http.Request) {
	a, _, ok := s.authKey(w, r, false)
	if !ok {
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
	d, err := s.svc.VerifyAgentRequest(a.ID, in.DelegationID, in.Message, in.Signature)
	if err != nil {
		writeJSON(w, 200, map[string]any{"valid": false, "error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"valid": true, "nullifier": d.Nullifier, "label": d.Label, "permissions": d.Permissions, "notAfter": d.NotAfter})
}

func (s *Server) v1Human(w http.ResponseWriter, r *http.Request) {
	a, _, ok := s.authKey(w, r, false)
	if !ok {
		return
	}
	bs, err := s.svc.Bindings([]string{a.ID}, 5000)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	var out []app.Binding
	for _, b := range bs {
		if b.Nullifier == r.PathValue("nullifier") {
			out = append(out, b)
		}
	}
	if len(out) == 0 {
		writeErr(w, 404, "this pseudonym has never verified with your app")
		return
	}
	writeJSON(w, 200, map[string]any{"nullifier": r.PathValue("nullifier"), "bindings": out})
}
