package com.sundial.api;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.sundial.engine.Samples;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.mock.web.MockHttpSession;
import org.springframework.test.web.servlet.MockMvc;

import static org.hamcrest.Matchers.*;
import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.csrf;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@SpringBootTest(properties = {
        "spring.datasource.url=jdbc:h2:mem:sundialtest;MODE=PostgreSQL;DB_CLOSE_DELAY=-1",
        "sundial.monte-carlo-paths=200", "sundial.rate-limit=off"})
@AutoConfigureMockMvc
class ApiTest {
    @Autowired MockMvc mvc;
    @Autowired ObjectMapper json;

    @Test
    void planEndpointReturnsAFullPlan() throws Exception {
        mvc.perform(post("/api/plan").with(csrf()).contentType(MediaType.APPLICATION_JSON)
                        .content(json.writeValueAsString(Samples.couple())))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.summary.successRate", allOf(greaterThanOrEqualTo(0.0), lessThanOrEqualTo(1.0))))
                .andExpect(jsonPath("$.claiming.recommended", hasSize(2)))
                .andExpect(jsonPath("$.paycheck", hasSize(12)))
                .andExpect(jsonPath("$.strategies[?(@.recommended == true)]", hasSize(1)));
    }

    @Test
    void planRejectsBadInputWithReadableErrors() throws Exception {
        String bad = json.writeValueAsString(Samples.single()).replace("\"birthYear\":1960", "\"birthYear\":1890");
        mvc.perform(post("/api/plan").with(csrf()).contentType(MediaType.APPLICATION_JSON).content(bad))
                .andExpect(status().isUnprocessableEntity())
                .andExpect(jsonPath("$.details[0]", containsString("birth year")));
    }

    @Test
    void csrfIsRequiredForWrites() throws Exception {
        mvc.perform(post("/api/plan").contentType(MediaType.APPLICATION_JSON).content(json.writeValueAsString(Samples.single())))
                .andExpect(status().isForbidden());
    }

    @Test
    void signupSaveAndListPlans() throws Exception {
        MockHttpSession session = new MockHttpSession();
        mvc.perform(get("/api/me/plans").session(session)).andExpect(status().isUnauthorized());
        mvc.perform(post("/api/auth/signup").session(session).with(csrf()).contentType(MediaType.APPLICATION_JSON)
                        .content("{\"name\":\"Pat\",\"email\":\"pat@example.com\",\"password\":\"correct horse battery\"}"))
                .andExpect(status().isOk()).andExpect(jsonPath("$.email").value("pat@example.com"));
        String body = "{\"name\":\"Our plan\",\"household\":" + json.writeValueAsString(Samples.single()) + "}";
        mvc.perform(post("/api/me/plans").session(session).with(csrf()).contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isOk()).andExpect(jsonPath("$.successRate").exists());
        mvc.perform(get("/api/me/plans").session(session)).andExpect(status().isOk()).andExpect(jsonPath("$", hasSize(1)));
        mvc.perform(post("/api/auth/login").with(csrf()).contentType(MediaType.APPLICATION_JSON)
                        .content("{\"email\":\"pat@example.com\",\"password\":\"wrong password\"}"))
                .andExpect(status().isUnauthorized());
        mvc.perform(post("/api/auth/signup").with(csrf()).contentType(MediaType.APPLICATION_JSON)
                        .content("{\"name\":\"Pat\",\"email\":\"pat@example.com\",\"password\":\"correct horse battery\"}"))
                .andExpect(status().isConflict());
    }

    @Test
    void loginWithoutAnExistingSessionWorks() throws Exception {
        mvc.perform(post("/api/auth/signup").with(csrf()).contentType(MediaType.APPLICATION_JSON)
                        .content("{\"name\":\"Lee\",\"email\":\"lee@example.com\",\"password\":\"correct horse battery\"}"))
                .andExpect(status().isOk())
                .andExpect(request().sessionAttribute("SPRING_SECURITY_CONTEXT", notNullValue()));
        mvc.perform(post("/api/auth/login").with(csrf()).contentType(MediaType.APPLICATION_JSON)
                        .content("{\"email\":\"lee@example.com\",\"password\":\"correct horse battery\"}"))
                .andExpect(status().isOk()).andExpect(jsonPath("$.name").value("Lee"));
    }

    @Test
    void waitlistReferralsMoveYouUp() throws Exception {
        String first = mvc.perform(post("/api/waitlist").with(csrf()).contentType(MediaType.APPLICATION_JSON)
                        .content("{\"email\":\"a@example.com\",\"firstName\":\"Ann\",\"ageBand\":\"60-64\"}"))
                .andExpect(status().isOk()).andReturn().getResponse().getContentAsString();
        String code = json.readTree(first).get("code").asText();
        for (String e : new String[]{"b@example.com", "c@example.com"})
            mvc.perform(post("/api/waitlist").with(csrf()).contentType(MediaType.APPLICATION_JSON)
                    .content("{\"email\":\"" + e + "\"}")).andExpect(status().isOk());
        mvc.perform(post("/api/waitlist").with(csrf()).contentType(MediaType.APPLICATION_JSON)
                        .content("{\"email\":\"d@example.com\",\"referredBy\":\"" + code + "\"}"))
                .andExpect(status().isOk());
        mvc.perform(get("/api/waitlist/" + code)).andExpect(status().isOk())
                .andExpect(jsonPath("$.referrals").value(1)).andExpect(jsonPath("$.position").value(1));
        mvc.perform(get("/api/admin/stats")).andExpect(status().isNotFound());
        mvc.perform(get("/api/admin/stats").header("X-Admin-Token", "dev-admin")).andExpect(status().isOk());
    }

    @Test
    void annuityQuotesAndSecurityHeaders() throws Exception {
        mvc.perform(get("/api/annuities/quotes?premium=100000&age=65&sex=MALE"))
                .andExpect(status().isOk())
                .andExpect(header().string("Content-Security-Policy", containsString("script-src 'self'")))
                .andExpect(jsonPath("$[0].monthly", allOf(greaterThan(550.0), lessThan(720.0))));
        mvc.perform(get("/plan")).andExpect(status().isOk()).andExpect(forwardedUrl("/index.html"));
    }
}
