package com.sundial.engine;

import com.sundial.engine.Model.Annuity;
import com.sundial.engine.Model.Assumptions;
import com.sundial.engine.Model.Household;
import com.sundial.engine.Model.OneTime;
import com.sundial.engine.Model.Pension;
import com.sundial.engine.Model.Person;
import com.sundial.engine.Model.Strategy;

import java.util.ArrayList;
import java.util.List;

/**
 * Year-by-year cash-flow simulation of a household's retirement: wages, Social Security
 * (own, spousal, survivor), pensions, annuities, RMDs, withdrawals by strategy, Roth
 * conversions, federal + state tax, IRMAA, spending guardrails and investment returns.
 * Nominal dollars internally; results also reported in today's dollars.
 */
public final class Simulator {

    public record Settings(int[] claimMonths, Strategy strategy, double conversionBracket, boolean guardrails,
                           double spendingScale, boolean haircut, boolean irmaaGuard, double inflation,
                           double extraSpendingFromYear, int extraSpendingStart, int extraSpendingYears) {
        public Settings withScale(double s) {
            return new Settings(claimMonths, strategy, conversionBracket, guardrails, s, haircut, irmaaGuard, inflation,
                    extraSpendingFromYear, extraSpendingStart, extraSpendingYears);
        }
    }

    public record Year(int year, int[] ages, boolean[] alive, double wages, double socialSecurity, double pensions,
                       double annuities, double rmd, double fromCash, double fromTaxable, double fromTraditional,
                       double fromRoth, double conversion, double federalTax, double stateTax, double irmaa,
                       double spendingTarget, double spendingMet, double cash, double taxable, double traditional,
                       double roth, double marginalRate, double magi, double inflationIndex) {
        public double portfolio() { return cash + taxable + traditional + roth; }
        public double guaranteed() { return socialSecurity + pensions + annuities; }
    }

    public record Run(List<Year> years, boolean success, int depletionYear, double endingNominal, double endingReal,
                      double endingAfterTaxReal, double lifetimeTaxReal, double lifetimeIrmaaReal,
                      double lifetimeSpendingReal, double lifetimeSsReal) {}

    /** Returns per year (nominal), index 0 = first simulated year. */
    public record Path(double[] stock, double[] bond) {}

    private final Household h;
    private final int startYear;

    public Simulator(Household h, int startYear) {
        this.h = h;
        this.startYear = startYear;
    }

    public int startYear() { return startYear; }

    public int horizonYears() {
        int last = 0;
        for (Person p : h.people()) last = Math.max(last, p.birthYear() + h.assumptions().planToAge());
        return Math.max(1, last - startYear + 1);
    }

    public Path expectedPath() {
        Assumptions a = h.assumptions();
        int n = horizonYears();
        double[] s = new double[n], b = new double[n];
        // Median (geometric) return for a deterministic plan.
        double gs = a.stockReturn() - a.stockVol() * a.stockVol() / 2;
        double gb = a.bondReturn() - a.bondVol() * a.bondVol() / 2;
        java.util.Arrays.fill(s, gs);
        java.util.Arrays.fill(b, gb);
        return new Path(s, b);
    }

    public double[] planDeathAges() {
        double[] d = new double[h.people().size()];
        java.util.Arrays.fill(d, h.assumptions().planToAge() + 0.999);
        return d;
    }

