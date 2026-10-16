package web

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stbodurov/bingo/singleton/internal/app"
	"github.com/stbodurov/bingo/singleton/internal/proof"
)

func (s *Server) routes() {
	m := s.mux
	// pages
	m.HandleFunc("GET /{$}", s.pageLanding)
	m.HandleFunc("GET /docs", s.pageDocs)
	m.HandleFunc("GET /transparency", s.pageTransparency)
	m.HandleFunc("GET /transparency/tx/{hash}", s.pageTx)
	m.HandleFunc("GET /pass", s.pagePass)
	m.HandleFunc("GET /verify/{id}", s.pageVerify)
	m.HandleFunc("GET /static/", s.serveStatic)
	m.HandleFunc("GET /sdk/singleton.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.ServeFileFS(w, r, s.static, "js/sdk.js")
	})
	m.HandleFunc("GET /favicon.svg", func(w http.ResponseWriter, r *http.Request) { http.ServeFileFS(w, r, s.static, "img/favicon.svg") })
	m.HandleFunc("GET /.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		writeJSON(w, 200, s.svc.JWKS())
	})

	// public registry API (used by the wallet and by anyone auditing the network)
	m.HandleFunc("GET /api/epoch", s.apiEpoch)
	m.HandleFunc("GET /api/predicates", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, app.Sets) })
	m.HandleFunc("GET /api/attesters", s.apiAttesters)
	m.HandleFunc("GET /api/transparency", s.apiTransparency)
	m.HandleFunc("GET /api/sets/{epoch}/{set}", s.apiSet)
	m.HandleFunc("GET /api/sets/{epoch}/{set}/bucket", s.limit("bucket", 240, time.Minute, s.apiBucket))
	m.HandleFunc("GET /api/sandbox/issuers", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, app.SandboxIssuers) })
	m.HandleFunc("POST /api/sandbox/credential", s.limit("sandbox", 30, time.Minute, s.apiSandboxCredential))
	m.HandleFunc("POST /api/enroll/credential", s.limit("enroll", 20, time.Minute, s.apiEnrollCredential))
	m.HandleFunc("GET /api/enroll/{pub}", s.apiEnrollment)
	m.HandleFunc("POST /api/enroll/in-person", s.limit("enroll", 20, time.Minute, s.apiInPersonCode))
	m.HandleFunc("GET /api/enroll/in-person/{code}", s.apiInPersonStatus)
	m.HandleFunc("POST /api/vouch", s.limit("vouch", 30, time.Minute, s.apiVouch))
	m.HandleFunc("GET /api/vouch/{pub}", s.apiVouchStatus)
	m.HandleFunc("POST /api/revoke", s.limit("revoke", 10, time.Minute, s.apiRevoke))
	m.HandleFunc("GET /api/verify/{id}", s.apiSessionPublic)
	m.HandleFunc("POST /api/verify/{id}/present", s.limit("present", 60, time.Minute, s.apiPresent))
	m.HandleFunc("GET /api/waitlist/challenge", s.limit("pow", 60, time.Minute, s.apiPowChallenge))
	m.HandleFunc("POST /api/waitlist", s.limit("waitlist", 12, time.Hour, s.apiWaitlistJoin))
	m.HandleFunc("GET /api/waitlist/{code}", s.apiWaitlistStatus)

	// developer API
	m.HandleFunc("OPTIONS /v1/{path...}", s.v1Preflight)
	m.HandleFunc("POST /v1/sessions", s.limit("v1", 600, time.Minute, s.v1CreateSession))
	m.HandleFunc("GET /v1/sessions/{id}", s.limit("v1", 600, time.Minute, s.v1GetSession))
	m.HandleFunc("POST /v1/presentations/verify", s.limit("v1", 600, time.Minute, s.v1VerifyPresentation))
	m.HandleFunc("POST /v1/delegations", s.limit("v1", 120, time.Minute, s.v1CreateDelegation))
	m.HandleFunc("POST /v1/agents/verify", s.limit("v1", 1200, time.Minute, s.v1VerifyAgent))
	m.HandleFunc("GET /v1/humans/{nullifier}", s.limit("v1", 600, time.Minute, s.v1Human))

	s.consoleRoutes()
	s.demoRoutes()
}

// ---------------------------------------------------------------- public registry API

func (s *Server) apiEpoch(w http.ResponseWriter, r *http.Request) {
	ep, err := s.svc.CurrentEpoch()
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"epoch": ep, "minRing": s.svc.Cfg.MinRing, "grace": s.svc.Cfg.EpochGrace, "network": app.AnchorNetwork, "registry": app.RegistryAddress})
}

func (s *Server) apiAttesters(w http.ResponseWriter, r *http.Request) {
	a, err := s.svc.Attesters()
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, a)
}

