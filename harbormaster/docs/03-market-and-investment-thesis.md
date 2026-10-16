# Harbormaster — market & investment thesis

**One line:** the system of record for what imported goods really cost — starting with
the largest refund event in US trade history.

## 1. Why this is a venture-scale company

1. **Tariffs became a top-three cost line overnight.** US customs duties were ~$195B in
   FY2025 ([USAFacts](https://usafacts.org/answers/how-much-revenue-does-the-federal-government-collect-from-tariffs/country/united-states/)),
   several times the historical run-rate, and the regime has changed a dozen times since.
   Effective rates of 20–45% (our engine, China/SE Asia baskets, Oct 2026) turn duty from
   a rounding error into the margin.
2. **A forced, dated, nine-figure-per-customer-cohort buying event.** ~$166B of IEEPA duty
   is refundable; CAPE Phase 3 and the Oct 8 2026 class certification leave billions in
   finally liquidated entries that need active work, plus the contested s122 surcharge
   ([Phillips Lytle](https://phillipslytle.com/cbp-opens-phase-3-of-cape-for-ieepa-refunds/),
   [Clark Hill](https://www.clarkhill.com/news-events/news/section-122-tariffs-ruling/)).
   Deadlines create urgency no enterprise sales cycle can match.
3. **Permanent complexity after the refund.** Forced-labor §301 has no expiry, the §301
   excess-capacity investigation is open, 232 keeps expanding (pharma, semis, wood). Every
   change re-prices every SKU. That is a recurring job.
4. **No system of record exists for SMB and mid-market importers.** Brokers file, ERPs
   hold POs, nobody holds "duty by SKU by date with the law attached."

## 2. Market size — bottom-up

**Customers.** The Census Bureau's *Profile of U.S. Importing and Exporting Companies*
(latest released May 2026) identifies hundreds of thousands of importing firms, the
overwhelming majority small and medium-sized
([Census](https://www.census.gov/foreign-trade/Press-Release/edb/edbrel2024.pdf)). We size
on the segment that pays meaningful duty and lacks a trade department:

| Segment | Importers (est.) | SaaS ACV | Recovery fees (2026–28, per customer) | SaaS TAM |
|---|---|---|---|---|
| $1–5M imports | 60,000 | $4.8k (Starter, *next*) | $40–150k | $290M |
| $5–50M imports | 25,000 | $10.8k–$30k | $150k–1.5M | $500M |
| $50–500M imports | 4,000 | $60–150k | $1.5–15M | $400M |
| Broker partners (white-label) | 2,000 brokers | $25–100k | rev share | $100M |
| **Total** | | | | **≈ $1.3B SaaS** + API |

**Recovery pool we can address (2026–28):** finally liquidated IEEPA entries and
complex cases outside Phases 1–2 (~$36B of the $166B on one estimate), contested s122,
filing errors (1–3% of non-IEEPA duty ≈ $1–3B/yr at FY2025 duty levels) and drawback
(industry estimates of $2.4–15B/yr unclaimed —
[Purolator](https://www.purolatorinternational.com/u-s-businesses-fail-to-claim-more-than-2-4-billion-annually-in-drawback-reimbursements/)).
Capturing 1% of that at a ~11% blended fee is **$40–60M of fees** — enough to fund the
land-and-expand motion without dilutive capital.

**API.** Landed cost at checkout/PO time for ERPs, marketplaces and 3PLs; priced per call.
The engine is the product.

## 3. Why we win

| Competitor type | Their model | Our edge |
|---|---|---|
| Refund-only startups (Tarifflo, Trava, NEU forge) | Contingency fee, one-time | Same wedge, but the audit populates a permanent SKU-level ledger; customers stay for margin and alerts |
| Classification tools (Gaia, GingerControl) | Per-classification | Classification is one feature; we own the entry ledger it feeds |
| Digital brokers (Flexport, KlearNow) | Brokerage revenue | We're broker-neutral and sell *through* brokers |
| Big-4 / law firms | Hours | 100× cheaper for the 99% of importers below $100M |

**Moats that compound:** (1) the reconciled entry ledger (switching cost — it's the audit
trail), (2) the rules engine's regression suite (every measure × date × origin pinned by
tests — hard to copy correctly), (3) broker distribution, (4) proprietary data: actual
landed costs and error rates by lane, broker and HTS, which power benchmarks no one else has.

## 4. Customer-success likelihood

- **Time to value in days:** upload ES-003 → findings with deadlines the same day.
- **Value is cash, with a receipt:** CBP refund wires, not "insights."
- **Low risk to try:** free audit; recovery fee only on money received.
- **Daily habit after the refund:** landed cost and price-to-hold on every PO, alerts on
  every proclamation.

## 5. Financial model (base case)

| | 2027 | 2028 | 2029 | 2030 |
|---|---|---|---|---|
| Paying importers | 450 | 1,800 | 4,200 | 7,500 |
| Broker partners | 25 | 120 | 300 | 550 |
| SaaS ARR | $5M | $24M | $62M | $118M |
| Recovery fees (recognised) | $14M | $22M | $9M | $6M |
| API revenue | $0.5M | $3M | $9M | $18M |
| **Total revenue** | **$19.5M** | **$49M** | **$80M** | **$142M** |
| Gross margin | 82% | 84% | 86% | 87% |

Recovery fees front-load cash in 2027–28 (the refund wave) and fund growth; SaaS +
API carry the company to $100M+ ARR by 2030. At 12–15× forward ARR for vertical
AI/fintech infrastructure, the 2029 business clears a unicorn valuation.

## 6. Risks

| Risk | Mitigation |
|---|---|
| Tariffs get rolled back | Every rollback creates refunds and re-pricing; volatility is the product. Even MFN-only regimes need landed cost, drawback and classification |
| CBP automates refunds fully | Phase 1–2 already partly automatic; our value shifts to edge cases, s122, errors, drawback, and the ledger |
| Rules accuracy | Engine is test-pinned; rules carry status and authority; anything uncertain is flagged for review, never guessed |
| Liability | We draft, licensed brokers/attorneys file; E&O insurance; explicit reference-data disclaimers |
| Competition from refund startups | They'll fight over the wave; we convert recoveries into SaaS seats (Growth plan cuts the fee to 10%) |

## 7. The ask (illustrative)
$8M seed: 6 engineers on the rules engine and ACE/ERP connectors, 3 trade-compliance
specialists (ex-CBP / licensed brokers), broker-partner GTM. Milestones: 300 paying
importers, $250M of recoveries under management, $6M ARR in 15 months.
