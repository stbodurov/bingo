import { useEffect, useState } from 'react'
import { Link, Navigate } from 'react-router-dom'
import { api } from '../lib/api'
import { compact, pct } from '../lib/format'
import type { SavedPlan } from '../lib/types'
import { useAuth } from '../main'

export function AccountPage() {
  const { me, loading } = useAuth()
  const [plans, setPlans] = useState<SavedPlan[] | null>(null)
  useEffect(() => { if (me) api.plans().then(setPlans).catch(() => setPlans([])) }, [me])
  if (loading) return null
  if (!me) return <Navigate to="/login?next=/account" replace />
  const remove = async (id: number) => {
    if (!confirm('Delete this saved plan?')) return
    await api.deletePlan(id)
    setPlans((ps) => ps?.filter((p) => p.id !== id) ?? null)
  }
  return (
    <div className="wrap" style={{ maxWidth: 920 }}>
      <div className="page-head"><div><span className="eyebrow">Hi, {me.name}</span><h1>Your plans</h1></div><Link className="btn btn-primary" to="/plan">New plan</Link></div>
      {plans === null ? <div className="loading"><div className="sunspin" /></div> : plans.length === 0 ? (
        <div className="card"><p>No saved plans yet. <Link to="/plan">Build one</Link> and press “Save plan”.</p></div>
      ) : (
        <div className="grid g2">
          {plans.map((p) => (
            <section className="card" key={p.id}>
              <h3>{p.name}</h3>
              <p className="muted small">Updated {new Date(p.updatedAt).toLocaleDateString()}</p>
              <p>Confidence <strong>{p.successRate != null ? pct(p.successRate) : '—'}</strong> · safe spending <strong>{p.safeSpending != null ? compact(p.safeSpending) : '—'}</strong>/yr</p>
              <div style={{ display: 'flex', gap: 8 }}><Link className="btn btn-sun btn-sm" to={`/plan?id=${p.id}`}>Open</Link><button className="btn btn-ghost btn-sm" onClick={() => remove(p.id)}>Delete</button></div>
            </section>
          ))}
        </div>
      )}
    </div>
  )
}
