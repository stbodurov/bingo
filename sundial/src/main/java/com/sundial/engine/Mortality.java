package com.sundial.engine;

import com.sundial.engine.Model.Health;
import com.sundial.engine.Model.Sex;

import java.util.random.RandomGenerator;

/**
 * Gompertz law of mortality, μ(x) = (1/b)·e^{(x−m)/b}, with modal age m and dispersion b.
 *
 * <p>Parameters are calibrated so remaining life expectancy at 65 matches SSA cohort
 * projections for people turning 65 in the mid-2020s (≈ 19.9 years for men, ≈ 22.4 for
 * women, before health adjustment). Self-reported health shifts the modal age — the
 * single largest driver of when to claim Social Security and whether a life annuity pays.
 */
public final class Mortality {
    private final double m;
    private final double b;

    public Mortality(double modalAge, double dispersion) {
        this.m = modalAge;
        this.b = dispersion;
    }

    public static Mortality of(Sex sex, Health health) {
        double base = sex == Sex.MALE ? 87.6 : 90.4;
        double disp = sex == Sex.MALE ? 9.6 : 8.6;
        double shift = switch (health) {
            case EXCELLENT -> 3.5;
            case GOOD -> 1.5;
            case AVERAGE -> 0.0;
            case POOR -> -6.0;
        };
        return new Mortality(base + shift, disp);
    }

    /** Probability a person aged {@code x} survives {@code t} more years. */
    public double survival(double x, double t) {
        if (t <= 0) return 1.0;
        return Math.exp(Math.exp((x - m) / b) * (1 - Math.exp(t / b)));
    }

    /** Remaining life expectancy at age x (years). */
    public double lifeExpectancy(double x) {
        double sum = 0;
        double dt = 1.0 / 12;
        for (double t = dt / 2; t < 70; t += dt) sum += survival(x, t) * dt;
        return sum;
    }

    /** Age by which the person is still alive with the given probability (e.g. 0.25 → "1 in 4 live to"). */
    public double ageWithSurvival(double x, double p) {
        double t = 0;
        while (t < 70 && survival(x, t) > p) t += 0.1;
        return x + t;
    }

    /** Sample an age at death for someone alive at age x (inverse-transform). */
    public double sampleDeathAge(double x, RandomGenerator rng) {
        double u = rng.nextDouble();
        // Solve S(x, t) = u for t: e^{(x-m)/b}(1 - e^{t/b}) = ln u
        double k = Math.exp((x - m) / b);
        double t = b * Math.log(1 - Math.log(Math.max(u, 1e-12)) / k);
        return x + t;
    }
}
