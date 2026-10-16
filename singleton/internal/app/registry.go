package app

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/stbodurov/bingo/singleton/internal/proof"
	"golang.org/x/text/unicode/norm"
)

// ---------------------------------------------------------------- predicate sets

type SetDef struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Group       string `json:"group"`
}

// Sets are the predicates a relying party can ask for. A holder's key is placed in every
// set it qualifies for when an epoch is sealed.
var Sets = []SetDef{
	{"human", "Unique human", "Verified once by any trusted attester, including vouching.", "personhood"},
	{"human_doc", "Document-verified human", "Verified with a government credential or in person (≈ NIST IAL2).", "personhood"},
	{"age_over_13", "Over 13", "Holder's credential shows they turned 13.", "age"},
	{"age_over_16", "Over 16", "Required for Australian social media accounts.", "age"},
	{"age_over_18", "Over 18", "Adult content, gambling, alcohol in most markets.", "age"},
	{"age_over_21", "Over 21", "US alcohol, cannabis and some gaming.", "age"},
	{"region_us", "Credential issued in the US", "Coarse jurisdiction for compliance routing.", "region"},
	{"region_gb", "Credential issued in the UK", "For Online Safety Act age assurance.", "region"},
	{"region_au", "Credential issued in Australia", "For the Social Media Minimum Age obligation.", "region"},
	{"region_eu", "Credential issued in the EU", "Any EU member state.", "region"},
}

func SetByName(name string) (SetDef, bool) {
	for _, s := range Sets {
		if s.Name == name {
			return s, true
		}
	}
	return SetDef{}, false
}

var euCountries = map[string]bool{"AT": true, "BE": true, "BG": true, "HR": true, "CY": true, "CZ": true, "DK": true, "EE": true, "FI": true, "FR": true, "DE": true, "GR": true, "HU": true, "IE": true, "IT": true, "LV": true, "LT": true, "LU": true, "MT": true, "NL": true, "PL": true, "PT": true, "RO": true, "SK": true, "SI": true, "ES": true, "SE": true}

func regionFor(country string) string {
	c := strings.ToUpper(country)
	switch {
	case c == "US", c == "GB", c == "AU":
		return strings.ToLower(c)
	case euCountries[c]:
		return "eu"
	}
	return ""
}

// ---------------------------------------------------------------- attesters