    public Run run(Settings st, Path path, double[] deathAges) {
        List<Person> ps = h.people();
        int n = ps.size();
        boolean married = n == 2;
        Assumptions a = h.assumptions();
        double inf = st.inflation();
        double cash = h.accounts().cash(), taxable = h.accounts().taxable(), basis = h.accounts().taxableBasis();
        double trad = h.accounts().traditional(), roth = h.accounts().roth();
        double priorTradEnd = trad;
        List<Year> out = new ArrayList<>();
        double[] magiHistory = new double[horizonYears() + 2];
        boolean success = true;
        int depletion = 0;
        double guard = 1.0, initialRate = -1;
        double lifeTax = 0, lifeIrmaa = 0, lifeSpend = 0, lifeSs = 0;
        int[] diedYear = new int[n];
        java.util.Arrays.fill(diedYear, Integer.MAX_VALUE);
        for (int i = 0; i < n; i++) diedYear[i] = ps.get(i).birthYear() + (int) Math.floor(deathAges[i]) + 1;

        int years = horizonYears();
        for (int k = 0; k < years; k++) {
            int y = startYear + k;
            double I = Math.pow(1 + inf, k);
            int[] ages = new int[n];
            boolean[] alive = new boolean[n];
            int aliveCount = 0, over65 = 0;
            for (int i = 0; i < n; i++) {
                ages[i] = y - ps.get(i).birthYear();
                alive[i] = y < diedYear[i] && ages[i] <= a.planToAge();
                if (alive[i]) { aliveCount++; if (ages[i] >= 65) over65++; }
            }
            if (aliveCount == 0) break;
            boolean joint = married && aliveCount == 2;

            // ── Income ───────────────────────────────────────────────────────
            double wages = 0;
            for (int i = 0; i < n; i++)
                if (alive[i] && ages[i] < ps.get(i).retireAge()) wages += ps.get(i).salary() * I;

            double ss = 0;
            double cola = SocialSecurity.colaIndex(TaxCode.BASE_YEAR, y, inf) * SocialSecurity.payable(y, st.haircut());
            for (int i = 0; i < n; i++) {
                if (!alive[i]) continue;
                Person p = ps.get(i);
                int months = monthsCollecting(p, st.claimMonths()[i], y);
                if (married) {
                    int j = 1 - i;
                    Person sp = ps.get(j);
                    if (!alive[j]) {
                        int deathAgeMonths = (diedYear[j] - 1 - sp.birthYear()) * 12 + 6;
                        int survAge = (y - p.birthYear()) * 12;
                        double own = months > 0 ? SocialSecurity.ownBenefit(p, st.claimMonths()[i]) : 0;
                        int spClaim = st.claimMonths()[j] <= deathAgeMonths ? st.claimMonths()[j] : -1;
                        double surv = SocialSecurity.survivorBenefit(p, sp, spClaim, deathAgeMonths,
                                Math.max(survAge, 60 * 12), own);
                        ss += surv * 12 * cola; // survivor benefit can start regardless of own claim
                        continue;
                    }
                    boolean spouseFiled = monthsCollecting(sp, st.claimMonths()[j], y) > 0;
                    double monthly = spouseFiled ? SocialSecurity.withSpousal(p, st.claimMonths()[i], sp)
                            : SocialSecurity.ownBenefit(p, st.claimMonths()[i]);
                    ss += monthly * months * cola;
                } else {
                    ss += SocialSecurity.ownBenefit(p, st.claimMonths()[i]) * months * cola;
                }
            }

            double pensions = 0;
            for (Pension pe : h.pensions()) {
                int o = Math.min(pe.owner(), n - 1);
                Person owner = ps.get(o);
                int startY = owner.birthYear() + pe.startAge();
                if (y < startY) continue;
                double amt = pe.monthly() * 12 * (pe.cola() ? Math.pow(1 + inf, y - startYear) : 1);
                if (alive[o]) pensions += amt;
                else if (married && alive[1 - o]) pensions += amt * pe.survivorPct();
            }
            double annuities = 0;
            for (Annuity an : h.annuities()) {
                int o = Math.min(an.owner(), n - 1);
                int startY = ps.get(o).birthYear() + an.startAge();
                if (y < startY) continue;
                if (alive[o] || (an.joint() && married && alive[1 - o])) annuities += an.monthly() * 12;
            }

            // ── Spending ─────────────────────────────────────────────────────
            int oldest = 0;
            for (int i = 0; i < n; i++) if (alive[i]) oldest = Math.max(oldest, ages[i]);
            double phase = oldest >= 85 ? 0.70 : oldest >= 75 ? 0.85 : 1.0;
            double household = (married && aliveCount == 1) ? 0.72 : 1.0;
            double target = (h.essentialSpending() + h.discretionarySpending() * phase * guard) * household
                    * st.spendingScale() * I;
            for (OneTime o : h.oneTimes()) if (o.year() == y) target += o.amount() * I;
            if (st.extraSpendingYears() > 0 && y >= st.extraSpendingStart() && y < st.extraSpendingStart() + st.extraSpendingYears())
                target += st.extraSpendingFromYear() * I;

            // ── RMD (oldest owner of the combined IRA bucket) ─────────────────
            Person owner = ps.get(0);
            for (int i = 0; i < n; i++) if (alive[i] && (!alive[0] || ps.get(i).birthYear() < owner.birthYear())) owner = ps.get(i);
            double rmd = Math.min(trad, Rmd.amount(owner.birthYear(), y - owner.birthYear(), priorTradEnd));

            // IRMAA uses MAGI from two years earlier.
            double magiLookback = k >= 2 ? magiHistory[k - 2] : -1;

            // ── Withdrawals, conversions and tax (fixed-point on tax) ─────────
            double tax = 0, state = 0, irmaa = 0, marginal = 0, magi = 0;
            double wCash = 0, wTax = 0, wTrad = 0, wRoth = 0, conv = 0, met = 0;
            for (int iter = 0; iter < 7; iter++) {
                double incomeIn = wages + ss + pensions + annuities + rmd;
                double need = Math.max(0, target + tax + state + irmaa - incomeIn);
                double[] w = withdraw(st, need, cash, taxable, trad - rmd, roth, y, joint, over65, wages + pensions + annuities + rmd, ss, inf);
                wCash = w[0]; wTax = w[1]; wTrad = w[2]; wRoth = w[3];
                double gainFrac = taxable > 0 ? Math.max(0, 1 - basis / taxable) : 0;
                double ltcg = wTax * gainFrac;
                double ordinaryBase = wages + pensions + annuities * 0.5 /* exclusion ratio approx */;
                conv = conversion(st, y, joint, over65, ordinaryBase + rmd + wTrad, ss, ltcg, trad - rmd - wTrad, ages, magiLookback, inf);
                TaxCode.Result r = TaxCode.compute(new TaxCode.Input(y, joint, over65, ordinaryBase, rmd + wTrad + conv, ss, ltcg,
                        a.stateTaxRate(), inf));
                tax = r.federal();
                state = r.stateTax();
                marginal = r.marginalRate();
                magi = r.magi();
                double lookMagi = magiLookback >= 0 ? magiLookback : magi;
                irmaa = 0;
                for (int i = 0; i < n; i++) if (alive[i] && ages[i] >= 65) irmaa += TaxCode.irmaaPerPerson(y, joint, lookMagi, inf);
                double available = incomeIn + wCash + wTax + wTrad + wRoth;
                met = Math.min(target, available - tax - state - irmaa);
            }
            magiHistory[k] = magi;

            // Apply flows.
            double gainFrac = taxable > 0 ? Math.max(0, 1 - basis / taxable) : 0;
            basis -= wTax * (1 - gainFrac);
            cash -= wCash; taxable -= wTax; trad -= rmd + wTrad + conv; roth -= wRoth; roth += conv;
            double surplus = wages + ss + pensions + annuities + rmd + wCash + wTax + wTrad + wRoth - tax - state - irmaa - target;
            if (surplus > 1) { taxable += surplus; basis += surplus; } // reinvest excess income
            if (met < target * 0.99 && aliveCount > 0) {
                success = false;
                if (depletion == 0) depletion = y;
            }

            // Guardrails on discretionary spending.
            double portfolioStart = cash + taxable + trad + roth;
            double withdrawRate = portfolioStart > 0 ? (wCash + wTax + wTrad + wRoth) / portfolioStart : 1;
            if (st.guardrails()) {
                if (initialRate < 0 && withdrawRate > 0) initialRate = withdrawRate;
                else if (initialRate > 0) {
                    if (withdrawRate > initialRate * 1.2) guard = Math.max(0.5, guard * 0.9);
                    else if (withdrawRate < initialRate * 0.8) guard = Math.min(1.5, guard * 1.1);
                }
            }

            // Returns.
            double rs = path.stock()[Math.min(k, path.stock().length - 1)];
            double rb = path.bond()[Math.min(k, path.bond().length - 1)];
            double r = h.assumptions().stockShare() * rs + (1 - h.assumptions().stockShare()) * rb;
            double rc = inf + 0.005;
            cash = Math.max(0, cash) * (1 + rc);
            taxable = Math.max(0, taxable) * (1 + r);
            trad = Math.max(0, trad) * (1 + r);
            roth = Math.max(0, roth) * (1 + r);
            priorTradEnd = trad;

            out.add(new Year(y, ages, alive, wages, ss, pensions, annuities, rmd, wCash, wTax, wTrad, wRoth, conv, tax, state,
                    irmaa, target, Math.max(0, met), cash, taxable, trad, roth, marginal, magi, I));
            lifeTax += (tax + state) / I;
            lifeIrmaa += irmaa / I;
            lifeSpend += Math.max(0, met) / I;
            lifeSs += ss / I;
        }
        Year last = out.get(out.size() - 1);
        double endNom = last.portfolio();
        double endReal = endNom / last.inflationIndex();
        double afterTax = (last.cash() + last.taxable() + last.roth() + last.traditional() * 0.76) / last.inflationIndex();
        return new Run(out, success, depletion, endNom, endReal, afterTax, lifeTax, lifeIrmaa, lifeSpend, lifeSs);
    }

