# Singleton — monetization, ads & growth plan

## Principles
1. **People never pay and never see ads.** A proof-of-human product that monetises attention
   or data destroys the trust it sells. No ads in the Pass, no data sales, no token.
2. **Platforms pay per verified human**, priced against what they replace (KYC at $1.35–$2.30,
   fraud losses, regulatory fines), not against our cost (cents of compute).
3. **Reuse is the business model.** Verification happens once; every subsequent proof is ~pure
   margin. Price falls with volume but margin rises.

## Pricing (live in the product's billing page)
| Plan | Price | Includes | Who |
|---|---|---|---|
| Developer | $0 | 1,000 live verifications / mo, unlimited test mode, all predicates | Startups, hackathons, evaluation |
| Growth | $0.05 per verified human; +$0.02 per age/region predicate; first 1,000 free | Duplicate detection, action scopes, agent delegation, webhooks | Most platforms |
| Scale | $2,000 / mo commit incl. 80,000; overage $0.025 | Compliance evidence packs (AU/UK/US), 99.95% SLA, priority attester routing | Regulated launches |
| Enterprise | Annual contract (typ. $250k–$2M) | On-prem verifier, custom attesters/regions, DPA, audit support | Networks with 10M+ users |

**Billable event:** a completed session (verified or duplicate) in live mode. Failures,
abandonments and test mode are free — we only earn when the platform gets an answer.

## Additional revenue streams (years 2–4)
| Stream | Mechanism | Size at scale |
|---|---|---|
| Agent delegation (KYA) | $0.01 per delegation + $0.001 per verified agent request above 1M/mo | 10–15% of ARR |
| Compliance packs | Per-regime evidence exports, audit support, regulator-ready reports | Included in Scale; $25k/yr add-on for Growth |
| Attester network fees | In-person partners (pharmacies, post offices) are paid per enrolment; we charge platforms a premium for document-verified (`human_doc`) proofs | Margin-neutral; drives supply |
| On-prem verifier licence | Banks, governments, large networks running verification in their own VPC | Enterprise ACV uplift |

## Unit economics (Growth plan, year 2 assumptions)
| | per verified human proof |
|---|---|
| Price (blended with predicates) | $0.058 |
| Verifier compute + bandwidth | $0.0004 |
| Amortised enrolment cost (attestation $0.40 ÷ ~25 proofs per holder-year) | $0.016 |
| Payment processing, support | $0.003 |
| **Gross margin** | **≈ 66% early → 90%+ as reuse grows** |

Enrolment cost is the only meaningful COGS and it is amortised across every app the holder
uses — so each new platform improves margins for all the others.

Platform CAC target: < 9 months of gross profit payback (self-serve ~ $1.5k, mid-market ~ $18k,
enterprise ~ $120k). Net revenue retention target 140%+ (usage expands with platform growth and
new predicates/regions).

## Ads & paid acquisition plan
We *buy* ads; we never *sell* them. Two audiences, two playbooks.

### A. Platforms (B2B) — 65% of marketing spend
| Channel | Tactic | Budget (yr 1) | KPI |
|---|---|---|---|
| LinkedIn ads | Trust & Safety, Product, Legal leaders at social/dating/marketplace/gaming companies in AU, UK, US; creative = "A$49.5M per breach — prove age without storing IDs" | $240k | Cost per qualified platform lead < $350 |
| Google Search | High-intent terms: "age verification API", "Online Safety Act age assurance", "prevent duplicate accounts", "CAPTCHA alternative" | $180k | CPA < $250, demo rate > 20% |
| Developer channels | Sponsored newsletters (TLDR, Pointer), Hacker News launch, Product Hunt, GitHub sponsorships; free tier as the ad | $90k | 3k developer signups |
| Compliance webinars | With law firms in Sydney/London; regulator-deadline content | $60k | 150 qualified accounts |
| Marketplaces | Cloudflare, Auth0/Okta, Shopify app listings (bot-proof drops) | $30k | 15% of new platforms |

### B. People (consumer Pass) — 35% of marketing spend
Holders mostly arrive *through* platforms ("Verify with Singleton" is the ad). Paid support only:
| Channel | Tactic | Budget (yr 1) | KPI |
|---|---|---|---|
| Co-marketing with launch partners | Ticket drops and dating apps promote "skip the bots — verify once with Singleton" | $150k MDF | Cost per verified holder < $1.20 |
| Short-form social (TikTok/Reels/Shorts) | "I proved I'm human without showing my ID" demos; creators in privacy/tech niches | $120k | 3M views, < $2 per enrolment |
| In-person partner signage | QR codes at partner pharmacies/post offices | $40k | 50k in-person enrolments |
| Referral (waitlist) | Move up 5 places per friend; Founding Human at 10 referrals | $0 (built in) | k-factor > 0.4 pre-launch |

## Go-to-market sequence
1. **Months 0–6 — Deadline wedge.** AU social/dating platforms facing the under-16 law and UK
   platforms under Ofcom enforcement. 40 design partners, free Scale plan for 6 months in
   exchange for case studies.
2. **Months 6–18 — Fair-access wedge.** Ticketing, sneaker/merch drops, gaming launches, airdrops
   where "one per human" directly protects revenue. Shopify app + Cloudflare integration.
3. **Months 18–36 — Agentic internet.** Agent platforms and APIs adopt Know-Your-Agent; Singleton
   becomes the default way to say "a real person stands behind this bot".

## KPIs we run the company on
- Verified holders; % enrolled via each attester path; enrolment success rate
- Proofs per holder per month (reuse — the margin engine)
- Platforms live; median time from signup to first live verification
- Duplicate-catch rate and bot-signup reduction reported by platforms
- Gross margin; NRR; payback months
- Privacy: zero personal data retained (audited), median anonymity bucket size
