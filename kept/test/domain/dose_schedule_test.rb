require "test_helper"

class DoseScheduleTest < ActiveSupport::TestCase
  Log = Struct.new(:taken_at, :site, :skipped)

  def med(**attrs) = Medication.new({ drug: "zepbound", dose_mg: 5, started_on: Date.new(2026, 8, 1), dose_started_on: Date.new(2026, 8, 1) }.merge(attrs))

  test "next dose is a week after the last one" do
    last = Time.zone.local(2026, 9, 1, 9)
    s = DoseSchedule.new(med, [Log.new(last, "arm_left", false)], now: last + 2.days)
    assert_equal last + 7.days, s.next_due_at
    refute s.due_today?
    assert_nil s.missed_guidance
  end

  test "late within the label window says take it now; past it says skip" do
    last = Time.zone.local(2026, 9, 1, 9)
    late = DoseSchedule.new(med, [Log.new(last, nil, false)], now: last + 9.days)
    assert_match "as soon as you remember", late.missed_guidance
    very_late = DoseSchedule.new(med, [Log.new(last, nil, false)], now: last + 12.days)
    assert_match "skip this dose", very_late.missed_guidance
  end

  test "rotates to the least recently used site" do
    t = Time.zone.local(2026, 9, 1, 9)
    used = DrugCatalog::INJECTION_SITES.first(5).each_with_index.map { |s, i| Log.new(t + i.weeks, s, false) }
    assert_equal DrugCatalog::INJECTION_SITES.last, DoseSchedule.new(med, used, now: t + 5.weeks).suggested_site
  end

  test "escalation follows the label ladder and holds for severe side effects" do
    now = Time.zone.local(2026, 9, 15, 12)
    early = DoseSchedule.new(med(dose_started_on: Date.new(2026, 9, 1)), [], now:)
    assert_match "2 of 4 weeks", early.escalation_note
    ready = DoseSchedule.new(med(dose_started_on: Date.new(2026, 8, 1)), [], now:)
    assert_match "next step is 7.5 mg", ready.escalation_note
    assert_match "hold the dose", ready.escalation_note(recent_severe_side_effects: true)
    top = DoseSchedule.new(med(dose_mg: 15), [], now:)
    assert_equal "You're on the highest labelled dose.", top.escalation_note
  end

  test "adherence counts taken doses, not skipped ones" do
    now = Time.zone.local(2026, 9, 29, 12)
    logs = (1..8).map { |i| Log.new(now - i.weeks + 1.hour, nil, i == 3) }
    assert_equal 88, DoseSchedule.new(med, logs, now:).adherence_pct
  end
end
