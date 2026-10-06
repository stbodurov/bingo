# Startup Thesis: Web3 × AI, picked profitability-first

Oct 6, 2026 · working name **Clearlien** · supersedes the Lading draft (see "How it was chosen")

## The pick: Clearlien

**Build Clearlien, the collateral truth layer for private credit.**

- AI verifies every pledged loan, invoice and lease against primary sources (titles, VINs, bank collections, shipment data, the account debtor's real domain), continuously instead of once a year by sampling.
- Each verified asset is registered on-chain at pledge time as a privacy-preserving fingerprint, so a second lender sees a collision before funding.
- New originations can be issued as UCC Article 12 *controllable electronic records*, where the lender's **control beats any filing**, so the same asset cannot be pledged twice.

It sells the *work* lenders already pay consultants and field examiners for, and builds the shared registry as a byproduct. That is the order of operations that makes it profitable.

- **For lenders (banks, private credit funds, factors):** know that every asset in the borrowing base exists, is performing, and is pledged only to you, every day, at a fraction of today's exam-plus-consultant cost.
- **For borrowers (originators, sellers of receivables):** a "Clearlien-verified" tape gets funded faster and cheaper, because the lender's fraud haircut shrinks.

**The wedge:** warehouse and back-leverage facilities secured by auto and consumer loans. This is the Tricolor pattern; Auto ABS was $123B, 34% of 2025 ABS issuance.

**The durable business:** a cross-lender registry and a control layer, the "MERS/DART for everything that isn't a mortgage." Network effects compound, and comparable registry and data businesses run at 55–88% margins.

---

## Profitability check: is this a profitable niche?

The question to answer before anything else. Four tests, each with evidence:

| Test | Evidence | Result |
|---|---|---|
| **1. Money is already being spent on the problem** | Collateral-verification demand **+80% at Guidehouse** and **+50% at Vervent** after Tricolor; Setpoint and CBIZ also report more demand (Bloomberg, Aug 21, 2026). ABL field exams cost **$500–$2,000 per examiner-day**, 1–2 times a year per facility, paid by the borrower. ABL monitoring ≈ 0.5% of commitment. Every warehouse has a verification agent or backup servicer | ✅ Existing, growing budget lines |
| **2. The pain is large and fresh** | First Brands: $2.3B of receivables "vanished" (Sep 2025). Tricolor: $800M collateral hole; 3,225 loans put in two securitizations at once (SEC complaint, Aug 18, 2026); JPMorgan $170M and Fifth Third ~$200M charges. HPS/BNP: $430–500M lent against fake telecom receivables, confirmed by **spoofed-domain emails**. MFS (UK): £930M shortfall, Barclays £600M exposure (Feb 2026). Stenn: ~$1B owed (Dec 2024) | ✅ ~$5B+ of losses in about 12 months |
| **3. Price benchmarks support high revenue per customer** | MERS charges **$24.95 per loan registration**. Lender-finance books are huge: US bank loans to non-depository financial institutions (NDFIs) hit **$1.4T** at end-2025, growing 22.7% a year (FDIC); bank ABL commitments $369.5B (SFNet); US factoring +35.5% in 2025 (FCI) | ✅ 2–3 bps on monitored collateral = $0.5–1.5M a year per mid-size lender |
| **4. Comparable businesses are very profitable** | **Figure** (runs the DART lien/eNote registry on its own blockchain, Provenance): Q2 2026 revenue $226M (+113%), **54.6% adj. EBITDA, 38.8% net margin**. FICO Scores: **88%** segment margin. Verisk: 54% EBITDA. TriumphPay (a freight payment network that stops double factoring): **25.7% EBITDA margin**, $89M annualized revenue | ✅ Registry and verification economics are among the best in finance |

**The counter-evidence, stated plainly:**

- **MonetaGo** has done double-financing registries since ~2016 and is still small (~33 employees, ~$23M raised). It sold *only* a registry, mainly to bank associations in emerging markets, and in 2019–20 it moved *off* blockchain (Fabric, then Corda, then cloud) because of scalability issues. **Lesson:** a pure registry has a cold-start problem and weak pricing. Clearlien charges for verification from day one, so the registry never has to pay for itself.
- **TriumphPay** ran EBITDA-negative for years before turning to 25.7%. **Lesson:** network businesses are slow unless each customer gets standalone value on day one. Verify does; the registry is the bonus.
- **Crypto-native B2B tools are often unprofitable** (Chainalysis: ~$190M ARR, still not profitable). **Lesson:** sell to traditional-finance budgets with the chain under the hood, which is what Figure does.

**My answer:** this niche passes all four tests, and the counter-evidence shapes how the business is built rather than ruling it out. The earlier picks do *not* pass as cleanly; see the screen below.

---

## How it was chosen (from scratch, profitability-first)

I scanned web3 × AI in late 2026 for:

- where money is flowing,
- what broke recently,
- who pays today, and how much.

I scored 12 ideas. The filters, in order:

1. An existing budget line.
2. Gross margin structure.
3. Profitable comparables.
4. Market ceiling.
5. Competition.
6. Blockchain and AI both genuinely necessary.
7. Founder fit.

| Idea | Who pays today (evidence) | Margin structure | Profitable comps? | Ceiling / competition | Verdict |
|---|---|---|---|---|---|
| **Collateral verification + on-chain registry (Clearlien)** | Lenders: consultants +80% demand, field exams, verification agents | Software/data 70–80% | Figure 55% EBITDA; FICO 88%; MERS | ~$0.6B US Verify revenue pool, $1.5–2B global; no neutral US cross-lender registry | **Pick** |
| Prediction-market resolution and integrity | Venues (Kalshi $1.5B+ annualized revenue); 1,150+ Polymarket disputes Jan–May 2026; WSJ/CFTC probes | Software 70%+ | Solidus (private) | ~20 venues and Kalshi has ~80% share, so one buyer dominates; Chaos Labs ($75M raised, Kalshi as customer), Chainlink and UMA are already in Polymarket V2's OracleAggregator | Real pain, **low ceiling and concentrated buyers** |
| Lading: AI-verified stablecoin trade escrow (earlier pick) | SME importers (LC fees 1.2–3%) | Payments take 0.5–1%, squeezed by ramp costs | Payoneer (~10% net margin), Airwallex | **Tazapay already does escrow released on verified shipment** (Circle- and Coinbase-backed, $10B+ annualized volume) | **Dropped:** thin margins, incumbent exists |
| Crypto-tax CP2000 notice resolution (1099-DA gross-proceeds wave) | Consumers and CPAs | Services 50–70% | Koinly (profitable, private) | Peaks in 2026–27, then shrinks as basis reporting matures; consumer CAC | Profitable small business, **not venture-scale** |
| Dispute handling for agent stablecoin payments | Nobody yet: x402 has ~$53M *cumulative* volume, and OpenAI shut Instant Checkout in Mar 2026 | — | — | — | **No budget yet** |
| KYA (Know Your Agent) identity | Nobody at scale | — | — | Catena (a16z), Skyfire, Visa/Mastercard/Cloudflare | No budget, crowded |
| AI smart-contract audits | Protocols ($2.68B hacked YTD 2026) | Services | CertiK, OpenZeppelin | Octane, Almanax, Sherlock; price war | Crowded |
| GENIUS Act and crypto AML compliance | Issuers and banks | Software | TRM ~$100M ARR at $2B valuation; Chainalysis unprofitable | Two well-funded leaders | Crowded |
| Stablecoin rails / B2B payouts | Businesses | FX spread, thin | BVNK ($1.8B to Mastercard), Bridge ($1.1B to Stripe) | Consolidated | Late |
| Tokenized-stock corporate-actions operations | xStocks, Robinhood Chain and a few others | Software | — | Fewer than 10 buyers | Too narrow |
| DePIN data for robotics | AI labs | Hardware-heavy | Whole sector ~$82M on-chain revenue | Lightwheel and others are funded off-chain | Weak economics |
| AI-crawler pay-per-use | Publishers | Network fee | — | Cloudflare owns it | Incumbent |

**Why the earlier pick (Lading) lost:** it is a payments business with a 0.5–1% take rate that ramp fees erode, and a funded incumbent (Tazapay) already escrows payments until shipment is verified. Clearlien sells into a budget that just grew 50–80%, at software margins, with registry economics on top.

---

## Is it unique, and can it be a unicorn?

| Unicorn trait | Evidence | Clearlien |
|---|---|---|
| Timing (42% of the success gap; Idealab) | Three double-pledging collapses in about 12 months; SEC Tricolor charges (Aug 2026); FSB private-credit vulnerabilities report (May 2026); SFA fraud survey ranks double-pledging the **#1 concern** (Mar 2026) and set up asset-class working groups; New York's Article 12 took effect **Jun 3, 2026** (33 states + DC in total) | **Strong** |
| Non-consensus insight | Private credit thinks of crypto as a fraud vector, and the one registry startup left blockchain | **Strong:** see below |
| Market ceiling | ~$0.6B US Verify revenue pool (2.5 bps on ~$2T of US pledged collateral) + registration/transfer fees; global ≈ 2.5–3× | Venture-scale if it becomes the standard |
| Compounding moat | Network effects: each lender joining makes double-pledging detectable for all | **Two-sided registry** + an exceptions dataset |
| Speed | Enterprise sales (3–9 months), but private credit funds and factors move faster than banks | Medium |
| Competition | Setpoint (a16z; warehouse ops), Guidehouse/CBIZ (consulting), Vervent (servicer), MonetaGo (non-US registry), Figure DART (mortgages, Figure's own ecosystem), Wolters Kluwer (UCC search), Ocrolus/Inscribe (document fraud) | **Nobody offers neutral, cross-lender, primary-source verification for non-mortgage collateral in the US** |

**The non-consensus insight.** Every post-mortem says "lenders need better audits." But audits sample the borrower's own documents, and documents are now forgeable at zero cost: HPS was fooled by debtor confirmations from look-alike domains. The fix is to make collateral *self-proving*:

1. Check each asset against sources the borrower does not control (title records, VINs, bank receipts, shipment data, the debtor's authenticated domain).
2. Put its fingerprint in a place no single party owns, where a second pledge collides.
3. Where the law allows, make the lender's claim *technical control* rather than a filing.

The cheapest neutral place for (2) and (3) is a public blockchain, the technology most of private credit still distrusts. Figure has already proven the model in mortgages: it runs DART on Provenance and earns 55% EBITDA margins, and Goldman, Jefferies and Deutsche Bank use DART.

**Why MonetaGo's exit from blockchain doesn't kill the thesis.** MonetaGo left *permissioned* chains (Fabric, Corda) in 2019–20 for scalability reasons, and only needed duplicate *detection*. Clearlien needs a chain for two things a database cannot give:

- **Neutrality among competing lenders:** public commitments that courts, auditors and rivals can verify, so there is no vendor-owned honeypot of everyone's collateral.
- **Article 12 control:** "control beats filing" priority requires a record whose exclusive control is technically enforced.

Article 12 also creates a new risk for every filing-only lender: someone who later takes *control* primes them. That makes a registry of controllable records something filing-only lenders now need to monitor. Public L2s now handle this volume for cents.

**Odds.** Base rates (same as the Recourse thesis): ~1.3–2.5% of seed-funded startups become unicorns. My judgment: **~7–9%** for Clearlien, if the registry reaches lender-network critical mass in one asset class.

**What must be true by month 9:**

- 8+ paying lenders and $3M+ ARR.
- $20B+ of collateral monitored.
- At least one real cross-lender duplicate or phantom asset caught. This is network proof, and the best marketing in the industry.
- At least one originator registering voluntarily because a lender priced it in.

---

## Problem and customer

**How the frauds actually worked:**

- **Tricolor (auto loans):** the same loans were pledged to multiple warehouse lines and securitizations; $2.2B was pledged against ~$1.4B of real assets.
- **First Brands (trade receivables):** invoices were fabricated or inflated (reportedly up to 50×) and factored to multiple funders through off-balance-sheet programs. Raistone, the SCF platform that relied on First Brands for 80% of its revenue, liquidated in Feb 2026.
- **HPS/BNP (telecom receivables):** fake customers, backdated contracts, and confirmation emails from spoofed domains. *Verification by email is broken.*
- **MFS (UK bridging loans):** the same properties were pledged to multiple lenders; £930M missing.

**Why existing controls miss it:**

- UCC filings sit at the borrower level ("all accounts"), not the asset level.
- Field exams sample once or twice a year from the borrower's own systems.
- Verification agents reconcile the tape the borrower hands them.
- No lender can see another lender's collateral.
- Generative AI makes fabricated invoices, bank statements and emails trivial to produce (Inscribe, 2026 State of Document Fraud).

**Customer:** heads of lender finance, warehouse lending, ABF/ABL and portfolio risk at:

1. Private credit funds with ABF and back-leverage books.
2. Banks with NDFI exposure ($1.4T, 86% held by banks with $100B+ in assets).
3. Factors and SCF funders.

Secondary buyers: ABS investors, trustees and rating agencies, who want verified tapes. The economic buyer often passes the cost to the borrower, as field exams already are.

---

## Why now

| Date | Event | Why it matters |
|---|---|---|
| Sep 2025 | First Brands Ch. 11 ($2.3B of receivables missing); Tricolor Ch. 7 | Banks take losses: JPMorgan $170M, Fifth Third ~$200M |
| Nov 2025 | HPS/BNP telecom receivables fraud revealed ($430–500M) | Document and email verification shown to be forgeable |
| Dec 2025 | FDIC: bank NDFI loans $1.4T, +22.7% CAGR | Regulators now track this exposure |
| Feb 2026 | MFS administration (£930M shortfall); Raistone liquidates | The pattern repeats across asset classes and countries |
| Mar 31, 2026 | SFA/EY fraud survey ranks double-pledging #1; asset-class working groups formed | The industry is actively looking for a standard |
| May 6, 2026 | FSB report on private-credit vulnerabilities | Global regulators are focused on data gaps |
| Jun 3, 2026 | NY Article 12 takes effect (33 states + DC) | "Control beats filing" becomes law where most financings are governed |
| Aug 18–21, 2026 | SEC charges Tricolor execs; Bloomberg reports verification demand +50–80% | Budgets are moving **now** |

**The slower shifts:**

- ABF is the fastest-growing private credit strategy (16.7% CAGR forecast).
- On-chain private credit reached ~$36B represented value, with Figure ~60% of it.
- LLM cost per unit of capability keeps falling, so verifying 100% of assets is now cheaper than sampling 5% by hand.

---

## Product and value proposition

> **Clearlien tells a lender, every day, that every asset in the borrowing base exists, performs, and is pledged only to them, and it proves it from sources the borrower doesn't control.**

### 1. Verify (the wedge, sold as work)

- **Ingest:** the borrower's loan tape or receivables aging, plus servicing-system exports and read-only bank collections (a standard borrower covenant).
- **Primary-source checks by asset class:**
  - **Auto loans:** VIN decode, lien and title record (state title data and NMVTIS), obligor identity, payments matched to actual bank receipts, dealer and vehicle sanity checks.
  - **Trade receivables:** account-debtor existence; confirmation through the debtor's *authenticated* domain or AP portal, not look-alikes; PO, shipment and customs evidence (this reuses the Recourse invoice engine); dilution and payment history.
  - **Bridge and real-estate loans:** land-registry charges and existing liens.
- **AI analysis:** tape anomalies (duplicated assets across facilities, impossible payment histories, 50× invoice inflation, round-tripped collections) and generative-forgery detection on documents.
- **Output:** a daily exception report with an evidence pack per exception, replacing sampled field exams and monthly verification-agent reports. Humans review the exceptions, not the whole tape.

### 2. Registry (the network)

- **Fingerprints:** at pledge, each verified asset gets a fingerprint built from its *real-world anchors* (VIN + obligor; debtor + PO + amount + date; title number), not from the forgeable document.
- **Privacy:** fingerprints are computed with a threshold oblivious-PRF so they cannot be brute-forced, and committed on a public L2. The commitment reveals nothing about the deal.
- **Collisions:** a match alerts both lenders through Clearlien without revealing either lender's identity until it escalates.
- **Free to query** for lenders, to grow the network. Originators pay a per-asset registration fee that lenders require as a covenant (benchmark: MERS at $24.95; Clearlien at **$1–5 per asset**).

### 3. Control (the Article 12 upgrade)

- New originations can be issued as controllable electronic records: the obligor agrees in the contract to pay whoever controls the record, which makes it a controllable payment intangible or controllable account.
- The lender takes on-chain control, giving priority over any filing and making a second pledge technically impossible. Moving the asset to a securitization or back-leverage facility is a control transfer in minutes, not a re-papering.
- **Fees:** transfer and release fees, which is how MERS and DART make money.

### Why web3 and why AI

- **Web3:** neutral cross-lender commitments, Article 12 control, and composability with on-chain private credit, the model Figure proved. Verify revenue does not depend on the chain; the moat does.
- **AI:** checking every asset across heterogeneous, partly adversarial sources is reasoning work. Generative forgery makes document review alone obsolete. And per Sequoia's "Services: The New Software," this is outsourced, mostly-intelligence work (consultants and field examiners) sold as an outcome.

---

## Market size

| Layer | Calculation | Size |
|---|---|---|
| US pledged collateral in lender-finance, warehouse, ABL, ABF and factoring structures | Bank NDFI loans $1.4T (FDIC), bank ABL commitments $369.5B (SFNet), plus private-credit ABF and back leverage and factoring, de-duplicated (estimate) | ~$2T |
| **Verify revenue pool (US)** | × 2.5 bps a year (versus ~50 bps all-in ABL monitoring and borrower-paid exams) | **~$500M/yr** |
| Registry fees (US) | Tens of millions of new loans and invoices pledged a year × $1–5 (estimate) | ~$50–150M/yr |
| Control / transfer fees | Per transfer, as with MERS/DART | Upside |
| **Global** | UK/EU (factoring alone is €2.66T of annual turnover in Europe), APAC | **~$1.5–2B/yr** |

$100M ARR is about 15–20% of the US pool. That is plausible for the category standard in a market with network effects; MERS is effectively the whole mortgage market. *The bps price, the de-duplicated collateral base and registry volumes are assumptions to test with the first 10 lenders.*

---

## Business model and unit economics

**Revenue lines:**

1. **Verify:** 2–3 bps a year on monitored collateral, with a floor per facility. Often reimbursed by the borrower, like field exams.
2. **Registry:** $1–5 per registered asset, paid by originators.
3. **Control:** per-transfer and per-release fees.
4. **Forensic sprints:** fixed-fee reviews when a lender suspects a problem, the work Guidehouse sees +80% demand for. A paid entry point that converts into Verify.

**Per customer (illustrative lender-finance group: 20 facilities, $4B of collateral):**

| | Assumption | Year 1 | Year 2+ |
|---|---|---|---|
| Verify | 2.5 bps × $4B | $1.0M | $1.2M (book growth plus more facilities) |
| Registry | Borrowers register ~300k assets × $2 | $0.6M | $0.7M |
| What it replaces | 20 facilities × ~2 exams × ~$15k, plus verification-agent fees, plus forensic consultants | ~$0.6–1.0M of existing spend, now with 100% daily coverage | |
| Fraud loss avoided | One Tricolor-style hit = $170–200M at a single bank | Insurance value, not counted | |

- **Gross margin target 70–80%.** Costs: inference, data (title, VIN and lien lookups), bank-data access, human exception review. Starts lower (~60%) while human review is heavy.
- **Early ACVs are lower:** $250–500k for a first lender, expanding as facilities are added. **CAC ~$120k** (enterprise AE plus solutions engineer, 4–6 month cycle), **payback ~5–7 months** at $300k ACV and 70% gross margin. The B2B SaaS median is 16 months.
- **Breakeven at ~$20M ARR:** a 60-person team at ~$250k loaded is ~$15M opex, plus COGS. That is about 50 lenders at $400k, roughly a quarter of the 200+ US banks, funds and factors with material lender-finance books.
- **Lumpiness risk is low:** revenue is recurring and tracks collateral size, not fraud headlines.

---

## Competition and moat

| Alternative | What it does | Gap Clearlien closes |
|---|---|---|
| Field examiners, Guidehouse, CBIZ | Sampled, periodic, human; reads the borrower's documents | 100% daily coverage from primary sources |
| Verification agents / backup servicers (Vervent, trust banks) | Reconcile the tape they are given | Independent sources; cross-lender visibility |
| **Setpoint** (a16z, Wells Fargo and Citi invested) | Warehouse operations OS for originators and lenders | Closest threat. Setpoint is the originator's system of record; Clearlien is a neutral check *across* lenders and against outside sources. Could partner (Setpoint as a data source) |
| Cardo AI, Cascade, Finley | ABF data, modeling, monitoring | Analytics on the borrower's data, not proof of existence or uniqueness |
| MonetaGo | Hash registry for bank associations (Singapore, India; FCI) | Registry only, non-US, no AI verification, no control layer |
| Figure DART | On-chain lien/eNote registry for Figure-ecosystem mortgages and HELOCs | Clearlien covers auto, consumer, SMB, receivables and bridge loans, and is lender-neutral. Figure is a partner or acquirer, not a direct competitor |
| Wolters Kluwer / CSC | UCC search and filing | Borrower-level filings; blind to asset-level double pledges and to Article 12 control |
| Ocrolus, Inscribe | Document-fraud detection | Checks the document, not whether the asset exists or is unique |

**What compounds:**

1. **The registry network:** each lender makes double-pledging detectable for every other lender. This is winner-take-most, as with MERS.
2. **The exceptions dataset:** labeled fraud patterns across facilities make detection better than any single lender's.
3. **Primary-source integrations** (title data, bank data, carrier and customs data) are slow and expensive to rebuild.
4. **The control layer:** once assets are controllable records, Clearlien is how they move.

---

## Go-to-market

**Lead with the fear that already has a budget.**

- **First 5 customers: forensic sprints.** Offer private-credit lender-finance funds a fixed-fee, two-week verification of one auto-loan facility, with the exception report as the demo. Funds burned by Tricolor or First Brands, or exposed to similar originators, buy quickly. Convert to Verify.
- **First 25: asset-class standard.** Join the SFA Auto ABS and Consumer Loan working groups, offer the registry schema as an open standard, and publish anonymized "phantom collateral" findings quarterly.
- **First 100: borrower pull.** Lenders write Clearlien registration into facility covenants, originators register once and reuse it across lenders, and verified originators get better advance rates. Then add trade receivables and factoring, where the founder's supply-chain and customs edge is strongest and US factoring grew 35.5% in 2025.
- **Channel:** ABS trustees and rating agencies can require verified tapes. Figure-style on-chain private-credit platforms are natural integrators for the control layer.

---

## Success likelihood and risks

**My estimate:** a **35–40% chance of raising a Series A within 24 months** (base rate 15–20%, Carta via ValueAdd VC), because buyers' budgets have measurably moved and ACVs are high.

| Outcome | Estimate | Reasoning |
|---|---|---|
| $1M ARR within 12 months | 55% | Three or four funds at $250–350k; forensic sprints bring cash in the first quarter |
| Series A within 24 months | 35–40% | Needs $8–10M ARR, a network proof point (a caught duplicate) and a bank logo |
| $100M ARR (venture scale) | ~7–9% | Requires becoming the registry standard in at least two asset classes |

*These are judgment calls, not model outputs.*

| Risk | Mitigation |
|---|---|
| **Setpoint (a16z) adds cross-lender checks** | Neutrality (Setpoint is the originator's system), primary-source depth, an open registry standard; partner where possible; pitch to Sequoia and a16z crypto rather than a16z fintech |
| **Lenders won't share data** | Commitments reveal nothing; free queries; collisions are revealed only to the two parties; standard set through the SFA |
| **Data access** (title, VIN, bank data) | Established commercial title and lien data vendors; bank data under existing borrower covenants; start with auto, where anchors are strongest |
| **Liability if fraud slips through** | Sold as verification with evidence, not a guarantee; contractual caps; E&O insurance |
| **"Blockchain" scares buyers** | The chain is invisible in Verify and Registry; Control is opt-in for Article 12-ready deals. Figure proves banks will use it |
| **The fear fades** | Regulators are institutionalizing scrutiny (FDIC NDFI reporting, FSB, SEC); AI forgery keeps rising; contractual covenants lock in usage |
| **Slow bank sales cycles** | Start with private credit funds and factors; banks follow once peers adopt |

**Kill criteria at day 120:**

- fewer than 3 paying lenders at $100k+ ACV, or
- zero material exceptions found across the first 5 facilities (the value would then be insurance only), or
- originators refuse registration covenants in more than 50% of cases.

---

## 18-month plan and fundraising

**Raise a $5–6M seed after Gate 1** (enterprise build plus data costs). **Raise a $25–35M Series A at $8–10M ARR** with a network proof point.

| Phase | Months | Build | Gate |
|---|---|---|---|
| 1. Forensic wedge | 0–4 | Auto-loan Verify (tape plus title/VIN plus bank receipts), exception workflow, registry v0 (fingerprints plus L2 commitments) | 3 paying funds, ≥$750k ARR, real exceptions found |
| 2. Network | 5–10 | Consumer and SMB loans, open registry API, SFA working-group standard, first bank customer, covenant templates | $3M ARR, $20B monitored, first cross-lender collision caught |
| 3. Standard | 11–18 | Trade receivables and factoring (shipment and customs proof via the Recourse engine), Article 12 control pilot on a NY-law facility, trustee and rating-agency integrations | $8–10M ARR, NRR >130%, ≥1 control-layer facility live |

**What a16z (crypto) or Sequoia will need at the A:**

- $8–10M ARR with gross margin at 70% or above and human-review hours per asset falling each quarter.
- **Network proof:** share of new facilities whose originator is already registered, and collisions caught.
- Net revenue retention above 130% (facilities and asset classes added).
- At least one bank and one ABS trustee or rating-agency integration.
- Zero material misses on verified facilities.

**Founding team:**

1. An engineer from supply-chain compliance software. The Recourse builder already shipped invoice parsing, a dated rulebook and per-line SHA-256 sealed proofs, which is the same primitive as asset fingerprints.
2. A co-founder from lender finance or warehouse lending (ex-bank lender-finance or private-credit operations) who owns the buyer relationships.
3. A cryptography and smart-contract engineer (OPRF commitments, Article 12 control records).

Advisor: a UCC/secured-transactions lawyer.

---

## Sources

All found via search, Oct 6, 2026. Many primary pages (Bloomberg, SEC, FDIC, a16z, ABF Journal) could not be opened directly from this environment, so figures come from search summaries and the listed coverage.

- **The frauds:** [First Brands (Wikipedia)](https://en.wikipedia.org/wiki/First_Brands_Group) · [First Brands off-balance-sheet financing (ABF Journal)](https://www.abfjournal.com/middle-market-debt-weekly-first-brands-collapse-spurs-scrutiny-of-off-balance-sheet-financing/) · [Raistone liquidation (Bloomberg Law)](https://news.bloomberglaw.com/bankruptcy-law/trade-financer-raistone-to-liquidate-after-first-brands-blowup) · [SEC Tricolor charges](https://www.newsfilecorp.com/release/310343/SEC-Charges-Former-Executives-With-Fraud-in-Connection-With-1.9-Billion-Collapse-of-Subprime-Auto-Lender-Tricolor) · [Tricolor $800M hole (InvestmentNews)](https://www.investmentnews.com/regulation-legal-compliance/sec-accuses-tricolor-executives-of-hiding-800-million-collateral-hole/267915) · [Tricolor bank losses (Auto Finance News)](https://www.autofinancenews.net/allposts/risk-management/jpmorgan-fifth-third-among-banks-facing-tricolor-losses/) · [HPS/BNP telecom fraud (US News/FT)](https://www.usnews.com/news/top-news/articles/2025-11-17/us-probes-telecom-firms-after-blackrocks-hps-uncovers-alleged-400m-fraud-financial-times-reports) · [Brahmbhatt lessons](https://sunandoroy.org/2026/02/27/when-documents-lie-five-hard-lessons-from-the-brahmbhatt-blackrock-fraud/) · [MFS collapse (Manatt)](https://www.manatt.com/insights/newsletters/client-alert/mfs-collapse-refuels-double-pledging-concerns-within-the-private-credit-industry) · [Stenn (GTR)](https://www.gtreview.com/news/europe/stenn-claims-uk-trade-finance-firm-sold-it-fake-invoices/)
- **Demand and budgets:** [Bloomberg: collateral checkers boom](https://www.bloomberg.com/news/articles/2026-08-21/tricolor-fallout-fuels-boom-in-business-for-collateral-checkers) · [Auto Finance News: same](https://www.autofinancenews.net/allposts/risk-management/tricolor-fallout-fuels-boom-in-business-for-collateral-checkers/) · [Backup servicers' new role](https://www.autofinancenews.net/allposts/risk-management/tricolor-a-year-later-backup-servicers-tasked-with-more-proactive-oversight/) · [Field exam fee clauses (Law Insider)](https://www.lawinsider.com/clause/field-exams) · [ABL rates and fees 2026](https://onlinebusinesslineofcredit.com/asset-based-lending-rates-fees-2026/)
- **Market data:** [FDIC on NDFI lending (PYMNTS)](https://www.pymnts.com/legal/bank-regulation/2026/fdic-finds-steep-rise-in-loans-from-banks-to-nonbanks) · [ABA Banking Journal on NDFIs](https://bankingjournal.aba.com/2026/02/loans-to-non-depository-financial-institutions-new-granularity-and-a-rapidly-growing-segment/) · [SFNet ABL 2025](https://www.sfnet.com/home/industry-data-publications/the-secured-lender/magazine/tsl-article-detail/a-market-of-contrasts--abl-closes-2025-strong-amid-slowing-growth-and-structural-shifts) · [FCI 2025 factoring statistics](https://fci.nl/en/news/fci-releases-2025-world-industry-statistics-global-factoring-market-surpasses-eu4-trillion?language_content_entity=en) · [US ABS 2025 (Diamond Hill)](https://www.diamond-hill.com/insights/a-871/infographics/securitization-in-focus-november-2025/) · [Private credit $1.7T](https://ctacquisitions.com/private-credit-market-2026/) · [ABF growth (GMI)](https://www.gminsights.com/industry-analysis/private-credit-market) · [On-chain private credit 2026](https://paragraph.com/@yielddesk/on-chain-private-credit-in-2026-what-the-headline-numbers-hide-and-where-the-real-yield-lives)
- **Regulation and standards:** [SFA fraud mitigation survey](https://structuredfinance.org/resources/structured-finance-association-releases-fraud-mitigation-survey-findings/) · [FSB private credit report](https://www.fsb.org/2026/05/report-on-vulnerabilities-in-private-credit/) · [NY Article 12 (Morgan Lewis)](https://www.morganlewis.com/pubs/2026/06/new-york-adopts-2022-ucc-amendments-a-new-era-of-digital-assets-electronic-commerce-and-secured-transactions) · ["Control beats filing" (Manatt)](https://www.manatt.com/insights/newsletters/client-alert/control-beats-filing-new-york-adopts-ucc-article-12-to-open-a-new-era-in-digital-asset-lending) · [Article 12 explainer (Cadwalader)](https://www.cadwalader.com/fund-finance-friday/index.php?eid=1736&nid=230)
- **Profitability comparables:** [Figure Q2 2026 (TradingView)](https://www.tradingview.com/news/tradingview:320450e7bcd73:0-figure-technology-solutions-posts-record-q2-revenue-226m-net-income-87-4m-adjusted-ebitda-119-4m/) · [Figure DART adoption](https://www.nationalmortgagenews.com/news/figure-invites-lenders-to-use-its-dart-system-a-competitor-to-mers) · [MERS pricing schedule](https://www.mersinc.org/products-services/mers-system/pricing-schedule) · [FICO 10-Q](https://www.sec.gov/Archives/edgar/data/814547/000081454726000011/fico-20251231.htm) · [Verisk FY2025 10-K](https://www.sec.gov/Archives/edgar/data/1442145/000143774926004452/vrsk20251231_10k.htm) · [TriumphPay EBITDA (FreightWaves)](https://www.freightwaves.com/news/triumphpay-gets-back-to-a-positive-ebitda) · [Triumph Q2 2026 letter](https://www.sec.gov/Archives/edgar/data/0001539638/000153963826000028/tfin-shareholderletterx2q26.htm) · [TriumphPay double-factoring prevention](https://www.freightwaves.com/news/leveraging-network-effects-to-mitigate-risk) · [Chainalysis ARR (Sacra)](https://sacra.com/research/chainalysis-at-190m-arr/) · [TRM Labs $2B (Fortune)](https://fortune.com/2026/09/09/trmlabs-valuation/)
- **Competitors and counter-evidence:** [Setpoint Series A (a16z)](https://techcrunch.com/2022/12/07/andreessen-horowitz-leads-43m-series-a-for-setpoint-which-aims-to-be-the-stripe-for-credit) · [Setpoint Series B](https://therealdeal.com/texas/austin/2024/08/15/setpoint-raises-31m-led-by-andreesen-horowitz/) · [MonetaGo leaves blockchain (Finextra)](https://www.finextra.com/newsarticle/40421/monetago-pivots-from-blockchain-to-cloud-citing-scalability-issues) · [MonetaGo × FCI](https://www.gtreview.com/news/fintech/fci-partners-with-monetago-to-fight-invoice-finance-fraud/) · [MonetaGo profile (Tracxn)](https://tracxn.com/d/companies/monetago/__ovXclDP76YxzYk_u-bgBOh56SkTFiKDBf27J61_KdsI) · [Cardo AI](https://www.abfjournal.com/cardo-ai-launches-cash-flow-modeling-tool-for-asset-based-finance/) · [Inscribe 2026 document fraud report](https://www.inscribe.ai/reports/2026-document-fraud-report)
- **Rejected ideas:** [Prediction market Sept volume](https://www.publish0x.com/prediction-graph/prediction-markets-hit-a-record-dollar-782b-in-september-q3-xnjpnem) · [WSJ on UMA judges](https://www.securitiesdocket.com/2026/05/18/the-mysterious-crypto-judges-who-settle-polymarket-disputes-wsj/) · [Polymarket Protocol V2](https://cryptotimes.io/2026/10/06/polymarket-introduces-protocol-v2-with-new-market-infrastructure) · [Chaos Labs profile (Tracxn)](https://tracxn.com/d/companies/chaoslabs/__mrpDlbic32YrAHAdZ8GgDWBXjS2w5nlDpj-z6fgrE8A) · [Kalshi fees and revenue](https://www.mexc.com/learn/article/kalshi-vs-polymarket-2026-same-1-75-fee-cap-different-everything-else-and-which-one-can-you-actually-use-/1) · [Tazapay Series B](https://www.crowdfundinsider.com/2026/03/269700-singapores-tazapay-secures-36m-series-b-to-expand-cross-border-payment-rails/) · [1099-DA CP2000 wave](https://www.cryptotaxaudit.com/blog/cp2000-avalanche-crypto-investors-irs-notices-1099-da) · [x402 adoption (Chainalysis)](https://www.chainalysis.com/blog/x402-agentic-payments-adoption/) · [Agentic commerce status](https://www.digitalapplied.com/blog/agentic-commerce-standards-ucp-acp-ap2-2026-merchant-guide) · [Crypto hacks 2026 (CryptoSlate)](https://cryptoslate.com/crypto-hackers-have-taken-2-7-billion-in-2026-and-the-losses-are-alarmingly-concentrated/) · [DePIN revenue](https://www.rzlt.io/blog/7-depin-projects-generating-10m-revenue-(and-what-you-can-learn-from-them)) · [Crypto VC H1 2026 (CoinGecko)](https://www.coingecko.com/learn/crypto-vc-funding-h1-2026)
- **Theses:** [Sequoia, Services: The New Software (summary)](https://rogerwong.me/2026/03/services-new-software) · [a16z crypto 2026 big ideas](https://a16zcrypto.com/posts/article/big-ideas-things-excited-about-crypto-2026/) · base rates and unicorn patterns carried over from the Recourse thesis.
