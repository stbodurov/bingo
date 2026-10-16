# Everything the app needs about one person, computed once per request.
class Overview
  attr_reader :user, :date

  def initialize(user, date: user.today)
    @user = user
    @date = date
  end

  # The current medicine, or the last one after stopping (maintenance mode keeps working).
  def medication = @medication ||= (user.medication || user.last_medication)
  def on_medication? = medication&.active?
  def drug = medication && DrugCatalog.fetch(medication.drug)
  def weights = @weights ||= user.weight_entries.order(:measured_on).to_a
  def trend = @trend ||= WeightTrend.new(weights)
  def weight_kg = trend.latest_weight || 80.0
  def targets = @targets ||= Targets.new(user, weight_kg)
  def dose_logs = @dose_logs ||= (medication ? medication.dose_logs.order(:taken_at).to_a : [])
  def schedule = medication && (@schedule ||= DoseSchedule.new(medication, dose_logs))
  def weeks_off = (on_medication? || medication&.stopped_on.nil?) ? nil : ((date - medication.stopped_on).to_i / 7)
  def side_effects = @side_effects ||= user.side_effect_logs.where(logged_on: date - 60..date).order(:logged_on).to_a
  def coach = @coach ||= SideEffectCoach.new(side_effects, dose_logs: dose_logs)
  def workouts = @workouts ||= user.workouts.where(performed_on: date - 60..date).order(:performed_on).to_a

  def food_today = user.food_entries.where(eaten_on: date).order(:created_at)
  def protein_today = food_today.sum(:protein_g).to_f
  def fiber_today = food_today.sum(:fiber_g).to_f
  def water_today = user.water_entries.where(drank_on: date).sum(:ml)

  def protein_by_day(days = 14)
    sums = user.food_entries.where(eaten_on: date - (days - 1)..date).group(:eaten_on).sum(:protein_g)
    (0...days).map { |k| d = date - (days - 1 - k); [d, sums[d].to_f] }
  end

  def protein_avg_7d
    days = protein_by_day(7).map(&:last)
    logged = days.reject(&:zero?)
    logged.empty? ? 0.0 : (logged.sum / logged.size).round
  end

  def strength_14d = workouts.count { |w| w.kind == "strength" && w.performed_on > date - 14 }
  def strength_this_week = workouts.count { |w| w.kind == "strength" && w.performed_on >= date.beginning_of_week }

  def muscle
    @muscle ||= MuscleGuard.new(user:, drug:, trend:, protein_avg_g: protein_avg_7d, protein_target_g: targets.protein_g,
                                strength_sessions_14d: strength_14d, body_comp: weights)
  end

  def check_in_today = user.check_ins.find_by(checked_on: date)
  def recent_check_ins = user.check_ins.where(checked_on: date - 56..date).order(:checked_on)

  def cost
    return nil unless medication
    @cost ||= CostOptions.new(drug_key: medication.drug, dose_mg: medication.dose_mg, coverage: user.coverage)
  end

  def maintenance
    return nil unless medication
    @maintenance ||= MaintenancePlan.new(medication, targets:, trend:, from_dose_mg: user.maintenance_from_dose_mg)
  end

  def off_ramp
    return nil unless maintenance && user.maintenance_path && user.maintenance_started_on
    @off_ramp ||= maintenance.progress(user.maintenance_path, started_on: user.maintenance_started_on,
                                       baseline_kg: user.maintenance_baseline_kg, today: date)
  end

  # Ordered nudges for the Today screen: safety first, then the day's actions.
  def alerts
    a = []
    urgent = side_effects.select { |l| l.logged_on >= date - 2 && l.severity >= 3 }
      .map { |l| coach.advice_for(l.symptom, l.severity) }.find(&:urgent)
    a << [:critical, urgent.label, urgent.red_flag] if urgent
    if on_medication? && schedule&.missed_guidance
      a << [:warning, "Missed dose", schedule.missed_guidance]
    end
    a << [:warning, "Losing fast", "You're losing more than 1.5% of your weight a week. Faster loss takes more muscle — mention it to your prescriber and prioritise protein."] if trend.too_fast?
    if off_ramp&.tripped
      a << [:critical, "Regain guardrail", "Your trend is #{off_ramp.drift_pct}% above where you started your #{off_ramp.path.label.downcase} plan. That's the point to act — share your Kept report with your prescriber this week."]
    end
    a << [:warning, "Trend rising", "Your weight trend is up more than 2% over four weeks. Look at food noise, protein and your dose schedule — and tell your clinician."] if trend.regaining? && !off_ramp&.tripped
    if on_medication? && schedule && trend.plateau?(schedule.weeks_on_medication)
      a << [:info, "Plateau", "Your trend has been flat for a month. That's normal around months 6–12; it's a good time to review dose, protein and the maintenance plan."]
    end
    a
  end
end
