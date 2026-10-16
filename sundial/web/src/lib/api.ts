import type { Household, Me, Plan, Quote, SavedPlan } from './types'

export class ApiError extends Error {
  constructor(public status: number, message: string, public details: string[] = []) { super(message) }
}

function xsrf(): string {
  const m = document.cookie.match(/(?:^|; )XSRF-TOKEN=([^;]*)/)
  return m ? decodeURIComponent(m[1]) : ''
}

let primed = false
async function prime() {
  if (primed && xsrf()) return
  await fetch('/api/auth/csrf', { credentials: 'same-origin' }).catch(() => undefined)
  primed = true
}

async function request<T>(method: string, url: string, body?: unknown): Promise<T> {
  if (method !== 'GET') await prime()
  const res = await fetch(url, {
    method,
    credentials: 'same-origin',
    headers: { 'Content-Type': 'application/json', 'X-XSRF-TOKEN': xsrf() },
    body: body === undefined ? undefined : JSON.stringify(body),
  })
  if (res.status === 204) return undefined as T
  const text = await res.text()
  const data = text ? JSON.parse(text) : undefined
  if (!res.ok) throw new ApiError(res.status, data?.error ?? `Request failed (${res.status})`, data?.details ?? [])
  return data as T
}

export const api = {
  plan: (h: Household) => request<Plan>('POST', '/api/plan', h),
  claiming: (people: unknown[]) => request<ClaimingAnswer>('POST', '/api/claiming', { people }),
  sample: (name: 'couple' | 'single') => request<Household>('GET', `/api/samples/${name}`),
  quotes: (q: { premium: number; age: number; sex: string; spouseSex?: string; spouseAge?: number }) =>
    request<Quote[]>('GET', `/api/annuities/quotes?${new URLSearchParams(Object.entries(q).filter(([, v]) => v !== undefined).map(([k, v]) => [k, String(v)]))}`),
  me: () => request<Me | undefined>('GET', '/api/auth/me'),
  signup: (name: string, email: string, password: string) => request<Me>('POST', '/api/auth/signup', { name, email, password }),
  login: (email: string, password: string) => request<Me>('POST', '/api/auth/login', { email, password }),
  logout: () => request<void>('POST', '/api/auth/logout'),
  plans: () => request<SavedPlan[]>('GET', '/api/me/plans'),
  savePlan: (name: string, household: Household, id?: number) =>
    id ? request<SavedPlan>('PUT', `/api/me/plans/${id}`, { name, household }) : request<SavedPlan>('POST', '/api/me/plans', { name, household }),
  deletePlan: (id: number) => request<void>('DELETE', `/api/me/plans/${id}`),
  waitlist: (body: Record<string, unknown>) => request<WaitlistStatus>('POST', '/api/waitlist', body),
  waitlistStatus: (code: string) => request<WaitlistStatus>('GET', `/api/waitlist/${code}`),
  advisor: (body: Record<string, unknown>) => request<{ status: string }>('POST', '/api/advisor-requests', body),
  quoteRequest: (body: Record<string, unknown>) => request<{ status: string }>('POST', '/api/quote-requests', body),
}

export interface WaitlistStatus { code: string; firstName?: string; position: number; total: number; referrals: number; shareUrl: string }

export interface ClaimOption { claimMonths: number[]; expectedPv: number; monthlyAtStart: number }
export interface ClaimingAnswer { best: number[]; bestMonthly: number; gainVsEarliest: number; breakevenAge: number; compare: ClaimOption[]; lifeExpectancy: number[] }
