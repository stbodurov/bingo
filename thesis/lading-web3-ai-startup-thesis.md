# Startup Thesis: Web3 × AI, Venture-Scale Idea Selection

Oct 6, 2026 · working name **Lading**

## The pick: Lading

**Build Lading, verified settlement for cross-border trade.** The importer's payment is locked in a stablecoin escrow contract. It is released to the supplier only when AI has checked the shipment evidence: purchase order, commercial invoice, packing list, bill of lading, carrier tracking, inspection report and certificate of origin, all agreeing with each other and with independent data.

It is the letter of credit rebuilt for the 80% of world trade that runs on open account today with no protection at all. The lock costs cents instead of 1.2% to 3%, and the document check takes minutes instead of days.

It sells to both sides:

- **Importers (buyers):** stop wiring 30% deposits to suppliers you cannot verify. You pay when the goods are proven shipped, are what you ordered, and come from where the supplier says, so a transshipment does not cost you a 40% penalty tariff.
- **Exporters (suppliers):** see the buyer's money locked on-chain before you start production. You get paid the day the container sails instead of 30 to 90 days after arrival, and you can finance the locked receivable at near-risk-free rates.

**The wedge:** US SME importers buying from China and Vietnam ($503B of 2025 imports). Many already pay by wire with deposits, and since August 2025 they also carry origin risk under the 40% transshipment penalty.

**The durable business:** every verified shipment builds a portable on-chain trade record for the supplier. Credit capital prices against that record, and every new buyer in the network can trust it.

**Why it can be a unicorn:** it is a two-sided network. Each buyer who demands Lading pulls their suppliers in, and each verified supplier makes Lading the default for that supplier's other buyers. Revenue scales with trade volume (GMV), not with seats.

---

## Is it unique, and can it be a unicorn?

As "stablecoin payments for importers", no. That market is consolidated: Stripe bought Bridge for $1.1B and Mastercard bought BVNK for $1.8B. As **the condition layer above the rails**, yes. Nobody owns the proof that releases the money.

| Unicorn trait | What the research says | "Stablecoin rails for trade" | **Lading (verified settlement)** |
|---|---|---|---|
| Timing | Timing explains 42% of the gap between successes and failures (Bill Gross, Idealab) | Late: rails consolidated in 2025–26 | **Strong:** GENIUS Act law, B2B stablecoin flows up 733% YoY, 40% transshipment penalty, HK and Japan trade-settlement pilots in 2026 |
| Non-consensus insight | Winners are "non-consensus and right" (Mike Maples) | Consensus: everyone is building it | **Strong:** "blockchain trade finance is dead" is the consensus after we.trade, Marco Polo, Contour and TradeLens failed. They failed for reasons that no longer hold (see below) |
| Market ceiling | $1B needs ~$100M+ revenue | Thin margins (FX spread) | ~$800M US wedge revenue pool; $13B–$20B global pool across SME cross-border prepayment trade (estimates) |
| Compounding moat | Network effects created 70% of tech value since 1994 (NFX) | None; switching costs are low | Two-sided buyer↔supplier network, plus the supplier's on-chain performance record, which credit capital prices against |
| Speed | Top AI companies reached $1M annualized revenue in a median 11.5 months (Stripe) | Fast but commoditized | Medium: revenue per transaction from the first order; GMV ramps with order cycles |
| Competition | — | Bridge/Stripe, BVNK/Mastercard, Rain, Conduit, OpenFX ($94M), Airwallex | Banks (LCs) are too expensive for SMEs; Alibaba Trade Assurance works only inside Alibaba; WSPN/TradeGo serve commodity traders and need eBLs. **Nobody does conditional settlement for open-account SME trade** |

**The non-consensus insight.** Stablecoins made moving money nearly free, so the money movement is no longer the scarce part of trade. The scarce part was always the *condition*: "pay only if the goods really shipped." The letter of credit was the original programmable money. It failed SMEs because a human had to check every document, and 60–75% of first presentations are refused for discrepancies, almost all clerical (ICC Banking Commission). AI makes the check almost free, and a smart contract makes the lock almost free. Together they rebuild the LC for the 80% of trade that runs on open account.

**Why the 2017–2022 blockchain trade projects failed, and why this one does not depend on what killed them:**

