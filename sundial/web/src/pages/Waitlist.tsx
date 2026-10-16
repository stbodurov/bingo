import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api, type WaitlistStatus } from '../lib/api'

export function WaitlistPage() {
  const { code = '' } = useParams()
  const [s, setS] = useState<WaitlistStatus | null>(null)
  const [missing, setMissing] = useState(false)
  const [copied, setCopied] = useState(false)
  useEffect(() => { api.waitlistStatus(code).then(setS).catch(() => setMissing(true)) }, [code])
  if (missing) return <div className="wrap"><div className="page-head"><h1>We couldn’t find that spot</h1></div><Link to="/#waitlist">Join the waitlist</Link></div>
  if (!s) return <div className="loading"><div className="sunspin" /></div>
  const copy = async () => { try { await navigator.clipboard.writeText(s.shareUrl); setCopied(true) } catch { /* clipboard blocked */ } }
  return (
    <div className="wrap" style={{ maxWidth: 820 }}>
      <div className="page-head"><div><span className="eyebrow">You’re on the list</span><h1>Thanks{s.firstName ? `, ${s.firstName}` : ''}.</h1></div></div>
      <div className="grid g3">
        <div className="card stat"><div className="label">Your place</div><div className="value">#{s.position}</div><div className="sub">of {s.total}</div></div>
        <div className="card stat"><div className="label">Friends joined</div><div className="value">{s.referrals}</div><div className="sub">each moves you up 10 places</div></div>
        <div className="card stat"><div className="label">Meanwhile</div><div className="value">Free</div><div className="sub"><Link to="/plan">Build your plan now</Link></div></div>
      </div>
      <div className="card" style={{ marginTop: 20 }}>
        <h3>Share your link</h3>
        <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap' }}>
          <input readOnly value={s.shareUrl} aria-label="Your referral link" style={{ flex: 1, minWidth: 240 }} />
          <button className="btn btn-primary" onClick={copy}>{copied ? 'Copied' : 'Copy link'}</button>
        </div>
      </div>
    </div>
  )
}
