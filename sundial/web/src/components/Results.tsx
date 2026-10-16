import { useMemo, useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { api, ApiError } from '../lib/api'
import { age, ageShort, compact, confidenceWord, monthName, pct, usd } from '../lib/format'
import type { Household, Plan } from '../lib/types'
import { ClaimHeat, FanChart, HBars, IncomeArea, Legend, Meter, SERIES } from './charts'
import { useAuth } from '../main'

const TABS = ['Overview', 'Social Security', 'Confidence', 'Taxes', 'Guaranteed income', 'Paycheck', 'Stress tests', 'Year by year'] as const
type Tab = typeof TABS[number]

function Confidence({ p }: { p: number }) {
  const cls = p >= 0.8 ? 'b-good' : p >= 0.65 ? 'b-warn' : 'b-crit'
  const icon = p >= 0.8 ? '✓' : p >= 0.65 ? '!' : '✕'
  return <span className={`badge ${cls}`}><span aria-hidden="true">{icon}</span>{confidenceWord(p)}</span>
}

function download(plan: Plan) {
  const head = ['year', 'ages', 'wages', 'social_security', 'pensions', 'annuities', 'rmd', 'from_cash', 'from_taxable', 'from_ira', 'from_roth',
    'roth_conversion', 'taxes', 'irmaa', 'spending', 'portfolio', 'ira', 'roth', 'taxable_and_cash']
  const rows = plan.years.map((y) => [y.year, y.ages.join('/'), y.wages, y.socialSecurity, y.pensions, y.annuities, y.rmd, y.fromCash, y.fromTaxable,
    y.fromTraditional, y.fromRoth, y.conversion, y.taxes, y.irmaa, y.spending, y.portfolio, y.traditional, y.roth, y.taxable]
    .map((v) => typeof v === 'number' ? Math.round(v) : v).join(','))
  const blob = new Blob([[head.join(','), ...rows].join('\n')], { type: 'text/csv' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = 'sundial-plan.csv'
  a.click()
  URL.revokeObjectURL(a.href)
}

function SavePanel({ household, savedId, onSaved }: { household: Household; savedId?: number; onSaved: (id: number) => void }) {
  const { me } = useAuth()
  const [name, setName] = useState('Our retirement plan')
  const [msg, setMsg] = useState('')
  if (!me) return <div className="notice">Want to keep this plan and get alerts when tax rules change? <Link to="/signup?next=/plan">Create a free account</Link> — your inputs are kept on this device until you do.</div>
  const save = async (e: FormEvent) => {
    e.preventDefault()
    try { const s = await api.savePlan(name, household, savedId); onSaved(s.id); setMsg('Saved.') } catch (x) { setMsg(x instanceof ApiError ? x.message : 'Could not save.') }
  }
  return (
    <form className="notice" onSubmit={save} style={{ display: 'flex', gap: 10, alignItems: 'end', flexWrap: 'wrap' }}>
      <div style={{ flex: 1, minWidth: 220 }}><label htmlFor="pname">Plan name</label><input id="pname" value={name} maxLength={120} onChange={(e) => setName(e.target.value)} /></div>
      <button className="btn btn-primary btn-sm" type="submit">{savedId ? 'Update saved plan' : 'Save plan'}</button>
      {msg && <span role="status" className="small">{msg}</span>}
    </form>
  )
}

function AdvisorForm() {
  const { me } = useAuth()
  const [f, setF] = useState({ name: me?.name ?? '', email: me?.email ?? '', phone: '', topic: 'Review my plan', message: '' })
  const [done, setDone] = useState(false)
  const [err, setErr] = useState('')
  const submit = async (e: FormEvent) => {
    e.preventDefault(); setErr('')
    try { await api.advisor(f); setDone(true) } catch (x) { setErr(x instanceof ApiError ? [x.message, ...x.details].join(' ') : 'Could not send.') }
  }
  if (done) return <p className="ok">Thanks — a fee-only planner will email you within one business day.</p>
  return (
    <form onSubmit={submit}>
      <div className="row">
        <div className="field"><label htmlFor="af-n">Name</label><input id="af-n" required value={f.name} onChange={(e) => setF({ ...f, name: e.target.value })} /></div>
        <div className="field"><label htmlFor="af-e">Email</label><input id="af-e" type="email" required value={f.email} onChange={(e) => setF({ ...f, email: e.target.value })} /></div>
        <div className="field"><label htmlFor="af-t">Topic</label><select id="af-t" value={f.topic} onChange={(e) => setF({ ...f, topic: e.target.value })}>
          {['Review my plan', 'Social Security timing', 'Roth conversions', 'Annuities', 'Managing the portfolio'].map((t) => <option key={t}>{t}</option>)}</select></div>
      </div>
      {err && <p className="error">{err}</p>}
      <button className="btn btn-primary btn-sm" type="submit">Request a call</button>
    </form>
  )
}

function QuoteForm({ kind, premium, age: a }: { kind: string; premium: number; age: number }) {
  const { me } = useAuth()
  const [email, setEmail] = useState(me?.email ?? '')
  const [state, setState] = useState('')
  const [done, setDone] = useState(false)
  const [err, setErr] = useState('')
  const submit = async (e: FormEvent) => {
    e.preventDefault(); setErr('')
    try { await api.quoteRequest({ email, kind, premium: Math.round(premium), age: a, state: state || undefined }); setDone(true) }
    catch (x) { setErr(x instanceof ApiError ? [x.message, ...x.details].join(' ') : 'Could not send.') }
  }
  if (done) return <p className="ok">Request received — you’ll get firm quotes from A-rated insurers by email, with commissions shown.</p>
  return (
    <form onSubmit={submit} style={{ display: 'flex', gap: 10, alignItems: 'end', flexWrap: 'wrap' }}>
      <div style={{ flex: 2, minWidth: 220 }}><label htmlFor="q-e">Email for firm quotes</label><input id="q-e" type="email" required value={email} onChange={(e) => setEmail(e.target.value)} /></div>
      <div style={{ flex: 1, minWidth: 90 }}><label htmlFor="q-s">State</label><input id="q-s" maxLength={2} value={state} onChange={(e) => setState(e.target.value.toUpperCase())} placeholder="CA" /></div>
      <button className="btn btn-sun btn-sm" type="submit">Get firm quotes</button>
      {err && <p className="error" style={{ width: '100%' }}>{err}</p>}
    </form>
  )
}

export function Results({ plan, household, savedId, onSaved, onEdit, onRebuild }:
  { plan: Plan; household: Household; savedId?: number; onSaved: (id: number) => void; onEdit: () => void; onRebuild: (h: Household) => void }) {
  const [tab, setTab] = useState<Tab>('Overview')
  const s = plan.summary
  const years = plan.years.map((y) => y.year)
  const series = useMemo(() => [
    { key: 'ss', label: 'Social Security', color: SERIES[0], values: plan.years.map((y) => y.socialSecurity) },
    { key: 'pen', label: 'Pensions', color: SERIES[1], values: plan.years.map((y) => y.pensions) },
    { key: 'ann', label: 'Annuities', color: SERIES[2], values: plan.years.map((y) => y.annuities) },
    { key: 'wage', label: 'Work', color: SERIES[3], values: plan.years.map((y) => y.wages) },
    { key: 'port', label: 'From savings', color: SERIES[4], values: plan.years.map((y) => y.fromCash + y.fromTaxable + y.fromTraditional + y.fromRoth + y.rmd) },
  ].filter((x) => x.values.some((v) => v > 0)), [plan])
  const spending = plan.years.map((y) => y.spending + y.taxes + y.irmaa)
  const rec = plan.strategies.find((x) => x.recommended)!
  const firstName = (i: number) => plan.people[i]?.name || (i === 0 ? 'You' : 'Spouse')

  return (
    <div className="wrap">
      <div className="page-head">
        <div><span className="eyebrow">Your retirement paycheck</span><h1>{plan.generatedFor ? `${plan.generatedFor}’s plan` : 'Your plan'}</h1></div>
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
          <button className="btn btn-ghost btn-sm" onClick={onEdit}>Edit inputs</button>
          <button className="btn btn-ghost btn-sm" onClick={() => download(plan)}>Download CSV</button>
          <button className="btn btn-ghost btn-sm" onClick={() => window.print()}>Print</button>
        </div>
      </div>

      <div className="grid g4">
        <div className="card stat"><div className="label">Monthly paycheck</div><div className="value">{usd(s.monthlyPaycheck)}</div>
          <div className="sub">after tax, from {s.firstRetirementYear}, in today’s dollars</div></div>
        <div className="card stat"><div className="label">Confidence</div><div className="value">{pct(s.successRate)}</div>
          <div className="sub"><Confidence p={s.successRate} /> money lasts as long as you do</div></div>
        <div className="card stat"><div className="label">Safe spending</div><div className="value">{compact(s.safeSpending)}</div>
          <div className="sub">a year at 85% confidence · you planned {compact(s.plannedSpending)}</div></div>
        <div className="card stat"><div className="label">Guaranteed</div><div className="value">{pct(s.guaranteedShare)}</div>
          <div className="sub">of spending paid by Social Security, pensions & annuities</div></div>
      </div>

      <div style={{ margin: '20px 0' }}><SavePanel household={household} savedId={savedId} onSaved={onSaved} /></div>

      <div className="tabs" role="tablist" aria-label="Plan sections">
        {TABS.map((t) => <button key={t} role="tab" aria-selected={tab === t} onClick={() => setTab(t)}>{t}</button>)}
      </div>

      {tab === 'Overview' && (
        <div className="grid g2" style={{ alignItems: 'start' }}>
          <section className="card">
            <h3>What to do</h3>
            {plan.insights.map((i) => (
              <div className="insight" key={i.kind}>
                <div className="icon" aria-hidden="true">{{ claiming: '☀', tax: '⇅', floor: '▣', spending: '◎', haircut: '⚑', irmaa: '＄' }[i.kind] ?? '•'}</div>
                <div><h4>{i.title}</h4><p>{i.body}</p></div>
              </div>
            ))}
          </section>
          <section className="card">
            <h3>Where your income comes from</h3>
            <Legend items={[...series.map((x) => ({ label: x.label, color: x.color })), { label: 'Spending + taxes', color: 'var(--ink)', dashed: true }]} />
            <IncomeArea years={years} series={series} spending={spending} />
            <p className="hint">Today’s dollars. Hover to see each year.</p>
            <details className="data"><summary>Talk it through with a fee-only planner</summary><div style={{ marginTop: 12 }}><AdvisorForm /></div></details>
          </section>
        </div>
      )}

      {tab === 'Social Security' && (
        <div className="grid" style={{ gap: 20 }}>
          <div className={`grid ${plan.people.length > 1 ? 'g2' : ''}`}>
            {plan.people.map((p, i) => (
              <section className="card" key={i}>
                <span className="eyebrow">{firstName(i)}</span>
                <h3>Claim at {age(plan.claiming.recommended[i] ?? p.recommendedClaimMonths)}</h3>
                <p className="stat"><span className="value">{usd(p.monthlyAtClaim)}</span><span className="muted"> a month in today’s dollars</span></p>
                <p className="muted">Full retirement age {age(p.fraMonths)}. Likely to live to about {Math.round(p.lifeExpectancyAge)}; 1 in 4 people like {i === 0 && plan.people.length === 1 ? 'you' : 'them'} reach {Math.round(p.oneInFourAge)}. RMDs start at {p.rmdStartAge}.</p>
              </section>
            ))}
          </div>
          <section className="card">
            <h3>Expected lifetime benefits by strategy</h3>
            <HBars items={plan.claiming.compare.map((r) => ({ label: `${r.label.replace('As early as possible', 'Earliest').replace('At full retirement age', 'Full retirement age')} · ${r.claimMonths.map(ageShort).join(' & ')}`, value: r.expectedPv }))} highlight={0} />
            <p className="hint">Today’s dollars, weighted month by month by the chance each of you is alive, including spousal and survivor benefits. Waiting until 70 pays off if you live past about {plan.claiming.breakevenAge}.</p>
            <p className="notice" style={{ marginTop: 12 }}>If Social Security pays 78% from 2033, these benefits fall by about <strong>{usd(plan.claiming.haircutPvLoss)}</strong>. {household.assumptions.socialSecurityHaircut ? 'Your plan already assumes this.' : <button className="link" onClick={() => onRebuild({ ...household, assumptions: { ...household.assumptions, socialSecurityHaircut: true } })}>Rebuild my plan assuming the cut</button>}</p>
          </section>
          {plan.people.length > 1 && plan.claiming.grid.length > 1 && (
            <section className="card"><h3>Every combination of claiming ages</h3>
              <ClaimHeat grid={plan.claiming.grid} start={plan.claiming.gridStart} best={plan.claiming.recommended} step={3} /></section>
          )}
        </div>
      )}

      {tab === 'Confidence' && (
        <div className="grid g2" style={{ alignItems: 'start' }}>
          <section className="card">
            <h3>Your savings over time, across 1,000 futures</h3>
            <Legend items={[{ label: 'Typical (median)', color: 'var(--s1)' }, { label: '10th–90th percentile', color: 'var(--s1-soft)' }]} />
            <FanChart years={plan.monteCarlo.fanYears} p10={plan.monteCarlo.fanP10} p50={plan.monteCarlo.fanP50} p90={plan.monteCarlo.fanP90} />
            <p className="hint">Today’s dollars. Each future draws market returns and how long each of you lives.</p>
          </section>
          <section className="card">
            <h3>{pct(plan.monteCarlo.successRate)} of futures never run short</h3>
            <p className="muted">In the futures that do, money most often runs short after 85 — {pct(plan.monteCarlo.probRunOutBefore85)} run short earlier.</p>
            <table><tbody>
              <tr><td>Money left at the end, typical</td><td className="r">{usd(plan.monteCarlo.medianEndingReal)}</td></tr>
              <tr><td>Money left, poor markets (10th)</td><td className="r">{usd(plan.monteCarlo.p10EndingReal)}</td></tr>
              <tr><td>Money left, good markets (90th)</td><td className="r">{usd(plan.monteCarlo.p90EndingReal)}</td></tr>
              <tr><td>Safe spending at 85% confidence</td><td className="r">{usd(s.safeSpending)} / yr</td></tr>
            </tbody></table>
            <h4 style={{ marginTop: 18 }}>Guardrails {household.guardrails ? 'are on' : 'are off'}</h4>
            <p className="muted">After a bad year that pushes your withdrawal rate 20% above where it started, Sundial trims travel and extras by 10%; after a good year it gives it back. Essentials are never cut.</p>
          </section>
        </div>
      )}

      {tab === 'Taxes' && (
        <div className="grid" style={{ gap: 20 }}>
          <section className="card">
            <h3>Recommended: {rec.label.charAt(0).toLowerCase() + rec.label.slice(1)}</h3>
            <div className="table-wrap"><table>
              <thead><tr><th>Withdrawal strategy</th><th className="r">Confidence</th><th className="r">Lifetime taxes</th><th className="r">Medicare surcharges</th><th className="r">Left to heirs after tax</th></tr></thead>
              <tbody>{plan.strategies.map((st) => (
                <tr key={st.key} className={st.recommended ? 'highlight' : ''}><td>{st.recommended ? '★ ' : ''}{st.label}</td><td className="r">{pct(st.successRate)}</td>
                  <td className="r">{usd(st.lifetimeTax)}</td><td className="r">{usd(st.lifetimeIrmaa)}</td><td className="r">{usd(st.afterTaxLegacy)}</td></tr>))}</tbody>
            </table></div>
            <p className="hint">Today’s dollars through age {household.assumptions.planToAge}, median markets. Heirs are assumed to pay 24% on inherited IRA money.</p>
          </section>
          <section className="card">
            <h3>Roth conversions and your bracket, year by year</h3>
            <div className="table-wrap"><table>
              <thead><tr><th>Year</th><th>Ages</th><th className="r">IRA withdrawals</th><th className="r">Roth conversion</th><th className="r">Taxes</th><th className="r">IRMAA</th><th className="r">Marginal rate</th></tr></thead>
              <tbody>{plan.years.slice(0, 18).map((y) => (
                <tr key={y.year}><td>{y.year}</td><td>{y.ages.join(' / ')}</td><td className="r">{usd(y.fromTraditional + y.rmd)}</td><td className="r">{y.conversion ? usd(y.conversion) : '—'}</td>
                  <td className="r">{usd(y.taxes)}</td><td className="r">{y.irmaa ? usd(y.irmaa) : '—'}</td><td className="r">{pct(y.marginalRate)}</td></tr>))}</tbody>
            </table></div>
          </section>
        </div>
      )}

      {tab === 'Guaranteed income' && (
        <div className="grid" style={{ gap: 20 }}>
          <section className="card">
            <h3>Essentials covered by income that can’t run out</h3>
            <Meter value={plan.floor.guaranteed / Math.max(1, plan.floor.essential)} label="Share of essential spending covered" />
            <p style={{ marginTop: 10 }}><strong>{usd(plan.floor.guaranteed)}</strong> of <strong>{usd(plan.floor.essential)}</strong> a year ({pct(Math.min(1, plan.floor.guaranteed / Math.max(1, plan.floor.essential)))}) once Social Security starts.</p>
            {plan.floor.gap > 500
              ? <p className="muted">Closing the {usd(plan.floor.gap)} gap with a lifetime annuity would cost about {usd(plan.floor.premiumToCloseGap)}. Sundial suggests no more than {usd(plan.floor.recommendedPremium)} (at most 35% of savings). Confidence: {pct(plan.floor.successWithout)} → {pct(plan.floor.successWith)}.</p>
              : <p className="muted">Your essentials are already covered. An annuity is optional — consider one only for longevity insurance past 85.</p>}
          </section>
          <section className="card">
            <h3>Indicative quotes for {usd(plan.floor.quotes[0]?.premium ?? 0)}</h3>
            <div className="table-wrap"><table>
              <thead><tr><th>Product</th><th className="r">Monthly income</th><th className="r">Starts at</th><th className="r">Payout rate</th><th className="r">Commission</th></tr></thead>
              <tbody>{plan.floor.quotes.map((q) => (
                <tr key={q.kind}><td><strong>{q.label}</strong><div className="small muted">{q.note}</div></td><td className="r">{usd(q.monthly)}</td><td className="r">{q.startAge}</td>
                  <td className="r">{pct(q.payoutRate, 1)}</td><td className="r">{q.commissionPct}%</td></tr>))}</tbody>
            </table></div>
            <p className="hint">Priced by Sundial’s model from insurer mortality tables and mid-2026 bond yields. Firm quotes come from licensed, A-rated insurers; guarantees depend on the insurer.</p>
            <QuoteForm kind={plan.floor.quotes[0]?.kind ?? 'SPIA_LIFE'} premium={plan.floor.quotes[0]?.premium ?? 100000} age={Math.max(65, plan.people[0].age)} />
          </section>
        </div>
      )}

      {tab === 'Paycheck' && (
        <>
          <p className="muted">Your first twelve months in retirement, deposit by deposit (today’s dollars).</p>
          <div className="grid g3">
            {plan.paycheck.map((m) => (
              <section className="month" key={m.month}>
                <h4><span>{monthName(m.month)}</span><span className="num">{usd(m.total)}</span></h4>
                <ul>{m.lines.map((l, i) => (
                  <li key={i}><span><span className="small muted">{l.date.slice(5)}</span> {l.source}{l.note && <span className="small muted"> · {l.note}</span>}</span><b className="num">{usd(l.amount)}</b></li>
                ))}</ul>
              </section>
            ))}
          </div>
        </>
      )}

      {tab === 'Stress tests' && (
        <div className="grid g2">
          {plan.scenarios.map((sc) => (
            <section className={`card scenario${sc.success ? '' : ' bad'}`} key={sc.key}>
              <div style={{ display: 'flex', justifyContent: 'space-between', gap: 10 }}><h3>{sc.label}</h3>
                <span className={`badge ${sc.success ? 'b-good' : 'b-crit'}`}><span aria-hidden="true">{sc.success ? '✓' : '✕'}</span>{sc.success ? 'Plan holds' : `Short at ${sc.runsOutAge}`}</span></div>
              <p className="muted">{sc.description}</p>
              <p className="small">Money left at the end: <strong>{usd(sc.endingReal)}</strong> · lifetime spending {compact(sc.lifetimeSpending)}</p>
            </section>
          ))}
        </div>
      )}

      {tab === 'Year by year' && (
        <section className="card table-wrap">
          <table>
            <thead><tr><th>Year</th><th>Ages</th><th className="r">Work</th><th className="r">Social Security</th><th className="r">Pensions</th><th className="r">From savings</th>
              <th className="r">Taxes</th><th className="r">Spending</th><th className="r">Savings left</th></tr></thead>
            <tbody>{plan.years.map((y) => (
              <tr key={y.year}><td>{y.year}</td><td>{y.ages.join(' / ')}</td><td className="r">{y.wages ? compact(y.wages) : '—'}</td><td className="r">{compact(y.socialSecurity)}</td>
                <td className="r">{y.pensions ? compact(y.pensions) : '—'}</td><td className="r">{compact(y.fromCash + y.fromTaxable + y.fromTraditional + y.fromRoth + y.rmd)}</td>
                <td className="r">{compact(y.taxes + y.irmaa)}</td><td className="r">{compact(y.spending)}</td><td className="r">{compact(y.portfolio)}</td></tr>))}</tbody>
          </table>
          <p className="hint">Today’s dollars, median markets, everyone living to {household.assumptions.planToAge}.</p>
        </section>
      )}
    </div>
  )
}
