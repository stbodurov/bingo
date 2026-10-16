package com.sundial.engine;

import com.sundial.engine.Model.Health;
import com.sundial.engine.Model.Person;
import com.sundial.engine.Model.Sex;
import org.junit.jupiter.api.Test;

import java.time.LocalDate;
import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.within;

class EngineTest {

    @Test
    void fullRetirementAgeByBirthYear() {
        assertThat(SocialSecurity.fraMonths(1954)).isEqualTo(66 * 12);
        assertThat(SocialSecurity.fraMonths(1957)).isEqualTo(66 * 12 + 6);
        assertThat(SocialSecurity.fraMonths(1960)).isEqualTo(67 * 12);
    }

    @Test
    void claimingFactorsMatchSsaTables() {
        int fra = 67 * 12;
        assertThat(SocialSecurity.retirementFactor(62 * 12, fra)).isCloseTo(0.70, within(1e-9));
        assertThat(SocialSecurity.retirementFactor(65 * 12, fra)).isCloseTo(0.8667, within(1e-4));
        assertThat(SocialSecurity.retirementFactor(70 * 12, fra)).isCloseTo(1.24, within(1e-9));
        assertThat(SocialSecurity.spousalFactor(62 * 12, fra)).isCloseTo(0.65, within(1e-9));
    }

    @Test
    void spousalAndSurvivorBenefits() {
        Person low = new Person("A", 1962, 1, 5, Sex.FEMALE, Health.AVERAGE, 800, 0, 0, 0);
        Person high = new Person("B", 1962, 1, 5, Sex.MALE, Health.AVERAGE, 3000, 0, 0, 0);
        // At FRA the lower earner gets their own 800 + excess spousal (1500 - 800).
        assertThat(SocialSecurity.withSpousal(low, 67 * 12, high)).isCloseTo(1500, within(1e-6));
        // Survivor of someone who waited to 70 inherits the delayed credits.
        double surv = SocialSecurity.survivorBenefit(low, high, 70 * 12, 75 * 12, 75 * 12, 800);
        assertThat(surv).isCloseTo(3000 * 1.24, within(1e-6));
        // Widow(er)'s limit: early claimer's survivor gets at least 82.5% of PIA.
        double early = SocialSecurity.survivorBenefit(low, high, 62 * 12, 70 * 12, 70 * 12, 800);
        assertThat(early).isCloseTo(0.825 * 3000, within(1e-6));
    }

    @Test
    void mortalityIsCalibrated() {
        assertThat(Mortality.of(Sex.MALE, Health.AVERAGE).lifeExpectancy(65)).isBetween(18.5, 21.0);
        assertThat(Mortality.of(Sex.FEMALE, Health.AVERAGE).lifeExpectancy(65)).isBetween(21.0, 23.5);
        assertThat(Mortality.of(Sex.MALE, Health.POOR).lifeExpectancy(65))
                .isLessThan(Mortality.of(Sex.MALE, Health.EXCELLENT).lifeExpectancy(65));
    }

    @Test
    void annuityPricingMatchesMarket() {
        double monthly = Annuities.monthlyFor(100_000, 65, Sex.MALE, 0, 0, null, null);
        assertThat(monthly).isBetween(600.0, 680.0);
        double joint = Annuities.monthlyFor(100_000, 65, Sex.MALE, 0, 0, Sex.FEMALE, 65);
        assertThat(joint).isLessThan(monthly);
        double deferred = Annuities.monthlyFor(100_000, 65, Sex.MALE, 15, 0, null, null);
        assertThat(deferred).isGreaterThan(monthly * 2.5);
    }

    @Test
    void taxCode2026() {
        // Married, both 65+, $40k IRA + $50k Social Security.
        var r = TaxCode.compute(new TaxCode.Input(2026, true, 2, 0, 40_000, 50_000, 0, 0, 0.025));
        double taxableSs = TaxCode.taxableSocialSecurity(true, 40_000, 50_000);
        assertThat(taxableSs).isCloseTo(0.85 * (65_000 - 44_000) + 6_000, within(1e-6));
        // Deduction: 32,200 + 2*1,650 + 2*6,000 (senior deduction, under phase-out).
        assertThat(r.deduction()).isCloseTo(47_500, within(1e-6));
        assertThat(r.federal()).isGreaterThan(0);
        // Single, $30k ordinary income, under 65: 10% on 12,400 + 12% on the rest of (30,000 - 16,100).
        var s = TaxCode.compute(new TaxCode.Input(2026, false, 0, 30_000, 0, 0, 0, 0, 0.025));
        assertThat(s.ordinaryTax()).isCloseTo(1_240 + 0.12 * (13_900 - 12_400), within(1e-6));
        // Senior deduction ends after 2028.
        assertThat(TaxCode.seniorDeduction(2029, true, 2, 50_000)).isZero();
        // IRMAA first tier.
        assertThat(TaxCode.irmaaPerPerson(2026, true, 230_000, 0.025)).isCloseTo((81.20 + 14.50) * 12, within(1e-6));
    }

