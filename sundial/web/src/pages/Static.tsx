import { Link } from 'react-router-dom'

export function MethodologyPage() {
  return (
    <div className="wrap" style={{ maxWidth: 860 }}>
      <div className="page-head"><div><span className="eyebrow">Methodology</span><h1>How Sundial does the math</h1></div></div>
      <div className="card">
        <h3>Social Security</h3>
        <p>Full retirement age by birth year; early reduction of 5/9 of 1% per month for the first 36 months and 5/12 of 1% after; delayed credits of 2/3 of 1% per month to 70. Spousal benefits up to 50% of the worker’s PIA, reduced 25/36 of 1% per month early, no delayed credits. Survivor benefits inherit the deceased’s actual benefit (at least 82.5% of PIA if they claimed early), reduced if taken before the survivor’s full retirement age. The 2027 COLA of 3.6% is applied, then your inflation assumption. The optional depletion scenario pays 78% of scheduled benefits from 2033 (2026 Trustees Report).</p>
        <h3>Claiming optimizer</h3>
        <p>Every combination of claiming ages (quarterly, then monthly around the best) is valued as the expected present value of household benefits: each future month is weighted by the probability each spouse is alive and which benefit is then paid, discounted at a real rate near the TIPS yield (1.8% by default).</p>
        <h3>Longevity</h3>
        <p>Gompertz mortality calibrated to Social Security cohort life expectancy at 65 (about 85 for men and 87 for women of average health), shifted by self-reported health. Monte Carlo draws an age at death for each person in every scenario.</p>
        <h3>Taxes</h3>
        <p>2026 federal brackets, standard deduction and the age-65 addition, the $6,000-per-person senior deduction for 2025–2028 with its 6% phase-out, taxation of Social Security (provisional income thresholds of $25,000/$32,000 and $34,000/$44,000, not indexed), long-term capital gains brackets, the 3.8% NIIT, and Medicare IRMAA tiers on income from two years earlier. Thresholds are indexed with your inflation assumption. State tax is a flat rate on income other than Social Security.</p>
        <h3>Withdrawals, RMDs and conversions</h3>
        <p>Required minimum distributions start at 73 (born 1951–59) or 75 (born 1960 or later) using the IRS Uniform Lifetime Table. Strategies compared: taxable-first, pro-rata, and “bracket fill” — drawing IRA money and converting to Roth up to the top of the 12%, 22% or 24% bracket, optionally capped below the first IRMAA tier. The recommendation maximises confidence first, then after-tax money left to heirs.</p>
        <h3>Markets</h3>
        <p>Stocks: 8.5% expected nominal return, 17% volatility. Bonds: 4.5%, 6%. Correlation 0.1. Lognormal annual returns; 1,000 scenarios per plan. Deterministic views use median returns. Cash earns inflation plus 0.5%.</p>
        <h3>Spending</h3>
        <p>Discretionary spending tapers 15% from 75 and 30% from 85 (the retirement “spending smile”). A surviving spouse needs about 72% of the couple’s spending. Guardrails trim discretionary spending 10% when the withdrawal rate rises 20% above its starting level and restore it when it falls 20% below.</p>
        <h3>Annuities</h3>
        <p>Indicative prices discount expected payments at a 5.35% insurer yield with insurer (healthy-annuitant) mortality and a 3.5% load, calibrated to mid-2026 market quotes (about $625–$665 a month per $100,000 for a 65-year-old man). QLAC premiums are capped at $210,000.</p>
        <p className="muted">These are planning estimates, not guarantees. <Link to="/plan">Build your plan</Link>.</p>
      </div>
    </div>
  )
}

export function PrivacyPage() {
  return (
    <div className="wrap" style={{ maxWidth: 820 }}>
      <div className="page-head"><h1>Privacy</h1></div>
      <div className="card">
        <p>You can build a plan without an account; your inputs are kept in your own browser until you choose to save them.</p>
        <p>If you create an account, we store your email, a hashed password and the plans you save. We never sell or rent personal data, never share it for marketing, and delete it when you ask.</p>
        <p>We use a session cookie to keep you signed in and a CSRF cookie to protect forms. No advertising trackers.</p>
      </div>
    </div>
  )
}

export function NotFound() {
  return <div className="wrap"><div className="page-head"><h1>That page isn’t here</h1></div><Link className="btn btn-primary" to="/">Go home</Link></div>
}
