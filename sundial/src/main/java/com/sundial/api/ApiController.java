package com.sundial.api;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.sundial.engine.Annuities;
import com.sundial.engine.Model;
import com.sundial.engine.Model.Household;
import com.sundial.engine.Planner;
import com.sundial.engine.Samples;
import com.sundial.store.Entities;
import com.sundial.store.Repos;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import jakarta.validation.Valid;
import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.Max;
import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.AuthenticationException;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.security.web.context.SecurityContextRepository;
import org.springframework.security.web.csrf.CsrfToken;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.server.ResponseStatusException;

import java.security.MessageDigest;
import java.security.SecureRandom;
import java.time.Instant;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api")
public class ApiController {
    private final PlanService plans;
    private final Repos.Users users;
    private final Repos.Plans saved;
    private final Repos.Waitlist waitlist;
    private final Repos.AdvisorRequests advisor;
    private final Repos.QuoteRequests quotes;
    private final PasswordEncoder encoder;
    private final AuthenticationManager auth;
    private final SecurityContextRepository contexts;
    private final ObjectMapper json;
    private final String adminToken;
    private static final SecureRandom RNG = new SecureRandom();

    public ApiController(PlanService plans, Repos.Users users, Repos.Plans saved, Repos.Waitlist waitlist,
                         Repos.AdvisorRequests advisor, Repos.QuoteRequests quotes, PasswordEncoder encoder,
                         AuthenticationManager auth, SecurityContextRepository contexts, ObjectMapper json,
                         @Value("${sundial.admin-token}") String adminToken) {
        this.plans = plans; this.users = users; this.saved = saved; this.waitlist = waitlist; this.advisor = advisor;
        this.quotes = quotes; this.encoder = encoder; this.auth = auth; this.contexts = contexts; this.json = json;
        this.adminToken = adminToken;
    }

    // ── Planning ─────────────────────────────────────────────────────────────

    @PostMapping("/plan")
    public Planner.Plan plan(@RequestBody Household household) {
        return plans.plan(household);
    }

    public record ClaimingRequest(List<Model.Person> people, Double realDiscount) {}

    public record ClaimingAnswer(int[] best, double bestMonthly, double gainVsEarliest, int breakevenAge,
                                 List<com.sundial.engine.ClaimingOptimizer.Option> compare, double[] lifeExpectancy) {}

