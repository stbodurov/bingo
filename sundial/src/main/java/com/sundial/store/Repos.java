package com.sundial.store;

import com.sundial.store.Entities.AdvisorRequest;
import com.sundial.store.Entities.QuoteRequest;
import com.sundial.store.Entities.SavedPlan;
import com.sundial.store.Entities.UserAccount;
import com.sundial.store.Entities.WaitlistEntry;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;
import java.util.Optional;

public final class Repos {
    private Repos() {}

    public interface Users extends JpaRepository<UserAccount, Long> {
        Optional<UserAccount> findByEmailIgnoreCase(String email);
    }

    public interface Plans extends JpaRepository<SavedPlan, Long> {
        List<SavedPlan> findByUserIdOrderByUpdatedAtDesc(Long userId);
        Optional<SavedPlan> findByIdAndUserId(Long id, Long userId);
    }

    public interface Waitlist extends JpaRepository<WaitlistEntry, Long> {
        Optional<WaitlistEntry> findByEmailIgnoreCase(String email);
        Optional<WaitlistEntry> findByCode(String code);
        long countByReferredBy(String code);
        List<WaitlistEntry> findAllByOrderByCreatedAtAsc();
    }

    public interface AdvisorRequests extends JpaRepository<AdvisorRequest, Long> {}

    public interface QuoteRequests extends JpaRepository<QuoteRequest, Long> {}
}
