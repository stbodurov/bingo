require "test_helper"

class MuscleGuardTest < ActiveSupport::TestCase
  Entry = Struct.new(:measured_on, :weight_kg, :body_fat_pct)

  def trend(loss_per_day = 0.08) = WeightTrend.new((0...120).map { |i| Entry.new(Date.current - 119 + i, 100 - i * loss_per_day, nil) })
  def guard(protein:, strength:, body_comp: [], user: User.new(birth_year: 1985), t: trend)
    MuscleGuard.new(user:, drug: DrugCatalog.fetch("wegovy"), trend: t, protein_avg_g: protein, protein_target_g: 100,
                    strength_sessions_14d: strength, body_comp:)
  end

  test "protein and lifting cut the estimated lean share" do
    none = guard(protein: 40, strength: 0)
    full = guard(protein: 110, strength: 6)
    assert_operator full.lean_share, :<, none.lean_share
    assert_operator full.score, :>, none.score
    assert_equal :good, full.status
    assert_equal :critical, none.status
  end

  test "lean share never goes below the floor and rises for older adults" do
    assert_operator guard(protein: 200, strength: 12).lean_share, :>=, MuscleGuard::FLOOR
    older = guard(protein: 50, strength: 1, user: User.new(birth_year: Date.current.year - 66))
    younger = guard(protein: 50, strength: 1)
    assert_in_delta 0.03, older.lean_share - younger.lean_share, 0.001
  end

  test "body-composition readings override the estimate" do
    readings = [Entry.new(Date.current - 100, 100, 40.0), Entry.new(Date.current, 90, 37.8)]
    g = guard(protein: 50, strength: 0, body_comp: readings)
    assert g.from_readings?
    # lean 60 -> 55.98 over 10 kg lost = 40%
    assert_in_delta 0.402, g.shown_share, 0.001
    assert_in_delta g.shown_share * trend.total_loss_kg, g.lean_lost_kg, 0.1
  end

  test "advice speaks to what's missing" do
    tips = guard(protein: 60, strength: 2).advice
    assert tips.any? { |t| t.include?("40 g of protein") }
    assert tips.any? { |t| t.include?("2 more strength sessions") }
    assert_equal ["Muscle is well protected. Keep the protein and the lifting going."], guard(protein: 100, strength: 6).advice
  end
end
