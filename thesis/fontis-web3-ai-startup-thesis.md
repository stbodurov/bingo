# Startup Thesis: Fontis

Oct 6, 2026 · picked for: lots of paying users, proven monetization, high odds of success, genuinely web3 + AI. It supersedes Clearlien (too niche) and Lading (thin margins).

## The pick

**Fontis lets anyone prove their income, work hours, balances or employment straight from the website that holds them** (a payroll portal, gig-app dashboard, bank, marketplace or tax account), instead of uploading documents.

- **Cryptography (zkTLS):** proves the numbers came from the real site and weren't edited, and lets the applicant share only the facts needed ("income ≥ 3× rent"), not the whole statement.
- **AI:** reads any portal in any format, works out income and hours under the verifier's rules, and walks the applicant through it.
- **Who pays:** verifiers (property managers, lenders, health plans, later employers) pay per verified proof. Applicants keep a free, reusable proof wallet.

> **Stop reading documents. Get the proof.**

**The wedge:** the roughly 1 in 3 applicants that payroll databases can't verify, mostly gig workers, freelancers, creators and the self-employed (74.9M Americans do independent work). Fontis plugs into existing verification flows as the step that works when the payroll lookup returns nothing, so verifiers don't have to replace anything.

---

## Does it meet the brief?