func (s *Server) transparencyData() (map[string]any, error) {
	ep, err := s.svc.CurrentEpoch()
	if err != nil {
		return nil, err
	}
	epochs, _ := s.svc.Epochs(12)
	txs, _ := s.svc.ChainTxs(12)
	atts, _ := s.svc.Attesters()
	revoked := 0
	for _, a := range atts {
		revoked += a.Revoked
	}
	return map[string]any{
		"epoch": ep, "epochs": epochs, "txs": txs, "attesters": atts, "revoked": revoked,
		"registry": app.RegistryAddress, "anchor": app.AnchorAddress, "network": app.AnchorNetwork, "sets": app.Sets,
	}, nil
}

func (s *Server) apiTransparency(w http.ResponseWriter, r *http.Request) {
	d, err := s.transparencyData()
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	writeJSON(w, 200, d)
}

func pathEpoch(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("epoch"), 10, 64)
}

// apiSet lets anyone download a set snapshot and recompute its anchored Merkle root.
func (s *Server) apiSet(w http.ResponseWriter, r *http.Request) {
	epoch, err := pathEpoch(r)
	if err != nil {
		writeErr(w, 400, "bad epoch")
		return
	}
	raw, err := s.svc.Members(epoch, r.PathValue("set"))
	if err != nil {
		writeErr(w, 404, err.Error())
		return
	}
	ep, _ := s.svc.Epoch(epoch)
	members := make([]string, len(raw))
	for i, m := range raw {
		members[i] = fmt.Sprintf("%x", m)
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	writeJSON(w, 200, map[string]any{"epoch": epoch, "set": r.PathValue("set"), "root": ep.Sets[r.PathValue("set")].Root, "anchorTx": ep.TxHash, "members": members})
}

func (s *Server) apiBucket(w http.ResponseWriter, r *http.Request) {
	epoch, err := pathEpoch(r)
	if err != nil {
		writeErr(w, 400, "bad epoch")
		return
	}
	b, err := s.svc.BucketFor(epoch, r.PathValue("set"), r.URL.Query().Get("pub"))
	if err != nil {
		writeErr(w, 404, err.Error())
		return
	}
	writeJSON(w, 200, b)
}

// ---------------------------------------------------------------- enrollment API

func (s *Server) apiSandboxCredential(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Issuer string     `json:"issuer"`
		Claims app.Claims `json:"claims"`
	}
	if err := decode(r, &in); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	c, err := s.svc.IssueSandboxCredential(in.Issuer, in.Claims)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, c)
}

func (s *Server) apiEnrollCredential(w http.ResponseWriter, r *http.Request) {
	var in struct {
		PubKey     string         `json:"pubkey"`
		Credential app.Credential `json:"credential"`
	}
	if err := decode(r, &in); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	e, err := s.svc.EnrollWithCredential(in.PubKey, &in.Credential)
	if err != nil {
		status := 400
		if errors.Is(err, app.ErrDuplicateHuman) || errors.Is(err, app.ErrKeyEnrolled) {
			status = 409
		}
		writeErr(w, status, err.Error())
		return
	}
	s.svc.MaybeSeal(time.Now().UTC(), true)
	writeJSON(w, 200, e)
}

func (s *Server) apiEnrollment(w http.ResponseWriter, r *http.Request) {
	e, err := s.svc.Enrollment(r.PathValue("pub"))
	if err != nil {
		writeErr(w, 404, "not enrolled")
		return
	}
	writeJSON(w, 200, e)
}

func (s *Server) apiInPersonCode(w http.ResponseWriter, r *http.Request) {
	var in struct {
		PubKey string `json:"pubkey"`
	}
	if err := decode(r, &in); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	code, err := s.svc.CreateInPersonCode(in.PubKey)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"code": code})
}

func (s *Server) apiInPersonStatus(w http.ResponseWriter, r *http.Request) {
	c, err := s.svc.InPersonStatus(r.PathValue("code"))
	if err != nil {
		writeErr(w, 404, "unknown code")
		return
	}
	writeJSON(w, 200, c)
}

func (s *Server) apiVouch(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Vouchee      string             `json:"vouchee"`
		Presentation proof.Presentation `json:"presentation"`
	}
	if err := decode(r, &in); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	n, e, err := s.svc.AddVouch(in.Vouchee, &in.Presentation)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	if e != nil {
		s.svc.MaybeSeal(time.Now().UTC(), true)
	}
	writeJSON(w, 200, map[string]any{"vouches": n, "required": app.VouchesRequired, "enrolled": e != nil})
}

func (s *Server) apiVouchStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"vouches": s.svc.VouchCount(r.PathValue("pub")), "required": app.VouchesRequired})
}