| Why they died | What changed |
|---|---|
| Built *for banks by bank consortia*; corporates never came (Marco Polo, we.trade) | Lading sells to the importer CFO and the supplier. No bank consortium is needed because the stablecoin is the settlement asset |
| Needed digitized title documents; eBLs are still only ~2% of bills of lading | AI reads the PDFs and photos people already exchange, then cross-checks them against carrier tracking and inspection data. eBLs are a bonus, not a prerequisite |
| Needed every participant on one permissioned ledger (TradeLens) | Public chains plus USDC/USDT, which suppliers in Yiwu, Shenzhen and Ho Chi Minh City already accept |
| No money moved on-chain, so it was just data sharing | The money itself sits in the contract |

**Odds.** About 1.3% (CB Insights 2008–14 cohorts) to 2.5% (AngelList 2013–19) of seed-funded companies become unicorns. My judgment: about 5–6% for Lading if the supplier side adopts. These are the same base rates used in the Recourse thesis.

**What must be true by month 9:**

- $50M+ cumulative escrowed GMV.
- At least 30% of new importers arrive through a supplier invitation (proof that the network is two-sided).
- Median release within 24h of the vessel departing, with an auto-release accuracy of at least 98% against human review.
- At least one credit partner financing escrow-backed supplier receivables below 10% APR.

---

## How it was chosen

I scored nine web3 × AI ideas, using the same filters as the Recourse thesis:

- **Sequoia's "Services: The New Software"** (Julien Bek, Mar 5, 2026): sell the work, not the tool. The work should be mostly intelligence rather than judgment, and already outsourced. Trade document checking is exactly that work, and banks already charge for it.
- **a16z crypto's 2026 big ideas:** stablecoins as the infrastructure of internet finance, better on/off ramps, and identity and payments for agents.
- **The exit path is proven:** Stripe bought Bridge for $1.1B (2025), Mastercard bought BVNK for $1.8B (Aug 2026), and Tempo raised $500M at $5B with Sequoia participating.

| Idea | Money pool | Who already serves it | Why now | Verdict |
|---|---|---|---|---|
| **Verified settlement for SME cross-border trade** | $503B US imports from CN+VN alone; 80% of world trade on open account | Banks (LCs, 1.2–3%), Alibaba (in-marketplace only), WSPN/TradeGo (commodities, eBL-only) | GENIUS Act; B2B stablecoin $226B/yr; 40% transshipment penalty; LLM doc checking | **Pick** |
| Chargebacks/disputes for agent stablecoin payments | x402: ~$53M *cumulative* volume, about half test or bot traffic | Circle Refund Protocol (open source), card networks, Amex Agent Purchase Protection | Agentic commerce | Too early: the volume is not there yet |
| KYA (Know Your Agent) identity | Unclear | Catena Labs ($30M a16z), Skyfire, Visa/Mastercard/Cloudflare agent protocols | a16z 2026 big idea | Crowded with incumbents and a16z's own portfolio |
| Prediction market integrity and surveillance | ~20 venues | Solidus Labs (Kalshi partner), Polysights | Congressional insider-trading probe, May 2026 | Low ceiling, first customer already taken |
| AI smart contract auditing | Audit budgets | Octane, Almanax, Sherlock, CertiK, OpenZeppelin | LLM code reasoning | Price war |
| GENIUS Act issuer compliance | 152 issuers | Chainalysis, TRM, Paxos and Anchorage white-label | Rules due Nov 2026 onward | Few buyers, rules not final |
| Stablecoin B2B rails / treasury | $226B/yr B2B flows | Bridge (Stripe), BVNK (Mastercard), Rain, Conduit, OpenFX, Airwallex | — | Consolidated |
| Tokenized RWA servicing / fund admin | $26B–$46B on-chain RWA | Issuer-integrated stacks (BlackRock BUIDL, Circle USYC), Securitize | Tokenized Treasuries | Slow, incumbent-led |
| On-chain tax (1099-DA) reconciliation | Retail and pro traders | CoinTracker, Koinly, TaxBit | 1099-DA reporting begins | Crowded consumer market |

**Founder fit.** This idea scores highest for an engineer from supply-chain compliance software, and it reuses the Recourse rulebook. Recourse's dated tariff rulebook (authority × origin × HTS × date) becomes Lading's origin and duty check at release time. The importer gets a payment that only releases when the origin evidence would survive a CBP transshipment inquiry.

