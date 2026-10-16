import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useEffect } from 'react'
import { useAuth } from '../main'
import { api } from '../lib/api'

export function Logo() {
  return (
    <svg viewBox="0 0 32 32" aria-hidden="true">
      <rect width="32" height="32" rx="16" fill="var(--dusk)" />
      <path d="M6 21a10 10 0 0 1 20 0" fill="var(--sun)" />
      <path d="M16 21 L22 11" stroke="var(--on-dusk)" strokeWidth="2.2" strokeLinecap="round" />
      <path d="M4 21h24" stroke="var(--on-dusk)" strokeWidth="2" strokeLinecap="round" />
    </svg>
  )
}

export function Layout() {
  const { me, setMe } = useAuth()
  const nav = useNavigate()
  const loc = useLocation()
  useEffect(() => {
    if (!loc.hash) window.scrollTo(0, 0)
  }, [loc.pathname, loc.hash])
  const logout = async () => {
    await api.logout().catch(() => undefined)
    setMe(null)
    nav('/')
  }
  return (
    <>
      <a className="sr-only" href="#main">Skip to content</a>
      <header className="site-head">
        <div className="wrap">
          <Link to="/" className="brand"><Logo />Sundial</Link>
          <nav className="nav" aria-label="Primary">
            <a className="hide-sm" href="/#how">How it works</a>
            <a className="hide-sm" href="/#pricing">Pricing</a>
            {me ? (
              <>
                <Link className="hide-sm" to="/account">My plans</Link>
                <button className="link" onClick={logout}>Sign out</button>
              </>
            ) : <Link className="hide-sm" to="/login">Sign in</Link>}
            <Link className="btn btn-primary btn-sm" to="/plan">Build my plan</Link>
          </nav>
        </div>
      </header>
      <main id="main"><Outlet /></main>
      <footer className="foot">
        <div className="wrap">
          <div>
            <Link to="/" className="brand"><Logo />Sundial</Link>
            <p style={{ marginTop: 12 }}>Retirement income, engineered. Planning tools are educational; personalised advice and managed accounts are offered through a registered investment adviser. Annuities are issued by licensed insurers; guarantees depend on the insurer's claims-paying ability.</p>
          </div>
          <div><strong>Product</strong><a href="/plan">Build a plan</a><a href="/#features">Features</a><a href="/#pricing">Pricing</a><a href="/methodology">Methodology</a></div>
          <div><strong>Company</strong><a href="/#waitlist">Join the waitlist</a><a href="/privacy">Privacy</a><a href="/#faq">FAQ</a></div>
        </div>
      </footer>
    </>
  )
}
