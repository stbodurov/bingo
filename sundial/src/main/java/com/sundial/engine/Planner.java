package com.sundial.engine;

import com.sundial.engine.Model.Annuity;
import com.sundial.engine.Model.Household;
import com.sundial.engine.Model.Person;
import com.sundial.engine.Model.Strategy;

import java.time.DayOfWeek;
import java.time.LocalDate;
import java.time.temporal.TemporalAdjusters;
import java.util.ArrayList;
import java.util.List;

/** Builds the full retirement-income plan for a household. */
public final class Planner {

    public record PersonView(String name, int age, int fraMonths, double lifeExpectancyAge, double oneInFourAge,
                             int recommendedClaimMonths, double monthlyAtClaim, int rmdStartAge) {}

    public record Summary(double successRate, double safeSpending, double plannedSpending, double monthlyPaycheck,
                          double guaranteedShare, double lifetimeTax, double medianEnding, double p10Ending,
                          double portfolio, int firstRetirementYear) {}

    public record ClaimingView(int[] recommended, double recommendedMonthly, double gainVsEarliest, int breakevenAge,
                               List<Row> compare, double haircutPvLoss, double[][] grid, int[] gridStart) {
        public record Row(String label, int[] claimMonths, double monthly, double expectedPv) {}
    }

    public record StrategyView(String key, String label, Strategy strategy, double conversionBracket, boolean irmaaGuard,
                               double successRate, double lifetimeTax, double lifetimeIrmaa, double afterTaxLegacy,
                               boolean recommended) {}

    public record FloorView(double essential, double guaranteed, double gap, double premiumToCloseGap,
                            double recommendedPremium, double successWithout, double successWith,
                            List<Annuities.Quote> quotes) {}

    public record Scenario(String key, String label, String description, boolean success, Integer runsOutAge,
                           double endingReal, double lifetimeSpending) {}

    public record YearView(int year, int[] ages, double wages, double socialSecurity, double pensions, double annuities,
                           double rmd, double fromCash, double fromTaxable, double fromTraditional, double fromRoth,
                           double conversion, double taxes, double irmaa, double spending, double portfolio,
                           double traditional, double roth, double taxable, double marginalRate) {}

    public record PaycheckLine(String date, String source, double amount, String note) {}

    public record PaycheckMonth(String month, List<PaycheckLine> lines, double total) {}

    public record Insight(String kind, String title, String body, double value) {}

    public record Plan(List<PersonView> people, Summary summary, ClaimingView claiming, List<StrategyView> strategies,
                       FloorView floor, MonteCarlo.Summary monteCarlo, List<Scenario> scenarios, List<YearView> years,
                       List<PaycheckMonth> paycheck, List<Insight> insights, String generatedFor) {}

    private final Household h;
    private final LocalDate today;
    private final int paths;

    public Planner(Household h, LocalDate today, int paths) {
        this.h = h;
        this.today = today;
        this.paths = paths;
    }

