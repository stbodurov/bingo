# Harbormaster — revenue model & ads plan

## 1. Revenue lines

| Line | Price | Unit | Why customers accept it |
|---|---|---|---|
| **Recovery fee** | 15% (no plan) · 10% (Growth) · 7% (Scale) | % of refunds *received* | Zero risk: nothing if we find nothing; below refund-agency norms of 20–35% |
| **Growth plan** | $900 / month (annual: $9,720) | per workspace, 5 seats | Landed cost & margin, sourcing, alerts, classification, copilot; pays for itself on one repriced SKU |
| **Scale plan** | $60–150k / year | per company | ERP integrations, SSO, API, 7% fee, dedicated trade specialist |
| **Broker white-label** | $25k–$100k / yr + 20% rev share on recoveries | per brokerage | Brokers offer recovery & landed-cost reports under their brand |
| **Landed-cost API** | $0.02 / call after 10k free; volume tiers to $0.004 | per `POST /api/v1/stack` | Checkout and PO-time duty for ERPs, 3PLs, marketplaces |
| **Drawback filing** | 20% of drawback received | % | Specialist work, rarely done for SMBs |

The success-fee ledger and plan logic are live in the product (`services/cases.py`,
Settings & billing page).

## 2. Unit economics (Growth customer, $6M imports — the demo workspace)

| | Year 1 | Year 2 |
|---|---|---|
| Recovery fees (10% of ~$1.0M received) | $100k | $10k (errors, drawback) |
| Subscription | $10.8k | $10.8k + seat expansion |
| COGS: compute + Claude usage (~2k calls/mo) | $1.6k | $1.6k |
| COGS: trade-specialist review (4 hrs/filing × 6 filings) | $3.6k | $1k |
| **Gross profit** | **$105k** | **$18k** |
| CAC (blended: content + broker channel + paid) | $6–9k | — |
| **Payback** | **< 1 month** (fees) | |

The refund wave makes CAC payback near-instant in 2026–28; the SaaS line is what we
optimise retention for (target NRR ≥ 115% via seats, API and broker expansion).

## 3. Ads & acquisition plan

We buy ads; we never sell them. Importers and brokers are the audience; the estimator
is the ad.

### Budget (first 12 months): $1.1M

| Channel | Tactic | Budget | Primary KPI |
|---|---|---|---|
| Google Search | High-intent terms: "IEEPA refund", "CAPE declaration", "tariff refund liquidated entries", "protest deadline customs", "landed cost calculator", "section 232 derivative metal content" → estimator | $360k | Cost per completed estimate < $40; estimate→waitlist > 25% |
| LinkedIn | CFOs/controllers/ops at importers (NAICS 42x wholesale, 44x–45x retail, 33x manufacturing; 20–1,000 staff); creative = "Your protest window closes in 180 days. Do you know which entries liquidated?" | $280k | Qualified pipeline at < $450/opportunity |
| Amazon/Shopify seller ecosystem | Sponsorships of seller newsletters, podcasts and communities; Shopify app listing (landed cost at product level) | $150k | 1,500 seller signups; DDP-refund pass-through education |
| Broker partnerships | Co-marketing funds for brokers who white-label; booths at NCBFAA events | $140k | 40 broker partners, 25% of new revenue via brokers |
| CPA / law-firm referrals | Referral fee (10% of first-year fees) to firms advising importers on CAPE/CIT | $90k | 200 referred importers |
| Content & SEO | Daily tariff-change briefs auto-generated from the engine's timeline; per-HTS landing pages with live rates ("6110.20.20 duty from Vietnam today") | $80k | 50k organic visits/mo by month 9 |

### Messaging by stage
1. **Urgency (now → mid-2027):** "Your IEEPA refund has a deadline." Lead with the
   estimator and the deadline calendar.
2. **Margin (2027 →):** "Know the duty before the PO." Lead with landed cost and the
   sourcing simulator.
3. **Platform (2028 →):** "The landed-cost API." Lead with developers and ERPs.

### Waitlist mechanics (live)
- Estimator result is saved with the signup → sales sees refund size on day one.
- Onboarding order = refund size × deadline urgency (not first-come), so the most
  valuable, most urgent accounts convert first.
- Referral: both sides get the first $25k recovered fee-free, and each referral moves the
  referrer up the queue.

## 4. Guardrails
- No data resale, no advertising inside the product, no training on customer data.
- Fees invoiced only against refunds visible in ACE/bank records.
- Every filing reviewed by a licensed broker or attorney; Harbormaster carries E&O cover.