---

## Problem and customer

**The importer's problem.** A $20M-a-year US importer buying from 30 suppliers in China and Vietnam usually pays 30% by wire at order and 70% before or at shipment. That leaves them exposed in three ways:

1. **Fraud and non-performance.** Deposit-and-vanish scams and changed-bank-detail emails are standard. Business email compromise alone cost US businesses $3.04B in 2025 across 24,768 complaints, about $123k per incident, and 86% of it moved by wire or ACH (FBI IC3 2025).
2. **Wrong goods.** Quality and quantity disputes surface after arrival, when the 70% is already gone.
3. **Origin liability.** Since Aug 7, 2025, goods CBP finds transshipped to evade tariffs carry an extra 40% tariff with no mitigation, plus §1592 penalties and False Claims Act exposure. The penalty was reissued under Section 122 on Feb 24, 2026, and CBP raided China-linked Vietnam factories in July 2026. *"The documentation that defends an origin claim has to exist before the shipment sails."*

**Their alternative is a letter of credit:** 1.2–3% all-in on a $50k LC, plus amendment fees and days of document back-and-forth. 60–75% of LCs are refused on first presentation. SMEs mostly skip LCs and take the risk.

**The supplier's problem.** They do not trust new buyers either, which is why they demand deposits. Under open account they wait 30–90 days after shipment to be paid and fund production themselves. The global trade finance gap is **$2.5T**, about 10% of world trade, and it is concentrated in SMEs (ADB 2025).

**Customer:** the owner, CFO or controller of a US importer with $5M–$500M of annual imports, mostly consumer goods, home and furniture, electronics accessories, apparel and light industrial. SMBs are 96–97% of US importers by count and 32% of import value, and 40% of the import value from China (Census).

---

## Why now

| Date | Event | Why it matters |
|---|---|---|
| Jul 18, 2025 | GENIUS Act signed; implementing rules proposed Dec 2025–Jun 2026, OCC final rule targeted Nov 2026 | USDC is a regulated payment instrument that a CFO can hold |
| Aug 7, 2025 | 40% transshipment penalty tariff | Origin proof becomes a payment condition, not a paperwork afterthought |
| Aug 1, 2025 → Apr 2026 | HK Stablecoins Ordinance takes effect; first issuer licenses (HSBC, Anchorpoint) | A regulated off-ramp on the exporter side of the China corridor |
| Jan 14, 2026 | WSPN + TradeGo run the first eBL-collateralized stablecoin trade finance | Proves the mechanics with commodity traders; SMEs remain unserved |
| Feb 24, 2026 | Transshipment penalty moves to Section 122 | The penalty survives the IEEPA ruling |
| Mar 17, 2026 | Mastercard buys BVNK for $1.8B (closed Aug 3) | Rails are commoditized; the open layer is the conditions above them |
| Sep 2026 | Japan FSA trade-settlement stablecoin pilot (TradeWaltz with MUFG, SMBC, Mizuho) begins | Incumbents validate the category, starting with large exporters and banks |
| Sep 16, 2026 | HK Policy Address: trade-finance stablecoin pilots with the mainland before year end | The China corridor opens for regulated settlement |
| Oct 1, 2026 | Stablecoin supply $300.9B; Visa reports B2B as the most cross-border category of business stablecoin flows (43%) | Suppliers increasingly hold a stablecoin wallet already |

Three slower shifts make this a company rather than a feature:

- **B2B stablecoin payments are real and growing:** about $226B in 2025, up 733% YoY, and over 60% of genuine stablecoin payments (McKinsey × Artemis).
- **Reading every trade document is now nearly free.** LLM inference cost for constant capability falls about 10× a year (a16z, "LLMflation"). The 60–75% discrepancy problem is clerical, which is exactly what models do well.
- **On-chain credit wants short-dated, real-economy collateral.** Huma has processed $8.8B–$12B of PayFi volume, and Obligate/TradeFlow's oTFY puts $200M of trade receivables into Solana DeFi as collateral. That capital needs better-underwritten trade assets.

---

## Product and value proposition

> **Lading is the only way to pay an overseas supplier where the money is provably locked before production and provably released only against verified shipment, at under 1% instead of 1.2–3%.**

### Wedge product: Verified Order