type Attester struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Assurance   string `json:"assurance"`
	Region      string `json:"region"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Enrolled    int    `json:"enrolled"`
	Revoked     int    `json:"revoked"`
	Rejected    int    `json:"rejected"`
}

var defaultAttesters = []Attester{
	{ID: "att_mdl", Name: "Mobile driver's licences", Kind: "mdl", Assurance: "IAL2", Region: "US", Description: "ISO/IEC 18013-5 mDL presented from Apple, Google or state wallets; issuer signatures checked against the AAMVA Digital Trust Service."},
	{ID: "att_passport", Name: "ePassport chip", Kind: "passport", Assurance: "IAL2", Region: "Global", Description: "ICAO 9303 passive authentication of the NFC chip (DG1/SOD) against country signing CAs from the ICAO PKD."},
	{ID: "att_eudi", Name: "EU Digital Identity Wallet", Kind: "eudi", Assurance: "IAL2", Region: "EU", Description: "Person Identification Data presented over OpenID4VP from a member-state EUDI wallet."},
	{ID: "att_inperson", Name: "In-person partner network", Kind: "in_person", Assurance: "IAL2", Region: "US · UK · AU", Description: "Trained agents at partner pharmacies and post offices inspect a physical ID and approve a one-time code. Nothing is photographed or kept."},
	{ID: "att_vouch", Name: "Vouching circle", Kind: "vouch", Assurance: "IAL1", Region: "Global", Description: "Three document-verified humans vouch for you. Grants 'human' only — the inclusive path for people without documents."},
}

func (s *Service) ensureAttesters() error {
	for _, a := range defaultAttesters {
		if _, err := s.DB.Exec(`INSERT INTO attesters(id, name, kind, assurance, region, description, created_at) VALUES(?,?,?,?,?,?,?)
			ON CONFLICT(id) DO UPDATE SET name=excluded.name, description=excluded.description, region=excluded.region`,
			a.ID, a.Name, a.Kind, a.Assurance, a.Region, a.Description, Now()); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Attesters() ([]Attester, error) {
	rows, err := s.DB.Query(`SELECT a.id, a.name, a.kind, a.assurance, a.region, a.status, a.description,
		(SELECT COUNT(*) FROM enrollments e WHERE e.attester_id = a.id AND e.status = 'active'),
		(SELECT COUNT(*) FROM enrollments e WHERE e.attester_id = a.id AND e.status = 'revoked'),
		(SELECT COUNT(*) FROM enrollment_attempts t WHERE t.attester_id = a.id AND t.outcome != 'enrolled')
		FROM attesters a ORDER BY a.rowid`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Attester
	for rows.Next() {
		var a Attester
		if err := rows.Scan(&a.ID, &a.Name, &a.Kind, &a.Assurance, &a.Region, &a.Status, &a.Description, &a.Enrolled, &a.Revoked, &a.Rejected); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------- sandbox credentials

// SandboxIssuer simulates a credential issuer (a DMV, a passport authority, an EU PID
// provider). Production replaces signature checks with mdoc COSE / ICAO SOD / SD-JWT VC
// verification against real trust lists; the enrollment logic after that is identical.
type SandboxIssuer struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Country  string `json:"country"`
	Attester string `json:"attester"`
}

var SandboxIssuers = []SandboxIssuer{
	{"sbx-dmv-ca", "California DMV (sandbox)", "mdl", "US", "att_mdl"},
	{"sbx-dmv-ny", "New York DMV (sandbox)", "mdl", "US", "att_mdl"},
	{"sbx-csca-gb", "HM Passport Office (sandbox)", "passport", "GB", "att_passport"},
	{"sbx-csca-au", "Australian Passport Office (sandbox)", "passport", "AU", "att_passport"},
	{"sbx-pid-de", "German EUDI PID provider (sandbox)", "eudi", "DE", "att_eudi"},
}

func sandboxIssuer(id string) (SandboxIssuer, bool) {
	for _, i := range SandboxIssuers {
		if i.ID == id {
			return i, true
		}
	}
	return SandboxIssuer{}, false
}

type Claims struct {
	GivenName      string `json:"given_name"`
	FamilyName     string `json:"family_name"`
	BirthDate      string `json:"birth_date"`
	DocumentNumber string `json:"document_number"`
	IssuingCountry string `json:"issuing_country"`
}

type Credential struct {
	Issuer    string `json:"issuer"`
	Type      string `json:"type"`
	Claims    Claims `json:"claims"`
	IssuedAt  string `json:"issued_at"`
	Signature string `json:"signature"`
}

func (c Credential) signingInput() []byte {
	b, _ := json.Marshal(struct {
		Issuer   string `json:"issuer"`
		Type     string `json:"type"`
		Claims   Claims `json:"claims"`
		IssuedAt string `json:"issued_at"`
	}{c.Issuer, c.Type, c.Claims, c.IssuedAt})
	return b
}

var dateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func (c Claims) validate() error {
	if strings.TrimSpace(c.GivenName) == "" || strings.TrimSpace(c.FamilyName) == "" {
		return errors.New("credential is missing a name")
	}
	if !dateRe.MatchString(c.BirthDate) {
		return errors.New("credential birth_date must be YYYY-MM-DD")
	}
	b, err := time.Parse("2006-01-02", c.BirthDate)
	if err != nil || b.After(time.Now()) || b.Before(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)) {
		return errors.New("credential birth_date is not plausible")
	}
	if len(strings.TrimSpace(c.DocumentNumber)) < 5 {
		return errors.New("credential document_number is too short")
	}
	if len(c.IssuingCountry) != 2 {
		return errors.New("credential issuing_country must be ISO 3166 alpha-2")
	}
	return nil
}

// IssueSandboxCredential signs claims as a sandbox issuer (demo/testing only).
func (s *Service) IssueSandboxCredential(issuerID string, claims Claims) (*Credential, error) {
	iss, ok := sandboxIssuer(issuerID)
	if !ok {
		return nil, errors.New("unknown sandbox issuer")
	}
	claims.IssuingCountry = iss.Country
	if err := claims.validate(); err != nil {
		return nil, err
	}
	c := &Credential{Issuer: iss.ID, Type: iss.Kind, Claims: claims, IssuedAt: Now()}
	c.Signature = base64.RawURLEncoding.EncodeToString(ed25519.Sign(s.issuerSKs[iss.ID], c.signingInput()))
	return c, nil
}

func (s *Service) verifyCredential(c *Credential) (SandboxIssuer, error) {
	iss, ok := sandboxIssuer(c.Issuer)
	if !ok {
		return iss, errors.New("credential issuer is not on the trust list")
	}
	if c.Type != iss.Kind {
		return iss, errors.New("credential type does not match issuer")
	}
	sig, err := base64.RawURLEncoding.DecodeString(c.Signature)
	if err != nil || !ed25519.Verify(s.issuers[iss.ID], c.signingInput(), sig) {
		return iss, errors.New("issuer signature is invalid — the credential was altered or forged")
	}
	if c.Claims.IssuingCountry != iss.Country {
		return iss, errors.New("issuing country does not match issuer")
	}
	return iss, c.Claims.validate()
}

// ---------------------------------------------------------------- enrollment

var (
	ErrDuplicateHuman = errors.New("this person already holds an active Singleton credential — recover that pass instead of creating a second one")
	ErrKeyEnrolled    = errors.New("this key is already enrolled")
	ErrBadKey         = errors.New("public key is not a valid ristretto255 point")
)

type Enrollment struct {
	ID        string   `json:"id"`
	PubKey    string   `json:"pubkey"`
	Attester  string   `json:"attester"`
	Assurance string   `json:"assurance"`
	Region    string   `json:"region"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"createdAt"`
	Sets      []string `json:"sets"`
	// Upcoming age thresholds the holder will cross (dates in the past are not stored).
	Upcoming map[string]string `json:"upcoming,omitempty"`
}

func normalizeName(s string) string {
	t := norm.NFKD.String(strings.ToLower(s))
	var b strings.Builder
	for _, r := range t {
		if unicode.IsLetter(r) && r < unicode.MaxASCII || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// thresholdDates returns, for each age threshold, the date the holder crosses it — or
// "past" if already crossed, so adults' birthdays are never retained.
func thresholdDates(birth string, now time.Time) map[int]string {
	b, _ := time.Parse("2006-01-02", birth)
	out := map[int]string{}
	for _, n := range []int{13, 16, 18, 21} {
		at := b.AddDate(n, 0, 0)
		if !at.After(now) {
			out[n] = "past"
		} else {
			out[n] = at.Format("2006-01-02")
		}
	}
	return out
}

func validPub(pub string) bool {
	_, err := proof.DecodeElementHex(pub)
	return err == nil
}

type enrollInput struct {
	pub, attester, assurance, docHash, personHash, region string
	ages                                                   map[int]string
	synthetic                                              bool
	createdAt                                              string
}

func (s *Service) insertEnrollment(in enrollInput) (*Enrollment, error) {
	if !validPub(in.pub) {
		return nil, ErrBadKey
	}
	if count(s.DB, `SELECT COUNT(*) FROM enrollments WHERE pubkey = ?`, in.pub) > 0 {
		return nil, ErrKeyEnrolled
	}
	if in.docHash != "" && count(s.DB, `SELECT COUNT(*) FROM enrollments WHERE status='active' AND (doc_hash = ? OR person_hash = ?)`, in.docHash, in.personHash) > 0 {
		s.recordAttempt(in.attester, "duplicate", "document or person already enrolled")
		return nil, ErrDuplicateHuman
	}
	id := newID("enr")
	created := in.createdAt
	if created == "" {
		created = Now()
	}
	age := func(n int) any {
		if v, ok := in.ages[n]; ok {
			if v == "past" {
				return "1900-01-01"
			}
			return v
		}
		return nil
	}
	_, err := s.DB.Exec(`INSERT INTO enrollments(id, pubkey, attester_id, assurance, doc_hash, person_hash, region, age13_at, age16_at, age18_at, age21_at, synthetic, created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, id, in.pub, in.attester, in.assurance, nullable(in.docHash), nullable(in.personHash), in.region,
		age(13), age(16), age(18), age(21), boolInt(in.synthetic), created)
	if err != nil {
		return nil, err
	}
	if !in.synthetic {
		s.recordAttempt(in.attester, "enrolled", "")
		s.DB.Audit("registry", "enrolled", fmt.Sprintf("%s via %s (%s)", proofFingerprint(in.pub), in.attester, in.assurance))
	}
	_ = s.DB.SetMeta("registry_dirty", "1")
	return s.Enrollment(in.pub)
}