    public Plan build() {
        List<Person> ps = h.people();
        int n = ps.size();
        Simulator sim = new Simulator(h, today.getYear());
        MonteCarlo mc = new MonteCarlo(h, sim, 20261016L);
        double inf = h.assumptions().inflation();
        boolean haircut = h.assumptions().socialSecurityHaircut();

        // 1. Social Security.
        ClaimingOptimizer opt = new ClaimingOptimizer(ps, today.getYear(), today.getMonthValue(), h.assumptions().realDiscount(), haircut);
        ClaimingOptimizer.Result cr = opt.optimise();
        int[] claims = new int[n];
        boolean userChose = true;
        for (int i = 0; i < n; i++) {
            claims[i] = ps.get(i).claimAgeMonths() > 0 ? ps.get(i).claimAgeMonths() : cr.best().claimMonths()[i];
            if (ps.get(i).claimAgeMonths() <= 0) userChose = false;
            claims[i] = Math.max(opt.minClaim(ps.get(i)), Math.min(SocialSecurity.LATEST, claims[i]));
        }
        ClaimingOptimizer haircutOpt = new ClaimingOptimizer(ps, today.getYear(), today.getMonthValue(), h.assumptions().realDiscount(), true);
        double haircutLoss = opt.expectedPv(cr.best().claimMonths()) - haircutOpt.expectedPv(cr.best().claimMonths());
        List<ClaimingView.Row> rows = List.of(
                new ClaimingView.Row("Recommended", cr.best().claimMonths(), cr.best().monthlyAtStart(), cr.best().expectedPv()),
                new ClaimingView.Row("As early as possible", cr.earliest().claimMonths(), cr.earliest().monthlyAtStart(), cr.earliest().expectedPv()),
                new ClaimingView.Row("At full retirement age", cr.fra().claimMonths(), cr.fra().monthlyAtStart(), cr.fra().expectedPv()),
                new ClaimingView.Row("At 70", cr.latest().claimMonths(), cr.latest().monthlyAtStart(), cr.latest().expectedPv()));
        int[] gridStart = new int[n];
        for (int i = 0; i < n; i++) gridStart[i] = opt.minClaim(ps.get(i));
        ClaimingView claiming = new ClaimingView(cr.best().claimMonths(), cr.best().monthlyAtStart(), cr.gainVsEarliest(),
                cr.breakevenAge(), rows, haircutLoss, cr.grid(), gridStart);

        // 2. Withdrawal & conversion strategies.
        Simulator.Settings base = settings(claims, h.strategy(), h.conversionBracket(), false, 1.0, haircut, inf);
        int smallPaths = Math.max(100, paths / 3);
        List<StrategyView> strategies = new ArrayList<>();
        Object[][] candidates = {
                {"conventional", "Taxable first, then IRA, then Roth", Strategy.CONVENTIONAL, 0.0, false},
                {"proportional", "Pro-rata from every account", Strategy.PROPORTIONAL, 0.0, false},
                {"fill12", "Fill the 12% bracket with IRA withdrawals + Roth conversions", Strategy.BRACKET_FILL, 0.12, true},
                {"fill22", "Fill the 22% bracket (IRMAA-aware)", Strategy.BRACKET_FILL, 0.22, true},
                {"fill24", "Fill the 24% bracket (aggressive conversions)", Strategy.BRACKET_FILL, 0.24, false},
        };
        StrategyView best = null;
        for (Object[] c : candidates) {
            Simulator.Settings s = settings(claims, (Strategy) c[2], (double) c[3], (boolean) c[4], 1.0, haircut, inf);
            Simulator.Run det = sim.run(s, sim.expectedPath(), sim.planDeathAges());
            double success = mc.run(s, smallPaths).successRate();
            StrategyView v = new StrategyView((String) c[0], (String) c[1], (Strategy) c[2], (double) c[3], (boolean) c[4],
                    success, det.lifetimeTaxReal(), det.lifetimeIrmaaReal(), det.endingAfterTaxReal(), false);
            strategies.add(v);
            if (best == null || score(v) > score(best)) best = v;
        }
        final StrategyView chosen = best;
        strategies.replaceAll(v -> v == chosen ? new StrategyView(v.key(), v.label(), v.strategy(), v.conversionBracket(),
                v.irmaaGuard(), v.successRate(), v.lifetimeTax(), v.lifetimeIrmaa(), v.afterTaxLegacy(), true) : v);
        Simulator.Settings plan = settings(claims, chosen.strategy(), chosen.conversionBracket(), chosen.irmaaGuard(), 1.0, haircut, inf);
        plan = new Simulator.Settings(plan.claimMonths(), plan.strategy(), plan.conversionBracket(), h.guardrails(),
                1.0, haircut, plan.irmaaGuard(), inf, 0, 0, 0);

        // 3. Monte Carlo on the recommended plan + safe spending.
        MonteCarlo.Summary summaryMc = mc.run(plan, paths);
        double safeScale = mc.safeScale(plan, 0.85, Math.max(100, paths / 4));
        Simulator.Run det = sim.run(plan, sim.expectedPath(), sim.planDeathAges());
        double plannedSpending = h.essentialSpending() + h.discretionarySpending();

        // 4. Income floor.
        int refYear = Math.max(today.getYear(), ps.get(0).birthYear() + 70);
        Simulator.Year ref = det.years().stream().filter(y -> y.year() == refYear).findFirst().orElse(det.years().get(det.years().size() - 1));
        double guaranteedReal = ref.guaranteed() / ref.inflationIndex();
        double gap = Math.max(0, h.essentialSpending() - guaranteedReal);
        Person a = ps.get(0);
        int ageA = today.getYear() - a.birthYear();
        Model.Sex spouseSex = n > 1 ? ps.get(1).sex() : null;
        Integer spouseAge = n > 1 ? today.getYear() - ps.get(1).birthYear() : null;
        int buyAge = Math.max(ageA, 65);
        double premiumGap = gap > 0 ? Annuities.premiumFor(gap / 12, buyAge, a.sex(), spouseSex, spouseAge == null ? null : spouseAge + (buyAge - ageA)) : 0;
        double liquid = h.accounts().total();
        double recommended = Math.min(premiumGap, 0.35 * liquid);
        double successWith = summaryMc.successRate();
        if (recommended > 5_000) {
            double monthly = Annuities.monthlyFor(recommended, buyAge, a.sex(), 0, 0, spouseSex, spouseAge == null ? null : spouseAge + (buyAge - ageA));
            Household withAnn = withAnnuity(recommended, monthly, buyAge, n > 1);
            Simulator simA = new Simulator(withAnn, today.getYear());
            successWith = new MonteCarlo(withAnn, simA, 20261016L).run(plan, Math.max(100, paths / 2)).successRate();
        }
        double quotePremium = recommended > 5_000 ? Math.round(recommended / 1000.0) * 1000 : 100_000;
        FloorView floor = new FloorView(h.essentialSpending(), guaranteedReal, gap, premiumGap, recommended,
                summaryMc.successRate(), successWith,
                Annuities.quotes(quotePremium, buyAge, a.sex(), spouseSex, spouseAge, h.accounts().traditional()));

        // 5. Scenarios (median markets unless stated).
        List<Scenario> scenarios = scenarios(sim, plan);

        // 6. Views.
        List<YearView> years = new ArrayList<>();
        for (Simulator.Year y : det.years()) {
            double I = y.inflationIndex();
            years.add(new YearView(y.year(), y.ages(), y.wages() / I, y.socialSecurity() / I, y.pensions() / I, y.annuities() / I,
                    y.rmd() / I, y.fromCash() / I, y.fromTaxable() / I, y.fromTraditional() / I, y.fromRoth() / I,
                    y.conversion() / I, (y.federalTax() + y.stateTax()) / I, y.irmaa() / I, y.spendingMet() / I,
                    y.portfolio() / I, y.traditional() / I, y.roth() / I, (y.taxable() + y.cash()) / I, y.marginalRate()));
        }
        int firstRetired = det.years().stream().filter(y -> y.wages() <= 0).mapToInt(Simulator.Year::year).findFirst().orElse(today.getYear());
        Simulator.Year firstYear = det.years().stream().filter(y -> y.year() == firstRetired).findFirst().orElse(det.years().get(0));
        double monthlyPaycheck = firstYear.spendingMet() / firstYear.inflationIndex() / 12;
        double guaranteedShare = firstYear.spendingMet() > 0 ? Math.min(1, firstYear.guaranteed() / firstYear.spendingMet()) : 0;

        List<PersonView> pv = new ArrayList<>();
        for (int i = 0; i < n; i++) {
            Person p = ps.get(i);
            Mortality m = Mortality.of(p.sex(), p.health());
            int age = today.getYear() - p.birthYear();
            double monthly = SocialSecurity.ownBenefit(p, claims[i]);
            pv.add(new PersonView(p.name(), age, SocialSecurity.fraMonths(p.birthYear()), age + m.lifeExpectancy(age),
                    m.ageWithSurvival(age, 0.25), claims[i], monthly, Rmd.startAge(p.birthYear())));
        }
        Summary summary = new Summary(summaryMc.successRate(), plannedSpending * safeScale, plannedSpending, monthlyPaycheck,
                guaranteedShare, det.lifetimeTaxReal(), summaryMc.medianEndingReal(), summaryMc.p10EndingReal(), liquid, firstRetired);

        List<PaycheckMonth> paycheck = paycheck(firstYear, claims, plan);
        List<Insight> insights = insights(summary, claiming, strategies, chosen, floor, scenarios, det, userChose);
        return new Plan(pv, summary, claiming, strategies, floor, summaryMc, scenarios, years, paycheck, insights,
                ps.stream().map(Person::name).reduce((x, y) -> x + " & " + y).orElse(""));
    }

