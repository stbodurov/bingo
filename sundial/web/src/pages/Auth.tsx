import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { api, ApiError } from '../lib/api'
import { useAuth } from '../main'

function useNext() {
  const [p] = useSearchParams()
  const n = p.get('next') ?? '/account'
  return n.startsWith('/') && !n.startsWith('//') ? n : '/account'
}

export function LoginPage() {
  const { setMe } = useAuth()
  const nav = useNavigate()
  const next = useNext()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [err, setErr] = useState('')
  const submit = async (e: FormEvent) => {
    e.preventDefault(); setErr('')
    try { setMe(await api.login(email, password)); nav(next) } catch (x) { setErr(x instanceof ApiError ? x.message : 'Could not sign in.') }
  }
  return (
    <div className="wrap" style={{ maxWidth: 480 }}>
      <div className="page-head"><h1>Welcome back</h1></div>
      <form className="card" onSubmit={submit}>
        <div className="field"><label htmlFor="em">Email</label><input id="em" type="email" required autoComplete="username" value={email} onChange={(e) => setEmail(e.target.value)} /></div>
        <div className="field"><label htmlFor="pw">Password</label><input id="pw" type="password" required autoComplete="current-password" value={password} onChange={(e) => setPassword(e.target.value)} /></div>
        {err && <p className="error" role="alert">{err}</p>}
        <button className="btn btn-primary" type="submit" style={{ width: '100%' }}>Sign in</button>
        <p className="hint" style={{ textAlign: 'center' }}>New here? <Link to={`/signup?next=${encodeURIComponent(next)}`}>Create a free account</Link></p>
      </form>
    </div>
  )
}

export function SignupPage() {
  const { setMe } = useAuth()
  const nav = useNavigate()
  const next = useNext()
  const [f, setF] = useState({ name: '', email: '', password: '' })
  const [err, setErr] = useState<string[]>([])
  const submit = async (e: FormEvent) => {
    e.preventDefault(); setErr([])
    try { setMe(await api.signup(f.name, f.email, f.password)); nav(next) }
    catch (x) { setErr(x instanceof ApiError ? (x.details.length ? x.details : [x.message]) : ['Could not create the account.']) }
  }
  return (
    <div className="wrap" style={{ maxWidth: 480 }}>
      <div className="page-head"><div><h1>Keep your plan</h1><p className="muted" style={{ margin: '8px 0 0' }}>Free. Save plans, update them as life changes, and get alerts when tax rules move.</p></div></div>
      <form className="card" onSubmit={submit}>
        <div className="field"><label htmlFor="nm">First name</label><input id="nm" required maxLength={80} autoComplete="given-name" value={f.name} onChange={(e) => setF({ ...f, name: e.target.value })} /></div>
        <div className="field"><label htmlFor="em2">Email</label><input id="em2" type="email" required autoComplete="email" value={f.email} onChange={(e) => setF({ ...f, email: e.target.value })} /></div>
        <div className="field"><label htmlFor="pw2">Password</label><input id="pw2" type="password" required minLength={10} autoComplete="new-password" value={f.password} onChange={(e) => setF({ ...f, password: e.target.value })} />
          <p className="hint">At least 10 characters. A short phrase works well.</p></div>
        {err.length > 0 && <div className="error" role="alert">{err.map((x) => <div key={x}>{x}</div>)}</div>}
        <button className="btn btn-primary" type="submit" style={{ width: '100%' }}>Create account</button>
        <p className="hint" style={{ textAlign: 'center' }}>Already have one? <Link to={`/login?next=${encodeURIComponent(next)}`}>Sign in</Link></p>
      </form>
    </div>
  )
}
