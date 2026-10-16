# Sundial — the numbers first

> **Method.** Sundial's research started from the arithmetic, not from interviews or
> regulation: we coded the rules that decide a retiree's income (Social Security, tax,
> RMDs, Medicare surcharges, mortality, annuity pricing) and measured which decisions move
> lifetime outcomes by six figures. Those decisions became the product. Every figure below
> is reproducible with the engine in `src/main/java/com/sundial/engine` (sample households
> in `Samples.java`; run `EngineTest` or the API).

## 1. The population and its problem

| Fact | Figure | Source |
|---|---|---|
| Americans turning 65 each year through 2027 ("Peak 65") | > 4.1 million (~11,200 a day) | [Alliance for Lifetime Income](https://protectedincome.org/peak65/) |
| Share of Peak-65ers without a defined-benefit pension | the majority | same |
| Retirement trust fund (OASI) depletion | Q4 2032; then 78% of scheduled benefits payable | [2026 Trustees Report](https://www.ssa.gov/oact/tr/2026/II_D_project.html) |
| 2026 / 2027 COLA | 2.8% / 3.6% | [SSA](https://www.ssa.gov/blog/en/posts/2026-10-14.html) |
| Individual annuity sales, 2025 | $461B (fourth straight record) | [LIMRA via Insurance Business](https://www.insurancebusinessmag.com/us/news/breaking-news/us-annuity-sales-hit-record-461-billion-as-indexed-products-surge--limra-565533.aspx) |
| Lifetime SPIA payout, male 65, $100k | ≈ $625–$665 / month (mid-2026) | [market survey](https://myannuitystore.com/annuities/immediate-annuity/) |

The defined-benefit pension used to make three decisions for people: when income starts,
how much is guaranteed, and how it's taxed. A 401(k) leaves all three to the retiree.

## 2. Decision 1 — when to claim Social Security

Rules coded (`SocialSecurity.java`): early reduction 5/9% per month (first 36) and 5/12%
beyond; delayed credits 2/3% per month to 70; spousal up to 50% of the worker's PIA
(25/36% reduction, no delayed credits); survivor inherits the deceased's actual benefit
(≥ 82.5% of PIA if they claimed early).

Valuation (`ClaimingOptimizer.java`): expected present value of household benefits, each
month weighted by joint survival probabilities, discounted at a 1.8% real rate.

| Household (engine output) | Best ages | vs. claiming ASAP | Note |
|---|---|---|---|
| Couple: Ellen (63, PIA $2,380, good health) + Marcus (61, PIA $3,140, average) | 66y1m & 70 | **+$95k** expected lifetime benefits | Delaying the higher earner raises the survivor's check |
| Single: Rosa (66, PIA $2,210, average) | 69y7m | +$15k | Breakeven of 62 vs 70 ≈ age 80 |
| Single man, poor health | before FRA | — | Optimizer moves earlier as life expectancy drops (test `claimingOptimizerPrefersDelayForHealthyHigherEarner`) |

**Insight:** the value is concentrated in couples with unequal earnings, where the claiming
decision is really a *survivor-insurance* decision. Most calculators ignore it.

## 3. Decision 2 — which account to spend, and when to convert

Tax rules coded (`TaxCode.java`): 2026 brackets and standard deduction, +$1,650/$2,050 at
65, the $6,000 senior deduction (2025–2028, 6% phase-out above $75k/$150k MAGI), the
§86 Social Security "tax torpedo" (thresholds not indexed), LTCG brackets, NIIT, IRMAA
tiers on MAGI two years back, RMDs at 73/75 (SECURE 2.0).

| Couple sample (median markets, to 95) | Lifetime tax + IRMAA | After-tax legacy |
|---|---|---|
| Taxable first, then IRA, then Roth | ≈ $343k | ≈ $0.46M |
| Pro-rata | ≈ $306k | ≈ $0.57M |
| **Fill the 12% bracket + Roth conversions** | **≈ $257k** | **≈ $1.16M** |

**Insight:** the "gap years" between retirement and RMDs/Social Security are a low-bracket
window; filling it with conversions cuts lifetime tax by ~$85k and roughly doubles the
after-tax legacy for this household — without lowering spending confidence.

## 4. Decision 3 — how much income to guarantee

Pricing (`Annuities.java`): PV of expected payments at a 5.35% insurer yield with
healthy-annuitant mortality and a 3.5% load → $100k buys ~$640/month for a 65-year-old man
(calibrated to market). Joint and deferred (DIA/QLAC, $210k limit) priced the same way.

The floor rule: cover *essential* spending with Social Security + pensions + annuities;
cap annuitisation at 35% of liquid savings. For the single sample, a ~$157k life annuity
closes the essentials gap and raises 85%-confidence spending odds from 84% to 87%.

## 5. Risk that people underestimate

| Risk | How it's modelled | Sample result |
|---|---|---|
| Longevity | Gompertz mortality calibrated to SSA cohort LE; sampled per person per scenario | 1 in 4 of Ellen's peers reach 95 |
| Sequence of returns | −35% then −10% stocks in the first two years | Couple survives; single runs short at 89 |
| Inflation | 4% for life | Single runs short at 92 |
| Trust-fund cut | 78% of benefits from 2033 | −$204k expected benefits for the couple |
| Long-term care | $110k/yr for 3 years at 82 | Single runs short at 83 |
| Widowhood | First death at 76; survivor files single, keeps larger benefit | Couple holds |

## 6. What this told us to build
1. A claiming optimizer that treats couples as one decision (spousal + survivor).
2. A withdrawal engine that compares strategies on *after-tax* outcomes, IRMAA-aware.
3. An income-floor tool with transparent, commission-disclosed annuity pricing.
4. Monte Carlo over markets **and lifetimes** — success means money outlasts the people.
5. Turning the plan into a monthly paycheck, because the plan only matters if it's executed.