    private static double score(StrategyView v) {
        // Success first, then after-tax legacy (taxes and IRMAA are already inside it).
        return v.successRate() * 2_000_000 + v.afterTaxLegacy();
    }

    private Simulator.Settings settings(int[] claims, Strategy s, double bracket, boolean irmaa, double scale, boolean haircut, double inf) {
        return new Simulator.Settings(claims, s, bracket, h.guardrails(), scale, haircut, irmaa, inf, 0, 0, 0);
    }

    private Household withAnnuity(double premium, double monthly, int startAge, boolean joint) {
        Model.Accounts ac = h.accounts();
        double fromTaxable = Math.min(ac.taxable() + ac.cash(), premium);
        double fromIra = premium - fromTaxable;
        double cashLeft = Math.max(0, ac.cash() - fromTaxable);
        double taxableLeft = Math.max(0, ac.taxable() - Math.max(0, fromTaxable - ac.cash()));
        double basisLeft = ac.taxable() > 0 ? ac.taxableBasis() * taxableLeft / ac.taxable() : 0;
        Model.Accounts next = new Model.Accounts(cashLeft, taxableLeft, basisLeft, Math.max(0, ac.traditional() - fromIra), ac.roth());
        List<Annuity> anns = new ArrayList<>(h.annuities());
        anns.add(new Annuity("SPIA", 0, premium, monthly, startAge, joint, fromIra > 0));
        return new Household(h.people(), next, h.essentialSpending(), h.discretionarySpending(), h.pensions(), h.oneTimes(),
                anns, h.strategy(), h.conversionBracket(), h.guardrails(), h.assumptions());
    }

