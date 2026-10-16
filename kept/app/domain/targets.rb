# Daily protein, fiber and water targets.
#
# Protein follows the 2025 joint advisory for people losing weight on GLP-1 medicines:
# 1.2–1.6 g per kg per day (Kept defaults to 1.4), using adjusted body weight when BMI ≥ 30
# so targets stay achievable when appetite is low. At least 60 g/day, split across meals
# (≈ 25–40 g per meal is what muscle protein synthesis responds to).
class Targets
  attr_reader :user, :weight_kg

  def initialize(user, weight_kg)
    @user = user
    @weight_kg = weight_kg.to_f
  end

  def ideal_weight_kg
    h = user.height_cm.to_f
    return weight_kg if h.zero?
    base = user.sex == "male" ? 50.0 : 45.5
    [base + 0.9 * (h - 152.4), 40].max
  end

  def bmi
    h = user.height_cm.to_f / 100
    h.zero? ? nil : (weight_kg / (h * h)).round(1)
  end

  def reference_weight_kg
    return weight_kg if bmi.nil? || bmi < 30
    ideal_weight_kg + 0.4 * (weight_kg - ideal_weight_kg)
  end

  def protein_g = [(user.protein_factor.to_f * reference_weight_kg).round, 60].max
  def protein_per_meal_g = (protein_g.to_f / [user.meals_per_day, 1].max).round
  def fiber_g = user.sex == "male" ? 34 : 28
  def water_ml = [[(35 * weight_kg).round(-2), 1_800].max, 3_700].min
  def strength_sessions_per_week = 3
end
