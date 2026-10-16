import { StrictMode, createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Route, Routes } from 'react-router-dom'
import './styles.css'
import { api } from './lib/api'
import type { Me } from './lib/types'
import { Layout } from './components/Layout'
import { Landing } from './pages/Landing'
import { PlannerPage } from './pages/Planner'
import { LoginPage, SignupPage } from './pages/Auth'
import { AccountPage } from './pages/Account'
import { WaitlistPage } from './pages/Waitlist'
import { MethodologyPage, PrivacyPage, NotFound } from './pages/Static'

interface AuthState { me: Me | null; loading: boolean; setMe: (m: Me | null) => void }
const AuthContext = createContext<AuthState>({ me: null, loading: true, setMe: () => undefined })
export const useAuth = () => useContext(AuthContext)

function AuthProvider({ children }: { children: ReactNode }) {
  const [me, setMe] = useState<Me | null>(null)
  const [loading, setLoading] = useState(true)
  useEffect(() => {
    api.me().then((m) => setMe(m ?? null)).catch(() => setMe(null)).finally(() => setLoading(false))
  }, [])
  return <AuthContext.Provider value={{ me, loading, setMe }}>{children}</AuthContext.Provider>
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <AuthProvider>
      <BrowserRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
        <Routes>
          <Route element={<Layout />}>
            <Route path="/" element={<Landing />} />
            <Route path="/plan" element={<PlannerPage />} />
            <Route path="/login" element={<LoginPage />} />
            <Route path="/signup" element={<SignupPage />} />
            <Route path="/account" element={<AccountPage />} />
            <Route path="/waitlist/:code" element={<WaitlistPage />} />
            <Route path="/methodology" element={<MethodologyPage />} />
            <Route path="/privacy" element={<PrivacyPage />} />
            <Route path="*" element={<NotFound />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  </StrictMode>,
)
