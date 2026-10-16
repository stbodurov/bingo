package com.sundial.api;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.sundial.engine.Model;
import com.sundial.engine.Model.Household;
import com.sundial.engine.Model.Person;
import com.sundial.engine.Planner;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.time.LocalDate;
import java.time.ZoneId;
import java.util.ArrayList;
import java.util.HexFormat;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/** Validates households and builds plans, with a small LRU cache keyed by the input. */
@Service
public class PlanService {
    private final ObjectMapper json;
    private final int paths;
    private final Map<String, Planner.Plan> cache = new LinkedHashMap<>(64, 0.75f, true) {
        @Override
        protected boolean removeEldestEntry(Map.Entry<String, Planner.Plan> e) { return size() > 200; }
    };

    public PlanService(ObjectMapper json, @Value("${sundial.monte-carlo-paths:1000}") int paths) {
        this.json = json;
        this.paths = paths;
    }

    public static LocalDate today() {
        String override = System.getenv("SUNDIAL_TODAY");
        return override != null ? LocalDate.parse(override) : LocalDate.now(ZoneId.of("America/New_York"));
    }

    public Planner.Plan plan(Household h) {
        List<String> errors = validate(h);
        if (!errors.isEmpty()) throw new InvalidHousehold(errors);
        Household normalized = normalize(h);
        String key = key(normalized);
        synchronized (cache) {
            Planner.Plan hit = cache.get(key);
            if (hit != null) return hit;
        }
        Planner.Plan plan = new Planner(normalized, today(), paths).build();
        synchronized (cache) { cache.put(key, plan); }
        return plan;
    }

    private String key(Household h) {
        try {
            byte[] digest = MessageDigest.getInstance("SHA-256").digest(json.writeValueAsString(h).getBytes(StandardCharsets.UTF_8));
            return HexFormat.of().formatHex(digest) + today();
        } catch (JsonProcessingException | NoSuchAlgorithmException e) {
            throw new IllegalStateException(e);
        }
    }

    /** Fill optional collections and defaults so the engine never sees nulls. */
    static Household normalize(Household h) {
        Model.Assumptions a = h.assumptions() == null ? Model.Assumptions.defaults() : h.assumptions();
        Model.Accounts ac = h.accounts();
        double basis = ac.taxableBasis() <= 0 ? ac.taxable() * 0.7 : Math.min(ac.taxableBasis(), ac.taxable());
        return new Household(h.people(), new Model.Accounts(ac.cash(), ac.taxable(), basis, ac.traditional(), ac.roth()),
                h.essentialSpending(), h.discretionarySpending(),
                h.pensions() == null ? List.of() : h.pensions(),
                h.oneTimes() == null ? List.of() : h.oneTimes(),
                h.annuities() == null ? List.of() : h.annuities(),
                h.strategy() == null ? Model.Strategy.CONVENTIONAL : h.strategy(),
                h.conversionBracket(), h.guardrails(), a);
    }

    public static List<String> validate(Household h) {
        List<String> e = new ArrayList<>();
        if (h == null) return List.of("Household is required.");
        if (h.people() == null || h.people().isEmpty() || h.people().size() > 2) e.add("Add one or two people.");
        int year = today().getYear();
        if (h.people() != null) for (int i = 0; i < h.people().size(); i++) {
            Person p = h.people().get(i);
            String who = p.name() == null || p.name().isBlank() ? "Person " + (i + 1) : p.name();
            if (p.birthYear() < year - 90 || p.birthYear() > year - 40) e.add(who + ": birth year must make them 40–90 years old.");
            if (p.birthMonth() < 1 || p.birthMonth() > 12) e.add(who + ": birth month must be 1–12.");
            if (p.birthDay() < 1 || p.birthDay() > 31) e.add(who + ": birth day must be 1–31.");
            if (p.sex() == null || p.health() == null) e.add(who + ": choose sex at birth and health (they drive life expectancy).");
            if (p.pia() < 0 || p.pia() > 6_000) e.add(who + ": monthly Social Security at full retirement age must be $0–$6,000.");
            if (p.salary() < 0 || p.salary() > 5_000_000) e.add(who + ": salary out of range.");
            if (p.retireAge() < 0 || p.retireAge() > 80) e.add(who + ": retirement age must be 0–80.");
            if (p.claimAgeMonths() != 0 && (p.claimAgeMonths() < 62 * 12 || p.claimAgeMonths() > 70 * 12)) e.add(who + ": claiming age must be 62–70.");
            if (p.name() != null && p.name().length() > 40) e.add("Names are limited to 40 characters.");
        }
        Model.Accounts a = h.accounts();
        if (a == null) e.add("Add your account balances.");
        else if (a.cash() < 0 || a.taxable() < 0 || a.traditional() < 0 || a.roth() < 0 || a.total() > 100_000_000) e.add("Account balances must be $0–$100M.");
        if (h.essentialSpending() < 0 || h.discretionarySpending() < 0 || h.essentialSpending() + h.discretionarySpending() > 2_000_000)
            e.add("Spending must be $0–$2M a year.");
        if (h.essentialSpending() + h.discretionarySpending() <= 0) e.add("Enter what you spend in a year.");
        if (h.pensions() != null && h.pensions().size() > 6) e.add("Up to six pensions.");
        if (h.oneTimes() != null && h.oneTimes().size() > 20) e.add("Up to twenty one-time expenses.");
        if (h.conversionBracket() < 0 || h.conversionBracket() > 0.37) e.add("Conversion bracket must be between 0% and 37%.");
        Model.Assumptions as = h.assumptions();
        if (as != null) {
            if (as.inflation() < 0 || as.inflation() > 0.1) e.add("Inflation must be 0–10%.");
            if (as.stockShare() < 0 || as.stockShare() > 1) e.add("Stock share must be 0–100%.");
            if (as.planToAge() < 85 || as.planToAge() > 110) e.add("Plan-to age must be 85–110.");
            if (as.stateTaxRate() < 0 || as.stateTaxRate() > 0.15) e.add("State tax rate must be 0–15%.");
            if (as.stockReturn() < -0.05 || as.stockReturn() > 0.15 || as.bondReturn() < -0.05 || as.bondReturn() > 0.1)
                e.add("Return assumptions are outside a plausible range.");
        }
        return e;
    }

    public static class InvalidHousehold extends RuntimeException {
        public final List<String> errors;
        public InvalidHousehold(List<String> errors) {
            super(String.join(" ", errors));
            this.errors = errors;
        }
    }
}
