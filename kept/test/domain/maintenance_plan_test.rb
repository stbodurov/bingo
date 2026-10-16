require "test_helper"

class MaintenancePlanTest < ActiveSupport::TestCase
  Entry = Struct.new(:measured_on, :weight_kg, :body_fat_pct)

  test "taper steps down the label ladder every four weeks to zero" do
    med = Medication.new(drug: "wegovy", dose_mg: 2.4, started_on: Date.current - 300, dose_started_on: Date.current - 200)
    trend = WeightTrend.new([Entry.new(Date.current - 300, 110, nil), Entry.new(Date.current, 92, nil)])
    targets = Targets.new(User.new(height_cm: 170, sex: "female"), 92)
    plan = MaintenancePlan.new(med, targets:, trend:)
    assert plan.ready?
    taper = plan.paths.find { |p| p.key == "taper" }
    assert_equal [1.7, 1.0, 0.5, 0], taper.weeks.map(&:dose_mg)
    assert_equal [1, 5, 9, 13], taper.weeks.map(&:week)
    assert_equal %i[low moderate high], plan.paths.map(&:risk)
    assert plan.guardrails.first.include?("#{targets.protein_g} g")
  end
end
