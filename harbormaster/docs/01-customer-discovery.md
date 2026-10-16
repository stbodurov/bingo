# Harbormaster — customer discovery

> **Method.** Harbormaster's research starts from the customer's job and the dollars
> attached to it, not from the technology. We (1) read what importers, brokers, CPAs and
> trade lawyers published about the 2025–26 tariff regime, (2) coded every recurring
> complaint into jobs-to-be-done, (3) sized each job in dollars for a reference importer,
> and (4) only then decided what to build. Personas below are **composites** synthesised
> from public sources (law-firm alerts, CPA guides, broker FAQs, vendor case studies,
> trade press) — not interviews with named people. Quotes are paraphrases of themes,
> marked as such. The interview guide at the end is what design-partner calls use.

## 1. What changed for importers (Feb 2025 → Oct 2026)

| Date | Event | What the importer had to do |
|---|---|---|
| Feb–Mar 2025 | IEEPA "fentanyl" duties on China (10→20%), Canada and Mexico (25%) | Re-price China goods; prove USMCA qualification to escape CA/MX duty |
| Apr 2025 | IEEPA reciprocal 10% baseline; China to 125% for a month | Decide whether to land goods at all; track in-transit exemptions |
| Jun 2025 | 232 steel/aluminum to 50%; derivatives on metal content | Report metal content per line or pay on full value |
| Aug 2025 | Country-specific reciprocal rates (EU/JP 15% all-in, VN 20%, IN 25+25%…), 407 derivative codes added | Rebuild landed cost for every SKU and origin |
| Oct–Nov 2025 | 232 on lumber, upholstered furniture, cabinets; trucks; ag exemptions | New stacking (232 excludes reciprocal) |
| **Feb 20 2026** | **Supreme Court: IEEPA does not authorise tariffs** | Claim ~$166B of refunds — but CBP decides how |
| Feb 24 2026 | 10% Section 122 surcharge for 150 days | New layer, new exemptions |
| Mar 4 2026 | CIT: every IOR who paid IEEPA is entitled to refunds | — |
| Apr 6 2026 | 232 metals on full value; 50/25/15% classes | Re-price metal-containing goods again |
| Apr 20 2026 | **CAPE Phase 1** (unliquidated + ≤80 days post-liquidation) | File CAPE declarations in ACE |
| May 7 2026 | CIT holds s122 unlawful, relief for plaintiffs only | Protest to preserve s122 refunds |
| Jul 24 2026 | s122 expires; **forced-labor §301** 10–12.5% on ~60 economies | Another layer, stacking on everything |
| **Oct 6 2026** | **CAPE Phase 3** — finally liquidated entries, CIT plaintiffs only | Decide: sue, join the class (certified Oct 8), or lose it |