1. **Lock.** The importer's PO becomes an escrow contract. The importer funds it in USDC from a bank wire through a licensed ramp partner (Bridge, BVNK or Circle Mint), so they never have to handle crypto. Milestones mirror today's terms, for example 30% at production start (released on an inspection photo or report) and 70% on shipment evidence.
2. **Show the supplier.** The supplier gets a link that proves the funds are locked, cannot be withdrawn by the buyer during the window, and will release on stated conditions. In practice this is an LC confirmation in seconds, for cents.
3. **Verify (the AI part).** Agents reconcile the PO, commercial invoice, packing list and B/L against each other: quantities, HS codes, consignee, ports and dates. They then cross-check independent sources: carrier container milestones, AIS vessel position, the inspection agency report (QIMA, SGS, V-Trust) and the certificate of origin. Ambiguous cases go to a human trade-ops reviewer. Every release ships with an evidence pack.
4. **Origin gate.** Using the Recourse rulebook, it scores transshipment risk from the supplier's factory location, bill-of-materials signals, routing and price-per-unit against peers. High-risk shipments hold for an origin dossier before release.
5. **Release or dispute.** A clean evidence pack releases funds automatically; the supplier is paid out in USDC or USD/CNH through a ramp partner. A dispute goes to a bounded arbiter (the Circle Refund Protocol pattern): the arbiter can split or refund, but can never take custody.

### Supplier side: Verified Supplier

- **A trade record:** every completed, verified shipment becomes a signed attestation tied to the supplier's identity (EAS-style, privacy-preserving: amounts and buyers hidden, performance provable). Buyers on Lading see "142 verified shipments, 0.7% dispute rate" before ordering.
- **Get paid at shipment or at production start.** The locked escrow is a receivable whose credit risk is already removed, because the buyer's money is in the contract. Only performance risk remains, and the AI verification measures it. On-chain or private credit partners finance it at a fraction of factoring cost.
- **Invite your buyers.** Suppliers want faster payment and fewer deposit fights, so they bring buyers onto Lading. That is the supplier-led half of the network.

### Why web3 (the question every partner will ask)

1. **Neutral, non-custodial lock.** Without a chain you need a licensed escrow agent trusted in both countries, which is why Alibaba only offers Trade Assurance inside its own marketplace. A public contract is trusted by both sides without either side's bank.
2. **Verifiable by the supplier in real time.** This is the "confirmation" banks charge 0.1–0.5% for, and it is free on-chain.
3. **Composable receivables.** The escrow claim is natively financeable by on-chain credit.
4. **A portable record.** The supplier owns their trade history across buyers and platforms.

### Why AI

- The verification is the product, and it is document reasoning across 6–10 heterogeneous files per shipment, in English and Chinese, often scanned.
- It is sold as work ("we check every shipment"), not as a tool. That is Sequoia's services-as-software pattern.
- Each dispute outcome trains the release model, so model improvements accrue to Lading's margin.

### Expansion path

- **Corridors:** Mexico and LatAm buyers paying China (USDT is already common there), then African importers paying China (Sub-Saharan Africa processed $50B of stablecoin transactions in Q1 2026).
- **Agentic procurement:** AI purchasing agents need a safe way to pay unknown suppliers. A Verified Order is the "KYA-safe" payment primitive for physical goods.
- **Trade credit marketplace:** the supplier record becomes underwriting data. This links to the $2.5T trade finance gap.

---

## Market size

The wedge is ~$120B of annual US SME prepayment GMV, about an $800M revenue pool. The global pool is $13B–$20B a year.

| Layer | Calculation | Size |
|---|---|---|
| US imports from China + Vietnam (2025) | $308.7B + $193.9B (Census) | $503B |
| SME share | ~35% (SMEs hold 40% of China import value, 32% overall) | ~$175B |
| Paid by deposit/TT without an LC (wedge GMV) | ~70% (estimate; open account plus prepayment ≈ 80% of world trade) | **~$120B/yr** |
| US wedge revenue pool | × 0.65% blended take | **~$800M/yr** |
| Global SME cross-border prepayment trade | $2T–$3T (estimate; 80% of world trade is open account, $2.5T financing gap) | |
| Global revenue pool | × 0.65% | **$13B–$20B/yr** |

