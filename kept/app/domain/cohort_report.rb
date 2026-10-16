# What an employer or health plan sees: whether the GLP-1 spend is working. De-identified,
# aggregate only, and suppressed below 10 members so no individual can be singled out.
class CohortReport
  MIN_COHORT = 10
  # Real-world benchmark without support: ~35% of people without diabetes still on therapy at
  # 12 months (2025 cohort of 125,474 adults: 64.8% stopped within a year).
  BENCHMARK_PERSISTENCE_12M = 0.35

  def initialize(employer, today: Date.current)
    @employer = employer
    @today = today
    @members = employer.users.includes(:medications, :weight_entries).to_a
  end

  def suppressed? = members_on_medication.size < MIN_COHORT
  def members = @members
  def members_on_medication = @members.select { |u| u.medications.any? }

  def active?(u) = DoseLog.where(user_id: u.id).where("taken_at >= ?", @today - 14).exists?

  def active_count = @active_count ||= members_on_medication.count { |u| active?(u) }

  # Persistence at N months among members who started at least N months ago.
  def persistence(months)
    eligible = members_on_medication.select { |u| (s = u.medications.map(&:started_on).min) && s <= @today << months }
    return nil if eligible.size < MIN_COHORT
    still = eligible.count { |u| DoseLog.where(user_id: u.id).where("taken_at >= ?", u.medications.map(&:started_on).min >> months).exists? }
    (still.to_f / eligible.size).round(3)
  end

  def avg_loss_pct
    vals = members_on_medication.filter_map { |u| t = WeightTrend.new(u.weight_entries.to_a); t.total_loss_pct unless t.empty? }
    vals.empty? ? 0.0 : (vals.sum / vals.size).round(1)
  end

  def protein_adherence
    vals = members_on_medication.filter_map do |u|
      w = u.weight_entries.max_by(&:measured_on)
      next unless w
      target = Targets.new(u, w.weight_kg).protein_g
      days = FoodEntry.where(user_id: u.id, eaten_on: @today - 13..@today).group(:eaten_on).sum(:protein_g).values
      next if days.empty?
      [days.sum / days.size / target, 1.0].min
    end
    vals.empty? ? 0.0 : (vals.sum / vals.size).round(2)
  end

  def strength_weekly
    total = Workout.where(user_id: members_on_medication.map(&:id), kind: "strength", performed_on: @today - 27..@today).count
    members_on_medication.empty? ? 0.0 : (total / 4.0 / members_on_medication.size).round(1)
  end

  def side_effect_risk_count
    SideEffectLog.where(user_id: members_on_medication.map(&:id), logged_on: @today - 13..@today).where(severity: 3).distinct.count(:user_id)
  end

  def monthly_drug_spend = (active_count * @employer.monthly_drug_cost.to_f).round
  def monthly_kept_cost = (@employer.covered_lives * @employer.pmpm.to_f).round

  # Drug spend protected: each extra person still on therapy at 12 months versus the
  # benchmark is spend that keeps producing results instead of being followed by regain.
  def persistence_lift_12m
    p = persistence(12)
    p && (p - BENCHMARK_PERSISTENCE_12M).round(3)
  end

  def eligible_12m = members_on_medication.count { |u| (s = u.medications.map(&:started_on).min) && s <= @today << 12 }

  # Extra members still on therapy at 12 months beyond the benchmark, and the drug spend
  # already invested in them (12 months each) that keeps producing results instead of regain.
  def retained_beyond_benchmark
    lift = persistence_lift_12m
    lift && lift.positive? ? (lift * eligible_12m).round : 0
  end

  def spend_kept_productive = retained_beyond_benchmark * 12 * @employer.monthly_drug_cost.to_f

  def by_drug
    members_on_medication.group_by { |u| u.medications.max_by(&:started_on).drug }
      .transform_values(&:size).sort_by { |_, n| -n }.map { |k, n| [DrugCatalog.fetch(k).brand, n] }
  end

  def starts_by_month(months = 12)
    (0...months).map do |k|
      m = (@today << (months - 1 - k)).beginning_of_month
      n = members_on_medication.count { |u| (s = u.medications.map(&:started_on).min) && s.beginning_of_month == m }
      [m, n]
    end
  end
end
