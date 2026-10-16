package com.sundial.engine;

/** Required minimum distributions: SECURE 2.0 start ages and the IRS Uniform Lifetime Table. */
public final class Rmd {
    private Rmd() {}

    private static final double[] DIVISORS = { // ages 72..120
            27.4, 26.5, 25.5, 24.6, 23.7, 22.9, 22.0, 21.1, 20.2, 19.4, 18.5, 17.7, 16.8, 16.0, 15.2, 14.4,
            13.7, 12.9, 12.2, 11.5, 10.8, 10.1, 9.5, 8.9, 8.4, 7.8, 7.3, 6.8, 6.4, 6.0, 5.6, 5.2, 4.9, 4.6,
            4.3, 4.1, 3.9, 3.7, 3.5, 3.4, 3.3, 3.1, 3.0, 2.9, 2.8, 2.7, 2.5, 2.3, 2.0};

    /** 73 for those born 1951–1959, 75 for 1960 and later (72 earlier). */
    public static int startAge(int birthYear) {
        if (birthYear >= 1960) return 75;
        if (birthYear >= 1951) return 73;
        return 72;
    }

    public static double divisor(int age) {
        int i = Math.max(0, Math.min(DIVISORS.length - 1, age - 72));
        return DIVISORS[i];
    }

    /** RMD for the year the owner turns {@code age}, given the prior year-end balance. */
    public static double amount(int birthYear, int age, double priorYearEndBalance) {
        if (age < startAge(birthYear) || priorYearEndBalance <= 0) return 0;
        return priorYearEndBalance / divisor(age);
    }
}