$100M of revenue needs about $15B of GMV, which is about 12% of the US wedge or under 1% of the global pool. Comparables for cross-border SMB payments at scale: Airwallex reached $1.3B annualized revenue in Mar 2026 at an $11B valuation, and Payoneer crossed $1.05B of revenue in 2025. The SME importer and exporter segment supports $1B+ revenue companies; Lading charges for verification, which is higher-margin than FX.

*Segment shares, the prepayment share and the take rate are assumptions to test with the first 20 importers.*

---

## Business model and unit economics

Three revenue lines:

1. **Verification fee: 0.5% of escrowed GMV**, paid by the importer. Benchmark: LCs at 1.2–3% all-in.
2. **Supplier early-payment spread: ~0.3% of financed GMV.** The supplier pays about 1% per 30 days advanced, the credit partner keeps most of it, and Lading keeps the origination spread. Assumes 50% uptake, so ~0.15% of total GMV.
3. **Origin dossier / compliance tier:** $500–$2,000 per month for importers with high transshipment exposure. This is upside and is not counted below.

**Per customer (illustrative $20M-imports US importer):**

| | Assumption | Year 1 | Year 2+ |
|---|---|---|---|
| GMV through Lading | 50% of imports year 1, 80% after | $10M | $16M |
| Verification fee | 0.5% | $50k | $80k |
| Supplier financing spread | 0.15% of GMV | $15k | $24k |
| **Lading revenue** | | **$65k** | **$104k** |
| Value to the importer | One avoided bad deposit (median BEC loss ~$123k) plus LC fees avoided (1.2–3%) | >$120k | |

- **Gross margin target 65–75%.** Costs: ramp fees (0.05–0.15%, partly passed through), inference (cents per shipment), human review on 10–20% of releases (falling with the dispute data), arbiter operations. Ramp fees are the main drag, which is why margin sits below the SaaS target.
- **CAC ~$12k (assumption), payback ~2–3 months.** Channels with low CAC: customs brokers and freight forwarders on a revenue share (they already hold the importer relationship and the documents), and supplier invitations. The median private B2B SaaS payback is 16 months (Aleph × Benchmarkit).
- **Credit risk stays off the balance sheet.** Lading originates and verifies; partners (on-chain pools, trade-finance funds such as TradeFlow) carry the capital.
- **Structural risk:** GMV is lumpy with order cycles and seasons (Q3 peak). Mitigate with a broad customer base and the compliance tier as recurring revenue.

---

## Competition and moat

Rails are crowded, and the conditional-settlement layer for SME trade is empty.

| Alternative | How it works | Gap Lading closes |
|---|---|---|
| Bank letters of credit | Banks check paper; 1.2–3%; 60–75% first-presentation refusals | 10× cheaper, minutes not days, works for $20k orders |
| Wire / TT with deposit | No protection | Buyer protection plus supplier certainty |
| Alibaba Trade Assurance (160M+ orders protected) | Custodial escrow inside Alibaba only | Works for any supplier, off-marketplace, and the record is portable |
| Stablecoin rails (Bridge, BVNK, Rain, Conduit, OpenFX, Airwallex) | Move money fast | They move money; none of them knows *whether it should move* |
| WSPN/TradeGo, Japan megabank pilot, Obligate/TradeFlow | Commodity traders; eBL-dependent; capital side | SME manufactured goods, without needing eBLs; Lading can be a deal source for this capital |
| AI doc-checkers (Cleareye.ai, Infrrd, V7) | Sell software to bank trade desks | Lading sells the outcome to the corporate, and owns the money flow |
| Escrow.com-type agents | Manual, custodial, single-country | Programmable, non-custodial, cross-border |

**What compounds:**

1. **Two-sided network.** Each verified supplier lowers the cost of trust for every future buyer; each buyer brings their supplier list.
2. **The supplier record.** On-chain attestations become the credit file for SMEs nobody underwrites today. Credit partners price against it, which is a data network effect.
3. **Dispute outcome data.** Which evidence patterns predict a clean delivery sets the auto-release thresholds and lowers human review hours each quarter.
4. **Origin rulebook** (from Recourse). It is dated, multi-authority and hard to rebuild, and each tariff regime change makes it more valuable.

**Likely entrants:** Airwallex or Payoneer adding escrow, or Stripe/Bridge adding conditional transfers. The defense is to own the supplier network and the verification data before they arrive. They are likelier acquirers than builders, as Bridge and BVNK showed.

