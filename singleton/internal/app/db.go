// Package app holds Singleton's services: the human registry (enrollment, epochs, sets,
// revocation), the verifier (sessions, presentations, bindings, tokens), the relying-party
// console, on-chain anchoring, webhooks and the waitlist. All state lives in SQLite.
package app

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const schema = `
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;

CREATE TABLE IF NOT EXISTS meta (k TEXT PRIMARY KEY, v TEXT NOT NULL);

CREATE TABLE IF NOT EXISTS attesters (
  id TEXT PRIMARY KEY, name TEXT NOT NULL, kind TEXT NOT NULL, assurance TEXT NOT NULL,
  region TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'active', description TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS enrollments (
  id TEXT PRIMARY KEY, pubkey TEXT NOT NULL UNIQUE, attester_id TEXT NOT NULL REFERENCES attesters(id),
  assurance TEXT NOT NULL, doc_hash TEXT, person_hash TEXT, region TEXT NOT NULL DEFAULT '',
  age13_at TEXT, age16_at TEXT, age18_at TEXT, age21_at TEXT,
  status TEXT NOT NULL DEFAULT 'active', revoked_reason TEXT, revoked_at TEXT,
  synthetic INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS enrollments_doc ON enrollments(doc_hash) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS enrollments_person ON enrollments(person_hash) WHERE status = 'active';

CREATE TABLE IF NOT EXISTS enrollment_attempts (
  id TEXT PRIMARY KEY, attester_id TEXT, outcome TEXT NOT NULL, reason TEXT, created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS in_person_codes (
  code TEXT PRIMARY KEY, pubkey TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'pending',
  attester_id TEXT, created_at TEXT NOT NULL, decided_at TEXT, enrollment_id TEXT
);

CREATE TABLE IF NOT EXISTS vouches (
  id TEXT PRIMARY KEY, vouchee_pubkey TEXT NOT NULL, voucher_nullifier TEXT NOT NULL, period TEXT NOT NULL,
  created_at TEXT NOT NULL, UNIQUE(vouchee_pubkey, voucher_nullifier)
);

CREATE TABLE IF NOT EXISTS epochs (
  id INTEGER PRIMARY KEY, sealed_at TEXT NOT NULL, holders INTEGER NOT NULL, tx_hash TEXT
);
CREATE TABLE IF NOT EXISTS epoch_sets (
  epoch INTEGER NOT NULL, set_name TEXT NOT NULL, size INTEGER NOT NULL, root TEXT NOT NULL,
  PRIMARY KEY (epoch, set_name)
);
CREATE TABLE IF NOT EXISTS epoch_members (
  epoch INTEGER NOT NULL, set_name TEXT NOT NULL, pubkey TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS epoch_members_idx ON epoch_members(epoch, set_name);

CREATE TABLE IF NOT EXISTS chain_txs (
  hash TEXT PRIMARY KEY, block INTEGER NOT NULL, from_addr TEXT NOT NULL, to_addr TEXT NOT NULL,
  method TEXT NOT NULL, calldata TEXT NOT NULL, summary TEXT NOT NULL, created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS orgs (id TEXT PRIMARY KEY, name TEXT NOT NULL, plan TEXT NOT NULL DEFAULT 'developer', demo INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY, org_id TEXT NOT NULL REFERENCES orgs(id), email TEXT NOT NULL UNIQUE, name TEXT NOT NULL,
  password_hash TEXT NOT NULL, role TEXT NOT NULL DEFAULT 'owner', created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS console_sessions (token_hash TEXT PRIMARY KEY, user_id TEXT NOT NULL, csrf TEXT NOT NULL, expires_at TEXT NOT NULL);

CREATE TABLE IF NOT EXISTS apps (
  id TEXT PRIMARY KEY, org_id TEXT NOT NULL REFERENCES orgs(id), name TEXT NOT NULL, scope TEXT NOT NULL UNIQUE,
  category TEXT NOT NULL DEFAULT 'social', mode TEXT NOT NULL DEFAULT 'test', origins TEXT NOT NULL DEFAULT '',
  required TEXT NOT NULL DEFAULT 'human', duplicate_policy TEXT NOT NULL DEFAULT 'flag',
  webhook_url TEXT NOT NULL DEFAULT '', webhook_secret TEXT NOT NULL, color TEXT NOT NULL DEFAULT '#d4ff4f',
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS api_keys (
  id TEXT PRIMARY KEY, app_id TEXT NOT NULL REFERENCES apps(id), kind TEXT NOT NULL, prefix TEXT NOT NULL,
  key_hash TEXT NOT NULL UNIQUE, created_at TEXT NOT NULL, last_used_at TEXT, revoked_at TEXT
);

CREATE TABLE IF NOT EXISTS verify_sessions (
  id TEXT PRIMARY KEY, app_id TEXT NOT NULL REFERENCES apps(id), challenge TEXT NOT NULL, scope TEXT NOT NULL,
  required TEXT NOT NULL, action TEXT NOT NULL DEFAULT '', reference TEXT NOT NULL DEFAULT '', origin TEXT NOT NULL DEFAULT '',
  return_url TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'pending', nullifier TEXT, predicates TEXT,
  ring_sizes TEXT, error TEXT, duplicate_of TEXT, is_returning INTEGER NOT NULL DEFAULT 0, duration_ms INTEGER, synthetic INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL, completed_at TEXT, expires_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS verify_sessions_app ON verify_sessions(app_id, created_at);

CREATE TABLE IF NOT EXISTS bindings (
  scope TEXT NOT NULL, nullifier TEXT NOT NULL, app_id TEXT NOT NULL, reference TEXT NOT NULL DEFAULT '',
  first_seen TEXT NOT NULL, last_seen TEXT NOT NULL, uses INTEGER NOT NULL DEFAULT 1,
  PRIMARY KEY (scope, nullifier)
);

CREATE TABLE IF NOT EXISTS delegations (
  id TEXT PRIMARY KEY, app_id TEXT NOT NULL, nullifier TEXT NOT NULL, agent TEXT NOT NULL, label TEXT NOT NULL,
  permissions TEXT NOT NULL, not_after TEXT NOT NULL, created_at TEXT NOT NULL, revoked_at TEXT
);

CREATE TABLE IF NOT EXISTS webhook_deliveries (
  id TEXT PRIMARY KEY, app_id TEXT NOT NULL, event TEXT NOT NULL, payload TEXT NOT NULL, status TEXT NOT NULL,
  response_code INTEGER, attempts INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS waitlist (
  id TEXT PRIMARY KEY, email TEXT NOT NULL UNIQUE, track TEXT NOT NULL, company TEXT NOT NULL DEFAULT '',
  role TEXT NOT NULL DEFAULT '', use_case TEXT NOT NULL DEFAULT '', volume TEXT NOT NULL DEFAULT '', region TEXT NOT NULL DEFAULT '',
  code TEXT NOT NULL UNIQUE, referred_by TEXT, referrals INTEGER NOT NULL DEFAULT 0, priority INTEGER NOT NULL DEFAULT 0,
  pow_bits INTEGER NOT NULL DEFAULT 0, ip_hash TEXT NOT NULL DEFAULT '', invited_at TEXT, created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS demo_records (
  app TEXT NOT NULL, rkey TEXT NOT NULL, nullifier TEXT NOT NULL, value TEXT NOT NULL, created_at TEXT NOT NULL,
  PRIMARY KEY (app, rkey, nullifier)
);

CREATE TABLE IF NOT EXISTS audit (
  id INTEGER PRIMARY KEY AUTOINCREMENT, actor TEXT NOT NULL, action TEXT NOT NULL, detail TEXT NOT NULL, created_at TEXT NOT NULL
);
`