func (s *Service) recordAttempt(attester, outcome, reason string) {
	_, _ = s.DB.Exec(`INSERT INTO enrollment_attempts(id, attester_id, outcome, reason, created_at) VALUES(?,?,?,?,?)`, newID("att"), attester, outcome, reason, Now())
}

// EnrollWithCredential is the mDL / passport / EUDI path.
func (s *Service) EnrollWithCredential(pub string, c *Credential) (*Enrollment, error) {
	iss, err := s.verifyCredential(c)
	if err != nil {
		s.recordAttempt(iss.Attester, "rejected", err.Error())
		return nil, err
	}
	cl := c.Claims
	docHash := s.keyedHash("doc", strings.ToUpper(cl.IssuingCountry+"|"+c.Type+"|"+strings.ReplaceAll(cl.DocumentNumber, " ", "")))
	personHash := s.keyedHash("person", normalizeName(cl.GivenName)+"|"+normalizeName(cl.FamilyName)+"|"+cl.BirthDate+"|"+strings.ToUpper(cl.IssuingCountry))
	return s.insertEnrollment(enrollInput{
		pub: pub, attester: iss.Attester, assurance: "IAL2", docHash: docHash, personHash: personHash,
		region: regionFor(cl.IssuingCountry), ages: thresholdDates(cl.BirthDate, time.Now().UTC()),
	})
}

