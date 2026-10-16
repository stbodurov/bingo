package com.sundial.engine;

import com.sundial.engine.Model.Person;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;

/**
 * Chooses Social Security claiming ages by maximising the expected present value of the
 * household's lifetime benefits, weighting every future month by the probability each
 * spouse is alive and paying the right benefit in each state (both alive with spousal top-
 * ups, either survivor with the survivor benefit). Searches every month from 62 to 70.
 */
public final class ClaimingOptimizer {

    public record Option(int[] claimMonths, double expectedPv, double monthlyAtStart) {}

    public record Result(Option best, Option earliest, Option fra, Option latest, List<Option> top,
                         double gainVsEarliest, int breakevenAge, double[][] grid) {}

    private final List<Person> people;
    private final int currentYear;
    private final int currentMonth;
    private final double realDiscount;
    private final boolean haircut;

    public ClaimingOptimizer(List<Person> people, int currentYear, int currentMonth, double realDiscount, boolean haircut) {
        this.people = people;
        this.currentYear = currentYear;
        this.currentMonth = currentMonth;
        this.realDiscount = realDiscount;
        this.haircut = haircut;
    }

    int ageMonthsNow(Person p) {
        return (currentYear - p.birthYear()) * 12 + (currentMonth - p.birthMonth());
    }

    /** Earliest claim age still available (can't claim in the past). */
    int minClaim(Person p) {
        return Math.min(SocialSecurity.LATEST, Math.max(SocialSecurity.EARLIEST, ageMonthsNow(p)));
    }

    public double expectedPv(int[] claim) {
        Person a = people.get(0);
        Person b = people.size() > 1 ? people.get(1) : null;
        Mortality ma = Mortality.of(a.sex(), a.health());
        Mortality mb = b == null ? null : Mortality.of(b.sex(), b.health());
        double ageA = ageMonthsNow(a) / 12.0, ageB = b == null ? 0 : ageMonthsNow(b) / 12.0;
        double v = Math.pow(1 + realDiscount, -1.0 / 12);
        double pv = 0, disc = 1;
        for (int t = 1; t <= 12 * 50; t++) {
            disc *= v;
            int amA = ageMonthsNow(a) + t;
            double pa = ma.survival(ageA, t / 12.0);
            double cut = SocialSecurity.payable(currentYear + (currentMonth - 1 + t) / 12, haircut);
            if (b == null) {
                if (amA >= claim[0]) pv += disc * pa * SocialSecurity.ownBenefit(a, claim[0]) * cut;
                continue;
            }
            int amB = ageMonthsNow(b) + t;
            double pb = mb.survival(ageB, t / 12.0);
            double aOn = amA >= claim[0] ? 1 : 0, bOn = amB >= claim[1] ? 1 : 0;
            double both = 0;
            if (aOn > 0) both += bOn > 0 ? SocialSecurity.withSpousal(a, claim[0], b) : SocialSecurity.ownBenefit(a, claim[0]);
            if (bOn > 0) both += aOn > 0 ? SocialSecurity.withSpousal(b, claim[1], a) : SocialSecurity.ownBenefit(b, claim[1]);
            // Survivor states: approximate the deceased's death at the midpoint of the elapsed time.
            double aOwn = aOn > 0 ? SocialSecurity.ownBenefit(a, claim[0]) : 0;
            double bOwn = bOn > 0 ? SocialSecurity.ownBenefit(b, claim[1]) : 0;
            int bDeath = ageMonthsNow(b) + t / 2, aDeath = ageMonthsNow(a) + t / 2;
            double aSurv = SocialSecurity.survivorBenefit(a, b, claim[1] <= bDeath ? claim[1] : -1, bDeath, Math.max(amA, 720), aOwn);
            double bSurv = SocialSecurity.survivorBenefit(b, a, claim[0] <= aDeath ? claim[0] : -1, aDeath, Math.max(amB, 720), bOwn);
            if (amA < 720) aSurv = aOwn; // survivor benefits from 60
            if (amB < 720) bSurv = bOwn;
            pv += disc * cut * (pa * pb * both + pa * (1 - pb) * aSurv + (1 - pa) * pb * bSurv);
        }
        return pv;
    }

    private Option option(int[] claim) {
        double start = 0;
        Person a = people.get(0);
        if (people.size() == 1) start = SocialSecurity.ownBenefit(a, claim[0]);
        else {
            start = SocialSecurity.withSpousal(a, claim[0], people.get(1)) + SocialSecurity.withSpousal(people.get(1), claim[1], a);
        }
        return new Option(claim, expectedPv(claim), start);
    }

    public Result optimise() {
        Person a = people.get(0);
        int minA = minClaim(a);
        List<Option> all = new ArrayList<>();
        double[][] grid;
        if (people.size() == 1) {
            grid = new double[1][SocialSecurity.LATEST - minA + 1];
            for (int m = minA; m <= SocialSecurity.LATEST; m++) {
                Option o = option(new int[]{m});
                grid[0][m - minA] = o.expectedPv();
                all.add(o);
            }
        } else {
            Person b = people.get(1);
            int minB = minClaim(b);
            // Coarse (quarterly) search, then refine monthly around the best.
            grid = new double[(SocialSecurity.LATEST - minA) / 3 + 1][(SocialSecurity.LATEST - minB) / 3 + 1];
            for (int i = 0; i < grid.length; i++)
                for (int j = 0; j < grid[0].length; j++) {
                    Option o = option(new int[]{minA + 3 * i, minB + 3 * j});
                    grid[i][j] = o.expectedPv();
                    all.add(o);
                }
            Option coarse = all.stream().max(Comparator.comparingDouble(Option::expectedPv)).orElseThrow();
            for (int da = -2; da <= 2; da++)
                for (int db = -2; db <= 2; db++) {
                    int ca = coarse.claimMonths()[0] + da, cb = coarse.claimMonths()[1] + db;
                    if (ca < minA || ca > SocialSecurity.LATEST || cb < minB || cb > SocialSecurity.LATEST) continue;
                    all.add(option(new int[]{ca, cb}));
                }
        }
        all.sort(Comparator.comparingDouble(Option::expectedPv).reversed());
        Option best = all.get(0);
        int n = people.size();
        int[] early = new int[n], fra = new int[n], late = new int[n];
        for (int i = 0; i < n; i++) {
            early[i] = minClaim(people.get(i));
            fra[i] = Math.max(early[i], SocialSecurity.fraMonths(people.get(i).birthYear()));
            late[i] = SocialSecurity.LATEST;
        }
        Option e = option(early), f = option(fra), l = option(late);
        return new Result(best, e, f, l, all.subList(0, Math.min(5, all.size())),
                best.expectedPv() - e.expectedPv(), breakeven(a), grid);
    }

    /** Age at which cumulative benefits from claiming at 70 overtake claiming at 62 (single, no discounting). */
    int breakeven(Person p) {
        double early = SocialSecurity.ownBenefit(p, 62 * 12), late = SocialSecurity.ownBenefit(p, 70 * 12);
        double cumE = 0, cumL = 0;
        for (int m = 62 * 12; m < 110 * 12; m++) {
            cumE += early;
            if (m >= 70 * 12) cumL += late;
            if (cumL > cumE) return m / 12;
        }
        return 110;
    }
}
