# Estimates how much of the weight lost is lean mass and scores how well it's being protected.
#
# Without intervention, trials put lean mass at roughly 26% of weight lost on tirzepatide and
# up to ~39% on semaglutide; adequate protein plus resistance training can bring it toward
# ~10%. Kept scales the drug's baseline by the person's protein adherence and strength
# training over the last two weeks. When body-composition readings exist, those win.
class MuscleGuard
  FLOOR = 0.10

  def initialize(user:, drug:, trend:, protein_avg_g:, protein_target_g:, strength_sessions_14d:, body_comp: [])
    @user, @drug, @trend = user, drug, trend
    @protein_avg_g, @protein_target_g = protein_avg_g.to_f, protein_target_g.to_f
    @strength = strength_sessions_14d.to_i
    @body_comp = body_comp.select { |e| e.body_fat_pct.present? }.sort_by(&:measured_on)
  end

  def protein_adherence = @protein_target_g.zero? ? 0.0 : [@protein_avg_g / @protein_target_g, 1.0].min
  def strength_adherence = [@strength / 6.0, 1.0].min

  def lean_share
    base = @drug&.base_lean_share || 0.33
    share = base * (1 - 0.45 * protein_adherence - 0.25 * strength_adherence)
    share = [share, FLOOR].max
    share += 0.05 if @trend.too_fast?
    age = @user.birth_year ? Date.current.year - @user.birth_year : 45
    share += 0.03 if age >= 60
    share.round(3)
  end

  def measured?
    @body_comp.size >= 2
  end

  # Lean share of the loss between the first and latest body-composition readings.
  def measured_share
    a, b = @body_comp.first, @body_comp.last
    lost = a.weight_kg.to_f - b.weight_kg.to_f
    return nil if lost < 1.0
    lean_a = a.weight_kg.to_f * (1 - a.body_fat_pct.to_f / 100)
    lean_b = b.weight_kg.to_f * (1 - b.body_fat_pct.to_f / 100)
    ((lean_a - lean_b) / lost).clamp(0.0, 1.0).round(3)
  end

  # What the person sees: measured when there are usable readings, otherwise estimated.
  def shown_share = (measured? && measured_share) || lean_share
  def from_readings? = measured? && !measured_share.nil?

  def lean_lost_kg = (@trend.total_loss_kg.to_f * shown_share).round(1)

  def score
    s = 100 * (0.55 * protein_adherence + 0.35 * strength_adherence) + 10
    s -= 15 if @trend.too_fast?
    s.clamp(0, 100).round
  end

  def status
    case score
    when 75.. then :good
    when 50...75 then :warning
    else :critical
    end
  end

  def advice
    tips = []
    tips << "Add about #{(@protein_target_g - @protein_avg_g).round} g of protein a day — lead each meal with it." if protein_adherence < 0.9
    if @strength < 6
      more = 3 - [@strength / 2, 3].min
      tips << "Add #{more} more strength #{more == 1 ? 'session' : 'sessions'} a week — full-body, 2–3 sets close to failure."
    end
    tips << "You're losing faster than 1.5% a week. Faster loss takes more muscle — tell your prescriber." if @trend.too_fast?
    tips << "Muscle is well protected. Keep the protein and the lifting going." if tips.empty?
    tips
  end
end