---

## Go-to-market

**Sell order by order.** Every new PO is a moment when the buyer is about to wire a deposit to someone they cannot fully verify.

- **First 10 importers: through customs brokers and forwarders.** Brokers see every entry and already hold the documents. Offer them 15% of fees. Run concierge orders: Lading staff plus AI verify, and the importer wires to the ramp.
- **First 100: supplier-led growth.** Every supplier onboarded is asked to invite their other US buyers, with "get paid at shipment" as the hook. Track the supplier-invited share as the network KPI.
- **First 1,000: self-serve, plus integrations.** Import POs from NetSuite, QuickBooks and Zoho Inventory. List in freight platforms (Flexport-style marketplaces). Run a free "Supplier Check" (origin risk plus trade record lookup) as the top of the funnel.
- **Tone with suppliers.** Do not frame it as distrust. Frame it as "your buyer's money is already locked; ship and get paid the same day." Suppliers can be paid in fiat through ramp partners and never need to touch crypto.
- **Financing as a closer.** Credit partners advance against locked escrows, so suppliers accept terms they would normally refuse, such as a lower deposit.

---

## Success likelihood and risks

**My estimate:** a 25–30% chance of raising a Series A within 24 months, versus a 15–20% base rate (Carta, via ValueAdd VC). The case is above base rate because transaction revenue starts with the first order and the category has strategic acquirers.

| Outcome | Estimate | Reasoning |
|---|---|---|
| $1M revenue run-rate within 12 months | 35% | Needs ~$150M annualized GMV; order cycles slow the ramp more than contingency revenue would |
| Series A within 24 months | 25–30% | Needs GMV growth of 15%+ month over month and visible supplier-side pull |
| $100M revenue (venture scale) | ~5–6% | Requires the multi-corridor network plus credit marketplace |

*These are judgment calls, not model outputs.*

