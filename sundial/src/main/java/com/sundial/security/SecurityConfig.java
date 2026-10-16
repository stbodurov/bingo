package com.sundial.security;

import com.sundial.store.Repos;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.data.jpa.repository.config.EnableJpaRepositories;
import org.springframework.http.HttpStatus;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.authentication.ProviderManager;
import org.springframework.security.authentication.dao.DaoAuthenticationProvider;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.core.userdetails.User;
import org.springframework.security.core.userdetails.UserDetailsService;
import org.springframework.security.core.userdetails.UsernameNotFoundException;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.HttpStatusEntryPoint;
import org.springframework.security.web.context.HttpSessionSecurityContextRepository;
import org.springframework.security.web.context.SecurityContextRepository;
import org.springframework.security.web.header.writers.ReferrerPolicyHeaderWriter;
import org.springframework.security.web.authentication.www.BasicAuthenticationFilter;
import org.springframework.security.web.csrf.CookieCsrfTokenRepository;
import org.springframework.security.web.csrf.CsrfToken;
import org.springframework.security.web.csrf.CsrfTokenRequestAttributeHandler;
import org.springframework.security.web.csrf.CsrfTokenRequestHandler;
import org.springframework.security.web.csrf.XorCsrfTokenRequestAttributeHandler;
import org.springframework.util.StringUtils;
import org.springframework.web.filter.OncePerRequestFilter;

import java.util.function.Supplier;

import java.io.IOException;
import java.util.ArrayDeque;
import java.util.Deque;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Configuration
@EnableJpaRepositories(basePackages = "com.sundial.store", considerNestedRepositories = true)
public class SecurityConfig {

    static final String CSP = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "
            + "font-src 'self' https://fonts.gstatic.com; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; "
            + "base-uri 'self'; form-action 'self'";

    @Bean
    SecurityFilterChain chain(HttpSecurity http, RateLimitFilter rateLimit) throws Exception {
        http
            .authorizeHttpRequests(a -> a
                .requestMatchers("/api/me/**").authenticated()
                .anyRequest().permitAll())
            .csrf(c -> c.csrfTokenRepository(CookieCsrfTokenRepository.withHttpOnlyFalse())
                .csrfTokenRequestHandler(new SpaCsrfTokenRequestHandler()))
            .addFilterAfter(new CsrfCookieFilter(), BasicAuthenticationFilter.class)
            .securityContext(s -> s.securityContextRepository(securityContextRepository()))
            .exceptionHandling(e -> e.authenticationEntryPoint(new HttpStatusEntryPoint(HttpStatus.UNAUTHORIZED)))
            .formLogin(f -> f.disable())
            .httpBasic(b -> b.disable())
            .logout(l -> l.logoutUrl("/api/auth/logout").logoutSuccessHandler((req, res, auth) -> res.setStatus(204)))
            .headers(h -> h
                .contentSecurityPolicy(c -> c.policyDirectives(CSP))
                .referrerPolicy(r -> r.policy(ReferrerPolicyHeaderWriter.ReferrerPolicy.SAME_ORIGIN))
                .frameOptions(f -> f.deny()))
            .addFilterBefore(rateLimit, org.springframework.security.web.csrf.CsrfFilter.class);
        return http.build();
    }

    @Bean
    SecurityContextRepository securityContextRepository() {
        return new HttpSessionSecurityContextRepository();
    }

    @Bean
    PasswordEncoder passwordEncoder() {
        return new BCryptPasswordEncoder(12);
    }

    @Bean
    UserDetailsService userDetailsService(Repos.Users users) {
        return email -> users.findByEmailIgnoreCase(email)
                .map(u -> User.withUsername(u.email).password(u.passwordHash).roles("USER").build())
                .orElseThrow(() -> new UsernameNotFoundException("no user"));
    }

    @Bean
    AuthenticationManager authenticationManager(UserDetailsService uds, PasswordEncoder encoder) {
        DaoAuthenticationProvider p = new DaoAuthenticationProvider(uds);
        p.setPasswordEncoder(encoder);
        return new ProviderManager(p);
    }

    /** Sliding-window rate limits per client IP on the expensive and abusable endpoints. */
    @Configuration
    static class RateLimitConfig {
        @Bean
        RateLimitFilter rateLimitFilter(@Value("${sundial.rate-limit:on}") String mode) {
            return new RateLimitFilter(!"off".equals(mode));
        }
    }

    public static class RateLimitFilter extends OncePerRequestFilter {
        private final boolean enabled;
        private final Map<String, Deque<Long>> hits = new ConcurrentHashMap<>();
        private static final Map<String, int[]> LIMITS = Map.of( // prefix -> {max, windowSeconds}
                "/api/plan", new int[]{40, 60},
                "/api/claiming", new int[]{60, 60},
                "/api/waitlist", new int[]{10, 600},
                "/api/auth/login", new int[]{10, 300},
                "/api/auth/signup", new int[]{5, 600},
                "/api/advisor-requests", new int[]{5, 600},
                "/api/quote-requests", new int[]{10, 600});

        public RateLimitFilter(boolean enabled) { this.enabled = enabled; }

        @Override
        protected void doFilterInternal(HttpServletRequest req, HttpServletResponse res, FilterChain chain)
                throws ServletException, IOException {
            if (enabled && "POST".equals(req.getMethod())) {
                for (var e : LIMITS.entrySet()) {
                    if (!req.getRequestURI().startsWith(e.getKey())) continue;
                    String key = e.getKey() + "|" + req.getRemoteAddr();
                    long now = System.currentTimeMillis();
                    Deque<Long> q = hits.computeIfAbsent(key, k -> new ArrayDeque<>());
                    synchronized (q) {
                        while (!q.isEmpty() && now - q.peekFirst() > e.getValue()[1] * 1000L) q.pollFirst();
                        if (q.size() >= e.getValue()[0]) {
                            res.setStatus(429);
                            res.setContentType("application/json");
                            res.getWriter().write("{\"error\":\"Too many requests — try again shortly.\"}");
                            return;
                        }
                        q.addLast(now);
                    }
                }
            }
            chain.doFilter(req, res);
        }
    }

    /**
     * CSRF for a single-page app (Spring Security reference, "Single-Page Applications"): the
     * token is read from the XSRF-TOKEN cookie and echoed in the X-XSRF-TOKEN header; BREACH
     * protection applies when the token is rendered into responses.
     */
    static final class SpaCsrfTokenRequestHandler implements CsrfTokenRequestHandler {
        private final CsrfTokenRequestHandler plain = new CsrfTokenRequestAttributeHandler();
        private final CsrfTokenRequestHandler xor = new XorCsrfTokenRequestAttributeHandler();

        @Override
        public void handle(HttpServletRequest request, HttpServletResponse response, Supplier<CsrfToken> csrfToken) {
            xor.handle(request, response, csrfToken);
            csrfToken.get();
        }

        @Override
        public String resolveCsrfTokenValue(HttpServletRequest request, CsrfToken csrfToken) {
            String header = request.getHeader(csrfToken.getHeaderName());
            return (StringUtils.hasText(header) ? plain : xor).resolveCsrfTokenValue(request, csrfToken);
        }
    }

    /** Forces the deferred token to load so the cookie is set on every response. */
    static final class CsrfCookieFilter extends OncePerRequestFilter {
        @Override
        protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain chain)
                throws ServletException, IOException {
            CsrfToken token = (CsrfToken) request.getAttribute(CsrfToken.class.getName());
            if (token != null) token.getToken();
            chain.doFilter(request, response);
        }
    }
}