    private List<Scenario> scenarios(Simulator sim, Simulator.Settings plan) {
        List<Scenario> out = new ArrayList<>();
        Simulator.Path base = sim.expectedPath();
        double[] deaths = sim.planDeathAges();
        Person a = h.people().get(0);
        out.add(scenario("base", "Median markets", "Typical returns; everyone lives to " + h.assumptions().planToAge() + ".", sim.run(plan, base, deaths), a));
        double[] s = base.stock().clone(), b = base.bond().clone();
        s[0] = -0.35; b[0] = 0.03; if (s.length > 1) { s[1] = -0.10; }
        out.add(scenario("crash", "Market crash in year one", "Stocks fall 35% then 10% — the sequence-of-returns risk that hurts most at the start.",
                sim.run(plan, new Simulator.Path(s, b), deaths), a));
        Simulator.Settings hiInf = new Simulator.Settings(plan.claimMonths(), plan.strategy(), plan.conversionBracket(), plan.guardrails(),
                1.0, plan.haircut(), plan.irmaaGuard(), 0.04, 0, 0, 0);
        out.add(scenario("inflation", "Inflation stays at 4%", "Prices rise 4% a year for the whole plan; returns don't keep up.", sim.run(hiInf, base, deaths), a));
        Simulator.Settings cut = new Simulator.Settings(plan.claimMonths(), plan.strategy(), plan.conversionBracket(), plan.guardrails(),
                1.0, true, plan.irmaaGuard(), plan.inflation(), 0, 0, 0);
        out.add(scenario("haircut", "Social Security pays 78% from 2033", "The 2026 Trustees Report's depletion scenario for the retirement trust fund.", sim.run(cut, base, deaths), a));
        int ltcStart = a.birthYear() + 82;
        Simulator.Settings ltc = new Simulator.Settings(plan.claimMonths(), plan.strategy(), plan.conversionBracket(), plan.guardrails(),
                1.0, plan.haircut(), plan.irmaaGuard(), plan.inflation(), 110_000, ltcStart, 3);
        out.add(scenario("ltc", "Long-term care at 82", "Three years of care at $110,000 a year (today's dollars).", sim.run(ltc, base, deaths), a));
        if (h.married()) {
            double[] early = deaths.clone();
            early[0] = 76;
            out.add(scenario("widow", h.people().get(0).name() + " dies at 76", "The survivor keeps the larger Social Security check, loses the smaller one, and files single — the 'widow's tax'.",
                    sim.run(plan, base, early), a));
        }
        return out;
    }

    private Scenario scenario(String key, String label, String desc, Simulator.Run r, Person a) {
        Integer runsOut = r.success() ? null : r.depletionYear() - a.birthYear();
        return new Scenario(key, label, desc, r.success(), runsOut, r.endingReal(), r.lifetimeSpendingReal());
    }

