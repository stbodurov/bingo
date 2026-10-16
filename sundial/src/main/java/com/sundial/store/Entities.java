package com.sundial.store;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.Table;

import java.time.Instant;

/** JPA entities. Kept as plain mutable classes in one file — the schema lives in Flyway. */
public final class Entities {
    private Entities() {}

    @Entity
    @Table(name = "user_account")
    public static class UserAccount {
        @Id @GeneratedValue(strategy = GenerationType.IDENTITY) public Long id;
        @Column(nullable = false, unique = true) public String email;
        @Column(nullable = false) public String name;
        @Column(name = "password_hash", nullable = false) public String passwordHash;
        @Column(name = "created_at", nullable = false) public Instant createdAt = Instant.now();
    }

    @Entity
    @Table(name = "saved_plan")
    public static class SavedPlan {
        @Id @GeneratedValue(strategy = GenerationType.IDENTITY) public Long id;
        @Column(name = "user_id", nullable = false) public Long userId;
        @Column(nullable = false) public String name;
        @Column(name = "household_json", nullable = false, columnDefinition = "text") public String householdJson;
        @Column(name = "success_rate") public Double successRate;
        @Column(name = "safe_spending") public Double safeSpending;
        @Column(name = "updated_at", nullable = false) public Instant updatedAt = Instant.now();
    }

    @Entity
    @Table(name = "waitlist_entry")
    public static class WaitlistEntry {
        @Id @GeneratedValue(strategy = GenerationType.IDENTITY) public Long id;
        @Column(nullable = false, unique = true) public String email;
        @Column(name = "first_name") public String firstName;
        @Column(name = "age_band") public String ageBand;
        @Column(name = "assets_band") public String assetsBand;
        public String status;
        public String interests;
        @Column(nullable = false, unique = true) public String code;
        @Column(name = "referred_by") public String referredBy;
        @Column(name = "plan_summary", columnDefinition = "text") public String planSummary;
        public String source;
        @Column(name = "created_at", nullable = false) public Instant createdAt = Instant.now();
    }

    @Entity
    @Table(name = "advisor_request")
    public static class AdvisorRequest {
        @Id @GeneratedValue(strategy = GenerationType.IDENTITY) public Long id;
        @Column(name = "user_id") public Long userId;
        @Column(nullable = false) public String name;
        @Column(nullable = false) public String email;
        public String phone;
        @Column(nullable = false) public String topic;
        public String message;
        @Column(name = "created_at", nullable = false) public Instant createdAt = Instant.now();
    }

    @Entity
    @Table(name = "quote_request")
    public static class QuoteRequest {
        @Id @GeneratedValue(strategy = GenerationType.IDENTITY) public Long id;
        @Column(name = "user_id") public Long userId;
        @Column(nullable = false) public String email;
        @Column(nullable = false) public String kind;
        @Column(nullable = false) public double premium;
        @Column(nullable = false) public int age;
        public String state;
        @Column(name = "created_at", nullable = false) public Instant createdAt = Instant.now();
    }
}