    @Test
    void rmdStartAgesAndDivisors() {
        assertThat(Rmd.startAge(1955)).isEqualTo(73);
        assertThat(Rmd.startAge(1962)).isEqualTo(75);
        assertThat(Rmd.amount(1953, 73, 265_000)).isCloseTo(10_000, within(1e-6));
        assertThat(Rmd.amount(1962, 73, 265_000)).isZero();
    }

    @Test
    void claimingOptimizerPrefersDelayForHealthyHigherEarner() {
        var people = List.of(
                new Person("A", 1963, 4, 14, Sex.FEMALE, Health.GOOD, 1_200, 0, 0, 0),
                new Person("B", 1963, 4, 14, Sex.MALE, Health.GOOD, 3_400, 0, 0, 0));
        var r = new ClaimingOptimizer(people, 2026, 10, 0.018, false).optimise();
        assertThat(r.best().claimMonths()[1]).isGreaterThanOrEqualTo(68 * 12);
        assertThat(r.gainVsEarliest()).isGreaterThan(0);
        var poor = List.of(new Person("C", 1963, 4, 14, Sex.MALE, Health.POOR, 2_000, 0, 0, 0));
        var rp = new ClaimingOptimizer(poor, 2026, 10, 0.03, false).optimise();
        assertThat(rp.best().claimMonths()[0]).isLessThan(67 * 12);
    }

    @Test
    void fullPlanIsConsistent() {
        var plan = new Planner(Samples.couple(), LocalDate.of(2026, 10, 16), 400).build();
        assertThat(plan.summary().successRate()).isBetween(0.0, 1.0);
        assertThat(plan.summary().safeSpending()).isPositive();
        assertThat(plan.years()).isNotEmpty();
        assertThat(plan.strategies()).anyMatch(Planner.StrategyView::recommended);
        assertThat(plan.paycheck()).hasSize(12);
        assertThat(plan.insights()).isNotEmpty();
        assertThat(plan.monteCarlo().fanP50().length).isGreaterThan(20);
        var single = new Planner(Samples.single(), LocalDate.of(2026, 10, 16), 300).build();
        assertThat(single.claiming().recommended()).hasSize(1);
    }

    @Test
    void simulatorIsDeterministicAndConservesMoney() {
        var h = Samples.single();
        var sim = new Simulator(h, 2026);
        var st = new Simulator.Settings(new int[]{70 * 12}, Model.Strategy.CONVENTIONAL, 0, false, 1.0, false, false, 0.025, 0, 0, 0);
        var a = sim.run(st, sim.expectedPath(), sim.planDeathAges());
        var b = sim.run(st, sim.expectedPath(), sim.planDeathAges());
        assertThat(a.endingNominal()).isEqualTo(b.endingNominal());
        var y0 = a.years().get(0);
        assertThat(y0.spendingMet()).isCloseTo(y0.spendingTarget(), within(1.0));
    }

    @Test
    void claimingTitleNamesEachSpouse() {
        Person ellen = new Person("Ellen", 1963, 4, 14, Sex.FEMALE, Health.GOOD, 2_380, 64, 118_000, 0);
        Person marcus = new Person("Marcus", 1965, 9, 3, Sex.MALE, Health.AVERAGE, 3_140, 65, 142_000, 0);
        assertThat(Planner.claimingTitle(List.of(ellen, marcus), new int[]{66 * 12 + 1, 70 * 12}))
                .isEqualTo("Claim Social Security: Ellen at 66 and 1 month, Marcus at 70");
        assertThat(Planner.claimingTitle(List.of(ellen), new int[]{67 * 12 + 6}))
                .isEqualTo("Claim Social Security at 67 and 6 months");
    }

    @Test
    void paycheckDepositsAddUpToTheHeadlinePaycheck() {
        var plan = new Planner(Samples.couple(), LocalDate.of(2026, 10, 16), 200).build();
        var month = plan.paycheck().get(0);
        System.out.printf("headline %.0f, deposits %.0f, lines %s%n", plan.summary().monthlyPaycheck(), month.total(), month.lines());
        // Ellen collects her own full benefit; it must not be diluted by Marcus, who claims later.
        var ellen = month.lines().stream().filter(l -> l.source().contains("Ellen")).findFirst().orElseThrow();
        assertThat(ellen.amount()).isGreaterThan(1_500);
        assertThat(month.total()).isCloseTo(plan.summary().monthlyPaycheck(), within(plan.summary().monthlyPaycheck() * 0.05));
    }
}