// Enrollment returns the registry view of a key (no personal data exists to return).
func (s *Service) Enrollment(pub string) (*Enrollment, error) {
	var e Enrollment
	var a13, a16, a18, a21 *string
	err := s.DB.QueryRow(`SELECT id, pubkey, attester_id, assurance, region, status, created_at, age13_at, age16_at, age18_at, age21_at FROM enrollments WHERE pubkey = ?`, pub).
		Scan(&e.ID, &e.PubKey, &e.Attester, &e.Assurance, &e.Region, &e.Status, &e.CreatedAt, &a13, &a16, &a18, &a21)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format("2006-01-02")
	e.Upcoming = map[string]string{}
	if e.Status == "active" {
		e.Sets = append(e.Sets, "human")
		if e.Assurance == "IAL2" {
			e.Sets = append(e.Sets, "human_doc")
		}
		for _, x := range []struct {
			n string
			v *string
		}{{"age_over_13", a13}, {"age_over_16", a16}, {"age_over_18", a18}, {"age_over_21", a21}} {
			if x.v == nil {
				continue
			}
			if *x.v <= now {
				e.Sets = append(e.Sets, x.n)
			} else {
				e.Upcoming[x.n] = *x.v
			}
		}
		if e.Region != "" {
			e.Sets = append(e.Sets, "region_"+e.Region)
		}
	}
	return &e, nil
}

// Revoke removes a key from all future epochs.
func (s *Service) Revoke(pub, reason, actor string) error {
	res, err := s.DB.Exec(`UPDATE enrollments SET status='revoked', revoked_reason=?, revoked_at=? WHERE pubkey=? AND status='active'`, reason, Now(), pub)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("no active enrollment for that key")
	}
	s.DB.Audit(actor, "revoked", fmt.Sprintf("%s — %s", proofFingerprint(pub), reason))
	return s.DB.SetMeta("registry_dirty", "1")
}

