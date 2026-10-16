import { useEffect, useState, type ReactNode } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { api, ApiError } from '../lib/api'
import type { Household, Health, OneTime, Pension, Person, Plan, Sex } from '../lib/types'
import { Results } from '../components/Results'
import { useAuth } from '../main'

const DRAFT_KEY = 'sundial.draft.v1'
const thisYear = new Date().getFullYear()

function blankPerson(name: string, sex: Sex, birthYear: number): Person {
  return { name, birthYear, birthMonth: 6, birthDay: 15, sex, health: 'GOOD', pia: 2400, retireAge: 65, salary: 90000, claimAgeMonths: 0 }
}

export function blankHousehold(): Household {
  return {
    people: [blankPerson('', 'FEMALE', thisYear - 62)],
    accounts: { cash: 30000, taxable: 150000, taxableBasis: 0, traditional: 450000, roth: 60000 },
    essentialSpending: 60000, discretionarySpending: 20000, pensions: [], oneTimes: [], annuities: [],
    strategy: 'CONVENTIONAL', conversionBracket: 0, guardrails: true,
    assumptions: { inflation: 0.025, stockReturn: 0.085, stockVol: 0.17, bondReturn: 0.045, bondVol: 0.06, stockShare: 0.55,
      realDiscount: 0.018, socialSecurityHaircut: false, planToAge: 95, stateTaxRate: 0 },
  }
}

function loadDraft(): Household | null {
  try { const s = localStorage.getItem(DRAFT_KEY); return s ? JSON.parse(s) : null } catch { return null }
}
function saveDraft(h: Household) {
  try { localStorage.setItem(DRAFT_KEY, JSON.stringify(h)) } catch { /* private mode */ }
}

function Money({ id, label, value, onChange, hint, step = 1000 }: { id: string; label: string; value: number; onChange: (v: number) => void; hint?: ReactNode; step?: number }) {
  return (
    <div className="field">
      <label htmlFor={id}>{label}</label>
      <input id={id} type="number" inputMode="numeric" min={0} step={step} value={Number.isFinite(value) ? value : 0}
        onChange={(e) => onChange(e.target.value === '' ? 0 : +e.target.value)} />
      {hint && <p className="hint">{hint}</p>}
    </div>
  )
}

function Seg<T extends string>({ label, value, options, onChange }: { label: string; value: T; options: [T, string][]; onChange: (v: T) => void }) {
  return (
    <div className="field">
      <p className="label-text">{label}</p>
      <div className="seg" role="group" aria-label={label}>
        {options.map(([v, l]) => <button type="button" key={v} aria-pressed={value === v} onClick={() => onChange(v)}>{l}</button>)}
      </div>
    </div>
  )
}

const STEPS = ['You', 'Work & Social Security', 'Savings', 'Spending & income', 'Assumptions']

