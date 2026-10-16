package com.sundial.config;

import com.sundial.security.SecurityConfig;
import org.springframework.boot.web.servlet.FilterRegistrationBean;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.GetMapping;

@Configuration
public class WebConfig {

    /** The rate limiter runs inside the security chain only, not as a second servlet filter. */
    @Bean
    FilterRegistrationBean<SecurityConfig.RateLimitFilter> rateLimitRegistration(SecurityConfig.RateLimitFilter f) {
        FilterRegistrationBean<SecurityConfig.RateLimitFilter> r = new FilterRegistrationBean<>(f);
        r.setEnabled(false);
        return r;
    }

    /** Client-side routes of the React app resolve to index.html. */
    @Controller
    static class SpaRoutes {
        @GetMapping({"/plan", "/plan/**", "/login", "/signup", "/account", "/account/**", "/waitlist/**", "/methodology", "/privacy"})
        String spa() {
            return "forward:/index.html";
        }
    }
}
