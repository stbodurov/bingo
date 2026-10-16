package com.sundial.engine;

import com.sundial.engine.Model.Sex;
import com.sundial.engine.Model.Health;

import java.util.ArrayList;
import java.util.List;

/**
 * Indicative pricing for income annuities from first principles: the premium equals the
 * present value of expected monthly payments (insurer pricing mortality, Treasury-plus-
 * spread discounting) grossed up by a load. Calibrated to the market: a $100,000 life-only
 * SPIA for a 65-year-old man pays roughly $620–$665 a month in mid-2026.
 *
 * <p>Insurers price with annuitant mortality (healthier than average), so pricing always
 * uses EXCELLENT-health tables regardless of the buyer's health — which is exactly why a
 * life annuity is worth more to someone in good health.
 */
public final class Annuities {
    private Annuities() {}

    /** Discount rate used by carriers (≈ A-rated corporate yield, mid-2026). */
    public static final double PRICING_RATE = 0.0535;
    public static final double LOAD = 0.035;
    public static final double QLAC_LIMIT = 210_000;

    public enum Kind { SPIA_LIFE, SPIA_10_CERTAIN, SPIA_JOINT, DIA, QLAC, MYGA }

    public record Quote(Kind kind, String label, double premium, double monthly, int startAge,
                        double payoutRate, String note, double commissionPct) {}

    static double factor(int age, Sex sex, int deferYears, int certainYears, Sex spouseSex, Integer spouseAge) {
        Mortality m = Mortality.of(sex, Health.EXCELLENT);
        Mortality ms = spouseSex == null ? null : Mortality.of(spouseSex, Health.EXCELLENT);
        double v = 1 / (1 + PRICING_RATE);
        double sum = 0;
        for (int k = deferYears * 12 + 1; k <= (deferYears + 60) * 12; k++) {
            double t = k / 12.0;
            double p1 = m.survival(age, t);
            double p = p1;
            if (ms != null) {
                double p2 = ms.survival(spouseAge, t);
                p = p1 + p2 - p1 * p2; // joint-and-100%-survivor: pays while either lives
            }
            if (t <= deferYears + certainYears) p = 1.0;
            sum += Math.pow(v, t) * p;
        }
        return sum;
    }

    public static double monthlyFor(double premium, int age, Sex sex, int deferYears, int certainYears,
                                    Sex spouseSex, Integer spouseAge) {
        return premium * (1 - LOAD) / factor(age, sex, deferYears, certainYears, spouseSex, spouseAge);
    }

    /** Premium needed to buy a given monthly life income starting now. */
    public static double premiumFor(double monthly, int age, Sex sex, Sex spouseSex, Integer spouseAge) {
        return monthly * factor(age, sex, 0, 0, spouseSex, spouseAge) / (1 - LOAD);
    }

    public static List<Quote> quotes(double premium, int age, Sex sex, Sex spouseSex, Integer spouseAge, double iraBalance) {
        List<Quote> out = new ArrayList<>();
        double life = monthlyFor(premium, age, sex, 0, 0, null, null);
        out.add(new Quote(Kind.SPIA_LIFE, "Lifetime income — single life", premium, life, age,
                12 * life / premium, "Highest income; stops at death.", 2.0));
        double certain = monthlyFor(premium, age, sex, 0, 10, null, null);
        out.add(new Quote(Kind.SPIA_10_CERTAIN, "Lifetime income — 10 years certain", premium, certain, age,
                12 * certain / premium, "Pays heirs the balance of 10 years if you die early.", 2.0));
        if (spouseSex != null) {
            double joint = monthlyFor(premium, age, sex, 0, 0, spouseSex, spouseAge);
            out.add(new Quote(Kind.SPIA_JOINT, "Joint lifetime income — 100% to survivor", premium, joint, age,
                    12 * joint / premium, "Pays until the second of you dies.", 2.0));
        }
        int deferTo = Math.max(age + 1, 80);
        double dia = monthlyFor(premium, age, sex, deferTo - age, 0, null, null);
        out.add(new Quote(Kind.DIA, "Longevity income starting at " + deferTo, premium, dia, deferTo,
                12 * dia / premium, "Cheap insurance against living past 85; nothing if you die before " + deferTo + ".", 2.5));
        if (iraBalance > 0 && age < 85) {
            double q = Math.min(QLAC_LIMIT, Math.min(premium, iraBalance));
            int start = Math.min(85, Math.max(age + 1, 80));
            double qm = monthlyFor(q, age, sex, start - age, 0, null, null);
            out.add(new Quote(Kind.QLAC, "QLAC from your IRA, starting at " + start, q, qm, start, 12 * qm / q,
                    "Up to $210,000 of IRA money is excluded from RMDs until income starts.", 2.5));
        }
        double myga = 0.0560;
        out.add(new Quote(Kind.MYGA, "5-year guaranteed rate (MYGA)", premium, premium * myga / 12, age, myga,
                "CD-like fixed rate for 5 years; principal back at maturity; tax-deferred.", 1.5));
        return out;
    }
}