    /** Months of the calendar year in which the person is at or past their claiming age. */
    static int monthsCollecting(Person p, int claimMonths, int year) {
        int claimIndex = p.birthYear() * 12 + (p.birthMonth() - 1) + claimMonths; // month index the claim starts
        int from = year * 12, to = year * 12 + 11;
        if (claimIndex > to) return 0;
        return to - Math.max(from, claimIndex) + 1;
    }

    /** Returns {cash, taxable, traditional, roth} withdrawals to cover {@code need}. */
    private double[] withdraw(Settings st, double need, double cash, double taxable, double trad, double roth,
                              int y, boolean joint, int over65, double ordinaryIn, double ss, double inf) {
        double[] w = new double[4];
        double rem = need;
        trad = Math.max(0, trad);
        switch (st.strategy()) {
            case PROPORTIONAL -> {
                double c = Math.min(cash, rem); w[0] = c; rem -= c;
                double tot = taxable + trad + roth;
                if (tot > 0 && rem > 0) {
                    double take = Math.min(rem, tot);
                    w[1] = take * taxable / tot; w[2] = take * trad / tot; w[3] = take * roth / tot;
                }
            }
            case BRACKET_FILL -> {
                double rate = st.conversionBracket() > 0 ? st.conversionBracket() : 0.12;
                double top = TaxCode.bracketTop(y, joint, rate, inf) + TaxCode.standardDeduction(y, joint, over65, inf);
                double room = Math.max(0, top - ordinaryIn - 0.85 * ss);
                double t = Math.min(Math.min(room, trad), rem); w[2] = t; rem -= t;
                double c = Math.min(cash, rem); w[0] = c; rem -= c;
                double tx = Math.min(taxable, rem); w[1] = tx; rem -= tx;
                double rt = Math.min(roth, rem); w[3] = rt; rem -= rt;
                double more = Math.min(trad - w[2], rem); w[2] += more;
            }
            default -> {
                double c = Math.min(cash, rem); w[0] = c; rem -= c;
                double tx = Math.min(taxable, rem); w[1] = tx; rem -= tx;
                double t = Math.min(trad, rem); w[2] = t; rem -= t;
                double rt = Math.min(roth, rem); w[3] = rt;
            }
        }
        return w;
    }

    /** Roth conversion amount that fills ordinary income to the target bracket (optionally below IRMAA). */
    private double conversion(Settings st, int y, boolean joint, int over65, double ordinary, double ss, double ltcg,
                              double tradLeft, int[] ages, double magiLookback, double inf) {
        if (st.conversionBracket() <= 0 || tradLeft <= 0) return 0;
        int youngest = Integer.MAX_VALUE;
        for (int a : ages) youngest = Math.min(youngest, a);
        if (youngest >= 80) return 0;
        double top = TaxCode.bracketTop(y, joint, st.conversionBracket(), inf) + TaxCode.standardDeduction(y, joint, over65, inf);
        double taxableSs = TaxCode.taxableSocialSecurity(joint, ordinary + ltcg, ss);
        double room = top - ordinary - taxableSs - ltcg;
        if (st.irmaaGuard()) {
            double cap = TaxCode.irmaaThreshold(y, joint, inf) * 0.97 - (ordinary + taxableSs + ltcg);
            room = Math.min(room, cap);
        }
        return Math.max(0, Math.min(room, tradLeft));
    }
}
