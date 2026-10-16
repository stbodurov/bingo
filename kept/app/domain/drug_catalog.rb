# The GLP-1 / GIP-GLP-1 medicines Kept supports, with the label facts the coach uses:
# schedule, dose-escalation ladder, missed-dose window and how to take it. Prices are
# self-pay list prices as of October 2026 for planning; insurers and pharmacies vary.
# Dose decisions always belong to the prescriber — Kept only tells people what the label
# says and when to ask.
module DrugCatalog
  Drug = Data.define(:key, :brand, :generic, :maker, :form, :frequency, :ladder, :step_weeks,
                     :missed_window_hours, :how_to_take, :base_lean_share, :cash_prices, :indication)

  DRUGS = {
    "zepbound" => Drug.new(
      key: "zepbound", brand: "Zepbound", generic: "tirzepatide", maker: "Lilly", form: :injection, frequency: :weekly,
      ladder: [2.5, 5, 7.5, 10, 12.5, 15], step_weeks: 4, missed_window_hours: 96,
      how_to_take: "Once a week, any time of day, with or without food. Rotate injection sites.",
      base_lean_share: 0.26, cash_prices: { 2.5 => 299, 5 => 399, 7.5 => 449, 10 => 449, 12.5 => 449, 15 => 449 },
      indication: "Chronic weight management; obstructive sleep apnea"
    ),
    "mounjaro" => Drug.new(
      key: "mounjaro", brand: "Mounjaro", generic: "tirzepatide", maker: "Lilly", form: :injection, frequency: :weekly,
      ladder: [2.5, 5, 7.5, 10, 12.5, 15], step_weeks: 4, missed_window_hours: 96,
      how_to_take: "Once a week, any time of day, with or without food. Rotate injection sites.",
      base_lean_share: 0.26, cash_prices: {}, indication: "Type 2 diabetes"
    ),
    "wegovy" => Drug.new(
      key: "wegovy", brand: "Wegovy", generic: "semaglutide", maker: "Novo Nordisk", form: :injection, frequency: :weekly,
      ladder: [0.25, 0.5, 1.0, 1.7, 2.4], step_weeks: 4, missed_window_hours: 120,
      how_to_take: "Once a week on the same day, any time, with or without food. Rotate injection sites.",
      base_lean_share: 0.39, cash_prices: { 0.25 => 349, 0.5 => 349, 1.0 => 349, 1.7 => 349, 2.4 => 349 },
      indication: "Chronic weight management; cardiovascular risk reduction; MASH"
    ),
    "ozempic" => Drug.new(
      key: "ozempic", brand: "Ozempic", generic: "semaglutide", maker: "Novo Nordisk", form: :injection, frequency: :weekly,
      ladder: [0.25, 0.5, 1.0, 2.0], step_weeks: 4, missed_window_hours: 120,
      how_to_take: "Once a week on the same day, any time, with or without food. Rotate injection sites.",
      base_lean_share: 0.39, cash_prices: { 0.25 => 499, 0.5 => 499, 1.0 => 499, 2.0 => 499 }, indication: "Type 2 diabetes"
    ),
    "wegovy_pill" => Drug.new(
      key: "wegovy_pill", brand: "Wegovy pill", generic: "oral semaglutide", maker: "Novo Nordisk", form: :pill, frequency: :daily,
      ladder: [1.5, 4, 9, 25], step_weeks: 4, missed_window_hours: 0,
      how_to_take: "Every morning on an empty stomach with up to 4 oz of water; wait 30 minutes before food, drink or other pills.",
      base_lean_share: 0.37, cash_prices: { 1.5 => 149, 4 => 149, 9 => 299, 25 => 299 }, indication: "Chronic weight management"
    ),
    "foundayo" => Drug.new(
      key: "foundayo", brand: "Foundayo", generic: "orforglipron", maker: "Lilly", form: :pill, frequency: :daily,
      ladder: [1, 3, 6, 12, 24, 36], step_weeks: 4, missed_window_hours: 0,
      how_to_take: "Once a day at any time, with or without food or water restrictions.",
      base_lean_share: 0.33, cash_prices: { 1 => 149, 3 => 149, 6 => 249, 12 => 299, 24 => 349, 36 => 349 },
      indication: "Chronic weight management"
    ),
    "saxenda" => Drug.new(
      key: "saxenda", brand: "Saxenda", generic: "liraglutide", maker: "Novo Nordisk", form: :injection, frequency: :daily,
      ladder: [0.6, 1.2, 1.8, 2.4, 3.0], step_weeks: 1, missed_window_hours: 0,
      how_to_take: "Once a day at any time. If you miss a day, take the next dose as usual; after 3+ missed days ask about restarting at 0.6 mg.",
      base_lean_share: 0.30, cash_prices: {}, indication: "Chronic weight management"
    )
  }.freeze

  INJECTION_SITES = %w[abdomen_left abdomen_right thigh_left thigh_right arm_left arm_right].freeze

  def self.fetch(key) = DRUGS.fetch(key.to_s)
  def self.options = DRUGS.values.map { |d| ["#{d.brand} (#{d.generic})", d.key] }
  def self.site_label(site) = site.to_s.tr("_", " ").sub(/\A./, &:upcase)
end
