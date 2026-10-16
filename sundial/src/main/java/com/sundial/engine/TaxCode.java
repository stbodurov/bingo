package com.sundial.engine;

/**
 * Federal income tax for retirees, tax year 2026 parameters indexed forward by the
 * inflation assumption: ordinary brackets, the standard deduction with the age-65
 * addition, the 2025–2028 senior deduction, taxation of Social Security benefits,
 * long-term capital gains brackets, NIIT, and Medicare IRMAA surcharges.
 */
public final class TaxCode {
    private TaxCode() {}

    public static final int BASE_YEAR = 2026;
    static final double[] RATES = {0.10, 0.12, 0.22, 0.24, 0.32, 0.35, 0.37};
    static final double[] SINGLE = {12_400, 50_400, 105_700, 201_775, 256_225, 640_600};
    static final double[] JOINT = {24_800, 100_800, 211_400, 403_550, 512_450, 768_700};
    static final double STD_SINGLE = 16_100, STD_JOINT = 32_200;
    static final double ADD65_SINGLE = 2_050, ADD65_JOINT = 1_650;
    static final double SENIOR_DED = 6_000, SENIOR_PHASE_SINGLE = 75_000, SENIOR_PHASE_JOINT = 150_000;
    static final double LTCG0_SINGLE = 49_450, LTCG0_JOINT = 98_900, LTCG15_SINGLE = 545_500, LTCG15_JOINT = 613_700;
    /** IRMAA tiers (MAGI two years earlier) and monthly Part B + Part D surcharges per person, 2026. */
    static final double[] IRMAA_SINGLE = {109_000, 137_000, 171_000, 205_000, 500_000};
    static final double[] IRMAA_JOINT = {218_000, 274_000, 342_000, 410_000, 750_000};
    static final double[] IRMAA_MONTHLY = {81.20 + 14.50, 202.90 + 37.50, 324.60 + 60.40, 446.30 + 83.30, 487.00 + 91.00};
    public static final double PART_B_2026 = 202.90;

    public record Input(int year, boolean joint, int over65, double wagesAndOrdinary, double traditionalWithdrawals,
                        double socialSecurity, double ltcg, double stateRate, double inflation) {}

    public record Result(double agi, double taxableSocialSecurity, double deduction, double taxableIncome,
                         double ordinaryTax, double capitalGainsTax, double niit, double stateTax,
                         double marginalRate, double magi) {
        public double federal() { return ordinaryTax + capitalGainsTax + niit; }
        public double total() { return federal() + stateTax; }
    }

    static double index(int year, double inflation) {
        return Math.pow(1 + inflation, Math.max(0, year - BASE_YEAR));
    }

    public static double[] brackets(int year, boolean joint, double inflation) {
        double[] base = joint ? JOINT : SINGLE;
        double f = index(year, inflation);
        double[] out = new double[base.length];
        for (int i = 0; i < base.length; i++) out[i] = base[i] * f;
        return out;
    }

    /** Top of the bracket with the given rate (e.g. 0.12) in that year's dollars. */
    public static double bracketTop(int year, boolean joint, double rate, double inflation) {
        double[] b = brackets(year, joint, inflation);
        for (int i = 0; i < RATES.length - 1; i++) if (Math.abs(RATES[i] - rate) < 1e-9) return b[i];
        return Double.POSITIVE_INFINITY;
    }

    public static double standardDeduction(int year, boolean joint, int over65, double inflation) {
        double f = index(year, inflation);
        return ((joint ? STD_JOINT : STD_SINGLE) + over65 * (joint ? ADD65_JOINT : ADD65_SINGLE)) * f;
    }

    /** Senior deduction (OBBBA §70103): $6,000 per person 65+, tax years 2025–2028, phased out at 6%. */
    public static double seniorDeduction(int year, boolean joint, int over65, double magi) {
        if (year > 2028 || over65 == 0) return 0;
        double threshold = joint ? SENIOR_PHASE_JOINT : SENIOR_PHASE_SINGLE;
        double per = Math.max(0, SENIOR_DED - 0.06 * Math.max(0, magi - threshold));
        return per * over65;
    }