    private List<PaycheckMonth> paycheck(Simulator.Year y, int[] claims, Simulator.Settings plan) {
        List<PaycheckMonth> months = new ArrayList<>();
        LocalDate start = LocalDate.of(y.year(), 1, 1);
        if (start.isBefore(today.withDayOfMonth(1))) start = today.withDayOfMonth(1).plusMonths(1);
        double withdraw = y.fromCash() + y.fromTaxable() + y.fromTraditional() + y.fromRoth() + y.rmd();
        double taxMonthly = (y.federalTax() + y.stateTax() + y.irmaa()) / 12;
        double I = y.inflationIndex();
        for (int k = 0; k < 12; k++) {
            LocalDate m = start.plusMonths(k);
            List<PaycheckLine> lines = new ArrayList<>();
            double total = 0;
            // The year's Social Security, split into each person's monthly check: weights are each
            // benefit (own or spousal, whichever is higher) times the months it's paid this year.
            int n = h.people().size();
            double[] weight = new double[n];
            double denom = 0;
            for (int i = 0; i < n; i++) {
                Person p = h.people().get(i);
                int paid = Simulator.monthsCollecting(p, claims[i], y.year());
                if (paid <= 0) continue;
                weight[i] = n == 1 ? 1 : SocialSecurity.withSpousal(p, claims[i], h.people().get(1 - i));
                denom += weight[i] * paid;
            }
            for (int i = 0; i < n; i++) {
                Person p = h.people().get(i);
                int claimIndex = p.birthYear() * 12 + p.birthMonth() - 1 + claims[i];
                if (weight[i] <= 0 || denom <= 0 || m.getYear() * 12 + m.getMonthValue() - 1 <= claimIndex) continue; // paid the month after
                LocalDate wed = m.with(TemporalAdjusters.dayOfWeekInMonth(SocialSecurity.depositWednesday(p.birthDay()), DayOfWeek.WEDNESDAY));
                double amt = y.socialSecurity() * weight[i] / denom / I;
                if (amt <= 0) continue;
                lines.add(new PaycheckLine(wed.toString(), "Social Security — " + p.name(), amt, "Deposited on the " + ordinal(SocialSecurity.depositWednesday(p.birthDay())) + " Wednesday"));
                total += amt;
            }
            LocalDate first = firstBusinessDay(m);
            if (y.pensions() > 0) { lines.add(new PaycheckLine(first.toString(), "Pension", y.pensions() / 12 / I, "")); total += y.pensions() / 12 / I; }
            if (y.annuities() > 0) { lines.add(new PaycheckLine(first.toString(), "Annuity income", y.annuities() / 12 / I, "Guaranteed for life")); total += y.annuities() / 12 / I; }
            if (withdraw > 0) {
                StringBuilder note = new StringBuilder();
                if (y.fromTaxable() + y.fromCash() > 0) note.append(String.format("brokerage %,.0f", (y.fromTaxable() + y.fromCash()) / 12 / I));
                if (y.fromTraditional() + y.rmd() > 0) note.append(note.length() > 0 ? " · " : "").append(String.format("IRA %,.0f", (y.fromTraditional() + y.rmd()) / 12 / I));
                if (y.fromRoth() > 0) note.append(note.length() > 0 ? " · " : "").append(String.format("Roth %,.0f", y.fromRoth() / 12 / I));
                LocalDate sweep = firstBusinessDay(m).plusDays(1);
                lines.add(new PaycheckLine(sweep.toString(), "Sundial paycheck transfer", withdraw / 12 / I, note.toString()));
                total += withdraw / 12 / I;
            }
            if (taxMonthly > 0) {
                lines.add(new PaycheckLine(first.toString(), "Tax & Medicare surcharge reserve", -taxMonthly / I, "Withheld so April has no surprises"));
                total -= taxMonthly / I;
            }
            if (y.conversion() > 0 && m.getMonthValue() == 11)
                lines.add(new PaycheckLine(m.withDayOfMonth(15).toString(), "Roth conversion (IRA → Roth)", y.conversion() / I, "Not spending — moves money to tax-free"));
            months.add(new PaycheckMonth(m.getMonth() + " " + m.getYear(), lines, total));
        }
        return months;
    }

    private static LocalDate firstBusinessDay(LocalDate m) {
        LocalDate d = m.withDayOfMonth(1);
        while (d.getDayOfWeek() == DayOfWeek.SATURDAY || d.getDayOfWeek() == DayOfWeek.SUNDAY) d = d.plusDays(1);
        return d;
    }

    private static String ordinal(int n) {
        return switch (n) { case 2 -> "second"; case 3 -> "third"; default -> "fourth"; };
    }

