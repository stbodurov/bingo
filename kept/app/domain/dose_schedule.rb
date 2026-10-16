# When the next dose is due, what to do about a missed one, where to inject next, and
# whether the label's escalation schedule says it's time to talk about the next step.
class DoseSchedule
  attr_reader :medication, :drug, :now

  def initialize(medication, logs, now: Time.current)
    @medication = medication
    @drug = DrugCatalog.fetch(medication.drug)
    @logs = logs.reject(&:skipped).sort_by(&:taken_at)
    @now = now
  end

  def last_dose = @logs.last

  def interval = drug.frequency == :weekly ? 7.days : 1.day

  def next_due_at
    if last_dose
      last_dose.taken_at + interval
    elsif drug.frequency == :weekly && medication.dose_weekday
      d = now.to_date
      d += 1 until d.wday == medication.dose_weekday
      d.in_time_zone.change(hour: 9)
    else
      now
    end
  end

  def due_today? = next_due_at.to_date <= now.to_date
  def overdue_hours = [((now - next_due_at) / 3600).floor, 0].max

  # Label-based missed-dose guidance.
  def missed_guidance
    return nil unless overdue_hours > 12
    if drug.missed_window_hours.zero?
      "Missed yesterday's #{drug.brand}? Skip it and take today's dose as usual — don't double up."
    elsif overdue_hours <= drug.missed_window_hours
      "Your #{drug.brand} dose is #{(overdue_hours / 24.0).round(1)} days late. The label says take it as soon as you remember " \
        "(within #{drug.missed_window_hours / 24} days), then continue weekly."
    else
      "It's been more than #{drug.missed_window_hours / 24} days. The label says skip this dose and take the next one on your " \
        "regular day. If you've missed two or more doses, ask your prescriber whether to restart at a lower dose."
    end
  end

  def weeks_at_dose = ((now.to_date - medication.dose_started_on).to_i / 7)
  def weeks_on_medication = ((now.to_date - medication.started_on).to_i / 7)
  def ladder_index = drug.ladder.index { |d| (d - medication.dose_mg.to_f).abs < 0.001 }
  def next_step = (i = ladder_index) && drug.ladder[i + 1]
  def max_dose? = ladder_index == drug.ladder.size - 1

  def escalation_note(recent_severe_side_effects: false)
    return "You're on the highest labelled dose." if max_dose?
    return nil unless next_step
    remaining = drug.step_weeks - weeks_at_dose
    if remaining.positive?
      "#{medication.dose_mg.to_f.to_s.sub(/\.0\z/, '')} mg for #{weeks_at_dose} of #{drug.step_weeks} weeks before the label's next step (#{fmt(next_step)} mg)."
    elsif recent_severe_side_effects
      "You've reached the usual #{drug.step_weeks} weeks, but side effects were strong this week — many prescribers hold the dose until they settle."
    else
      "You've been on this dose #{weeks_at_dose} weeks. The label's next step is #{fmt(next_step)} mg — worth discussing with your prescriber."
    end
  end

  def suggested_site
    return nil unless drug.form == :injection
    recent = @logs.last(DrugCatalog::INJECTION_SITES.size).map(&:site).compact
    DrugCatalog::INJECTION_SITES.min_by { |s| recent.rindex(s) || -1 }
  end

  def adherence_pct(days = 56)
    expected = drug.frequency == :weekly ? days / 7 : days
    taken = @logs.count { |l| l.taken_at >= now - days.days }
    expected.zero? ? 100 : [(100.0 * taken / expected).round, 100].min
  end

  private

  def fmt(v) = v.to_f.to_s.sub(/\.0\z/, "")
end
