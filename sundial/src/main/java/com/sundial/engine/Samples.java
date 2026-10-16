package com.sundial.engine;

import com.sundial.engine.Model.Accounts;
import com.sundial.engine.Model.Assumptions;
import com.sundial.engine.Model.Health;
import com.sundial.engine.Model.Household;
import com.sundial.engine.Model.OneTime;
import com.sundial.engine.Model.Pension;
import com.sundial.engine.Model.Person;
import com.sundial.engine.Model.Sex;
import com.sundial.engine.Model.Strategy;

import java.util.List;

/** Illustrative households used by the demo, the landing page and the tests (fictional people). */
public final class Samples {
    private Samples() {}

    /** A married couple, 63 and 61, $1.15M saved, one small pension. */
    public static Household couple() {
        return new Household(
                List.of(new Person("Ellen", 1963, 4, 14, Sex.FEMALE, Health.GOOD, 2_380, 64, 118_000, 0),
                        new Person("Marcus", 1965, 9, 3, Sex.MALE, Health.AVERAGE, 3_140, 65, 142_000, 0)),
                new Accounts(45_000, 230_000, 160_000, 760_000, 115_000),
                86_000, 42_000,
                List.of(new Pension(0, 1_150, 65, false, 0.5)),
                List.of(new OneTime(2027, 38_000, "New car"), new OneTime(2029, 25_000, "Kitchen")),
                List.of(), Strategy.CONVENTIONAL, 0, true, Assumptions.defaults());
    }

    /** A single retiree, 66, $640k saved, no pension. */
    public static Household single() {
        return new Household(
                List.of(new Person("Rosa", 1960, 6, 22, Sex.FEMALE, Health.AVERAGE, 2_210, 0, 0, 0)),
                new Accounts(35_000, 95_000, 70_000, 455_000, 55_000),
                44_000, 12_000, List.of(), List.of(), List.of(), Strategy.CONVENTIONAL, 0, true, Assumptions.defaults());
    }
}
