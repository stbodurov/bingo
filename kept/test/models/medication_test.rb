require "test_helper"

class MedicationTest < ActiveSupport::TestCase
  test "dose must be on the label's ladder" do
    user = users(:one)
    ok = user.medications.new(drug: "wegovy", dose_mg: 1.7, started_on: Date.current, dose_started_on: Date.current)
    assert ok.valid?
    bad = user.medications.new(drug: "wegovy", dose_mg: 3, started_on: Date.current, dose_started_on: Date.current)
    refute bad.valid?
    assert_equal "Wegovy 1.7 mg", ok.label
  end
end
