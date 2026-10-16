# Investment memo — Sundial

**Company:** Sundial — the retirement paycheck. **Stage:** pre-seed / seed.
**One line:** Sundial turns a 401(k) into a pension: it decides when to claim Social
Security, which account to draw from and how much income to guarantee — then pays the
household on the first of every month.

## Summary recommendation
Invest. Decumulation is the largest unserved problem in consumer finance ("the nastiest,
hardest problem in finance" — William Sharpe), the customer count is at an all-time peak,
and the revenue pools (assets, annuity premiums, subscriptions) are among the largest in
fintech. Sundial's wedge is a free plan that is genuinely better than what advisors hand
out, and its moat is owning the monthly paycheck — the cash-flow rail every retirement
dollar passes through.

## Market
- **Customers:** > 4.1M Americans turn 65 every year through 2027; tens of millions are in
  the 55–75 window where these decisions are made. Most have no pension.
- **Money in motion:** US retirement assets exceed $40T (Investment Company Institute);
  rollovers out of workplace plans run in the hundreds of billions a year; individual
  annuity sales hit **$461B in 2025** (LIMRA) — a fourth record year, showing demand for
  guarantees that is currently met mainly through commissioned agents.
- **Why now:** (1) Peak 65; (2) the 2026 Trustees Report puts OASI depletion in late 2032
  (78% payable), so households now need a plan *with* Social Security uncertainty;
  (3) higher rates make guaranteed income cheap again (≈ 7.5% payout at 65);
  (4) SECURE 2.0 and the 2025 tax law (RMD age 75, QLAC $210k, senior deduction through
  2028) created new, time-limited planning windows.

## Product (live in this repository)
Free plan in 10 minutes → Social Security optimizer (every month 62–70, spousal and survivor
rules, health-adjusted) → tax-smart withdrawal & Roth conversion strategy (IRMAA-aware) →
safe spending with guardrails (Monte Carlo over markets and lifetimes) → income floor with
indicative annuity pricing and disclosed commissions → stress tests → a 12-month paycheck
calendar. Accounts, saved plans, planner requests, quote requests, waitlist with referrals.

## Business model
| Line | Price | Margin | Year-3 mix |
|---|---|---|---|
| Paycheck subscription | $19/mo or $190/yr | ~85% | 20% |
| Managed accounts (RIA) | 0.30%/yr, capped at $3k | ~70% | 45% |
| Annuity distribution | ~2% of premium (carrier-paid, disclosed, capped) | ~80% | 25% |
| B2B: recordkeepers & employers (in-plan decumulation, white-label engine) | $2–6 per participant/yr | ~85% | 10% |

**Unit economics (managed customer, $800k):** revenue $2.4k/yr AUM + one-time ~$3k annuity
commission on a $150k purchase; CAC target $600–900 via the free plan; lifetime 15+ years.

## Competition
| | Advisors (Edelman, Fisher, RIAs) | Robo-advisors | Planning software (Boldin, Income Lab via advisors) | Annuity marketplaces | **Sundial** |
|---|---|---|---|---|---|
| Social Security optimisation (couples, survivor) | Varies | Minimal | Yes | No | **Yes** |
| Tax-aware withdrawals + conversions | Varies | Partial | Yes | No | **Yes, IRMAA-aware** |
| Guaranteed income | Commissioned | No | Advice only | Product only | **Integrated, disclosed** |
| Executes a monthly paycheck | Some | Some | No | No | **Yes** |
| Price | ~1% AUM | 0.25% | $10–$20/mo | Commission | Free → $19/mo → 0.30% |

## Why Sundial wins
1. **Free plan that outperforms a commissioned recommendation** — viral among the 55–75
   cohort and their adult children.
2. **The paycheck is the habit and the moat** — once deposits run through Sundial, moving
   away means re-plumbing a retiree's income.
3. **Neutral product shelf** — the engine never sees commissions; that trust is the brand.
4. **Data compounding** — real retirement spending, claiming and withdrawal behaviour at
   scale improves defaults and becomes a B2B asset for recordkeepers.

## Plan & milestones (seed, $7M)
- 0–6 months: SEC RIA registration, insurance agency licences (resident + non-resident),
  custody partner for the paycheck (sweeps, withholding), 3 annuity carriers.
- 6–18 months: 50k free plans, 4k Paycheck subscribers, $400M AUM, $60M annuity premium.
- 18–36 months: recordkeeper partnerships (in-plan decumulation), $2B AUM.

| | Y1 | Y2 | Y3 | Y4 |
|---|---|---|---|---|
| Free plans (cumulative) | 120k | 450k | 1.1M | 2.2M |
| Paycheck subscribers | 6k | 28k | 75k | 150k |
| AUM | $0.5B | $2.4B | $7B | $15B |
| Annuity premium placed | $90M | $420M | $1.1B | $2.2B |
| Revenue | $4.2M | $19M | $52M | $104M |

## Risks
| Risk | Mitigation |
|---|---|
| Trust (older users, scams) | Free, no-account plan; plain English; no cold calls; disclosed commissions; fiduciary advisers |
| Regulation (RIA, insurance suitability, Reg BI) | Licensed entities, written suitability for every annuity, recorded rationale from the engine |
| Incumbents copy | Their revenue depends on 1% AUM and commissions — a neutral, cheap shelf cannibalises them |
| Model error | Rules unit-tested (`EngineTest`), assumptions published (`/methodology`), plans re-run when rules change |
| CAC in an older demographic | Adult-children channel, employer/recordkeeper distribution, SEO for high-intent questions |
