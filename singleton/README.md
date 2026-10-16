# Singleton — prove you're human, reveal nothing else

> Privacy-preserving **proof-of-human and age assurance** for the AI internet.
> People verify once with an ID they already carry. Apps get a yes — and a pseudonym that
> is stable inside the app and unlinkable everywhere else. No orb, no ID uploads, no tracking.

**Category:** web3 identity infrastructure (non-financial) · **Stack:** Go 1.24, SQLite, Go→WebAssembly wallet,
server-rendered `html/template`, Solidity contract · **Status:** working MVP with a sandbox network.

| | |
|---|---|
| Why now | Bots passed humans on the web (57.5% of traffic, Cloudflare 2026); CAPTCHAs are solved by AI; Australia, the UK and US states now *require* age assurance; ID-upload vendors are getting breached. |
| Wedge | One API answering *human? unique here? old enough?* — priced at $0.05 instead of $1.35–$2.30 per KYC check, because the hard verification happens once and is reused. |
| Moat | Two-sided network (attesters ↔ holders ↔ platforms), anchored public sets, and the only option that is both AI-proof and private. |
| Validation | World (18M+ verified, $525M raised) and Humanity Protocol ($1.1B valuation) prove demand; both require body scans. YC's Fall 2026 RFS asks for exactly this ("Proving You're Human"). |

Read the research and the business case:

1. [`docs/01-threat-model-research.md`](docs/01-threat-model-research.md) — threat-model-first research (adversaries, attack economics, failing defenses, design goals)
2. [`docs/02-protocol.md`](docs/02-protocol.md) — the cryptographic protocol and its security argument
3. [`docs/03-investment-case.md`](docs/03-investment-case.md) — market, why now, competition, moat, path to $100M ARR, risks
4. [`docs/04-monetization-and-growth.md`](docs/04-monetization-and-growth.md) — pricing, unit economics, ads & acquisition plan, GTM, KPIs

---

## Run it

```bash
cd singleton
go run ./cmd/singleton            # http://localhost:4201  (first run seeds a sandbox network, ~2s)
go test ./...                     # crypto core + end-to-end service tests
make wasm                         # optional: prebuild the wallet core (otherwise built on first request)
```

Requires Go 1.24+. No other services. Data lives in `data/singleton.db` (delete it to reset).

| URL | What it is |
|---|---|
| `/` | Landing page with a live, no-CAPTCHA proof-of-work waitlist (people + platform tracks) |
| `/pass` | **Singleton Pass** — the holder wallet (key generated in-browser via WebAssembly) |
| `/demo/driftwood` | Demo social app: one account per human, 16+ |
| `/demo/frontrow` | Demo ticket drop: one ticket per human per drop |
| `/demo/commons` | Demo poll: one anonymous vote per human |
| `/console` | Relying-party console (demo: `demo@singleton.dev` / `singleton-demo`) |
| `/docs` | Developer docs, API reference, protocol summary |
| `/transparency` | Epochs, Merkle roots, anchor transactions, attesters, revocations |
| `/attester` | In-person partner console (demo agent button) |
| `/admin?token=dev-admin` | Waitlist admin + CSV export |

### Try the core loop in 60 seconds
1. Open `/pass` → **Create my Pass** → **Mobile driver's licence** → **Present credential** (sandbox issuer).
2. Open `/demo/driftwood`, pick a handle, **Sign up**. A consent popup builds a ring-signature proof on your device.
3. Sign up again with a different handle → *"This person already has a Driftwood account."*
4. Open `/demo/frontrow` → your pseudonym there is different (unlinkable), and you can claim exactly one ticket.
5. Open `/console` → see the verification, the duplicate, the signed token claims and billing.

---

## What's in the box

**Protocol & crypto (`internal/proof`)**
- ristretto255 holder keys; **scoped linkable ring signatures** (bLSAG variant) where the key image is `x·H(scope)` → one pseudonym per app, unlinkable across apps
- Merkle commitments for every predicate set + inclusion proofs; canonical **anonymity buckets** (prefix tree, stable across epochs)
- Presentations with multiple predicates bound to one holder; replay/scope/splice protection
- Schnorr signatures for self-revocation; **Shamir 3-of-5 social recovery** over GF(2⁸) that restores the same pseudonyms
- Agent **delegation statements** (Know-Your-Agent) signed inside the app's scope

**Registry & verifier (`internal/app`)**
- Enrollment paths: mDL, ePassport chip, EU wallet (sandbox issuers with real signature checks), in-person attestation, vouching (3 IAL2 vouchers, rate-limited by pseudonym)
- De-duplication by HMAC-peppered document and person hashes; adults' birthdays never retained (only threshold dates)
- Epoch sealing, set snapshots, revocation, **on-chain anchoring** with real ABI calldata (keccak selectors, EIP-55 addresses)
- Verification sessions, one-human-one-account bindings, duplicate detection with the existing account reference, action scopes
- EdDSA result tokens + JWKS, stateless verification, agent request verification, signed/retried webhooks
- Console: orgs, users (bcrypt), apps, publishable/secret keys (hashed), analytics, billing plans & invoices, compliance evidence (AU/UK/US)
- Waitlist with HMAC-signed hashcash challenges, people/platform tracks, referral queue and design-partner prioritisation

**Front end (`web/`)**
- Landing page, docs, transparency explorer, console, wallet, consent, attester and three demo relying parties
- `sdk/singleton.js` — popup flow, `Singleton.verify()` and `Singleton.button()`
- Accessible, dark visual system; SVG charts rendered server-side (≤24px columns, 2px gaps, hairline grid, hover tooltips)

**Contract (`contracts/HumanRegistry.sol`)** — epoch roots and revocation commitments.

## Layout

```
cmd/singleton/        server + `verify-root` audit command
cmd/wallet-wasm/      wallet core exported to JavaScript (GOOS=js GOARCH=wasm)
internal/proof/       cryptography (tested, benchmarked)
internal/app/         registry, verifier, console, chain, webhooks, waitlist, seed
web/                  HTTP handlers, templates, static assets, embedded at build
contracts/            HumanRegistry.sol
docs/                 research, protocol, investment case, monetization
```

## Configuration

| Env / flag | Default | |
|---|---|---|
| `-addr` / `ADDR` | `:4201` | listen address |
| `-db` / `DB_PATH` | `data/singleton.db` | SQLite file |
| `-base-url` / `BASE_URL` | `http://localhost:4201` | public URL used in tokens & links |
| `ADMIN_TOKEN` | `dev-admin` | waitlist admin |
| `ATTESTER_TOKEN` | `dev-attester` | in-person partner console |
| `-epoch` | `90s` | minimum time between sealed epochs |
| `-no-seed` | false | skip the sandbox network & demo workspace |
| `RATE_LIMIT=off` | | disable rate limits (tests) |

## Honest limitations of this MVP
- Credential issuers are **sandbox** issuers (Ed25519). Production swaps in ISO 18013-5 mdoc COSE verification (AAMVA DTS), ICAO 9303 passive authentication and SD-JWT VC — the enrollment logic after the signature check is unchanged.
- Rings are anonymity **buckets** (≥32 members) rather than the full set; v2 replaces them with a zk-SNARK membership proof.
- Anchoring is recorded in a local block log with real calldata; broadcasting to Base needs a funded signer.
- The wallet is a web app; production ships native iOS/Android wallets with keys in the secure enclave.