export function PlannerPage() {
  const [params] = useSearchParams()
  const nav = useNavigate()
  const { me } = useAuth()
  const [h, setH] = useState<Household>(() => loadDraft() ?? blankHousehold())
  const [step, setStep] = useState(0)
  const [plan, setPlan] = useState<Plan | null>(null)
  const [busy, setBusy] = useState(false)
  const [errors, setErrors] = useState<string[]>([])
  const [savedId, setSavedId] = useState<number | undefined>()

  useEffect(() => { saveDraft(h) }, [h])
  useEffect(() => {
    const id = params.get('id')
    const sample = params.get('sample')
    if (id && me) {
      api.plans().then((ps) => {
        const p = ps.find((x) => String(x.id) === id)
        if (p) { setH(p.household); setSavedId(p.id); build(p.household) }
      }).catch(() => undefined)
    } else if (sample === 'couple' || sample === 'single') {
      api.sample(sample).then((s) => { setH(s); build(s) }).catch(() => undefined)
    }
  }, [params, me]) // eslint-disable-line react-hooks/exhaustive-deps

  const build = async (house = h) => {
    setBusy(true); setErrors([])
    try {
      const p = await api.plan(house)
      setPlan(p)
      window.scrollTo({ top: 0, behavior: 'smooth' })
    } catch (e) {
      setErrors(e instanceof ApiError ? (e.details.length ? e.details : [e.message]) : ['Something went wrong. Please try again.'])
    }
    setBusy(false)
  }

  const people = h.people
  const married = people.length === 2
  const setPerson = (i: number, patch: Partial<Person>) => setH({ ...h, people: people.map((p, j) => j === i ? { ...p, ...patch } : p) })
  const setAcc = (k: keyof Household['accounts']) => (v: number) => setH({ ...h, accounts: { ...h.accounts, [k]: v } })
  const setAs = (patch: Partial<Household['assumptions']>) => setH({ ...h, assumptions: { ...h.assumptions, ...patch } })

  if (plan) {
    return <Results plan={plan} household={h} savedId={savedId} onSaved={setSavedId}
      onEdit={() => { setPlan(null); setStep(0) }} onRebuild={(next) => { setH(next); build(next) }} />
  }

  return (
    <div className="wrap" style={{ maxWidth: 920 }}>
      <div className="page-head">
        <div><span className="eyebrow">Your free plan</span><h1>Let’s build your paycheck.</h1>
          <p className="muted" style={{ margin: '8px 0 0' }}>About 10 minutes. Round numbers are fine — you can refine everything later.</p></div>
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
          <button className="btn btn-ghost btn-sm" onClick={() => nav('/plan?sample=couple')}>Try a sample couple</button>
          <button className="btn btn-ghost btn-sm" onClick={() => nav('/plan?sample=single')}>Try a sample single</button>
        </div>
      </div>
      <ol className="steps">
        {STEPS.map((s, i) => <li key={s} aria-current={i === step ? 'step' : undefined} className={i < step ? 'done' : ''}>
          <button className="link" style={{ color: 'inherit', textDecoration: 'none' }} onClick={() => setStep(i)}>{i + 1}. {s}</button></li>)}
      </ol>
      <div className="card">
        {step === 0 && (
          <>
            <Seg label="Who is this plan for?" value={married ? 'couple' : 'single'} options={[['single', 'Just me'], ['couple', 'Me and my spouse']]}
              onChange={(v) => setH({ ...h, people: v === 'couple' ? [people[0], people[1] ?? blankPerson('', people[0].sex === 'MALE' ? 'FEMALE' : 'MALE', people[0].birthYear - 2)] : [people[0]] })} />
            {people.map((p, i) => (
              <div className="person-card" key={i}>
                <h3>{i === 0 ? 'You' : 'Your spouse'}</h3>
                <div className="row">
                  <div className="field"><label htmlFor={`n${i}`}>First name</label><input id={`n${i}`} value={p.name} maxLength={40} onChange={(e) => setPerson(i, { name: e.target.value })} placeholder={i === 0 ? 'Alex' : 'Sam'} /></div>
                  <div className="field"><label htmlFor={`b${i}`}>Date of birth</label>
                    <input id={`b${i}`} type="date" value={`${p.birthYear}-${String(p.birthMonth).padStart(2, '0')}-${String(p.birthDay).padStart(2, '0')}`}
                      onChange={(e) => { const [y, m, d] = e.target.value.split('-').map(Number); if (y) setPerson(i, { birthYear: y, birthMonth: m, birthDay: d }) }} /></div>
                </div>
                <Seg label="Sex at birth (used for life expectancy)" value={p.sex} options={[['FEMALE', 'Female'], ['MALE', 'Male']]} onChange={(v) => setPerson(i, { sex: v })} />
                <Seg<Health> label="Health compared with others your age" value={p.health}
                  options={[['EXCELLENT', 'Excellent'], ['GOOD', 'Good'], ['AVERAGE', 'Average'], ['POOR', 'Poor']]} onChange={(v) => setPerson(i, { health: v })} />
              </div>
            ))}
          </>
        )}
        {step === 1 && people.map((p, i) => (
          <div className="person-card" key={i}>
            <h3>{p.name || (i === 0 ? 'You' : 'Your spouse')}</h3>
            <div className="row">
              <Money id={`pia${i}`} label="Monthly Social Security at full retirement age" value={p.pia} step={10} onChange={(v) => setPerson(i, { pia: v })}
                hint={<>Your “benefit at full retirement age” from <a href="https://www.ssa.gov/myaccount/" target="_blank" rel="noreferrer">ssa.gov/myaccount</a>.</>} />
              <div className="field"><label htmlFor={`claim${i}`}>When to claim</label>
                <select id={`claim${i}`} value={p.claimAgeMonths} onChange={(e) => setPerson(i, { claimAgeMonths: +e.target.value })}>
                  <option value={0}>Let Sundial choose the best age</option>
                  {Array.from({ length: 9 }, (_, k) => 62 + k).map((a) => <option key={a} value={a * 12}>At {a}</option>)}
                </select></div>
            </div>
            <div className="row">
              <Money id={`sal${i}`} label="Current salary (0 if retired)" value={p.salary} onChange={(v) => setPerson(i, { salary: v })} />
              <div className="field"><label htmlFor={`ret${i}`}>Age you’ll stop working</label>
                <input id={`ret${i}`} type="number" min={0} max={80} value={p.retireAge} onChange={(e) => setPerson(i, { retireAge: +e.target.value })} />
                <p className="hint">Enter 0 if already retired.</p></div>
            </div>
          </div>
        ))}
        {step === 2 && (
          <>
            <p className="muted">Combined for the household. Include 401(k)s and 403(b)s with IRAs.</p>
            <div className="row">
              <Money id="cash" label="Cash & savings" value={h.accounts.cash} onChange={setAcc('cash')} />
              <Money id="taxable" label="Brokerage (taxable) accounts" value={h.accounts.taxable} onChange={setAcc('taxable')} />
              <Money id="basis" label="…of which you originally invested" value={h.accounts.taxableBasis} onChange={setAcc('taxableBasis')} hint="Cost basis. Leave 0 to assume 70%." />
            </div>
            <div className="row">
              <Money id="trad" label="Traditional IRA / 401(k)" value={h.accounts.traditional} onChange={setAcc('traditional')} />
              <Money id="roth" label="Roth IRA / Roth 401(k)" value={h.accounts.roth} onChange={setAcc('roth')} />
            </div>
            <div className="field"><label htmlFor="stocks">Stocks vs. bonds: {Math.round(h.assumptions.stockShare * 100)}% stocks</label>
              <input id="stocks" type="range" min={0} max={100} step={5} value={h.assumptions.stockShare * 100} onChange={(e) => setAs({ stockShare: +e.target.value / 100 })} /></div>
          </>
        )}
        {step === 3 && (
          <>
            <div className="row">
              <Money id="ess" label="Essential spending per year" value={h.essentialSpending} onChange={(v) => setH({ ...h, essentialSpending: v })}
                hint="Housing, food, utilities, insurance, healthcare and Medicare premiums." />
              <Money id="disc" label="Everything else per year" value={h.discretionarySpending} onChange={(v) => setH({ ...h, discretionarySpending: v })}
                hint="Travel, gifts, hobbies, dining out. Usually tapers in your late 70s." />
            </div>
            <h3 style={{ marginTop: 10 }}>Pensions</h3>
            {h.pensions.map((pe, k) => (
              <div className="row" key={k}>
                <Money id={`pm${k}`} label="Monthly pension" value={pe.monthly} step={50} onChange={(v) => setH({ ...h, pensions: h.pensions.map((x, j) => j === k ? { ...x, monthly: v } : x) })} />
                <div className="field"><label htmlFor={`pa${k}`}>Starts at age</label><input id={`pa${k}`} type="number" min={50} max={80} value={pe.startAge}
                  onChange={(e) => setH({ ...h, pensions: h.pensions.map((x, j) => j === k ? { ...x, startAge: +e.target.value } : x) })} /></div>
                <div className="field"><label htmlFor={`ps${k}`}>Survivor keeps</label><select id={`ps${k}`} value={pe.survivorPct}
                  onChange={(e) => setH({ ...h, pensions: h.pensions.map((x, j) => j === k ? { ...x, survivorPct: +e.target.value } : x) })}>
                  {[0, 0.5, 0.75, 1].map((s) => <option key={s} value={s}>{s * 100}%</option>)}</select></div>
                <div className="field" style={{ alignSelf: 'end' }}>
                  <label className="check"><input type="checkbox" checked={pe.cola} onChange={(e) => setH({ ...h, pensions: h.pensions.map((x, j) => j === k ? { ...x, cola: e.target.checked } : x) })} /> Inflation-adjusted</label>
                  <button className="link small" onClick={() => setH({ ...h, pensions: h.pensions.filter((_, j) => j !== k) })}>Remove</button></div>
              </div>
            ))}
            <button className="btn btn-ghost btn-sm" onClick={() => setH({ ...h, pensions: [...h.pensions, { owner: 0, monthly: 1000, startAge: 65, cola: false, survivorPct: 0.5 } as Pension] })}>+ Add a pension</button>
            <h3 style={{ marginTop: 22 }}>Big one-time expenses</h3>
            {h.oneTimes.map((o, k) => (
              <div className="row" key={k}>
                <div className="field"><label htmlFor={`ol${k}`}>What</label><input id={`ol${k}`} value={o.label} maxLength={40} onChange={(e) => setH({ ...h, oneTimes: h.oneTimes.map((x, j) => j === k ? { ...x, label: e.target.value } : x) })} /></div>
                <div className="field"><label htmlFor={`oy${k}`}>Year</label><input id={`oy${k}`} type="number" min={thisYear} max={thisYear + 40} value={o.year} onChange={(e) => setH({ ...h, oneTimes: h.oneTimes.map((x, j) => j === k ? { ...x, year: +e.target.value } : x) })} /></div>
                <Money id={`oa${k}`} label="Amount (today’s $)" value={o.amount} onChange={(v) => setH({ ...h, oneTimes: h.oneTimes.map((x, j) => j === k ? { ...x, amount: v } : x) })}
                  hint={<button className="link small" onClick={() => setH({ ...h, oneTimes: h.oneTimes.filter((_, j) => j !== k) })}>Remove</button>} />
              </div>
            ))}
            <button className="btn btn-ghost btn-sm" onClick={() => setH({ ...h, oneTimes: [...h.oneTimes, { year: thisYear + 2, amount: 30000, label: 'New car' } as OneTime] })}>+ Add an expense</button>
          </>
        )}
        {step === 4 && (
          <>
            <p className="muted">Sensible defaults are already set. Change them if you have a reason to.</p>
            <div className="row">
              <div className="field"><label htmlFor="inf">Inflation: {(h.assumptions.inflation * 100).toFixed(1)}%</label>
                <input id="inf" type="range" min={1} max={5} step={0.1} value={h.assumptions.inflation * 100} onChange={(e) => setAs({ inflation: +e.target.value / 100 })} /></div>
              <div className="field"><label htmlFor="plan">Plan until age {h.assumptions.planToAge}</label>
                <input id="plan" type="range" min={85} max={105} value={h.assumptions.planToAge} onChange={(e) => setAs({ planToAge: +e.target.value })} /></div>
              <div className="field"><label htmlFor="state">State income tax: {(h.assumptions.stateTaxRate * 100).toFixed(1)}%</label>
                <input id="state" type="range" min={0} max={11} step={0.25} value={h.assumptions.stateTaxRate * 100} onChange={(e) => setAs({ stateTaxRate: +e.target.value / 100 })} /></div>
            </div>
            <label className="check"><input type="checkbox" checked={h.assumptions.socialSecurityHaircut} onChange={(e) => setAs({ socialSecurityHaircut: e.target.checked })} />
              Plan for Social Security paying only 78% from 2033 (Trustees’ depletion scenario)</label>
            <label className="check" style={{ marginTop: 10 }}><input type="checkbox" checked={h.guardrails} onChange={(e) => setH({ ...h, guardrails: e.target.checked })} />
              Use spending guardrails (trim extras 10% after bad years, restore after good ones)</label>
          </>
        )}
        {errors.length > 0 && <div className="error" role="alert" style={{ marginTop: 16 }}>Please check a few things:<ul>{errors.map((e) => <li key={e}>{e}</li>)}</ul></div>}
        <div className="wizard-nav">
          <button className="btn btn-ghost" disabled={step === 0} onClick={() => setStep(step - 1)}>Back</button>
          {step < STEPS.length - 1
            ? <button className="btn btn-primary" onClick={() => setStep(step + 1)}>Continue</button>
            : <button className="btn btn-sun" disabled={busy} onClick={() => build()}>{busy ? 'Running 1,000 scenarios…' : 'Build my plan'}</button>}
        </div>
      </div>
      {step < STEPS.length - 1 && <p className="hint" style={{ textAlign: 'center' }}>In a hurry? <button className="link" disabled={busy} onClick={() => build()}>{busy ? 'Building…' : 'Build my plan with what I’ve entered'}</button></p>}
    </div>
  )
}
