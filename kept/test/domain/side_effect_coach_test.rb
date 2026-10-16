require "test_helper"

class SideEffectCoachTest < ActiveSupport::TestCase
  Log = Struct.new(:logged_on, :symptom, :severity)
  Dose = Struct.new(:taken_at, :skipped)

  test "severe symptoms with a red flag are urgent" do
    coach = SideEffectCoach.new([])
    SideEffectCoach::SYMPTOMS.each do |key, s|
      next unless s[:red_flag]
      assert coach.advice_for(key, 3).urgent, "#{key} at severity 3 should be urgent"
    end
    refute coach.advice_for("nausea", 1).urgent
  end

  test "finds the day after the dose when symptoms peak" do
    doses = (0..5).map { |w| Dose.new((Date.current - 42 + w * 7).in_time_zone.change(hour: 9), false) }
    logs = doses.map { |d| Log.new(d.taken_at.to_date + 2, "nausea", 2) }
    coach = SideEffectCoach.new(logs, dose_logs: doses)
    assert_equal 12, coach.by_day_after_dose[2]
    assert_match "2 days after your dose", coach.pattern_note
  end
end
