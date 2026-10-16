package com.sundial.engine;

import java.util.List;

/** The household a plan is built for. All money is in today's dollars unless noted. */
public final class Model {
    private Model() {}

    public enum Sex { MALE, FEMALE }

    public enum Health { EXCELLENT, GOOD, AVERAGE, POOR }

    public enum Strategy { CONVENTIONAL, PROPORTIONAL, BRACKET_FILL }

    public record Person(
            String name,
            int birthYear,
            int birthMonth,
            int birthDay,
            Sex sex,
            Health health,
            /** Monthly benefit at full retirement age from the SSA statement (the PIA). */
            double pia,
            /** Age (years) the person stops working; 0 if already retired. */
            int retireAge,
            /** Annual wages until retirement. */
            double salary,
            /** Planned claiming age in months (e.g. 67*12); 0 = let Sundial choose. */
            int claimAgeMonths) {}

    public record Pension(int owner, double monthly, int startAge, boolean cola, double survivorPct) {}

    public record Accounts(double cash, double taxable, double taxableBasis, double traditional, double roth) {
        public double total() { return cash + taxable + traditional + roth; }
    }

    public record OneTime(int year, double amount, String label) {}

    public record Annuity(String type, int owner, double premium, double monthly, int startAge, boolean joint, boolean fromIra) {}

    public record Assumptions(
            double inflation,
            double stockReturn,
            double stockVol,
            double bondReturn,
            double bondVol,
            double stockShare,
            /** Real discount rate for valuing benefits (≈ TIPS yield). */
            double realDiscount,
            boolean socialSecurityHaircut,
            int planToAge,
            double stateTaxRate) {
        public static Assumptions defaults() {
            return new Assumptions(0.025, 0.085, 0.17, 0.045, 0.06, 0.55, 0.018, false, 95, 0.0);
        }
    }

    public record Household(
            List<Person> people,
            Accounts accounts,
            /** Essential spending per year (housing, food, health, insurance). */
            double essentialSpending,
            /** Discretionary spending per year (travel, gifts, hobbies). */
            double discretionarySpending,
            List<Pension> pensions,
            List<OneTime> oneTimes,
            List<Annuity> annuities,
            Strategy strategy,
            /** Roth conversions fill ordinary income up to this bracket rate (0 = none). */
            double conversionBracket,
            boolean guardrails,
            Assumptions assumptions) {

        public boolean married() { return people.size() == 2; }
    }
}