Sources: [KPMG on the CIT order](https://kpmg.com/us/en/taxnewsflash/news/2026/03/us-trade-court-orders-refund-removal-ieepa-duties.html),
[JD Supra CAPE guide](https://www.jdsupra.com/legalnews/a-guide-to-the-new-tariff-duty-refund-5709914/),
[Phillips Lytle on Phase 3](https://phillipslytle.com/cbp-opens-phase-3-of-cape-for-ieepa-refunds/),
[Clark Hill on s122](https://www.clarkhill.com/news-events/news/section-122-tariffs-ruling/),
[Honigman on forced-labor §301](https://www.honigman.com/alert-3462),
[Thompson Hine on 232 metals](https://www.thompsonhinesmartrade.com/2026/06/president-trump-modifies-section-232-tariffs-on-aluminum-copper-and-steel-imports/),
[Livingston on the wood-products delay](https://www.livingstonintl.com/section-232-duties-delayed-on-upholstered-furniture-kitchen-cabinets-and-vanities-2/).

## 2. Who feels it (composite personas)

**Dana — VP Finance, $6M/yr DTC home-goods brand (our demo workspace).** Imports from
ten countries through two brokers. Tariffs went from ~4% to ~33% of COGS in a year.
Theme: *"I can't close the month because I don't know what duty we owe, what we'll get
back, or what to accrue."* Owns the refund. Doesn't know which entries are liquidated.

**Rafael — Ops lead, same company.** Places POs. Theme: *"By the time the broker tells me
the duty, the goods are on the water and the price is on the website."* Wants landed cost
before the PO, not after the entry.

**Priya — Owner, $1.5M/yr Amazon seller.** Supplier ships DDP; she isn't the importer of
record. Theme: *"My supplier raised prices for tariffs. Now the tariffs are refunded — to
him."* CBP refunds only the IOR ([JD Supra](https://www.jdsupra.com/legalnews/a-guide-to-the-new-tariff-duty-refund-5709914/)).

**Marcus — Licensed customs broker, 140 SMB clients.** Filing volume doubled; fees didn't.
Theme: *"Every client wants me to find their refunds and I'm billing $95 an entry to file
CAPE uploads by hand."* Wants a tool he can white-label.

**Elena — Controller, $80M industrial distributor (steel/aluminum heavy).** Theme: *"Our
broker paid 232 on full value for nine months instead of metal content."* Vendor audits
report exactly this: one post-entry audit found $1.7M of stacking overcharges across 10,246
entries ([KlearNow case study](https://klearnow.ai/case-studies/post-entry-audit-finds-1-7m-in-duty-overcharges-hidden-by-tariff-stacking/)).

## 3. Jobs to be done, ranked by dollars

Reference importer: $6M/yr imports, 55% China, 20% Vietnam, 10% India, 15% Mexico.
Dollar figures from Harbormaster's own engine and estimator (`services/estimator.py`).

| # | Job | Frequency | $ at stake (reference importer) | Today's workaround | Harbormaster feature |
|---|---|---|---|---|---|
| J1 | Get back the IEEPA duty I paid, with interest | Once (2026–28) | ~$1.2–1.4M | Broker files what they can; liquidated entries forgotten | Recovery pipeline, CAPE CSV, protest packets |
| J2 | Don't miss a deadline (protest 180 days, PSC 300 days, CIT 2 yrs) | Weekly | Entire refund on affected entries | Spreadsheet, if anyone | Deadline engine + calendar |
| J3 | Know what each SKU costs landed *today* | Every PO | Margin on ~$6M COGS | Broker quote after the fact | Landed cost & margin, price-to-hold |
| J4 | Catch broker filing errors | Every entry | 1–3% of non-IEEPA duty (~$15–45k/yr) | Annual audit, if any | Entry audit, broker scorecard |
| J5 | Decide where to source next | Quarterly | 5–25% of landed cost | Gut feel, supplier quotes | Sourcing simulator + scenarios |
| J6 | Classify new products correctly | Per new SKU | Rate spread between candidate codes (often 5–20 pts) | Broker picks | AI classification w/ duty impact |
| J7 | Understand what a new proclamation does to us | Monthly | Varies | Law-firm alerts, not catalog-specific | Change timeline with SKU impact |
| J8 | Claim drawback on exports | Quarterly | 99% of eligible duty on exports | Never (complexity) — up to ~80% of drawback goes unclaimed per industry estimates ([Purolator](https://www.purolatorinternational.com/u-s-businesses-fail-to-claim-more-than-2-4-billion-annually-in-drawback-reimbursements/)) | Drawback matcher |
| J9 | Get refunds my DDP supplier received | Once | Pass-through share of the refund | Email the supplier | IOR warnings, pass-through quantification |

**Insight 1 — the refund is the wedge, the margin is the business.** J1 is the largest,
most urgent and most measurable job, but it ends. J3/J4/J5/J7 recur every PO and every
entry. The product has to win J1 and *keep* the customer on J3.

**Insight 2 — deadlines are the urgency engine.** Protest windows are jurisdictional
([19 U.S.C. 1514](https://gingercontrol.com/tools/protest-deadline-calculator)); a missed one
is money gone. That converts a "nice audit" into a this-week purchase.

**Insight 3 — the broker is a channel, not a competitor.** Brokers are swamped and paid
per entry. A tool that drafts their filings and lets them offer recovery is welcome.

**Insight 4 — trust needs explanations.** Every number must show its rate, base,
authority and the layers *not* applied, or finance teams won't book it. Hence the
engine's `Exclusion` records and the per-line "why" column.

## 4. Competitive scan (as of Oct 2026)

| Player | Focus | Gap we exploit |
|---|---|---|
| Tarifflo (YC S26), Trava (YC), NEU forge (YC) | Duty refund recovery, contingency fees | Recovery-only: churn when the wave ends |
| Gaia Dynamics ($7M seed), GingerControl ($2.1M seed) | HS/HTS classification, calculators | No entry ledger, no recovery |
| Flexport, KlearNow, digital brokers | Brokerage + visibility | Tied to their own brokerage; SMBs keep their broker |
| Zonos, Avalara cross-border | DDP checkout duty for e-commerce | Outbound/consumer, not importer ledgers |
| Law/CPA firms (KPMG, Big-4 trade) | Refund strategy for large importers | Priced for $100M+ importers |

Sources: [YC Tarifflo](https://yespress.io/tarifflo-yc-s26), [YC Trava](https://ycombinator.com/companies/trava),
[YC NEU forge](https://www.ycombinator.com/launches/NEU-forge-getting-companies-refunds-on-tariffs-duties),
[GingerControl seed](https://www.trysignalbase.com/news/funding/gingercontrol-secures-21m-seed).

## 5. Design-partner interview guide (30 minutes)

1. Walk me through the last PO you placed. When did you know the duty? Who told you?
2. How much duty did you pay in 2025? How do you know?
3. Who is importer of record on your entries? Any DDP suppliers?
4. Have you filed anything in CAPE? Which entries? Who decided which?
5. Do you know which of your 2025 entries have liquidated? When?
6. Last time your broker got something wrong — how did you find out?
7. What would you do with an extra 5 points of gross margin on your top 20 SKUs?
8. If we found $X, what would make you trust the number enough to file?
9. Who else would need to say yes (CFO, broker, counsel)?
10. What would you pay: % of recovery, monthly, both?

**Success criteria for the design-partner program:** 25 importers, ≥ 10 with a CAPE/protest
filing drafted by Harbormaster within 14 days of upload, ≥ 60% converting to the Growth
plan after their recovery, NPS ≥ 50 from finance users.
