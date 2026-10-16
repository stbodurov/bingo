package com.sundial.engine;

import com.sundial.engine.Model.Assumptions;
import com.sundial.engine.Model.Household;
import com.sundial.engine.Model.Person;

import java.util.Arrays;
import java.util.SplittableRandom;

/**
 * Monte Carlo over market returns *and* lifetimes. Each path draws correlated lognormal
 * stock/bond returns and an age at death for each person from their mortality curve, so
 * "success" means the money lasted as long as the people did — not to an arbitrary age.
 */
public final class MonteCarlo {

    public record Summary(int paths, double successRate, double medianEndingReal, double p10EndingReal,
                          double p90EndingReal, double[] depletionAgeHistogram, double probRunOutBefore85,
                          double medianLifetimeSpendingReal, double[] fanP10, double[] fanP50, double[] fanP90,
                          int[] fanYears) {}

    private final Household h;
    private final Simulator sim;
    private final long seed;

    public MonteCarlo(Household h, Simulator sim, long seed) {
        this.h = h;
        this.sim = sim;
        this.seed = seed;
    }

    public Summary run(Simulator.Settings st, int paths) {
        Assumptions a = h.assumptions();
        SplittableRandom rng = new SplittableRandom(seed);
        int n = sim.horizonYears();
        double[] endings = new double[paths];
        double[] spent = new double[paths];
        double[][] wealth = new double[n][paths];
        for (double[] row : wealth) Arrays.fill(row, Double.NaN);
        int ok = 0, before85 = 0;
        double[] hist = new double[8]; // depletion age buckets: <70,70-74,...,95+
        double muS = Math.log(1 + a.stockReturn()) - a.stockVol() * a.stockVol() / 2;
        double muB = Math.log(1 + a.bondReturn()) - a.bondVol() * a.bondVol() / 2;
        double rho = 0.1;
        Person first = h.people().get(0);
        for (int p = 0; p < paths; p++) {
            double[] s = new double[n], b = new double[n];
            for (int k = 0; k < n; k++) {
                double z1 = gaussian(rng), z2 = rho * z1 + Math.sqrt(1 - rho * rho) * gaussian(rng);
                s[k] = Math.exp(muS + a.stockVol() * z1) - 1;
                b[k] = Math.exp(muB + a.bondVol() * z2) - 1;
            }
            double[] deaths = new double[h.people().size()];
            for (int i = 0; i < deaths.length; i++) {
                Person per = h.people().get(i);
                double age = sim.startYear() - per.birthYear();
                deaths[i] = Math.min(a.planToAge() + 10, Mortality.of(per.sex(), per.health()).sampleDeathAge(age, rng));
            }
            Simulator.Run r = sim.run(st, new Simulator.Path(s, b), deaths);
            endings[p] = r.endingReal();
            spent[p] = r.lifetimeSpendingReal();
            for (int k = 0; k < r.years().size() && k < n; k++) {
                Simulator.Year y = r.years().get(k);
                wealth[k][p] = y.portfolio() / y.inflationIndex();
            }
            if (r.success()) ok++;
            else {
                int age = r.depletionYear() - first.birthYear();
                if (age < 85) before85++;
                int bucket = Math.max(0, Math.min(7, (age - 65) / 5));
                hist[bucket]++;
            }
        }
        Arrays.sort(endings);
        Arrays.sort(spent);
        int years = 0;
        while (years < n && countFinite(wealth[years]) > paths / 10) years++;
        double[] p10 = new double[years], p50 = new double[years], p90 = new double[years];
        int[] ys = new int[years];
        for (int k = 0; k < years; k++) {
            double[] col = Arrays.stream(wealth[k]).filter(v -> !Double.isNaN(v)).sorted().toArray();
            ys[k] = sim.startYear() + k;
            if (col.length == 0) continue;
            p10[k] = col[(int) (0.10 * (col.length - 1))];
            p50[k] = col[(int) (0.50 * (col.length - 1))];
            p90[k] = col[(int) (0.90 * (col.length - 1))];
        }
        for (int i = 0; i < hist.length; i++) hist[i] /= paths;
        return new Summary(paths, (double) ok / paths, pct(endings, 0.5), pct(endings, 0.1), pct(endings, 0.9), hist,
                (double) before85 / paths, pct(spent, 0.5), p10, p50, p90, ys);
    }

    private static int countFinite(double[] row) {
        int c = 0;
        for (double v : row) if (!Double.isNaN(v)) c++;
        return c;
    }

    private static double pct(double[] sorted, double q) {
        return sorted[(int) Math.round(q * (sorted.length - 1))];
    }

    private static double gaussian(SplittableRandom r) {
        double u1 = Math.max(r.nextDouble(), 1e-12), u2 = r.nextDouble();
        return Math.sqrt(-2 * Math.log(u1)) * Math.cos(2 * Math.PI * u2);
    }

    /** Highest spending scale that keeps the success rate at or above {@code target}. */
    public double safeScale(Simulator.Settings st, double target, int paths) {
        double lo = 0.2, hi = 3.0;
        if (run(st.withScale(lo), paths).successRate() < target) return lo;
        for (int i = 0; i < 12; i++) {
            double mid = (lo + hi) / 2;
            if (run(st.withScale(mid), paths).successRate() >= target) lo = mid;
            else hi = mid;
        }
        return lo;
    }
}
