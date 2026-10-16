# Harbormaster — product requirements (v1)

Status: **shipped in this repository** unless marked *(next)*. Each requirement lists the
job it serves (see `01-customer-discovery.md`) and how it is verified.

## 1. Principles
1. **Numbers come from the engine, words from the model.** Duty, deadlines and refund
   routes are deterministic and testable; Claude is used for language (classification,
   Q&A) and always through the engine's tools.
2. **Every number explains itself.** Rate, base, Chapter 99 heading, authority, status —
   and every layer considered and not applied, with the reason.
3. **We draft, licensed professionals file.** No feature submits to CBP on its own.
4. **Works without AI.** `HARBORMASTER_AI=off` keeps every page functional.

## 2. Tariff engine (`harbormaster/tariff/`)

| ID | Requirement | Job | Verified by |
|---|---|---|---|
| E1 | Compute MFN, FTA/USMCA preference, IEEPA (border, reciprocal, Brazil, India), s122, s301 (China lists + 2024 increases, forced-labor, Brazil), s232 (metals primary/derivative, autos/parts, trucks, wood, semis, pharma) for any line and date | J3, J4 | `tests/test_engine.py` (21 cases) |
| E2 | Non-stacking: 232 goods excluded from reciprocal/s122; CA/MX border duty not on 232 goods; derivative metal content splits the base before Apr 6 2026 | J4 | `test_232_derivative_on_metal_content_then_full_value` |
| E3 | In-transit relief (Apr 5 / Aug 7 2025 reciprocal; Jul 24 2026 forced-labor) | J4 | `test_in_transit_relief_for_country_rates`, `test_forced_labor_in_transit_exemption` |
| E4 | Exemptions: Annex II, Apr 11 electronics (retroactive), Nov 2025 agriculture | J4 | `test_electronics_exemption…` |
| E5 | Measure status by date (in force / invalidated / contested) | J1 | `test_status_by_date` |
| E6 | MPF min/max and HMF | J4 | `test_fees_capped` |
| E7 | Landed cost per unit incl. freight and fees | J3 | calculator + API tests |
| E8 | Rules loaded from a versioned rules service with CSMS diffing *(next)* | J7 | — |

## 3. Recovery (`tariff/recovery.py`, `services/audit.py`, `services/cases.py`)

| ID | Requirement | Verified by |
|---|---|---|
| R1 | Per entry: PSC close (min(entry+300, liq−15)), projected liquidation (summary+314), protest close (liq+180), CIT limit (+2 yrs), drawback (+5 yrs) | `test_deadlines_follow_the_statute` |
| R2 | IEEPA routing: unliquidated/≤80 days → CAPE; ≤180 days → protest; final → CAPE Phase 3 if plaintiff else class/§1581(i); closed after CIT limit | `test_unliquidated…`, `test_liquidated…`, `test_final_entries…` |
| R3 | s122: statutory amount = contested (preserve/protest/CIT); excess = filing error | `test_s122_split…` |
| R4 | Filing errors: overpayment → PSC or protest; underpayment → PSC or prior disclosure | `test_underpayment…` |
| R5 | Interest estimate on refunds | route tests |
| R6 | Bundle findings into filings; CAPE entry-list CSV; printable packet with statement of reasons; status workflow; success-fee ledger | `test_create_case…` |
| R7 | Entry import from ES-003/7501 CSV (per-measure columns or Chapter 99 pairs), re-audit on import | `test_upload_roundtrip` |
| R8 | ACE API connector & broker SFTP pickup *(next)* | — |

## 4. Margin & sourcing

| ID | Requirement | Verified by |
|---|---|---|
| M1 | Per-SKU landed cost, duty/unit, margin vs. target, price to hold target, one-click reprice | page render + `sku_economics` |
| M2 | Origin comparison with labour index, lane freight, duty, fees and scenario overlays (§301 excess capacity, FL raise, new global surcharge) | `test_sourcing_ranks…` |
| M3 | Tariff timeline with per-SKU annual impact computed by the engine | `test_timeline_impact_matches_engine` |
| M4 | Drawback: substitution matching, highest-duty designation, 5-year window, eligible measures only | `test_drawback…` |
| M5 | Broker scorecard: error rate, over/under by measure | page render |

## 5. AI (`harbormaster/ai/`)

| ID | Requirement | Verified by |
|---|---|---|
| A1 | Classification: Claude ranks candidates from the reference catalog via structured outputs (enum of codes), GRI reasoning, missing facts; engine attaches duty per candidate; human approval writes the SKU | `test_classifier_uses_claude…` |
| A2 | Refusal / API failure → deterministic catalog matcher | `test_classifier_falls_back_on_refusal` |
| A3 | Copilot: tool loop over 8 strict tools; every figure from tool results; history persisted | `test_copilot_tool_loop` |
| A4 | Server-side refusal fallbacks opted in (`fallbacks="default"`), typed error handling | `ai/llm.py` |

Model: `claude-opus-5-5` (override with `HARBORMASTER_MODEL`). Effort `high` for
classification, `medium` for the copilot.

## 6. Platform

| ID | Requirement |
|---|---|
| P1 | Session auth (scrypt), CSRF on every form and htmx request, strict CSP (no inline script), rate limits on public and AI endpoints |
| P2 | API keys (hashed) for `/api/v1/stack` and `/api/v1/findings`; OpenAPI at `/api/docs` |
| P3 | Waitlist: estimator results saved with the signup, onboarding priority by refund size × urgency, referral credit, admin export |
| P4 | Light/dark themes; mobile layout without horizontal scroll; charts with tooltips, legends and table views |
| P5 | SQLite for single-tenant/dev, any SQLAlchemy URL in production (`HARBORMASTER_DB`) |
| P6 | SSO/SAML, SOC 2 controls, per-workspace encryption keys *(next)* |

## 7. Non-goals (v1)
Filing directly to ACE as a broker; AD/CVD rate computation (flagged, not computed);
non-US customs regimes.