    private List<Insight> insights(Summary s, ClaimingView c, List<StrategyView> strategies, StrategyView chosen,
                                   FloorView floor, List<Scenario> scenarios, Simulator.Run det, boolean userChose) {
        List<Insight> out = new ArrayList<>();
        out.add(new Insight("claiming", claimingTitle(h.people(), c.recommended()),
                String.format("Compared with claiming as early as possible, this adds about $%,.0f of expected lifetime benefits in today's dollars, weighted by how long you're likely to live.%s", c.gainVsEarliest(),
                        h.married() ? " Delaying the higher earner also raises the survivor's check for life." : ""),
                c.gainVsEarliest()));
        StrategyView conventional = strategies.stream().filter(v -> v.key().equals("conventional")).findFirst().orElse(chosen);
        double legacyGain = chosen.afterTaxLegacy() - conventional.afterTaxLegacy();
        double taxSaved = conventional.lifetimeTax() + conventional.lifetimeIrmaa() - chosen.lifetimeTax() - chosen.lifetimeIrmaa();
        if (!chosen.key().equals("conventional"))
            out.add(new Insight("tax", "Draw down in a tax-smart order: " + Character.toLowerCase(chosen.label().charAt(0)) + chosen.label().substring(1),
                    String.format("Versus the usual 'taxable first' order, this %s lifetime taxes and Medicare surcharges by $%,.0f and leaves heirs $%,.0f more after tax.",
                            taxSaved >= 0 ? "cuts" : "adds", Math.abs(taxSaved), legacyGain), legacyGain));
        if (floor.gap() > 500)
            out.add(new Insight("floor", String.format("Your guaranteed income covers %.0f%% of essentials", 100 * Math.min(1, floor.guaranteed() / Math.max(1, floor.essential()))),
                    String.format("A lifetime annuity of about $%,.0f would close %s of a $%,.0f-a-year gap so groceries, housing and healthcare never depend on markets. %s",
                            floor.recommendedPremium(), floor.recommendedPremium() >= floor.premiumToCloseGap() - 1 ? "all" : "most",
                            floor.gap(), Math.abs(floor.successWith() - floor.successWithout()) >= 0.005
                                    ? String.format("Plan confidence moves from %.0f%% to %.0f%%.", 100 * floor.successWithout(), 100 * floor.successWith())
                                    : "Your confidence stays about the same — what you buy is certainty for the essentials."), floor.recommendedPremium()));
        else
            out.add(new Insight("floor", "Essentials are fully covered by guaranteed income",
                    "Social Security, pensions and annuities pay for your essential spending on their own — markets only fund the extras.", 0));
        out.add(new Insight("spending", String.format("You can spend about $%,.0f a year", s.safeSpending()),
                String.format("That's the level with an 85%% chance of lasting as long as you do, versus the $%,.0f you planned. Guardrails trim travel and extras by 10%% after a bad year and give it back after a good one.", s.plannedSpending()),
                s.safeSpending() - s.plannedSpending()));
        scenarios.stream().filter(x -> x.key().equals("haircut")).findFirst().ifPresent(x -> out.add(new Insight("haircut",
                x.success() ? "A 22% Social Security cut in 2033 doesn't break the plan" : "A 22% Social Security cut in 2033 would strain the plan",
                String.format("Under the Trustees' depletion scenario your expected benefits fall by about $%,.0f. %s", c.haircutPvLoss(),
                        x.success() ? "You'd still have money at the end of the plan." : "Money would run short around age " + x.runsOutAge() + "."),
                -c.haircutPvLoss())));
        double maxIrmaa = det.years().stream().mapToDouble(Simulator.Year::irmaa).max().orElse(0);
        if (maxIrmaa > 0 && det.lifetimeIrmaaReal() > 5_000)
            out.add(new Insight("irmaa", "Medicare surcharges (IRMAA) kick in some years",
                    String.format("Income two years earlier crosses an IRMAA tier; the plan pays $%,.0f in surcharges over its life. Sundial keeps conversions just under the first threshold where it pays to.", det.lifetimeIrmaaReal()),
                    -det.lifetimeIrmaaReal()));
        return out;
    }

    /** "Claim Social Security at 70", or for a couple "Claim Social Security: Ellen at 66 and 1 month, Marcus at 70". */
    static String claimingTitle(List<Model.Person> people, int[] months) {
        if (months.length == 1) return "Claim Social Security at " + age(months[0]);
        StringBuilder sb = new StringBuilder("Claim Social Security: ");
        for (int i = 0; i < months.length; i++) {
            if (i > 0) sb.append(", ");
            String name = i < people.size() && people.get(i).name() != null && !people.get(i).name().isBlank()
                    ? people.get(i).name() : (i == 0 ? "you" : "your spouse");
            sb.append(name).append(" at ").append(age(months[i]));
        }
        return sb.toString();
    }

    static String age(int months) {
        int y = months / 12, m = months % 12;
        return y + (m > 0 ? " and " + m + " month" + (m > 1 ? "s" : "") : "");
    }
}