| Risk | Mitigation |
|---|---|
| **Money transmission / licensing.** Holding buyer funds can make Lading a money transmitter | Non-custodial contract with a bounded arbiter (Circle Refund Protocol pattern; FinCEN's 2019 CVC guidance treats non-custodial software differently). Fiat legs go through licensed ramp partners. Get regulatory counsel before launch |
| **China capital controls / PBoC stance on stablecoins** | Suppliers can be paid in fiat through HK/SG ramp partners; the HK licensed-issuer path; spread corridors across Vietnam, India and Mexico |
| **AML: stablecoin flows disguised as trade** (FinCEN has flagged stablecoin transfers that look like ordinary commercial activity) | Lading's evidence pack *is* the proof of real goods, which makes it the cleanest stablecoin trade flow a bank or issuer can see. KYB on both sides, plus sanctions screening at lock and release |
| **Suppliers refuse escrow** | Lead with the supplier benefit (paid at shipment, financing); start with suppliers who already have Western buyers; fall back to milestone splits |
| **AI wrongly releases funds** | Human review below a confidence threshold, a dispute window before final release, and an insurance partner for verified-release errors |
| **An incumbent (Airwallex, Stripe/Bridge) copies it** | Win the supplier network and verification data first; be the acquirable layer |
| **Fraudulent documents** | Never trust documents alone; require independent carrier, AIS and inspection signals for release |

**Kill criteria at day 120:**

- Fewer than 10 importers running live orders, or
- suppliers decline escrow in more than 50% of asks, or
- auto-release accuracy below 95% against human review.

---

## 18-month plan and fundraising

**Raise a $4–5M seed after Gate 1. Raise a $25–35M Series A at about a $5M revenue run-rate, which is about $750M of annualized GMV.**

| Phase | Months | Build | Gate |
|---|---|---|---|
| 1. Concierge | 0–4 | Escrow contract (audited), ramp integration, AI doc reconciliation with human-in-the-loop, 10 broker-sourced importers | $5M GMV, 0 loss events, release accuracy ≥95% |
| 2. Two-sided | 5–10 | Supplier app and attestations, early-payment financing with one credit partner, origin gate (Recourse rulebook) | $50M cumulative GMV, ≥30% supplier-invited importers, a financing partner live |
| 3. Scale | 11–18 | Self-serve, ERP integrations, second corridor (Mexico/LatAm↔China), agentic-procurement API | $5M revenue run-rate, ≥98% auto-release accuracy |

**What a16z or Sequoia will need at the A:**

- GMV growing 15%+ month over month, with blended take holding at 0.6% or above.
- **Network proof:** supplier-invited share of new importers, and the share of new orders where the supplier already has a Lading record.
- Human review hours per shipment falling quarter over quarter; gross margin ≥65%.
- Zero principal losses on financed receivables, and a dispute rate below 2%.
- Net GMV retention above 130% as importers move more suppliers onto Lading.

**Founding team:**

- An engineer from supply-chain compliance software (rules engines, supplier data, audit trails; the Recourse builder).
- A smart-contract and stablecoin infrastructure engineer.
- A trade-operations lead from a bank trade desk or a freight forwarder.
- An advisor who is a licensed customs broker.

---

## Sources

All found via search, Oct 1–6, 2026. Some primary pages (a16z crypto, Visa, FinCEN) could not be opened directly from this environment, so figures for them come from search summaries and coverage.

- **Investor theses:** [a16z crypto, 17 things for 2026](https://a16zcrypto.com/posts/article/big-ideas-things-excited-about-crypto-2026/) · [a16z Big Ideas 2026 Part 3](https://a16z.com/newsletter/big-ideas-2026-part-3/) · [Sequoia, Services: The New Software (summary)](https://rogerwong.me/2026/03/services-new-software) · [Sequoia portfolio 2026 (ValueAdd VC)](https://valueaddvc.com/blog/sequoia-capital-portfolio-2026-biggest-bets-biggest-wins-and-what-they-re-investing-in-now)
- **Stablecoin market:** [Stablecoin market $300.9B (Crypto Times)](https://www.cryptotimes.io/2026/10/02/stablecoin-market-reaches-300-9b-as-tron-gains-5b-in-q3/) · [McKinsey × Artemis on real payments (KuCoin)](https://www.kucoin.com/news/flash/mckinsey-and-artemis-report-only-1-of-stablecoin-35t-volume-is-real-payments) · [Visa business stablecoin data](https://www.financialcontent.com/article/bizwire-2026-10-1-visa-data-shows-stablecoins-gaining-traction-in-business-payments) · [Yiwu USDT (PANews)](https://www.panewslab.com/en/articles/2c2de82c-2e8b-415f-9d7b-258a5927f2e3)
- **Rails consolidation:** [Mastercard–BVNK (CNBC)](https://www.cnbc.com/2026/03/17/mastercard-acquiring-stablecoin-startup-bvnk-in-crypto-bet.html) · [Mastercard IR](https://investor.mastercard.com/investor-news/investor-news-details/2026/Mastercard-to-Acquire-BVNK-to-Connect-On-Chain-Payments-and-Fiat-Rails/default.aspx) · [Stripe–Bridge (CNBC)](https://www.cnbc.com/2025/02/04/stripe-closes-1point1-billion-bridge-deal-prepares-for-stablecoin-push-.html) · [OpenFX $94M (CoinDesk)](https://www.coindesk.com/business/2026/03/31/forex-startup-openfx-raises-usd94-million-to-expand-stablecoin-powered-cross-border-payments) · [Airwallex Series H (CNBC)](https://www.cnbc.com/2026/06/26/airwallex-series-h-funding-11-billion-valuation-ai-finance.html) · [Payoneer FY2025](https://investor.payoneer.com/news-releases/news-release-details/payoneer-reports-fourth-quarter-and-full-year-2025-financial)
- **Regulation:** [OCC GENIUS proposal (Mayer Brown)](https://www.mayerbrown.com/en/insights/publications/2026/03/occ-proposes-comprehensive-rulemaking-to-implement-the-genius-act) · [GENIUS AML rules (HK Law)](https://www.hklaw.com/en/insights/publications/2026/04/fincen-and-ofac-propose-aml-sanctions-rules-for-stablecoin-issuers) · [CLARITY Act cloture fails (Troutman)](https://www.troutman.com/insights/in-the-wake-of-clarity-acts-failure-agencies-move-forward-without-congressional-action-or-certainty/) · [Hong Kong stablecoins (CNBC)](https://www.cnbc.com/2026/02/11/hong-kong-stablecoin-licenses-china-crypto-ban-usdt-usdc-cny-cnh-rmb-usd-fiat-currency.html) · [HK Policy Address](https://stablecoininsider.org/hong-kong-policy-address-stablecoin-trading/)
- **Trade finance:** [ADB Trade Finance Gap Survey 2025](https://www.adb.org/publications/adb-global-trade-finance-gap-survey) · [LC discrepancy rates](https://seasaw.ai/resources/lc-document-discrepancies/) · [LC costs](https://importivity.com/blog/letters-of-credit-trade-finance/) · [Open account share (Trade Finance Global)](https://www.tradefinanceglobal.com/letters-of-credit/) · [eBL adoption 2026](https://cleareye.ai/electronic-bill-of-lading-news-2026-adoption-standards/) · [Why Contour failed (TFG)](https://www.tradefinanceglobal.com/posts/contour-collapses-what-does-this-mean-for-digital-trade-finance/) · [Marco Polo liquidation (GTR)](https://www.gtreview.com/news/top-stories/marco-polo-brings-in-liquidators-as-funds-run-dry/) · [WSPN eBL stablecoin trade finance](https://www.prnewswire.com/news-releases/wspn-enables-stablecoin-based-trade-finance-using-digital-bills-of-lading-302661064.html) · [Japan FSA trade settlement pilot](https://www.digitaltoday.co.kr/en/view/109017/japan-fsa-approves-trade-settlement-stablecoin-pilot-with-three-megabanks) · [Obligate oTFY](https://solanacompass.com/news/obligate-brings-200m-trade-finance-rwa-token-otfy-to-solana-as-defi-lending-collateral) · [Huma PayFi volume](https://www.bitget.com/news/detail/12560605184034) · [Vixtra / Brazil importers](https://noticias.mobilemoneylatam.com/stablecoins-to-finance-the-import-sector/)
- **Importer risk:** [FBI IC3 2025 BEC](https://spycloud.com/blog/fbi-internet-crime-report-2025/) · [40% transshipment penalty (Snell & Wilmer)](https://www.swlaw.com/publication/new-reciprocal-tariff-rates-announced-but-the-real-risk-is-hidden-transshipment-enforcement-now-comes-with-an-additional-40-tariff/) · [Section 122 reissue (Tradlinx)](https://blogs.tradlinx.com/cbp-cracks-down-on-misdeclared-origin-40-tariff-penalty-explained/) · [CBP Vietnam raids](https://www.techtimes.com/articles/322039/20260729/cbp-raids-china-linked-vietnam-factories-over-transshipment-40-penalty-looms.htm) · [Alibaba Trade Assurance](https://tradeassurance.alibaba.com/)
- **US import data:** [Census 2025 China](https://www.census.gov/foreign-trade/balance/c5700.html) · [Census 2025 Vietnam](https://www.census.gov/foreign-trade/balance/c5520.html) · [Census importer profile](https://www.census.gov/foreign-trade/Press-Release/edb/2024prelimprofile.pdf)
- **Rejected ideas:** [x402 adoption (Chainalysis)](https://www.chainalysis.com/blog/x402-agentic-payments-adoption/) · [Circle Refund Protocol](https://www.circle.com/blog/refund-protocol-non-custodial-dispute-resolution-for-stablecoin-payments) · [Agent chargeback liability (Chargeflow)](https://www.chargeflow.io/blog/ai-agent-chargeback-liability) · [Catena Labs Series A (Fortune)](https://fortune.com/2026/05/20/catena-labs-series-a-sean-neville-ai-native-bank/) · [Kalshi × Solidus](https://www.businesswire.com/news/home/20260205995596/en/Kalshi-Partners-with-Solidus-Labs-to-Power-Next-Generation-Trade-Surveillance) · [House insider-trading probe (CNBC)](https://www.cnbc.com/2026/05/22/kalshi-polymarket-comer-insider-trading-probe-congress.html) · [Octane seed (SecurityWeek)](https://www.securityweek.com/octane-raises-6-75m-for-smart-contract-security-tech/) · [RWA market (PYMNTS)](https://www.pymnts.com/blockchain/2026/tokenized-real-world-asset-value-jumps-fourfold-to-26-billion/)
- **Base rates and unicorn patterns:** carried over from the Recourse thesis (CB Insights, AngelList, Carta via ValueAdd VC, Idealab/Bill Gross, Mike Maples, NFX, Stripe).