// DB wraps *sql.DB with a few helpers.
type DB struct{ *sql.DB }

func OpenDB(path string) (*DB, error) {
	dsn := path
	if path == ":memory:" {
		dsn = "file::memory:?cache=shared"
	}
	sqldb, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if path == ":memory:" {
		sqldb.SetMaxOpenConns(1)
	} else {
		sqldb.SetMaxOpenConns(4)
	}
	for _, stmt := range strings.Split(schema, ";\n") {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if _, err := sqldb.Exec(stmt); err != nil {
			return nil, fmt.Errorf("schema: %w (%s)", err, firstLine(stmt))
		}
	}
	return &DB{sqldb}, nil
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i > 0 {
		return s[:i]
	}
	return s
}

func (db *DB) Meta(k string) string {
	var v string
	_ = db.QueryRow(`SELECT v FROM meta WHERE k = ?`, k).Scan(&v)
	return v
}

func (db *DB) SetMeta(k, v string) error {
	_, err := db.Exec(`INSERT INTO meta(k, v) VALUES(?, ?) ON CONFLICT(k) DO UPDATE SET v = excluded.v`, k, v)
	return err
}

func (db *DB) Audit(actor, action, detail string) {
	_, _ = db.Exec(`INSERT INTO audit(actor, action, detail, created_at) VALUES (?, ?, ?, ?)`, actor, action, detail, Now())
}

// Now is RFC 3339 UTC; overridable in tests and seeding.
var Now = func() string { return time.Now().UTC().Format(time.RFC3339) }

func ts(t time.Time) string { return t.UTC().Format(time.RFC3339) }

func parseTS(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

func count(db *DB, q string, args ...any) int {
	var n int
	_ = db.QueryRow(q, args...).Scan(&n)
	return n
}
