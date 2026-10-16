export const usd = (v: number, digits = 0) =>
  (v < 0 ? '−' : '') + '$' + Math.abs(v).toLocaleString('en-US', { minimumFractionDigits: digits, maximumFractionDigits: digits })

export const compact = (v: number) => {
  const a = Math.abs(v), s = v < 0 ? '−' : ''
  if (a >= 1e6) return `${s}$${(a / 1e6).toFixed(a >= 1e7 ? 1 : 2)}M`
  if (a >= 1e4) return `${s}$${Math.round(a / 1e3)}k`
  if (a >= 1e3) return `${s}$${(a / 1e3).toFixed(1)}k`
  return `${s}$${Math.round(a)}`
}

export const pct = (v: number, digits = 0) => `${(v * 100).toFixed(digits)}%`

export const age = (months: number) => {
  const y = Math.floor(months / 12), m = months % 12
  return m ? `${y} and ${m} mo` : `${y}`
}

export const monthName = (s: string) => s.charAt(0) + s.slice(1).toLowerCase()

export const confidenceWord = (p: number) =>
  p >= 0.9 ? 'Very confident' : p >= 0.8 ? 'Confident' : p >= 0.65 ? 'Some risk' : 'At risk'

export const ageShort = (months: number) => {
  const y = Math.floor(months / 12), m = months % 12
  return m ? `${y}y ${m}m` : `${y}`
}