// SelfRevoke lets a holder revoke their own key with a Schnorr signature.
func (s *Service) SelfRevoke(pub, sig string) error {
	el, err := proof.DecodeElementHex(pub)
	if err != nil {
		return ErrBadKey
	}
	ss, err := proof.ParseSchnorr(sig)
	if err != nil || !proof.VerifySchnorr(el, []byte("singleton/revoke/v1:"+pub), ss) {
		return errors.New("signature does not prove control of this key")
	}
	return s.Revoke(pub, "revoked by holder", "holder")
}

// ---------------------------------------------------------------- in-person path

func (s *Service) CreateInPersonCode(pub string) (string, error) {
	if !validPub(pub) {
		return "", ErrBadKey
	}
	code := strings.ToUpper(randomHex(3)[:6])
	code = code[:3] + "-" + code[3:]
	_, err := s.DB.Exec(`INSERT INTO in_person_codes(code, pubkey, created_at) VALUES(?,?,?)`, code, pub, Now())
	return code, err
}

type InPersonCode struct {
	Code      string `json:"code"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	PubKey    string `json:"-"`
}

func (s *Service) InPersonStatus(code string) (*InPersonCode, error) {
	var c InPersonCode
	err := s.DB.QueryRow(`SELECT code, status, created_at, pubkey FROM in_person_codes WHERE code = ?`, strings.ToUpper(code)).Scan(&c.Code, &c.Status, &c.CreatedAt, &c.PubKey)
	return &c, err
}

func (s *Service) PendingInPerson() ([]InPersonCode, error) {
	rows, err := s.DB.Query(`SELECT code, status, created_at, pubkey FROM in_person_codes WHERE status='pending' ORDER BY created_at DESC LIMIT 50`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []InPersonCode
	for rows.Next() {
		var c InPersonCode
		if err := rows.Scan(&c.Code, &c.Status, &c.CreatedAt, &c.PubKey); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ApproveInPerson is called by a partner agent after inspecting a physical ID.
func (s *Service) ApproveInPerson(code string, cl Claims, docType string) (*Enrollment, error) {
	c, err := s.InPersonStatus(code)
	if err != nil {
		return nil, errors.New("unknown code")
	}
	if c.Status != "pending" {
		return nil, errors.New("code already used")
	}
	if err := cl.validate(); err != nil {
		return nil, err
	}
	if docType == "" {
		docType = "physical_id"
	}
	docHash := s.keyedHash("doc", strings.ToUpper(cl.IssuingCountry+"|"+docType+"|"+strings.ReplaceAll(cl.DocumentNumber, " ", "")))
	personHash := s.keyedHash("person", normalizeName(cl.GivenName)+"|"+normalizeName(cl.FamilyName)+"|"+cl.BirthDate+"|"+strings.ToUpper(cl.IssuingCountry))
	e, err := s.insertEnrollment(enrollInput{pub: c.PubKey, attester: "att_inperson", assurance: "IAL2", docHash: docHash, personHash: personHash,
		region: regionFor(cl.IssuingCountry), ages: thresholdDates(cl.BirthDate, time.Now().UTC())})
	status := "approved"
	if err != nil {
		status = "rejected"
	}
	_, _ = s.DB.Exec(`UPDATE in_person_codes SET status=?, decided_at=?, attester_id='att_inperson' WHERE code=?`, status, Now(), c.Code)
	return e, err
}

func (s *Service) RejectInPerson(code string) error {
	_, err := s.DB.Exec(`UPDATE in_person_codes SET status='rejected', decided_at=? WHERE code=? AND status='pending'`, Now(), strings.ToUpper(code))
	if err == nil {
		s.recordAttempt("att_inperson", "rejected", "agent rejected")
	}
	return err
}

// ---------------------------------------------------------------- vouching path

const (
	VouchScope         = "singleton:vouch"
	VouchesRequired    = 3
	VouchesPerVoucher  = 5 // per calendar month
)

// AddVouch records a document-verified human vouching for a new key. The voucher proves
// membership in human_doc inside the vouch scope, so the registry can rate-limit vouchers
// by pseudonym without learning who they are.
func (s *Service) AddVouch(vouchee string, p *proof.Presentation) (int, *Enrollment, error) {
	if !validPub(vouchee) {
		return 0, nil, ErrBadKey
	}
	if count(s.DB, `SELECT COUNT(*) FROM enrollments WHERE pubkey = ?`, vouchee) > 0 {
		return 0, nil, ErrKeyEnrolled
	}
	v, err := s.VerifyPresentation(p, VouchScope, "vouch:"+vouchee, []string{"human_doc"})
	if err != nil {
		return 0, nil, err
	}
	period := time.Now().UTC().Format("2006-01")
	if count(s.DB, `SELECT COUNT(*) FROM vouches WHERE voucher_nullifier=? AND period=?`, v.Nullifier, period) >= VouchesPerVoucher {
		return 0, nil, fmt.Errorf("you have already vouched %d times this month", VouchesPerVoucher)
	}
	if _, err := s.DB.Exec(`INSERT INTO vouches(id, vouchee_pubkey, voucher_nullifier, period, created_at) VALUES(?,?,?,?,?)`, newID("vch"), vouchee, v.Nullifier, period, Now()); err != nil {
		return 0, nil, errors.New("you already vouched for this person")
	}
	n := count(s.DB, `SELECT COUNT(*) FROM vouches WHERE vouchee_pubkey=?`, vouchee)
	if n >= VouchesRequired {
		e, err := s.insertEnrollment(enrollInput{pub: vouchee, attester: "att_vouch", assurance: "IAL1"})
		return n, e, err
	}
	return n, nil, nil
}

func (s *Service) VouchCount(vouchee string) int {
	return count(s.DB, `SELECT COUNT(*) FROM vouches WHERE vouchee_pubkey=?`, vouchee)
}

// ---------------------------------------------------------------- epochs & sets

type EpochInfo struct {
	ID       int64             `json:"epoch"`
	SealedAt string            `json:"sealedAt"`
	Holders  int               `json:"holders"`
	TxHash   string            `json:"txHash"`
	Sets     map[string]SetRef `json:"sets"`
}

type SetRef struct {
	Size int    `json:"size"`
	Root string `json:"root"`
}

// membersNow computes set membership from the live registry.
func (s *Service) membersNow(now time.Time) (map[string][]string, int, error) {
	rows, err := s.DB.Query(`SELECT pubkey, assurance, region, age13_at, age16_at, age18_at, age21_at FROM enrollments WHERE status='active' AND created_at <= ?`, ts(now))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	today := now.Format("2006-01-02")
	out := map[string][]string{}
	n := 0
	for rows.Next() {
		var pub, assurance, region string
		var a13, a16, a18, a21 *string
		if err := rows.Scan(&pub, &assurance, &region, &a13, &a16, &a18, &a21); err != nil {
			return nil, 0, err
		}
		n++
		out["human"] = append(out["human"], pub)
		if assurance == "IAL2" {
			out["human_doc"] = append(out["human_doc"], pub)
		}
		for _, x := range []struct {
			n string
			v *string
		}{{"age_over_13", a13}, {"age_over_16", a16}, {"age_over_18", a18}, {"age_over_21", a21}} {
			if x.v != nil && *x.v <= today {
				out[x.n] = append(out[x.n], pub)
			}
		}
		if region != "" {
			out["region_"+region] = append(out["region_"+region], pub)
		}
	}
	return out, n, rows.Err()
}

// SealEpoch snapshots every set, commits Merkle roots and anchors them on-chain.
func (s *Service) SealEpoch(now time.Time) (*EpochInfo, error) {
	s.sealMu.Lock()
	defer s.sealMu.Unlock()
	members, holders, err := s.membersNow(now)
	if err != nil {
		return nil, err
	}
	var last int64
	_ = s.DB.QueryRow(`SELECT COALESCE(MAX(id), 0) FROM epochs`).Scan(&last)
	epoch := last + 1
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	info := &EpochInfo{ID: epoch, SealedAt: ts(now), Holders: holders, Sets: map[string]SetRef{}}
	var roots []anchorRoot
	for _, def := range Sets {
		keys := members[def.Name]
		raw := make([][]byte, 0, len(keys))
		for _, k := range keys {
			b, _ := hex.DecodeString(k)
			raw = append(raw, b)
		}
		root := proof.BuildMerkle(raw).RootHex()
		info.Sets[def.Name] = SetRef{Size: len(keys), Root: root}
		roots = append(roots, anchorRoot{Set: def.Name, Root: root, Size: len(keys)})
		if _, err := tx.Exec(`INSERT INTO epoch_sets(epoch, set_name, size, root) VALUES(?,?,?,?)`, epoch, def.Name, len(keys), root); err != nil {
			return nil, err
		}
		stmt, err := tx.Prepare(`INSERT INTO epoch_members(epoch, set_name, pubkey) VALUES(?,?,?)`)
		if err != nil {
			return nil, err
		}
		for _, k := range keys {
			if _, err := stmt.Exec(epoch, def.Name, k); err != nil {
				stmt.Close()
				return nil, err
			}
		}
		stmt.Close()
	}
	if _, err := tx.Exec(`INSERT INTO epochs(id, sealed_at, holders) VALUES(?,?,?)`, epoch, ts(now), holders); err != nil {
		return nil, err
	}
	// Keep member snapshots for the epochs verifiers still accept; roots are kept forever.
	if _, err := tx.Exec(`DELETE FROM epoch_members WHERE epoch <= ?`, epoch-(s.Cfg.EpochGrace+2)); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	hash, err := s.anchorEpoch(epoch, roots, now)
	if err == nil {
		info.TxHash = hash
		_, _ = s.DB.Exec(`UPDATE epochs SET tx_hash=? WHERE id=?`, hash, epoch)
	}
	_ = s.DB.SetMeta("registry_dirty", "0")
	s.logf("sealed epoch %d: %d holders, anchored %s", epoch, holders, short(hash))
	return info, nil
}

func short(h string) string {
	if len(h) > 12 {
		return h[:10] + "…"
	}
	return h
}

// CurrentEpoch returns the latest sealed epoch (sealing the first one if needed).
func (s *Service) CurrentEpoch() (*EpochInfo, error) {
	var id int64
	err := s.DB.QueryRow(`SELECT COALESCE(MAX(id), 0) FROM epochs`).Scan(&id)
	if err != nil {
		return nil, err
	}
	if id == 0 {
		return s.SealEpoch(time.Now().UTC())
	}
	return s.Epoch(id)
}

func (s *Service) Epoch(id int64) (*EpochInfo, error) {
	info := &EpochInfo{ID: id, Sets: map[string]SetRef{}}
	var tx *string
	if err := s.DB.QueryRow(`SELECT sealed_at, holders, tx_hash FROM epochs WHERE id=?`, id).Scan(&info.SealedAt, &info.Holders, &tx); err != nil {
		return nil, err
	}
	if tx != nil {
		info.TxHash = *tx
	}
	rows, err := s.DB.Query(`SELECT set_name, size, root FROM epoch_sets WHERE epoch=?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		var r SetRef
		if err := rows.Scan(&name, &r.Size, &r.Root); err != nil {
			return nil, err
		}
		info.Sets[name] = r
	}
	return info, rows.Err()
}