    /** Taxable portion of Social Security benefits (IRC §86); thresholds are not indexed. */
    public static double taxableSocialSecurity(boolean joint, double otherIncome, double benefits) {
        double t1 = joint ? 32_000 : 25_000, t2 = joint ? 44_000 : 34_000;
        double provisional = otherIncome + 0.5 * benefits;
        if (provisional <= t1) return 0;
        if (provisional <= t2) return Math.min(0.5 * benefits, 0.5 * (provisional - t1));
        return Math.min(0.85 * benefits, 0.85 * (provisional - t2) + Math.min(0.5 * benefits, 0.5 * (t2 - t1)));
    }

    public static Result compute(Input in) {
        double ordinary = in.wagesAndOrdinary() + in.traditionalWithdrawals();
        double taxableSs = taxableSocialSecurity(in.joint(), ordinary + in.ltcg(), in.socialSecurity());
        double agi = ordinary + taxableSs + in.ltcg();
        double magi = agi;
        double deduction = standardDeduction(in.year(), in.joint(), in.over65(), in.inflation())
                + seniorDeduction(in.year(), in.joint(), in.over65(), magi);
        double taxable = Math.max(0, agi - deduction);
        double ordinaryTaxable = Math.max(0, taxable - in.ltcg());
        double[] b = brackets(in.year(), in.joint(), in.inflation());
        double tax = 0, prev = 0, marginal = RATES[0];
        for (int i = 0; i < RATES.length; i++) {
            double top = i < b.length ? b[i] : Double.POSITIVE_INFINITY;
            if (ordinaryTaxable > prev) {
                tax += (Math.min(ordinaryTaxable, top) - prev) * RATES[i];
                marginal = RATES[i];
            }
            prev = top;
        }
        double f = index(in.year(), in.inflation());
        double zeroTop = (in.joint() ? LTCG0_JOINT : LTCG0_SINGLE) * f;
        double fifteenTop = (in.joint() ? LTCG15_JOINT : LTCG15_SINGLE) * f;
        double gains = Math.min(in.ltcg(), taxable);
        double start = ordinaryTaxable;
        double at0 = Math.max(0, Math.min(start + gains, zeroTop) - start);
        double at15 = Math.max(0, Math.min(start + gains, fifteenTop) - Math.max(start, zeroTop));
        double at20 = Math.max(0, gains - at0 - at15);
        double cg = at15 * 0.15 + at20 * 0.20;
        double niit = 0.038 * Math.max(0, Math.min(in.ltcg(), magi - (in.joint() ? 250_000 : 200_000)));
        // Most states exempt Social Security; apply the flat state rate to the rest of AGI.
        double state = in.stateRate() * Math.max(0, agi - taxableSs);
        // Marginal rate including the Social Security "tax torpedo" (an extra $ of ordinary income
        // can make $0.85 of benefits taxable).
        double torpedo = taxableSocialSecurity(in.joint(), ordinary + in.ltcg() + 100, in.socialSecurity()) - taxableSs;
        double effMarginal = marginal * (1 + torpedo / 100.0);
        return new Result(agi, taxableSs, deduction, taxable, tax, cg, niit, state, effMarginal, magi);
    }

    /** Annual IRMAA surcharge per Medicare enrollee for a MAGI (from two years earlier). */
    public static double irmaaPerPerson(int year, boolean joint, double magi, double inflation) {
        double[] tiers = joint ? IRMAA_JOINT : IRMAA_SINGLE;
        double f = index(year, inflation);
        double surcharge = 0;
        for (int i = 0; i < tiers.length; i++) if (magi > tiers[i] * f) surcharge = IRMAA_MONTHLY[i] * f * 12;
        return surcharge;
    }

    /** First IRMAA threshold (useful as a Roth-conversion ceiling). */
    public static double irmaaThreshold(int year, boolean joint, double inflation) {
        return (joint ? IRMAA_JOINT[0] : IRMAA_SINGLE[0]) * index(year, inflation);
    }
}
