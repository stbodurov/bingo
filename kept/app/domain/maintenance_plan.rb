# The off-ramp. Most people stop GLP-1 medicines within a year and regain roughly 60% of the
# weight within a year of stopping. Kept turns stopping or reducing into a planned, clinician-
# led transition with habit targets and a regain guardrail instead of an abrupt cliff.
class MaintenancePlan
  Week = Data.define(:week, :dose_mg, :focus)
  Path = Data.define(:key, :label, :summary, :weeks, :risk)

  def initialize(medication, targets:, trend:)
    @med = medication
    @drug = DrugCatalog.fetch(medication.drug)
    @targets = targets
    @trend = trend
  end

  def ready?
    weeks = ((Date.current - @med.started_on).to_i / 7)
    weeks >= 26 || (@trend.plateau?(weeks) && @trend.total_loss_pct >= 10)
  end

  def paths
    ladder = @drug.ladder
    i = ladder.index { |d| (d - @med.dose_mg.to_f).abs < 0.001 } || (ladder.size - 1)
    taper = (0..3).map do |k|
      dose = k < 3 ? ladder[[i - 1 - k, 0].max] : 0
      Week.new(week: k * 4 + 1, dose_mg: dose, focus: TAPER_FOCUS[k])
    end
    [
      Path.new(key: "maintain", label: "Stay on a maintenance dose",
               summary: "Many people keep results on a lower dose long-term. Lowest regain risk; ongoing cost.",
               weeks: [Week.new(week: 1, dose_mg: ladder[[i - 1, 0].max], focus: "Step down one dose with your prescriber; keep protein and lifting.")],
               risk: :low),
      Path.new(key: "taper", label: "Taper over 12–16 weeks",
               summary: "Step down every 4 weeks while habits take over. Moderate regain risk.", weeks: taper, risk: :moderate),
      Path.new(key: "stop", label: "Stop now",
               summary: "Highest regain risk — trials show most weight returns within a year without a plan.",
               weeks: [Week.new(week: 1, dose_mg: 0, focus: "Weigh 3×/week; Kept alerts your clinician-share report if the trend rises 2%.")],
               risk: :high)
    ]
  end

  def guardrails
    [
      "Protein: #{@targets.protein_g} g a day, every day.",
      "Strength: #{@targets.strength_sessions_per_week} sessions a week — muscle is what keeps your metabolism up.",
      "Weigh at least 3 times a week; Kept watches the trend, not the noise.",
      "Regain guardrail: if your trend rises 2% over 4 weeks, Kept flags it so you and your prescriber can act early.",
      "Food noise check-ins weekly — returning food noise is the earliest warning sign."
    ]
  end

  TAPER_FOCUS = [
    "Lock in protein-first meals and three strength sessions.",
    "Appetite returns: plan meals in advance, keep high-volume foods on hand.",
    "Last step down: weekly weigh-in review with your clinician.",
    "Off medication: daily habits carry the result; guardrail active."
  ].freeze
end