func (s *Service) Epochs(limit int) ([]EpochInfo, error) {
	rows, err := s.DB.Query(`SELECT id FROM epochs ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		_ = rows.Scan(&id)
		ids = append(ids, id)
	}
	rows.Close()
	var out []EpochInfo
	for _, id := range ids {
		e, err := s.Epoch(id)
		if err != nil {
			return nil, err
		}
		out = append(out, *e)
	}
	return out, nil
}

// MaybeSeal seals a new epoch if the registry changed and the interval elapsed.
func (s *Service) MaybeSeal(now time.Time, force bool) {
	if !force && s.DB.Meta("registry_dirty") != "1" {
		return
	}
	cur, err := s.CurrentEpoch()
	if err == nil && !force && now.Sub(parseTS(cur.SealedAt)) < s.Cfg.EpochInterval {
		return
	}
	if _, err := s.SealEpoch(now); err != nil {
		s.logf("seal failed: %v", err)
	}
}

// ---------------------------------------------------------------- set access

type cachedSet struct {
	raw        [][]byte
	bucketKeys map[string][]byte
}

// Members implements proof.SetSource over epoch snapshots.
func (s *Service) Members(epoch int64, set string) ([][]byte, error) {
	cs, err := s.loadSet(epoch, set)
	if err != nil {
		return nil, err
	}
	return cs.raw, nil
}

func (s *Service) loadSet(epoch int64, set string) (*cachedSet, error) {
	key := fmt.Sprintf("%d/%s", epoch, set)
	if v, ok := s.setCache.Load(key); ok {
		return v.(*cachedSet), nil
	}
	if _, ok := SetByName(set); !ok {
		return nil, fmt.Errorf("unknown set %q", set)
	}
	if count(s.DB, `SELECT COUNT(*) FROM epochs WHERE id=?`, epoch) == 0 {
		return nil, fmt.Errorf("epoch %d does not exist", epoch)
	}
	rows, err := s.DB.Query(`SELECT pubkey FROM epoch_members WHERE epoch=? AND set_name=?`, epoch, set)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cs := &cachedSet{bucketKeys: map[string][]byte{}}
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err != nil {
			return nil, err
		}
		b, _ := hex.DecodeString(h)
		cs.raw = append(cs.raw, b)
	}
	if len(cs.raw) == 0 {
		var size int
		_ = s.DB.QueryRow(`SELECT size FROM epoch_sets WHERE epoch=? AND set_name=?`, epoch, set).Scan(&size)
		if size > 0 {
			return nil, fmt.Errorf("epoch %d has expired; refresh and prove against the current epoch", epoch)
		}
	}
	sort.Slice(cs.raw, func(i, j int) bool { return hex.EncodeToString(cs.raw[i]) < hex.EncodeToString(cs.raw[j]) })
	s.setCache.Store(key, cs)
	return cs, nil
}

type BucketResponse struct {
	Epoch   int64    `json:"epoch"`
	Set     string   `json:"set"`
	Bits    int      `json:"bits"`
	Prefix  string   `json:"prefix"`
	Members []string `json:"members"`
	Root    string   `json:"root"`
	SetSize int      `json:"setSize"`
}

// BucketFor serves the canonical anonymity bucket for a holder, identified only by the
// holder's own public key (which the holder will reveal as a ring member anyway).
func (s *Service) BucketFor(epoch int64, set, pub string) (*BucketResponse, error) {
	cs, err := s.loadSet(epoch, set)
	if err != nil {
		return nil, err
	}
	me, err := hex.DecodeString(pub)
	if err != nil || len(me) != 32 {
		return nil, ErrBadKey
	}
	bits, ring := proof.Bucket(cs.raw, me, proof.BucketTarget)
	found := false
	members := make([]string, len(ring))
	for i, r := range ring {
		members[i] = hex.EncodeToString(r)
		if members[i] == pub {
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("your key is not in the %q set for epoch %d", set, epoch)
	}
	var root string
	_ = s.DB.QueryRow(`SELECT root FROM epoch_sets WHERE epoch=? AND set_name=?`, epoch, set).Scan(&root)
	return &BucketResponse{Epoch: epoch, Set: set, Bits: bits, Prefix: proof.PrefixFor(me, bits), Members: members, Root: root, SetSize: len(cs.raw)}, nil
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func proofFingerprint(pub string) string {
	el, err := proof.DecodeElementHex(pub)
	if err != nil {
		return pub[:8]
	}
	return proof.Fingerprint(el)
}
