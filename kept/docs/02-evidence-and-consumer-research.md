# Evidence & consumer research

Kept started from two questions:
- **The clinical question:** what decides whether a GLP-1 result lasts?
- **The consumer question:** what do people on these medicines actually struggle with week to week?

Every feature traces back to one or both.

## 1. The clinical evidence

| Finding | Source | What Kept does about it |
|---|---|---|
| Adult GLP-1 use for weight rose from ~3% (2024) to ~11% (2026) | Gallup, via KFF Health News, 2026 | Mass-market positioning; supports every major drug, injection and pill |
| 64.8% of adults without diabetes stopped within 12 months (cohort of 125,474) | US Pharmacist summary of 2025 cohort study | Persistence is the headline metric for members and employers |
| ~60% of lost weight returns within a year of stopping (meta-analysis, 48 studies) | 2026 meta-analysis, via doctors.net.uk | Off-ramp planner: maintain / taper / stop, plus a regain guardrail |
| Cleveland Clinic real-world: 45% kept weight off at one year, many by restarting or switching | Cleveland Clinic | Off-ramp includes "stay on a maintenance dose"; the cost finder surfaces switches |
| Lean mass ≈ 25% (tirzepatide) to ~39% (semaglutide) of weight lost without intervention | SURMOUNT-1 and STEP-1 DXA sub-studies | MuscleGuard baseline lean share per drug |
| 2025 joint advisory (four societies): 1.2–1.6 g/kg/day protein, adjusted body weight above BMI 30, resistance training ≥ 3×/week | Joint advisory, 2025 | Targets engine; three-day strength program; muscle score |
| Labels define dose ladders, step intervals and missed-dose windows (tirzepatide 4 days, semaglutide 5 days) | US prescribing information | DrugCatalog + DoseSchedule; missed-dose guidance is label text, never a recommendation |
| Label warnings: pancreatitis, gallbladder disease, dehydration/AKI, hypoglycaemia with insulin or sulfonylureas, hypersensitivity | US prescribing information | Side-effect coach red flags; severity-3 red flags put a critical alert at the top of Today |
| Faster loss (> 1.5%/week) takes more lean mass | Body-composition literature | "Losing fast" alert, and a lean-share penalty in the muscle estimate |

### How the models work

**Weight trend.**
- Exponential moving average, α = 0.1, carried across missing days.
- Pace is the trend change over the last 28 days, in % of body weight per week.
- Flags: too fast (> 1.5%/week); plateau (≥ 12 weeks on medication and |pace| < 0.15%/week); regain (trend up more than 2% over four weeks).

**Lean-share estimate.**
- Formula: `base × (1 − 0.45·protein adherence − 0.25·strength adherence)`.
- Floor of 10%; +5 points if losing too fast; +3 points at age ≥ 60.
- Body-composition readings override the estimate whenever two or more readings span at least 1 kg of loss. The UI says which one it's showing.

**Dose schedule.**
- Next due time is the last dose plus the interval, or the dose weekday if no dose has been logged yet.
- Missed-dose text is chosen by the label's window.
- Injection-site rotation picks the least-recently-used site.
- Escalation notes respect the label's minimum weeks per step, and hold if side effects were severe that week.

**Side-effect pattern.**
- Symptom severity is summed by days since the preceding dose, for weekly drugs.
- Kept tells the member which day their symptoms peak, so they can plan lighter, lower-fat meals for that window.

**Cohort persistence.**
- Denominator: members whose first start date is at least N months ago.
- A member counts as persistent if they have any dose logged at or after start + N months.
- Every metric is suppressed below 10 members.

## 2. Consumer research

Method so far:
- Desk research only: public GLP-1 community discussions (r/Zepbound, r/Semaglutide, r/WegovyWeightLoss, r/Mounjaro, and GLP-1 creator comment threads), news coverage and employer benefit surveys.
- Themes are grouped by job-to-be-done.
- **No interviews have been run yet.** The validation plan at the end of this section lists them, and they are the first thing to do with the waitlist.

Recurring themes, in the order we'd expect them to matter:


1. **"Is this normal?" — side effects** (most frequent by far): nausea timing, constipation, sulfur burps, fatigue on shot day, "is this pain my gallbladder?" People want a fast, calm answer and to know when to worry.
   → Side-effect coach with red-flag triage, and a per-member pattern ("day 2 after your dose").
2. **Cost and coverage anxiety:** prior-auth denials, savings-card confusion, employers dropping coverage mid-year, whether to switch to the pill.
   → Cost finder by coverage type, annual savings figure, appeal letter drafted from logs.
3. **"Am I losing muscle?"** Fear driven by social video ("Ozempic face", "skinny fat"). People don't know how much protein, and can't eat enough of it with a small appetite.
   → Protein target from adjusted body weight, appetite-friendly quick-add foods, a three-day program, a muscle score.
4. **Dose logistics:** forgetting shot day, site rotation, what to do when late, when to step up.
   → Calendar reminders, site rotation, label missed-dose rule, escalation notes.
5. **"What happens when I stop?"** Fear of regain, and no plan from the prescriber.
   → Off-ramp planner and regain guardrail.
6. **Scale panic:** daily water swings read as failure.
   → Trend line over dots, pace as % per week.
7. **Talking to the prescriber:** 15-minute visits and no data.
   → One-page doctor report (print/PDF).

Hypotheses to validate:
- **Prescribers** want adherence and side effects between visits on one page, not an app login, and do not want an app nudging dose increases. That is why the report is a printable page, and why Kept never recommends dose changes.
- **Benefits leads** are being asked to justify the GLP-1 line item. They can't see whether members stay on therapy, and if they keep coverage it will be with conditions. That is why the dashboard leads with persistence against benchmark and "drug spend kept productive", and why pricing is a predictable per-covered-life fee.

Validation plan (first 30 days of the waitlist):
1. **Member interviews** — 15 members across stages: considering, starting, 3+ months, stopping and stopped. Rank the seven themes and test the willingness to pay $9.99/month or $69/year.
2. **Prescriber interviews** — 5 prescribers (PCP, obesity medicine, telehealth). React to the doctor report and the escalation wording.
3. **Benefits-lead interviews** — 5 benefits leads and consultants. React to the dashboard, the PMPM price, and Kept as a condition-of-coverage program.
4. **Waitlist analysis** — read the role, drug and stage fields captured on sign-up to size segments and pick the launch cohort.

## 3. What we deliberately didn't build

- **Prescribing or a pharmacy.** It would destroy neutrality and add drug margin to the incentives.
- **Compounded-drug sales or referrals.** That carries safety and regulatory risk, and members can't verify what they're getting.
- **Calorie-deficit tracking as the core loop.** On a GLP-1 the problem is eating *enough* protein and fiber, not eating less. Calories are optional.
- **Ads and data sales.** People are logging their health; the business must not depend on attention or resale.
