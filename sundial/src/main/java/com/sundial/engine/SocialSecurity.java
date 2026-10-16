package com.sundial.engine;

import com.sundial.engine.Model.Person;

/**
 * Social Security retirement, spousal and survivor benefit rules (42 U.S.C. 402, 20 CFR 404).
 *
 * <p>All benefits are expressed as monthly amounts in today's dollars relative to the
 * worker's primary insurance amount (PIA, the benefit at full retirement age).
 */
public final class SocialSecurity {
    private SocialSecurity() {}

    public static final int EARLIEST = 62 * 12;
    public static final int LATEST = 70 * 12;
    /** Share of scheduled benefits payable after the OASI trust fund is depleted (2026 Trustees Report). */
    public static final double PAYABLE_AFTER_DEPLETION = 0.78;
    public static final int DEPLETION_YEAR = 2033;
    /** COLAs already announced: 2.8% for 2026, 3.6% for 2027. */
    public static final double COLA_2027 = 0.036;

    /** Full retirement age in months for a birth year. */
    public static int fraMonths(int birthYear) {
        if (birthYear <= 1937) return 65 * 12;
        if (birthYear <= 1942) return 65 * 12 + 2 * (birthYear - 1937);
        if (birthYear <= 1954) return 66 * 12;
        if (birthYear <= 1959) return 66 * 12 + 2 * (birthYear - 1954);
        return 67 * 12;
    }

    /** People born on the 1st or 2nd of a month attain an age the month before (SSA rule); ignored here. */
    public static double retirementFactor(int claimMonths, int fraMonths) {
        if (claimMonths < fraMonths) {
            int early = fraMonths - claimMonths;
            return 1 - (5.0 / 9 / 100) * Math.min(early, 36) - (5.0 / 12 / 100) * Math.max(0, early - 36);
        }
        int late = Math.min(claimMonths, LATEST) - fraMonths;
        return 1 + (2.0 / 3 / 100) * late;
    }

    /** Spousal benefit factor applied to 50% of the worker's PIA. No delayed credits. */
    public static double spousalFactor(int claimMonths, int fraMonths) {
        if (claimMonths >= fraMonths) return 1.0;
        int early = fraMonths - claimMonths;
        return 1 - (25.0 / 36 / 100) * Math.min(early, 36) - (5.0 / 12 / 100) * Math.max(0, early - 36);
    }

    public static double ownBenefit(Person p, int claimMonths) {
        return p.pia() * retirementFactor(claimMonths, fraMonths(p.birthYear()));
    }

    /**
     * Monthly benefit for {@code p} while both spouses are alive and both have filed:
     * own retirement benefit plus any excess spousal benefit (deemed filing).
     */
    public static double withSpousal(Person p, int claimMonths, Person spouse) {
        double own = ownBenefit(p, claimMonths);
        if (spouse == null) return own;
        double excess = Math.max(0, 0.5 * spouse.pia() - p.pia());
        if (excess <= 0) return own;
        return own + excess * spousalFactor(claimMonths, fraMonths(p.birthYear()));
    }

    /**
     * Survivor benefit for {@code survivor} after {@code deceased} dies.
     *
     * @param deceasedClaimMonths age the deceased claimed, or -1 if they died before claiming
     * @param deceasedDeathMonths age of the deceased at death (months)
     * @param survivorAgeMonths   survivor's age when they start the survivor benefit
     */
    public static double survivorBenefit(Person survivor, Person deceased, int deceasedClaimMonths,
                                         int deceasedDeathMonths, int survivorAgeMonths, double survivorOwn) {
        int dFra = fraMonths(deceased.birthYear());
        double base;
        if (deceasedClaimMonths >= 0 && deceasedClaimMonths <= deceasedDeathMonths) {
            base = ownBenefit(deceased, deceasedClaimMonths);
            if (deceasedClaimMonths < dFra) base = Math.max(base, 0.825 * deceased.pia()); // widow(er)'s limit
        } else {
            base = deceased.pia() * retirementFactor(Math.max(dFra, Math.min(deceasedDeathMonths, LATEST)), dFra);
        }
        int sFra = fraMonths(survivor.birthYear());
        if (survivorAgeMonths < sFra) {
            // Survivor benefits reduce linearly to 71.5% at age 60.
            double span = sFra - 60 * 12;
            double early = Math.min(span, sFra - survivorAgeMonths);
            base *= 1 - 0.285 * early / span;
        }
        return Math.max(base, survivorOwn);
    }

    /** Annual inflation factor for benefits: known 2027 COLA, assumption afterwards. */
    public static double colaIndex(int fromYear, int toYear, double inflation) {
        double f = 1;
        for (int y = fromYear + 1; y <= toYear; y++) f *= 1 + (y == 2027 ? COLA_2027 : inflation);
        return f;
    }

    /** Payable share in a given year under the trust-fund depletion scenario. */
    public static double payable(int year, boolean haircut) {
        return haircut && year >= DEPLETION_YEAR ? PAYABLE_AFTER_DEPLETION : 1.0;
    }

    /** SSA deposit day: 2nd, 3rd or 4th Wednesday by birthday (for those who claimed after May 1997). */
    public static int depositWednesday(int birthDay) {
        if (birthDay <= 10) return 2;
        if (birthDay <= 20) return 3;
        return 4;
    }
}
