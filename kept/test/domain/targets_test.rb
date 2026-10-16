require "test_helper"

class TargetsTest < ActiveSupport::TestCase
  test "protein uses adjusted body weight when BMI is 30 or more" do
    user = User.new(sex: "female", height_cm: 165, protein_factor: 1.4, meals_per_day: 3)
    t = Targets.new(user, 110)
    assert_operator t.bmi, :>=, 30
    assert_operator t.reference_weight_kg, :<, 110
    assert_equal (1.4 * t.reference_weight_kg).round, t.protein_g
  end

  test "protein uses actual weight below BMI 30 and never drops under 60 g" do
    user = User.new(sex: "female", height_cm: 170, protein_factor: 1.2, meals_per_day: 3)
    assert_equal (1.2 * 65).round, Targets.new(user, 65).protein_g
    assert_equal 60, Targets.new(user, 40).protein_g
  end

  test "water is 35 ml per kg, clamped" do
    user = User.new(sex: "male", height_cm: 180)
    assert_equal 1_800, Targets.new(user, 40).water_ml
    assert_equal 3_500, Targets.new(user, 100).water_ml
    assert_equal 3_700, Targets.new(user, 160).water_ml
    assert_equal 34, Targets.new(user, 90).fiber_g
  end
end
