# Singleton — investment case

> Built from the threat model outward: the size of the market is the size of the problem
> platforms are now **legally and economically forced** to solve.

## 1. One-line pitch
Singleton is the privacy-preserving acceptance network for proof-of-human and age: people
verify once with an ID they already have, and every app gets a one-bit answer plus a
per-app pseudonym — at $0.05 instead of $1.35–$2.30 per check.

## 2. Why now (three curves crossed in 2025–26)
1. **Bots overtook people.** 57.5% of HTML traffic was automated in June 2026, and AI-agent
   traffic grew >1,700% YoY ([Cloudflare via TechTimes](https://www.techtimes.com/articles/317877/20260605/bot-traffic-passes-humans-online-cloudflare-says-agentic-ai-drove-575-share.htm),
   [Forbes](https://www.forbes.com/sites/jonmarkman/2026/10/01/cloudflare-says-ai-agent-traffic-grew-more-than-1700-in-the-past-year/)).
   Imitation tests (CAPTCHAs) are solved 100% by off-the-shelf models ([Decrypt](https://decrypt.co/251107/ai-can-best-googles-bot-detection-system-swiss-researchers-find)).
2. **Regulators mandate age assurance.** Australia's under-16 law (A$49.5M fines, in force
   Dec 2025), the UK Online Safety Act (10% of global revenue), US state laws and the EU
   wallet rollout ([eSafety](https://www.esafety.gov.au/about-us/industry-regulation/social-media-age-restrictions),
   [Ofcom summary](https://www.zyphe.com/resources/blog/age-verification-uk-online-safety-act)).
3. **Credentials went digital.** 22 US states/territories have TSA-accepted mobile IDs; Apple
   Wallet IDs are live in 14–17 jurisdictions; every EU state must ship a wallet by end-2026
   ([idscan](https://idscan.net/blog/the-state-of-digital-ids-in-2026-a-complete-guide/)). Strong
   verification no longer needs an upload — it needs an acceptance layer.

## 3. Market
**Top-down.** Identity verification is ~$16.5B in 2026, growing ~15% a year to ~$33B by 2031
([MarketsandMarkets via Security World Market](https://www.securityworldmarket.com/int/News/Business-News/id-verification-market-to-surpass-33-billion-by-2031)).
Age assurance and bot management are adjacent budgets being pulled into the same decision.

**Bottom-up (what platforms would pay us).**

| Segment | Verifications / yr (est.) | Price | Revenue pool |
|---|---|---|---|
| Social & community (AU/UK/US age + sybil) | 1.5B signups & re-checks | $0.05–0.07 | $75–105M |
| Dating | 300M | $0.10 (doc-verified) | $30M |
| Marketplaces & reviews | 600M | $0.05 | $30M |
| Ticketing, drops, gaming | 800M | $0.05 | $40M |
| AI-agent platforms (KYA) | 2B agent delegations / checks | $0.01–0.02 | $20–40M |
| Civic, polls, airdrops | 200M | $0.05 | $10M |
| **Total near-term SAM** | **~5.4B** | | **≈ $205–255M** |

These are deliberately conservative: Meta alone actioned ~3.5B fake accounts in 2025
([PPC Land / VAB](https://ppc.land/facebook-banned-3-5-billion-fake-accounts-in-2025-vab-analysis-finds/)) —
the demand for "is this a distinct person?" is an order of magnitude larger than today's KYC volume.
At scale, verification becomes a default step of every signup, like email confirmation.

## 4. Competition
| | World ID | Humanity Protocol | KYC vendors (Persona, Sumsub, Onfido…) | CAPTCHA / bot mgmt | **Singleton** |
|---|---|---|---|---|---|
| How a human is proven | Iris scan at an Orb | Palm scan | Document + selfie | Behaviour | Issuer-signed credentials, in person, or vouching |
| Hardware | Orb | Scanner / phone camera | Camera | — | None |
| Apps learn | Unique-human proof | Unique-human proof | Full identity | Score | Pseudonym + predicates |
| Age | No | No | Yes | No | **Yes** |
| Regulatory exposure | Biometrics (repeated enforcement) | Biometrics | Data honeypot | Tracking | Minimal data |
| Price to app | Per verification | Token-linked | $1.35–$2.30 | Low | $0.05 |

World and Humanity Protocol validate the category and the valuation (Humanity at **$1.1B** on
$20M — [Reuters](https://www.streetinsider.com/Reuters/Humanity+Protocol+valued+at+%241.1+billion+after+latest+fundraise/24249164.html);
World Foundation **$525M** raised — [HackerNoon](https://hackernoon.com/world-foundation-raises-$525-million-to-scale-proof-of-human-as-ai-transforms-the-internet)).
Both depend on body scans that regulators and users resist. KYC vendors can't go privacy-first
without cannibalising per-check revenue. That is the gap.

## 5. Moat
1. **Two-sided network.** Every attester added makes the Pass easier to get; every app that
   accepts it makes it more valuable to hold; holders verified once become free distribution to
   the next app. Platforms pay; holders never do.
2. **Public, anchored sets.** Trust is auditable rather than contractual — a credibility moat for
   regulators and privacy advocates that closed vendors can't match.
3. **Cost structure.** Verification is a few milliseconds of curve arithmetic; marginal gross
   margin >90%. Incumbents' cost per check is human review and document capture.
4. **Compliance evidence.** Regulators ask for evidence of method and data minimisation; our
   console produces it per regime.

## 6. Customer success likelihood
- **Immediate, measurable value:** duplicate accounts caught, bot signups prevented, age-gate
  pass rate, and the end of ID storage (breach liability removed).
- **Low switching cost in:** a button and a JWT. Starts in test mode for free.
- **Regulatory pull:** AU/UK deadlines turn "nice to have" into a board-level line item.
- **User experience:** median ~8 s to verify after the one-time enrolment; reusable everywhere.

## 7. Path to $100M ARR
| Year | Platforms (paying) | Verified holders | Verifications / yr | ARR |
|---|---|---|---|---|
| 1 | 40 design partners → 120 | 1.5M | 60M | $3M |
| 2 | 600 | 12M | 400M | $22M |
| 3 | 2,000 | 45M | 1.2B | $58M |
| 4 | 4,500 | 110M | 2.3B | $104M |

Blended price falls from $0.05 to ~$0.045 with volume tiers; enterprise minimums (Scale/Enterprise)
contribute ~35% of ARR by year 3. At 30–40× ARR (identity infra multiples), the year-3 business
clears a unicorn valuation.

## 8. Risks and mitigations
| Risk | Mitigation |
|---|---|
| Platform-owned wallets (Apple/Google) verify age directly | We *accept* those wallets; our value is cross-wallet acceptance, sybil-resistance per app and agent delegation, which wallets don't provide |
| Regulators specify a method we don't support | Attester model is pluggable; we add attesters (bank, telco, facial age estimation) without changing the protocol |
| Anonymity criticism (bucket size) | Minimum 32-member buckets now, zk-SNARK whole-set proofs in v2 |
| Attester fraud | Per-attester assurance levels, public fraud metrics, mass revocation by attester |
| Cold start | Start with platforms under deadlines (AU/UK social, dating, ticketing), subsidise first enrolments, in-person partner network |

## 9. The ask (illustrative)
Seed: $6M to ship native wallets, production mDL/ePassport verification, 3 regional
attester networks and 40 design partners (AU/UK social & dating first).