    /** The landing page's quick "when should I claim?" calculator. */
    @PostMapping("/claiming")
    public ClaimingAnswer claiming(@RequestBody ClaimingRequest r) {
        if (r.people() == null || r.people().isEmpty() || r.people().size() > 2)
            throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "One or two people.");
        Household probe = new Household(r.people(), new Model.Accounts(0, 0, 0, 0, 0), 1, 0, List.of(), List.of(), List.of(),
                Model.Strategy.CONVENTIONAL, 0, false, Model.Assumptions.defaults());
        List<String> errors = PlanService.validate(probe);
        if (!errors.isEmpty()) throw new PlanService.InvalidHousehold(errors);
        var today = PlanService.today();
        var opt = new com.sundial.engine.ClaimingOptimizer(r.people(), today.getYear(), today.getMonthValue(),
                r.realDiscount() == null ? 0.018 : r.realDiscount(), false);
        var res = opt.optimise();
        double[] le = r.people().stream().mapToDouble(p -> {
            int a = today.getYear() - p.birthYear();
            return a + com.sundial.engine.Mortality.of(p.sex(), p.health()).lifeExpectancy(a);
        }).toArray();
        return new ClaimingAnswer(res.best().claimMonths(), res.best().monthlyAtStart(), res.gainVsEarliest(), res.breakevenAge(),
                List.of(res.earliest(), res.fra(), res.latest(), res.best()), le);
    }

    @GetMapping("/samples/{name}")
    public Household sample(@PathVariable String name) {
        return switch (name) {
            case "couple" -> Samples.couple();
            case "single" -> Samples.single();
            default -> throw new ResponseStatusException(HttpStatus.NOT_FOUND);
        };
    }

    @GetMapping("/annuities/quotes")
    public List<Annuities.Quote> annuityQuotes(@RequestParam double premium, @RequestParam int age,
                                               @RequestParam Model.Sex sex, @RequestParam(required = false) Model.Sex spouseSex,
                                               @RequestParam(required = false) Integer spouseAge,
                                               @RequestParam(defaultValue = "0") double ira) {
        if (premium < 10_000 || premium > 5_000_000 || age < 50 || age > 90)
            throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "Premium $10k–$5M, age 50–90.");
        return Annuities.quotes(premium, age, sex, spouseSex, spouseAge, ira);
    }

    // ── Auth ─────────────────────────────────────────────────────────────────

    public record SignupRequest(@NotBlank @Size(max = 80) String name, @Email @NotBlank String email,
                                @Size(min = 10, max = 200) String password) {}

    public record LoginRequest(@NotBlank String email, @NotBlank String password) {}

    public record Me(Long id, String name, String email) {}

    @GetMapping("/auth/csrf")
    public Map<String, String> csrf(CsrfToken token) {
        return Map.of("headerName", token.getHeaderName(), "token", token.getToken());
    }

    @PostMapping("/auth/signup")
    @Transactional
    public Me signup(@Valid @RequestBody SignupRequest r, HttpServletRequest req, HttpServletResponse res) {
        if (users.findByEmailIgnoreCase(r.email()).isPresent())
            throw new ResponseStatusException(HttpStatus.CONFLICT, "An account with that email exists — sign in instead.");
        Entities.UserAccount u = new Entities.UserAccount();
        u.email = r.email().trim().toLowerCase();
        u.name = r.name().trim();
        u.passwordHash = encoder.encode(r.password());
        users.save(u);
        login(new LoginRequest(u.email, r.password()), req, res);
        return new Me(u.id, u.name, u.email);
    }

    @PostMapping("/auth/login")
    public Me login(@Valid @RequestBody LoginRequest r, HttpServletRequest req, HttpServletResponse res) {
        try {
            Authentication a = auth.authenticate(new UsernamePasswordAuthenticationToken(r.email().trim().toLowerCase(), r.password()));
            var ctx = SecurityContextHolder.createEmptyContext();
            ctx.setAuthentication(a);
            SecurityContextHolder.setContext(ctx);
            if (req.getSession(false) != null) req.changeSessionId(); // session-fixation defence
            else req.getSession(true);
            contexts.saveContext(ctx, req, res);
        } catch (AuthenticationException e) {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "That email and password don't match.");
        }
        var u = users.findByEmailIgnoreCase(r.email()).orElseThrow();
        return new Me(u.id, u.name, u.email);
    }

    @GetMapping("/auth/me")
    public ResponseEntity<Me> me(Authentication a) {
        if (a == null || !a.isAuthenticated() || "anonymousUser".equals(a.getName())) return ResponseEntity.noContent().build();
        return users.findByEmailIgnoreCase(a.getName()).map(u -> ResponseEntity.ok(new Me(u.id, u.name, u.email)))
                .orElse(ResponseEntity.noContent().build());
    }

    private Entities.UserAccount current(Authentication a) {
        return users.findByEmailIgnoreCase(a.getName()).orElseThrow(() -> new ResponseStatusException(HttpStatus.UNAUTHORIZED));
    }

    // ── Saved plans ──────────────────────────────────────────────────────────

    public record SavedPlanView(Long id, String name, Instant updatedAt, Double successRate, Double safeSpending, Household household) {}

    public record SaveRequest(@NotBlank @Size(max = 120) String name, Household household) {}

    private SavedPlanView view(Entities.SavedPlan p) {
        try {
            return new SavedPlanView(p.id, p.name, p.updatedAt, p.successRate, p.safeSpending, json.readValue(p.householdJson, Household.class));
        } catch (JsonProcessingException e) {
            throw new IllegalStateException(e);
        }
    }

    @GetMapping("/me/plans")
    public List<SavedPlanView> myPlans(Authentication a) {
        return saved.findByUserIdOrderByUpdatedAtDesc(current(a).id).stream().map(this::view).toList();
    }

    @PostMapping("/me/plans")
    @Transactional
    public SavedPlanView save(@Valid @RequestBody SaveRequest r, Authentication a) throws JsonProcessingException {
        return store(new Entities.SavedPlan(), r, current(a));
    }

    @PutMapping("/me/plans/{id}")
    @Transactional
    public SavedPlanView update(@PathVariable Long id, @Valid @RequestBody SaveRequest r, Authentication a) throws JsonProcessingException {
        var u = current(a);
        var p = saved.findByIdAndUserId(id, u.id).orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND));
        return store(p, r, u);
    }

    @DeleteMapping("/me/plans/{id}")
    @Transactional
    public ResponseEntity<Void> delete(@PathVariable Long id, Authentication a) {
        var p = saved.findByIdAndUserId(id, current(a).id).orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND));
        saved.delete(p);
        return ResponseEntity.noContent().build();
    }

    private SavedPlanView store(Entities.SavedPlan p, SaveRequest r, Entities.UserAccount u) throws JsonProcessingException {
        Planner.Plan plan = plans.plan(r.household());
        p.userId = u.id;
        p.name = r.name().trim();
        p.householdJson = json.writeValueAsString(r.household());
        p.successRate = plan.summary().successRate();
        p.safeSpending = plan.summary().safeSpending();
        p.updatedAt = Instant.now();
        saved.save(p);
        return view(p);
    }

    // ── Waitlist & leads ─────────────────────────────────────────────────────

    public record WaitlistRequest(@Email @NotBlank @Size(max = 200) String email, @Size(max = 80) String firstName,
                                  @Size(max = 20) String ageBand, @Size(max = 20) String assetsBand, @Size(max = 30) String status,
                                  List<@Size(max = 40) String> interests, @Size(max = 12) String referredBy,
                                  Map<String, Object> planSummary, @Size(max = 200) String website) {}

    public record WaitlistStatus(String code, String firstName, int position, int total, long referrals, String shareUrl) {}

    @PostMapping("/waitlist")
    @Transactional
    public WaitlistStatus join(@Valid @RequestBody WaitlistRequest r, HttpServletRequest req) throws JsonProcessingException {
        if (r.website() != null && !r.website().isBlank()) return new WaitlistStatus("", "", 0, 0, 0, ""); // honeypot
        var existing = waitlist.findByEmailIgnoreCase(r.email().trim());
        Entities.WaitlistEntry w = existing.orElseGet(Entities.WaitlistEntry::new);
        if (existing.isEmpty()) {
            w.email = r.email().trim().toLowerCase();
            w.code = code();
            w.referredBy = r.referredBy() == null || r.referredBy().isBlank() || waitlist.findByCode(r.referredBy()).isEmpty() ? null : r.referredBy();
            w.source = req.getHeader("Referer") == null ? "" : req.getHeader("Referer").substring(0, Math.min(120, req.getHeader("Referer").length()));
        }
        w.firstName = r.firstName();
        w.ageBand = r.ageBand();
        w.assetsBand = r.assetsBand();
        w.status = r.status();
        w.interests = r.interests() == null ? null : String.join(",", r.interests().stream().limit(8).toList());
        if (r.planSummary() != null) w.planSummary = json.writeValueAsString(r.planSummary());
        waitlist.save(w);
        return status(w, req);
    }

    @GetMapping("/waitlist/{code}")
    public WaitlistStatus waitlistStatus(@PathVariable String code, HttpServletRequest req) {
        return status(waitlist.findByCode(code).orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND)), req);
    }

    private WaitlistStatus status(Entities.WaitlistEntry w, HttpServletRequest req) {
        List<Entities.WaitlistEntry> all = waitlist.findAllByOrderByCreatedAtAsc();
        Map<String, Long> refs = new LinkedHashMap<>();
        for (var e : all) if (e.referredBy != null) refs.merge(e.referredBy, 1L, Long::sum);
        List<Entities.WaitlistEntry> ranked = new ArrayList<>(all);
        // Each referral moves you ahead of ten people; ties keep signup order.
        ranked.sort(Comparator.comparingLong((Entities.WaitlistEntry e) -> all.indexOf(e) - 10 * refs.getOrDefault(e.code, 0L)));
        int pos = ranked.indexOf(ranked.stream().filter(e -> e.id.equals(w.id)).findFirst().orElse(w)) + 1;
        String base = req.getScheme() + "://" + req.getServerName() + (req.getServerPort() == 80 || req.getServerPort() == 443 ? "" : ":" + req.getServerPort());
        return new WaitlistStatus(w.code, w.firstName, pos, all.size(), refs.getOrDefault(w.code, 0L), base + "/?ref=" + w.code);
    }

    private String code() {
        String alphabet = "abcdefghjkmnpqrstuvwxyz23456789";
        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < 8; i++) sb.append(alphabet.charAt(RNG.nextInt(alphabet.length())));
        return sb.toString();
    }

    public record AdvisorRequestBody(@NotBlank @Size(max = 120) String name, @Email @NotBlank String email,
                                     @Size(max = 40) String phone, @NotBlank @Size(max = 60) String topic,
                                     @Size(max = 1000) String message) {}

    @PostMapping("/advisor-requests")
    @Transactional
    public ResponseEntity<Map<String, String>> advisorRequest(@Valid @RequestBody AdvisorRequestBody b, Authentication a) {
        var r = new Entities.AdvisorRequest();
        r.userId = a == null ? null : users.findByEmailIgnoreCase(a.getName()).map(u -> u.id).orElse(null);
        r.name = b.name(); r.email = b.email(); r.phone = b.phone(); r.topic = b.topic(); r.message = b.message();
        advisor.save(r);
        return ResponseEntity.status(HttpStatus.CREATED).body(Map.of("status", "received"));
    }

    public record QuoteRequestBody(@Email @NotBlank String email, @NotBlank @Size(max = 30) String kind,
                                   @Min(10_000) @Max(5_000_000) double premium, @Min(50) @Max(90) int age,
                                   @Size(max = 2) String state) {}

    @PostMapping("/quote-requests")
    @Transactional
    public ResponseEntity<Map<String, String>> quoteRequest(@Valid @RequestBody QuoteRequestBody b, Authentication a) {
        var q = new Entities.QuoteRequest();
        q.userId = a == null ? null : users.findByEmailIgnoreCase(a.getName()).map(u -> u.id).orElse(null);
        q.email = b.email(); q.kind = b.kind(); q.premium = b.premium(); q.age = b.age(); q.state = b.state();
        quotes.save(q);
        return ResponseEntity.status(HttpStatus.CREATED).body(Map.of("status", "received"));
    }

    // ── Admin ────────────────────────────────────────────────────────────────

    @GetMapping("/admin/stats")
    public Map<String, Object> stats(@RequestHeader(value = "X-Admin-Token", required = false) String token) {
        requireAdmin(token);
        return Map.of("users", users.count(), "savedPlans", saved.count(), "waitlist", waitlist.count(),
                "advisorRequests", advisor.count(), "quoteRequests", quotes.count());
    }

    @GetMapping(value = "/admin/waitlist.csv", produces = "text/csv")
    public String waitlistCsv(@RequestHeader(value = "X-Admin-Token", required = false) String token) {
        requireAdmin(token);
        StringBuilder sb = new StringBuilder("created_at,email,first_name,age_band,assets_band,status,interests,code,referred_by\n");
        for (var w : waitlist.findAllByOrderByCreatedAtAsc())
            sb.append(String.join(",", List.of(String.valueOf(w.createdAt), csv(w.email), csv(w.firstName), csv(w.ageBand),
                    csv(w.assetsBand), csv(w.status), csv(w.interests), w.code, csv(w.referredBy)))).append('\n');
        return sb.toString();
    }

    private static String csv(String s) {
        if (s == null) return "";
        String v = s.replace("\"", "\"\"");
        if (v.startsWith("=") || v.startsWith("+") || v.startsWith("-") || v.startsWith("@")) v = "'" + v; // formula injection
        return "\"" + v + "\"";
    }

    private void requireAdmin(String token) {
        if (token == null || !MessageDigest.isEqual(token.getBytes(), adminToken.getBytes()))
            throw new ResponseStatusException(HttpStatus.NOT_FOUND);
    }
}
