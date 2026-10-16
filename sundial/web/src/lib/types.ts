export type Sex = 'MALE' | 'FEMALE'
export type Health = 'EXCELLENT' | 'GOOD' | 'AVERAGE' | 'POOR'
export type Strategy = 'CONVENTIONAL' | 'PROPORTIONAL' | 'BRACKET_FILL'

export interface Person {
  name: string
  birthYear: number
  birthMonth: number
  birthDay: number
  sex: Sex
  health: Health
  pia: number
  retireAge: number
  salary: number
  claimAgeMonths: number
}

export interface Pension { owner: number; monthly: number; startAge: number; cola: boolean; survivorPct: number }
export interface OneTime { year: number; amount: number; label: string }
export interface Accounts { cash: number; taxable: number; taxableBasis: number; traditional: number; roth: number }
export interface Assumptions {
  inflation: number; stockReturn: number; stockVol: number; bondReturn: number; bondVol: number; stockShare: number
  realDiscount: number; socialSecurityHaircut: boolean; planToAge: number; stateTaxRate: number
}

export interface Household {
  people: Person[]
  accounts: Accounts
  essentialSpending: number
  discretionarySpending: number
  pensions: Pension[]
  oneTimes: OneTime[]
  annuities: unknown[]
  strategy: Strategy
  conversionBracket: number
  guardrails: boolean
  assumptions: Assumptions
}

export interface PersonView {
  name: string; age: number; fraMonths: number; lifeExpectancyAge: number; oneInFourAge: number
  recommendedClaimMonths: number; monthlyAtClaim: number; rmdStartAge: number
}
export interface Summary {
  successRate: number; safeSpending: number; plannedSpending: number; monthlyPaycheck: number; guaranteedShare: number
  lifetimeTax: number; medianEnding: number; p10Ending: number; portfolio: number; firstRetirementYear: number
}
export interface ClaimRow { label: string; claimMonths: number[]; monthly: number; expectedPv: number }
export interface Claiming {
  recommended: number[]; recommendedMonthly: number; gainVsEarliest: number; breakevenAge: number; compare: ClaimRow[]
  haircutPvLoss: number; grid: number[][]; gridStart: number[]
}
export interface StrategyView {
  key: string; label: string; strategy: Strategy; conversionBracket: number; irmaaGuard: boolean; successRate: number
  lifetimeTax: number; lifetimeIrmaa: number; afterTaxLegacy: number; recommended: boolean
}
export interface Quote { kind: string; label: string; premium: number; monthly: number; startAge: number; payoutRate: number; note: string; commissionPct: number }
export interface Floor {
  essential: number; guaranteed: number; gap: number; premiumToCloseGap: number; recommendedPremium: number
  successWithout: number; successWith: number; quotes: Quote[]
}
export interface MonteCarlo {
  paths: number; successRate: number; medianEndingReal: number; p10EndingReal: number; p90EndingReal: number
  depletionAgeHistogram: number[]; probRunOutBefore85: number; medianLifetimeSpendingReal: number
  fanP10: number[]; fanP50: number[]; fanP90: number[]; fanYears: number[]
}
export interface Scenario { key: string; label: string; description: string; success: boolean; runsOutAge?: number; endingReal: number; lifetimeSpending: number }
export interface YearView {
  year: number; ages: number[]; wages: number; socialSecurity: number; pensions: number; annuities: number; rmd: number
  fromCash: number; fromTaxable: number; fromTraditional: number; fromRoth: number; conversion: number; taxes: number
  irmaa: number; spending: number; portfolio: number; traditional: number; roth: number; taxable: number; marginalRate: number
}
export interface PaycheckLine { date: string; source: string; amount: number; note: string }
export interface PaycheckMonth { month: string; lines: PaycheckLine[]; total: number }
export interface Insight { kind: string; title: string; body: string; value: number }

export interface Plan {
  people: PersonView[]; summary: Summary; claiming: Claiming; strategies: StrategyView[]; floor: Floor
  monteCarlo: MonteCarlo; scenarios: Scenario[]; years: YearView[]; paycheck: PaycheckMonth[]; insights: Insight[]
  generatedFor: string
}

export interface Me { id: number; name: string; email: string }
export interface SavedPlan { id: number; name: string; updatedAt: string; successRate?: number; safeSpending?: number; household: Household }
