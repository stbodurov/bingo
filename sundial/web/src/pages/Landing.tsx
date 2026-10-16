import { useEffect, useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { api, ApiError, type ClaimingAnswer } from '../lib/api'
import { age, ageShort, usd } from '../lib/format'
import type { Health, Plan, Sex } from '../lib/types'
import { HBars } from '../components/charts'

const SOURCE_COLORS: Record<string, string> = { 'Social Security': 'var(--s1)', Pension: 'var(--s2)', Annuity: 'var(--s3)', Sundial: 'var(--s5)' }

function PayCard() {
  const [plan, setPlan] = useState<Plan | null>(null)
  useEffect(() => {
    let live = true
    api.sample('couple').then((h) => api.plan(h)).then((p) => live && setPlan(p)).catch(() => undefined)
    return () => { live = false }
  }, [])
  const month = plan?.paycheck[0]
  const deposits = month?.lines.filter((l) => l.amount > 0 && !l.source.startsWith('Roth')) ?? []
  return (
    <figure className="paycard" aria-label="Example monthly paycheck from a Sundial plan">
      <div className="head">
        <div><div className="eyebrow" style={{ marginBottom: 4 }}>Example paycheck</div>
          <strong>{plan ? `${plan.generatedFor} · ${month?.month.charAt(0)}${month?.month.slice(1).toLowerCase()}` : 'Ellen & Marcus'}</strong></div>
        <span className="badge b-good">{plan ? `${Math.round(plan.summary.successRate * 100)}% confidence` : 'Planning…'}</span>
      </div>
      <div className="total">{plan ? usd(month?.total ?? 0) : '—'}<span className="muted small"> /month after tax</span></div>
      <div style={{ marginTop: 14 }}>
        {deposits.map((d, i) => {
          const day = new Date(d.date + 'T12:00:00')
          const key = d.source.startsWith('Social') ? 'Social Security' : d.source.startsWith('Pension') ? 'Pension' : d.source.startsWith('Annuity') ? 'Annuity' : 'Sundial'
          return (
            <div className="deposit" key={i} style={{ animationDelay: `${i * 120}ms` }}>
              <div className="day"><small>{day.toLocaleString('en-US', { month: 'short' }).toUpperCase()}</small>{day.getDate()}</div>
              <div><span className="dot" style={{ background: SOURCE_COLORS[key] }} />{d.source}<div className="small muted">{d.note}</div></div>
              <b>{usd(d.amount)}</b>
            </div>
          )
        })}
        {!plan && <div className="loading" style={{ minHeight: 180 }}><div><div className="sunspin" />Running 1,000 market and lifespan scenarios…</div></div>}
      </div>
      <figcaption className="small muted" style={{ marginTop: 10 }}>A sample couple’s plan, computed live by Sundial’s engine. Taxes and Medicare surcharges are set aside every month.</figcaption>
    </figure>
  )
}

function QuickClaim() {
  const [birthYear, setBirthYear] = useState(1963)
  const [pia, setPia] = useState(2600)
  const [sex, setSex] = useState<Sex>('FEMALE')
  const [health, setHealth] = useState<Health>('GOOD')
  const [married, setMarried] = useState(false)
  const [spousePia, setSpousePia] = useState(1400)
  const [res, setRes] = useState<ClaimingAnswer | null>(null)
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const run = async (e?: FormEvent) => {
    e?.preventDefault()
    setBusy(true); setErr('')
    const me = { name: 'You', birthYear, birthMonth: 6, birthDay: 15, sex, health, pia, retireAge: 0, salary: 0, claimAgeMonths: 0 }
    const people = married ? [me, { ...me, name: 'Spouse', sex: sex === 'MALE' ? 'FEMALE' : 'MALE', pia: spousePia, health: 'AVERAGE' }] : [me]
    try { setRes(await api.claiming(people)) } catch (x) { setErr(x instanceof ApiError ? [x.message, ...x.details].join(' ') : 'Something went wrong.') }
    setBusy(false)
  }
  useEffect(() => { run() }, []) // eslint-disable-line react-hooks/exhaustive-deps
  const labels = ['Earliest', 'Full retirement age', 'At 70', 'Sundial’s pick']
  return (
    <div className="quick">
      <form className="card" onSubmit={run}>
        <div className="row">
          <div className="field"><label htmlFor="qc-year">Birth year</label>
            <input id="qc-year" type="number" min={1936} max={1986} value={birthYear} onChange={(e) => setBirthYear(+e.target.value)} /></div>
          <div className="field"><label htmlFor="qc-pia">Benefit at full retirement age</label>
            <input id="qc-pia" type="number" min={0} max={6000} step={10} value={pia} onChange={(e) => setPia(+e.target.value)} />
            <p className="hint">From your statement at ssa.gov/myaccount</p></div>
        </div>
        <div className="field"><p className="label-text">Sex at birth</p>
          <div className="seg" role="group" aria-label="Sex at birth">
            {(['FEMALE', 'MALE'] as Sex[]).map((s) => <button type="button" key={s} aria-pressed={sex === s} onClick={() => setSex(s)}>{s === 'FEMALE' ? 'Female' : 'Male'}</button>)}
          </div></div>
        <div className="field"><p className="label-text">Health compared with others your age</p>
          <div className="seg" role="group" aria-label="Health">
            {(['EXCELLENT', 'GOOD', 'AVERAGE', 'POOR'] as Health[]).map((h) => <button type="button" key={h} aria-pressed={health === h} onClick={() => setHealth(h)}>{h.charAt(0) + h.slice(1).toLowerCase()}</button>)}
          </div></div>
        <label className="check"><input type="checkbox" checked={married} onChange={(e) => setMarried(e.target.checked)} /> I’m married</label>
        {married && <div className="field" style={{ marginTop: 12 }}><label htmlFor="qc-sp">Spouse’s benefit at full retirement age</label>
          <input id="qc-sp" type="number" min={0} max={6000} step={10} value={spousePia} onChange={(e) => setSpousePia(+e.target.value)} /></div>}
        <button className="btn btn-primary" type="submit" disabled={busy} style={{ marginTop: 12 }}>{busy ? 'Calculating…' : 'When should I claim?'}</button>
        {err && <p className="error" style={{ marginTop: 12 }}>{err}</p>}
      </form>
      <div className="card" aria-live="polite">
        {res ? (
          <>
            <div className="eyebrow">Sundial’s pick</div>
            <div className="result-big">Claim at {res.best.map(age).join(' & ')}</div>
            <p className="muted" style={{ marginTop: 10 }}>Worth about <strong style={{ color: 'var(--ink)' }}>{usd(res.gainVsEarliest)}</strong> more in expected lifetime benefits than claiming as early as possible — weighted by how long {married ? 'you and your spouse are' : 'you are'} likely to live (to about {Math.round(res.lifeExpectancy[0])}{married ? ` and ${Math.round(res.lifeExpectancy[1])}` : ''}).</p>
            <HBars items={res.compare.map((c, i) => ({ label: `${labels[i]} · ${c.claimMonths.map(ageShort).join(' & ')}`, value: c.expectedPv }))} highlight={3} />
            <p className="hint">Expected lifetime benefits in today’s dollars. Claiming at 70 instead of 62 pays off if you live past about {res.breakevenAge}.</p>
            <Link className="btn btn-sun" to="/plan" style={{ marginTop: 10 }}>See the rest of my plan</Link>
          </>
        ) : <div className="loading"><div><div className="sunspin" />Weighing every month from 62 to 70…</div></div>}
      </div>
    </div>
  )
}

function Waitlist() {
  const nav = useNavigate()
  const [params] = useSearchParams()
  const [form, setForm] = useState({ firstName: '', email: '', ageBand: '60-64', assetsBand: '500k-1m', status: 'retiring-soon', website: '' })
  const [interests, setInterests] = useState<string[]>(['Social Security timing'])
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const toggle = (i: string) => setInterests((xs) => xs.includes(i) ? xs.filter((x) => x !== i) : [...xs, i])
  const submit = async (e: FormEvent) => {
    e.preventDefault(); setBusy(true); setErr('')
    try {
      const s = await api.waitlist({ ...form, interests, referredBy: params.get('ref') ?? sessionStorage.getItem('ref') ?? undefined })
      nav(`/waitlist/${s.code}`)
    } catch (x) { setErr(x instanceof ApiError ? [x.message, ...x.details].join(' ') : 'Something went wrong.') }
    setBusy(false)
  }
  const set = (k: keyof typeof form) => (e: { target: { value: string } }) => setForm({ ...form, [k]: e.target.value })
  return (
    <form className="card" onSubmit={submit}>
      <div className="row">
        <div className="field"><label htmlFor="w-name">First name</label><input id="w-name" value={form.firstName} onChange={set('firstName')} autoComplete="given-name" /></div>
        <div className="field"><label htmlFor="w-email">Email</label><input id="w-email" type="email" required value={form.email} onChange={set('email')} autoComplete="email" /></div>
      </div>
      <div className="row">
        <div className="field"><label htmlFor="w-age">Your age</label><select id="w-age" value={form.ageBand} onChange={set('ageBand')}>
          {['50-54', '55-59', '60-64', '65-69', '70-74', '75+'].map((a) => <option key={a}>{a}</option>)}</select></div>
        <div className="field"><label htmlFor="w-assets">Savings & investments</label><select id="w-assets" value={form.assetsBand} onChange={set('assetsBand')}>
          {[['under-250k', 'Under $250k'], ['250k-500k', '$250k–$500k'], ['500k-1m', '$500k–$1M'], ['1m-3m', '$1M–$3M'], ['3m+', '$3M+']].map(([v, l]) => <option key={v} value={v}>{l}</option>)}</select></div>
        <div className="field"><label htmlFor="w-status">Where you are</label><select id="w-status" value={form.status} onChange={set('status')}>
          <option value="working">Working, 3+ years out</option><option value="retiring-soon">Retiring in the next 2 years</option><option value="retired">Already retired</option></select></div>
      </div>
      <fieldset className="field" style={{ border: 0, padding: 0 }}>
        <legend style={{ fontWeight: 700, marginBottom: 8, color: 'var(--ink-2)' }}>What do you want help with?</legend>
        <div className="seg" style={{ borderRadius: 18 }}>
          {['Social Security timing', 'Monthly paycheck', 'Taxes & Roth conversions', 'Guaranteed income', 'Talking to a planner'].map((i) =>
            <button type="button" key={i} aria-pressed={interests.includes(i)} onClick={() => toggle(i)}>{i}</button>)}
        </div>
      </fieldset>
      <div style={{ position: 'absolute', left: -10000 }} aria-hidden="true"><label htmlFor="w-site">Website</label><input id="w-site" tabIndex={-1} value={form.website} onChange={set('website')} autoComplete="off" /></div>
      {err && <p className="error">{err}</p>}
      <button className="btn btn-sun" type="submit" disabled={busy}>{busy ? 'Joining…' : 'Join the waitlist'}</button>
      <p className="hint">We’ll email you once, when your spot opens. No spam, no selling your address.</p>
    </form>
  )
}

const FEATURES: [string, string, string][] = [
  ['☀', 'Social Security optimizer', 'Every month from 62 to 70 for both of you, with spousal and survivor rules and your health built in.'],
  ['⇅', 'Tax-smart withdrawals', 'Which account to draw from each year, and how much to convert to Roth while your bracket is low.'],
  ['◎', 'Safe spending, with guardrails', 'The most you can spend with 85% confidence — and the rule that keeps it there when markets misbehave.'],
  ['▣', 'Guaranteed income floor', 'Cover groceries, housing and healthcare with income that can’t run out. Compare annuities with commissions shown.'],
  ['✉', 'A paycheck on the 1st', 'One deposit a month, with taxes and Medicare surcharges set aside, so April holds no surprises.'],
  ['⚑', 'Stress tests', 'A crash in year one, 4% inflation, long-term care, a 22% Social Security cut, losing a spouse.'],
  ['♥', 'Survivor planning', 'What happens to income, taxes and the house if one of you dies first — the “widow’s penalty”, quantified.'],
  ['＄', 'Medicare IRMAA guard', 'Keeps income under the surcharge cliffs two years ahead, where it pays to.'],
  ['☎', 'A planner when you want one', 'Fee-only CFP® professionals review your plan with you. They don’t earn more for any product.'],
]

export function Landing() {
  const [params] = useSearchParams()
  useEffect(() => { const r = params.get('ref'); if (r) sessionStorage.setItem('ref', r) }, [params])
  return (
    <>
      <section className="hero">
        <div className="wrap">
          <div>
            <span className="eyebrow">For the 4.1 million Americans turning 65 this year</span>
            <h1>Turn your savings into a <em>paycheck</em> for life.</h1>
            <p className="lede">Sundial tells you when to claim Social Security, which account to draw from each month, and how much you can safely spend — then pays you on the first of every month, taxes already handled.</p>
            <div className="cta-row">
              <Link className="btn btn-primary" to="/plan">Build my free plan</Link>
              <a className="btn btn-ghost" href="#waitlist">Join the waitlist</a>
            </div>
            <div className="reassure"><span>Your plan in 10 minutes</span><span>No account needed</span><span>We never sell your data</span></div>
          </div>
          <PayCard />
        </div>
      </section>

      <section className="band alt" aria-labelledby="why-h">
        <div className="wrap">
          <div className="band-head">
            <span className="eyebrow">Why now</span>
            <h2 id="why-h">The biggest retirement wave in history — and most of it without a pension.</h2>
            <p className="muted">Your parents had a pension that decided all of this for them. You have a 401(k), a Social Security statement and thirty years of decisions to make on your own.</p>
          </div>
          <div className="stats">
            <div><div className="stat-big">4.1M</div><p>Americans turn 65 every year through 2027 — over 11,000 a day.</p><a href="https://protectedincome.org/peak65/">Alliance for Lifetime Income</a></div>
            <div><div className="stat-big">78%</div><p>of scheduled retirement benefits payable once the trust fund runs dry in late 2032.</p><a href="https://www.ssa.gov/oact/tr/2026/II_D_project.html">2026 Trustees Report</a></div>
            <div><div className="stat-big">$461B</div><p>of annuities sold in 2025, a fourth straight record — people want guarantees.</p><a href="https://www.insurancebusinessmag.com/us/news/breaking-news/us-annuity-sales-hit-record-461-billion-as-indexed-products-surge--limra-565533.aspx">LIMRA via Insurance Business</a></div>
            <div><div className="stat-big">3.6%</div><p>Social Security COLA for 2027 — inflation is still a retirement risk.</p><a href="https://www.ssa.gov/blog/en/posts/2026-10-14.html">SSA</a></div>
          </div>
        </div>
      </section>

      <section className="band" id="how" aria-labelledby="dec-h">
        <div className="wrap">
          <div className="band-head">
            <span className="eyebrow">Three decisions worth six figures</span>
            <h2 id="dec-h">Get these right and the rest of retirement gets simple.</h2>
          </div>
          <div className="decision"><div className="n">1</div><div><h3>When to claim Social Security</h3><p className="muted">Claiming at 62 versus 70 changes your check by 77% — for life, and for your spouse after you. Sundial weighs every month for both of you against how long you’re each likely to live.</p></div></div>
          <div className="decision"><div className="n">2</div><div><h3>Which account to spend first</h3><p className="muted">Taxable, IRA, Roth — the order changes your lifetime tax bill by tens of thousands, and triggers (or dodges) Medicare surcharges and the tax on your Social Security.</p></div></div>
          <div className="decision"><div className="n">3</div><div><h3>How much income to guarantee</h3><p className="muted">Enough guaranteed income to cover the essentials lets you invest the rest without fear. Too much locks money away. Sundial sizes it to your budget, not a salesperson’s quota.</p></div></div>
        </div>
      </section>

      <section className="band alt" aria-labelledby="calc-h">
        <div className="wrap">
          <div className="band-head">
            <span className="eyebrow">Try it</span>
            <h2 id="calc-h">When should you claim? Ask the same engine that builds the full plan.</h2>
          </div>
          <QuickClaim />
        </div>
      </section>

      <section className="band" id="features" aria-labelledby="feat-h">
        <div className="wrap">
          <div className="band-head"><span className="eyebrow">Everything in one plan</span><h2 id="feat-h">A pension, rebuilt from the savings you already have.</h2></div>
          <div className="feature-grid">
            {FEATURES.map(([icon, t, d]) => <article className="card feature" key={t}><div className="icon" aria-hidden="true">{icon}</div><h3>{t}</h3><p>{d}</p></article>)}
          </div>
        </div>
      </section>

      <section className="band" aria-labelledby="prom-h">
        <div className="wrap">
          <div className="promise">
            <span className="eyebrow">Our promise</span>
            <h2 id="prom-h">The math doesn’t know who pays us.</h2>
            <ul>
              <li>Recommendations come from the engine, which never sees commissions or fees.</li>
              <li>Every annuity quote shows its commission. We cap ours at 2% and never sell surrender-charge-heavy products.</li>
              <li>Our planners are fee-only. They don’t earn more for anything you buy.</li>
              <li>Your data is yours: encrypted, never sold, deleted when you ask.</li>
            </ul>
          </div>
        </div>
      </section>

      <section className="band alt" id="pricing" aria-labelledby="price-h">
        <div className="wrap">
          <div className="band-head"><span className="eyebrow">Pricing</span><h2 id="price-h">Free to plan. Simple to keep.</h2></div>
          <div className="pricing">
            <div className="card plan-card"><h3>Plan</h3><div className="price">$0</div><p className="muted">The whole plan, forever free.</p>
              <ul><li>Social Security optimizer</li><li>Safe spending & stress tests</li><li>Tax-smart withdrawal order</li><li>Save and update your plan</li></ul></div>
            <div className="card plan-card featured"><h3>Paycheck</h3><div className="price">$19<span className="small muted">/month</span></div><p className="muted">or $190 a year</p>
              <ul><li>Automatic monthly paycheck</li><li>Tax & IRMAA withholding</li><li>RMDs and Roth conversions on schedule</li><li>Guardrail alerts after big market moves</li><li>Two planner sessions a year</li></ul></div>
            <div className="card plan-card"><h3>Managed</h3><div className="price">0.30%<span className="small muted">/year</span></div><p className="muted">of assets, capped at $3,000 a year</p>
              <ul><li>Everything in Paycheck</li><li>We manage the portfolio behind it</li><li>Dedicated fee-only CFP®</li><li>Estate & survivor coordination</li></ul></div>
          </div>
          <p className="hint" style={{ marginTop: 16 }}>Annuities are optional. If you buy one through Sundial, the insurer pays a disclosed commission (capped at 2%).</p>
        </div>
      </section>

      <section className="band faq" id="faq" aria-labelledby="faq-h">
        <div className="wrap">
          <span className="eyebrow">Questions</span>
          <h2 id="faq-h">Good questions, straight answers.</h2>
          <details><summary>Do I have to buy an annuity?</summary><p>No. Many plans don’t need one. Sundial shows how much guaranteed income would cover your essentials and what it would cost — you decide. Most people who do buy one use 10–30% of their savings.</p></details>
          <details><summary>What about the Social Security shortfall?</summary><p>The 2026 Trustees Report projects the retirement trust fund can pay 78% of scheduled benefits from late 2032 unless Congress acts. Every Sundial plan shows what happens to you in that scenario, and you can plan on it by default.</p></details>
          <details><summary>How accurate is this?</summary><p>Benefit rules, 2026 tax brackets, the senior deduction, IRMAA tiers and RMD tables are coded from the source rules, and every plan runs 1,000 simulations of markets and lifespans. Our <a href="/methodology">methodology</a> lists every assumption. Plans are estimates, not promises.</p></details>
          <details><summary>Is this financial advice?</summary><p>The planning tools are educational. Personalised advice and managed accounts are provided through a registered investment adviser, with a written agreement and fee-only planners.</p></details>
          <details><summary>How do you make money?</summary><p>Subscriptions for the paycheck service, an asset-based fee if we manage your portfolio, and disclosed insurer commissions if you choose an annuity through us. Never by selling your data.</p></details>
          <details><summary>Is my data safe?</summary><p>Your plan works without an account. If you save one, it’s encrypted in transit and at rest, never sold or shared for marketing, and deleted when you ask.</p></details>
        </div>
      </section>

      <section className="band alt" id="waitlist" aria-labelledby="wl-h">
        <div className="wrap quick">
          <div>
            <span className="eyebrow">Early access</span>
            <h2 id="wl-h">The paycheck service opens in waves.</h2>
            <p className="muted">Planning is free and open today. The automatic paycheck, managed portfolios and planner sessions open to the waitlist first — people retiring soonest go first.</p>
            <p className="muted">Invite a friend: each person who joins with your link moves you up ten places.</p>
          </div>
          <Waitlist />
        </div>
      </section>
    </>
  )
}