| Your criterion | Evidence | Verdict |
|---|---|---|
| **Monetizes with a lot of people** | Equifax's verification business (The Work Number) made **$2.18B in 2025** within a $2.58B segment at **44% operating margin**, charging per check across 188M active worker records. Plaid hit **$546M ARR** and is profitable. Snappt (rental document fraud) has an estimated **$50–100M** revenue. US tenant screening is a **$2.3–3.5B** market | ✅ Every renter, borrower and benefits applicant is a paid check |
| **High success likelihood** | The budget already exists and is growing. Incumbents are hated on price (The Work Number costs **$55–70 per check**, **+272% since 2012**, and faces an antitrust class action). The wedge segment is one incumbents structurally can't serve. No money transmission, no balance sheet, capital-light. Insert as a fallback, so there is no rip-and-replace sale | ✅ Higher than a consumer fintech; see odds below |
| **Genuinely web3 + AI** | zkTLS and verifiable credentials are crypto-native cryptography. AI is what makes "any website" possible without building integrations | ✅ Both are load-bearing; see "Why web3, why AI" |
| **a16z / Sequoia worthy** | a16z crypto invests in zk and data provenance (Opacity, through a16z's crypto accelerator). Verification and identity have produced large outcomes (Plaid, Persona, Equifax Workforce Solutions). AI fraud is a board-level topic | ✅ Category-sized outcome if the coverage moat holds |

**How it was chosen (final round):**

| Finalist | Many paying people | Success odds for this founder | Why it lost |
|---|---|---|---|
| **Fontis (proof from source)** | High: per-check revenue across tens of millions of applications | **Moderate–high** | **Pick** |
| AI + stablecoin money apps (Félix, KAST, Kontigo) | Very high | Low for this founder | Needs community and corridor ties; licensing-heavy; ruled out by you |
| Trading front-ends (Axiom, Phantom) | Very high | Low | Revenue swings with crypto prices; hyper-competitive |
| AI accountant for stablecoin businesses | Medium | Medium | Price collapse (Pilot's AI accountant is $99 a month; subledgers start at $29) |
| Consumer crypto tax | Medium | Medium | Weak monetization (CoinTracker: ~$23M revenue on $102M raised) |
| Clearlien | Low (few hundred buyers) | Medium | Too niche |

---

## The non-consensus insight

The industry is answering AI-forged documents with AI that detects forged documents. That is an arms race the detectors lose, because generation gets cheaper faster than detection gets better. A pay stub can be generated in seconds, and editable bank-statement templates sell for under $10.

The fix is to **stop accepting documents**. Get the data from the source, with a cryptographic proof that the source said it. Then use AI for a different job: reading every possible source and applying the rules.

Crypto's most useful invention for mainstream finance may not be a token. It is **zkTLS: a way to prove what a website showed you**, without that website's cooperation.

---

## Problem and customer

**The problem:**

- **Documents no longer count as evidence.** 1 in 16 documents in Inscribe's network is flagged as fraudulent, AI-generated document fraud rose about 5× in the second half of 2025, and bank statements are the top concern for 85.6% of fraud leaders (Inscribe, 2026).
- **Renters:** 12% of rental applications contain fraud, 93% of multifamily operators saw application fraud in 2025, and 55% of landlords have caught or suspected that the person who moved in wasn't the applicant (Checkr/Snappt, 2026).
- **Payroll databases miss the fastest-growing workforce.** The Work Number and payroll APIs instantly handle 60–70% of cases. Gig workers, the self-employed, small-employer W-2s and recent hires fall outside. Ocrolus calls gig income "the hardest verification problem in consumer lending." Argyle reports about 55% conversion on payroll connections.
- **Verification is expensive and slow.** The Work Number costs $55–70 per check, and a two-borrower mortgage can reach $280. Manual document review takes days and is now unreliable.

**Customers, in order:**

1. **Property managers and leasing teams**, starting with Sun Belt multifamily, where fraud is highest. They need a decision the same day and buy per-application tools today.
2. **Consumer and auto lenders, BNPL providers and fintechs** with large self-employed and gig applicant pools.
3. **Medicaid managed-care plans.** From Dec 31, 2026, about 20M expansion adults must show 80 hours of work or $580 of monthly income. Plans lose members, and revenue, when gig workers can't prove their hours, and state wage records don't capture gig hours.
4. **Later: employers and background-check firms** proving employment history and identity. Gartner projects 1 in 4 candidate profiles could be fake by 2028; 41% of organizations say they've unknowingly hired a fraudulent candidate; the FBI has documented 300+ US companies that hired North Korean operatives.

**The applicant** (free): proves once from their phone, in about a minute, and reuses the proof with the next landlord or lender.

---

## Why now

| Date | Event | Why it matters |
|---|---|---|
| 2025–26 | AI document forgery ~5× (Inscribe) | Document review is broken |
| 2026 | 74.9M independent workers (MBO Partners) | The unverifiable segment keeps growing |
| 2024–26 | Work Number price +272% since 2012; antitrust class action; mortgage-lender complaints | Buyers want an alternative |
| 2026 | zkTLS reaches production: Reclaim 1M+ verifications, 25+ clients; a 1.2MB response attested in 7.3s on a mid-range phone (IACR 2026); $40M+ invested across zkTLS teams | The technology is ready for mainstream workloads |
| 2026 | CFPB open-banking rule (§1033) enjoined and being rewritten | Bank-API access is uncertain; zkTLS needs no bank cooperation |
| Dec 31, 2026 | Medicaid work requirements live (some states earlier) | 20M people must prove hours or income, many from gig work |

---

## Product

### How a proof works

1. **Request.** The verifier sends a link from their leasing or loan system, or via API. Rules come with it, for example "monthly income ≥ $4,950" or "≥ 80 hours last month."
2. **Prove.** The applicant opens the link on their phone and signs in to the source they choose, in their own session. Fontis never sees or stores the password.
   - The **AI agent** finds the right pages (earnings, hours, balances, tenure) and extracts the fields.
   - **zkTLS** attests that those bytes came from the real site over an authentic TLS session.
   - **Selective disclosure** reveals only what the rule needs.
3. **Decide.** The verifier gets a proof report with the result against the rule, the source, the date range and a proof ID anyone can re-verify. Typical time is a couple of minutes, not days.
4. **Reuse.** The applicant keeps the proof in a Fontis wallet and can share it with the next verifier while it's fresh.

### What AI does and what cryptography does

| AI (the coverage engine) | Cryptography (the trust engine) |
|---|---|
| Navigates any portal: gig dashboards, payroll systems, banks, marketplaces, tax accounts | zkTLS proves the data came from the real site and wasn't altered |
| Writes a new source "recipe" when a site isn't covered yet, so coverage grows without integrations | Selective disclosure shares facts, not statements |
| Computes income and hours under the verifier's rules (3× rent, 80 hours, 24-month self-employed averaging) | Proofs are verifiable by anyone, and optional on-chain attestations (EAS) give audit trails and revocation |
| Guides the applicant in plain language; falls back to document forensics only when no source exists | Proofs live in an applicant-held wallet, portable and reusable across verifiers |

### Expansion

- **New proof types:** employment history and tenure, education, identity binding, platform reputation (ratings, tenure on gig apps), creator earnings and audience, on-chain income such as stablecoin payouts.
- **New verifiers:** insurers, employers, marketplaces and government programs.
- **Outside the US:** any bank or employer website anywhere works the same way, which payroll databases can't match.

---

## Market size

| Layer | Basis | Size |
|---|---|---|
| US income and employment verification spend | Equifax verification services $2.18B (2025), plus Experian Verify, Truv, Argyle, Pinwheel and Plaid income | ~$3B/yr |
| US tenant screening | Industry estimates | $2.3–3.5B/yr |
| Fontis serviceable market (US wedge) | ~150M+ income and hours checks a year across rentals, consumer and auto lending, mortgages and Medicaid redeterminations; ~30–40% outside payroll-database coverage; × $6–12 per proof | **~$400–600M/yr** (estimate) |
| Expansion | Hiring and identity, employment history, international, other proof types | Multi-$B |

$100M ARR requires roughly 10–15M paid proofs a year: a few percent of US verifications, or a third of the unverifiable segment. *The check counts and segment shares are estimates to validate with the first 20 customers.*

---

## Business model and unit economics

- **Price:**
  - $6–12 per verified proof for rental and lending; enterprise volume tiers.
  - Pay-per-success: no proof, no charge. This is cheaper than The Work Number ($55–70) and manual review.
  - Health plans pay per member verification.
- **Cost per proof:** cents of inference and attestation compute, plus support. **Gross margin target 85%+.** Comparables: Equifax Workforce Solutions runs a 44% *operating* margin, and Plaid is profitable.
- **Illustrative customer:** a 20,000-unit property manager processes ~12,000 applications a year. About 35% of them, roughly 4,200 proofs, at $9 is ~$38k a year, plus fraud savings. Snappt-style ROI: one avoided fraudulent tenancy costs a landlord thousands in lost rent and eviction.
- **Distribution economics:** integrations into leasing and loan systems put Fontis inside the existing verification waterfall. Each applicant who gets a proof can reuse it, so the next verifier who accepts it gets a free onboarding.

---

## Competition and moat

| Alternative | What it does | Gap Fontis closes |
|---|---|---|
| The Work Number (Equifax) | W-2 payroll database, $55–70 per check | Can't see gig or self-employed income; expensive |
| Truv, Argyle, Pinwheel | Consumer-permissioned payroll APIs | Integration-by-integration coverage; ~55% conversion; payroll-centric |
| Plaid (Consumer Report, Income) | Bank-data income inference | Shows deposits, not their source; depends on bank APIs and the §1033 rule |
| Snappt, Inscribe, Ocrolus | AI document forensics | Judges documents, which is the arms race being lost |
| SteadyIQ | Gig income verification for state agencies | Government-focused, integration-based |
| zkTLS infrastructure (Reclaim, zkPass, Opacity) | Developer SDKs mostly sold to crypto apps | Not packaged for landlords, lenders and health plans. Potential suppliers or partners |

**What compounds:**

1. **The coverage library:** AI-written source recipes, verified by use. The long tail is the moat.
2. **A two-sided network:** applicants reuse proofs, and verifiers accept proofs already in applicants' wallets.
3. **Outcome data:** which sources and patterns predict a real applicant.
4. **Compliance posture:** consent records, audit trails, CRA-ready operations.

---

## Go-to-market

1. **First 10 customers: fraud-heavy property managers.** Run a free pilot on 500 applications as a fallback after their existing screening. Measure verified versus "unverifiable" applicants, fraud caught and hours to decision.
2. **First 100: integrations and lenders.** List in property-management and leasing-system marketplaces and loan-origination systems. Close fintech lenders with gig-heavy books. Pilot with one Medicaid plan on hours proof for gig members.
3. **Then product-led growth.** A free tier for independent landlords, where the applicant shares an existing Fontis proof. Each proof request is a referral, and the applicant wallet drives verifier adoption.

---

## Success likelihood and risks

**My estimate:** a **30–35% chance of a Series A within 24 months**, against a 15–20% base rate (Carta via ValueAdd VC). The reasons: an existing budget, per-use revenue from the first pilot, and a wedge incumbents can't serve.

| Outcome | Estimate |
|---|---|
| $1M ARR in 12 months | ~45% |
| Series A in 24 months | 30–35% |
| $100M ARR | ~6–8% |

*These are judgment calls, not model outputs.*

| Risk | Mitigation |
|---|---|
| A source site blocks automated sessions or objects under its terms | Proofs run in the applicant's own session viewing their own data; diversify sources; partner where available; graceful fallback |
| FCRA: being treated as a consumer reporting agency | Launch as a consumer-permissioned data provider (the Argyle and Truv model); partner with CRAs; register as a CRA when volumes justify it (the Plaid Check path) |
| Plaid, Truv or Equifax copy it | Win the long-tail coverage library and the applicant wallet network first; the company becomes a natural acquisition target |
| zkTLS trust assumptions (notary or proxy collusion) | Multiple notaries, TEEs and open audits; publish the security model |
| Verifier inertia | Sold as a fallback inside existing waterfalls, priced below manual review, pay-per-success |
| Medicaid policy changes | One segment among several, not the core |

**Kill criteria at day 120:**

- fewer than 70% of "unverifiable" applicants complete a proof, or
- fewer than 5 paying verifiers, or
- the median time to proof is over 5 minutes.

---

## 18-month plan and fundraising

**Raise a $4–6M seed after the first pilots. Raise a $20–30M Series A at $5M+ ARR.**

| Phase | Months | Build | Gate |
|---|---|---|---|
| 1. Leasing fallback | 0–4 | Proof links and reports; top 50 gig, payroll and bank sources; AI recipe writer; 5 property-manager pilots | ≥70% completion on unverifiable applicants; first paid contracts |
| 2. Lending + network | 5–10 | Lending rules (self-employed averaging), loan-system integration, applicant wallet and reuse, Medicaid hours pilot | $1M ARR run-rate; ≥15% of proofs reused |
| 3. Scale | 11–18 | Leasing-marketplace integrations, self-serve tier, employment-history proofs, CRA readiness | $5M ARR; net revenue retention >130% |

**Founding team:**

1. A proof and compliance engineer. The Recourse builder shipped per-line SHA-256 sealed proofs and rule engines, which is the same mindset.
2. An applied cryptographer (zkTLS, MPC, TEEs).
3. A go-to-market lead from proptech or lending.

---

## Sources

All found via search, Oct 6, 2026. Many primary pages could not be opened directly from this environment, so figures come from search summaries.

- **Monetization comparables:** [Equifax Q4 2025 results](https://investor.equifax.com/news-events/press-releases/detail/1391/equifax-delivers-fourth-quarter-2025-revenue-growth-of-9) · [Equifax 10-K FY2025](https://www.sec.gov/Archives/edgar/data/33185/000003318526000010/efx-20251231.htm) · [Plaid $546M ARR (Sacra)](https://sacra.com/research/plaid-at-546m-arr-growing-40-yoy/) · [Snappt profile (Tracxn)](https://tracxn.com/d/companies/snappt/___P0bbcQYf9mVkIroApw_Tk9X92ZEs_ZxWiDf-ouQieA) · [Tenant screening market](https://marketintelo.com/report/tenant-screening-software-market)
- **Pricing and incumbent pain:** [Work Number pricing and complaints (HousingWire)](https://www.housingwire.com/articles/equifaxs-work-number-is-too-expensive-and-regulators-should-look-into-it-chla/) · [Truework on verification pricing](https://www.truework.com/resource-center/blog/understanding-income-and-employment-verification-pricing) · [Truv/Argyle pricing](https://slashdot.org/software/income-verification/) · [Coverage gaps and conversion](https://fintechspecs.com/blog/income-verification-api-lenders/) · [Ocrolus on gig income](https://www.ocrolus.com/blog/gig-income-verification-consumer-lending/)
- **Fraud:** [Inscribe 2026 State of Document Fraud](https://www.inscribe.ai/reports/2026-document-fraud-report) · [Checkr 2026 rental fraud report](https://checkr.com/resources/articles/2026-rental-fraud-report) · [Gartner fake candidates and DPRK hires](https://www.refontelearning.com/blog/ai-fake-candidate-fraud)
- **Workforce:** [MBO Partners State of Independence 2026](https://www.mbopartners.com/state-of-independence/) · [Staffing Industry on 2025 counts](https://www.staffingindustry.com/news/global-daily-news/number-of-independent-workers-rises-again-mbo-partners)
- **Technology:** [Reclaim Protocol paper (IACR 2026)](https://eprint.iacr.org/2026/2187) · [zkPass Series A](https://www.prnewswire.com/apac/news-releases/series-a-funding-secures-12-5m-for-zkpass-to-drive-zktls-oracle-protocol-302281736.html) · [Opacity seed](https://www.theblock.co/post/321160/opacity-network-funding-zk-data-verification) · [zkTLS overview](https://bex.co/blog/2026/01/13/zktls-verifiable-web-data-zero-knowledge-proofs)
- **Regulation:** [Section 1033 status (Cozen O'Connor)](https://www.cozen.com/news-resources/publications/2026/section-1033-compliance-date-open-banking-rule-enjoined-and-under-reconsideration) · [Non-FCRA consumer-permissioned data (Argyle)](https://www.argyle.com/blog/using-non-fcra-data-for-income-employment-verifications) · [Medicaid work requirements timeline](https://federalnewsnetwork.com/technology-main/2026/01/technology-emerges-as-key-for-states-to-meet-medicaid-work-verification-rules-before-december-2026-deadline/) · [CBPP vendor landscape](https://www.cbpp.org/research/health/assessing-the-medicaid-work-requirement-vendor-landscape)
- **Rejected finalists:** [Félix $200M round](https://valueaddvc.com/blog/felix-pago-valuation-2026-200m-series-c-1-4b-whatsapp-remittances) · [KAST Series A](https://www.blockhead.co/2026/03/10/kast-raises-80-million-for-stablecoin-neobank-at-600-million-valuation/) · [Kontigo seed](https://www.fintechfutures.com/blockchain-crypto-digital-assets/kontigo-raises-20m-in-seed-funding) · [Axiom (Forbes)](https://www.forbes.com/sites/boazsobrado/2026/02/26/axiom-became-ycs-biggest-hit-then-someone-was-watching/) · [Phantom statistics](https://sqmagazine.co.uk/phantom-wallet-statistics/) · [Pilot AI accountant at $99](https://beancount.io/forum/t/pilot-ai-accountant-does-end-to-end-bookkeeping-for-99-month-now-should-beancount-users-be-worried-or-is-this-proof-we-were-right-all-along/3660) · [CoinTracker revenue](https://getlatka.com/companies/cointracker.io) · [Rillet $1B](https://www.rillet.com/blog/rillet-raises-100m-series-c-at-1b-valuation-to-build-accounting-superintelligence)