func (s *Server) apiRevoke(w http.ResponseWriter, r *http.Request) {
	var in struct {
		PubKey    string `json:"pubkey"`
		Signature string `json:"signature"`
	}
	if err := decode(r, &in); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	if err := s.svc.SelfRevoke(in.PubKey, in.Signature); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	s.svc.MaybeSeal(time.Now().UTC(), true)
	writeJSON(w, 200, map[string]bool{"revoked": true})
}

// ---------------------------------------------------------------- hosted verification

func (s *Server) apiSessionPublic(w http.ResponseWriter, r *http.Request) {
	vs, err := s.svc.Session(r.PathValue("id"))
	if err != nil {
		writeErr(w, 404, "unknown session")
		return
	}
	a, _ := s.svc.App(vs.AppID)
	writeJSON(w, 200, map[string]any{
		"id": vs.ID, "app": map[string]string{"name": a.Name, "color": a.Color, "category": a.Category}, "scope": vs.Scope,
		"challenge": vs.Challenge, "required": vs.Required, "action": vs.Action, "status": vs.Status, "origin": vs.Origin,
		"returnUrl": vs.ReturnURL, "expiresAt": vs.ExpiresAt, "mode": a.Mode,
	})
}

func (s *Server) apiPresent(w http.ResponseWriter, r *http.Request) {
	var p proof.Presentation
	if err := decode(r, &p); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	vs, err := s.svc.CompleteSession(r.PathValue("id"), &p)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, vs)
}

// ---------------------------------------------------------------- waitlist API

func (s *Server) apiPowChallenge(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.svc.NewPowChallenge())
}

func (s *Server) apiWaitlistJoin(w http.ResponseWriter, r *http.Request) {
	var in app.WaitlistInput
	if err := decode(r, &in); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	st, err := s.svc.JoinWaitlist(in, ipHash(r))
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, st)
}

func (s *Server) apiWaitlistStatus(w http.ResponseWriter, r *http.Request) {
	st, err := s.svc.WaitlistStatus(r.PathValue("code"))
	if err != nil {
		writeErr(w, 404, err.Error())
		return
	}
	writeJSON(w, 200, st)
}

// ---------------------------------------------------------------- pages

func (s *Server) pageLanding(w http.ResponseWriter, r *http.Request) {
	d, _ := s.transparencyData()
	people, platforms := s.svc.WaitlistCounts()
	s.render(w, r, "landing", pageData{"T": d, "People": people, "Platforms": platforms, "Plans": app.Plans, "Ref": r.URL.Query().Get("ref")})
}

func (s *Server) pageDocs(w http.ResponseWriter, r *http.Request) {
	s.render(w, r, "docs", pageData{"Sets": app.Sets, "JWKS": s.svc.JWKS(), "KeyID": s.svc.KeyID()})
}

func (s *Server) pageTransparency(w http.ResponseWriter, r *http.Request) {
	d, err := s.transparencyData()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.render(w, r, "transparency", pageData{"T": d})
}

func (s *Server) pageTx(w http.ResponseWriter, r *http.Request) {
	tx, err := s.svc.ChainTx(r.PathValue("hash"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	s.render(w, r, "tx", pageData{"Tx": tx, "Words": calldataWords(tx.Calldata), "Network": app.AnchorNetwork})
}

func calldataWords(cd string) []string {
	cd = strings.TrimPrefix(cd, "0x")
	if len(cd) < 8 {
		return nil
	}
	out := []string{"selector  0x" + cd[:8]}
	for i := 8; i+64 <= len(cd); i += 64 {
		out = append(out, fmt.Sprintf("[%03d]     0x%s", (i-8)/64, cd[i:i+64]))
	}
	return out
}

func (s *Server) pagePass(w http.ResponseWriter, r *http.Request) {
	s.render(w, r, "pass", pageData{"Issuers": app.SandboxIssuers, "Sets": app.Sets})
}

func (s *Server) pageVerify(w http.ResponseWriter, r *http.Request) {
	vs, err := s.svc.Session(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	a, _ := s.svc.App(vs.AppID)
	var defs []app.SetDef
	for _, req := range vs.Required {
		if d, ok := app.SetByName(req); ok {
			defs = append(defs, d)
		}
	}
	s.render(w, r, "verify", pageData{"Session": vs, "App": a, "Defs": defs})
}

func writeCSV(w http.ResponseWriter, name string, header []string, rows [][]string) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	cw := csv.NewWriter(w)
	_ = cw.Write(header)
	for _, row := range rows {
		for i, c := range row {
			if len(c) > 0 && strings.ContainsAny(c[:1], "=+-@") {
				row[i] = "'" + c
			}
		}
		_ = cw.Write(row)
	}
	cw.Flush()
}
