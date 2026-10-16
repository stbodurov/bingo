# Practical self-care for common GLP-1 side effects, plus red flags that mean "call your
# clinician" or "seek care now". Content follows prescribing-information warnings
# (pancreatitis, gallbladder disease, dehydration/kidney injury, hypoglycaemia with
# insulin or sulfonylureas, allergic reactions) and consensus self-care advice.
class SideEffectCoach
  SYMPTOMS = {
    "nausea" => { label: "Nausea", tips: [
      "Eat smaller meals and stop at the first sign of fullness.",
      "Avoid greasy, fried and very sweet foods for the 2–3 days after your dose.",
      "Sip cold water, ginger or peppermint tea; try bland, protein-forward snacks (Greek yogurt, cottage cheese).",
      "Don't lie down for 30–60 minutes after eating."] },
    "vomiting" => { label: "Vomiting", tips: [
      "Sip fluids with electrolytes in small amounts every 15 minutes.",
      "Pause solid food for a few hours, then restart with bland foods."],
      red_flag: "If you can't keep fluids down for 24 hours, feel dizzy or are passing little urine, contact your clinician today — dehydration can harm your kidneys." },
    "constipation" => { label: "Constipation", tips: [
      "Work up to your fiber target with vegetables, berries, chia and beans.",
      "Drink to your water target — fiber without water makes it worse.",
      "A daily walk helps; ask your pharmacist about an osmotic laxative if needed."] },
    "diarrhea" => { label: "Diarrhea", tips: [
      "Replace fluids and electrolytes.",
      "Limit sugar alcohols, very fatty meals and large portions."] },
    "reflux" => { label: "Heartburn / reflux", tips: [
      "Finish eating 3 hours before bed and raise the head of your bed.",
      "Smaller, lower-fat meals; limit alcohol, coffee and carbonation."] },
    "fatigue" => { label: "Fatigue", tips: [
      "Check you're hitting protein and water targets — low intake is the usual cause.",
      "Keep strength sessions short but regular; energy usually returns after titration."] },
    "headache" => { label: "Headache", tips: ["Hydration and regular meals help; check whether it follows days with very little food."] },
    "injection_site" => { label: "Injection-site reaction", tips: [
      "Rotate sites every dose and let the pen reach room temperature first.",
      "Mild redness or itch usually fades in a day or two."] },
    "abdominal_pain" => { label: "Abdominal pain", tips: ["Mild cramping can occur with constipation or after large meals."],
      red_flag: "Severe pain in the upper abdomen that may spread to your back, with or without vomiting, can be pancreatitis. " \
                "Pain under the right ribs, fever or yellow eyes can be gallbladder disease. Stop the medicine and seek care now." },
    "dizziness" => { label: "Dizziness", tips: ["Stand up slowly, drink to target and don't skip meals."],
      red_flag: "If you take insulin or a sulfonylurea, check your blood sugar — GLP-1s raise the risk of lows." },
    "hair_loss" => { label: "Hair shedding", tips: [
      "Shedding 2–4 months into rapid weight loss is common and temporary.",
      "Protein, iron and overall calorie adequacy matter most — another reason to hit protein."] },
    "food_aversion" => { label: "Food aversion", tips: [
      "Lean on easy proteins: shakes, yogurt, eggs, cottage cheese, tofu, soups.",
      "Eat by the clock rather than by appetite so protein still happens."] },
    "allergic" => { label: "Swelling, rash or trouble breathing", tips: [],
      red_flag: "Swelling of the face, lips or throat, or trouble breathing, can be a serious allergic reaction. Call 911." }
  }.freeze

  Advice = Data.define(:symptom, :label, :severity, :tips, :red_flag, :urgent)

  def self.options = SYMPTOMS.map { |k, v| [v[:label], k] }
  def self.label(k) = SYMPTOMS.dig(k, :label) || k.to_s.humanize

  def initialize(logs, dose_logs: [])
    @logs = logs
    @dose_logs = dose_logs.reject(&:skipped)
  end

  def advice_for(symptom, severity)
    s = SYMPTOMS.fetch(symptom)
    urgent = s[:red_flag] && (severity >= 3 || %w[allergic].include?(symptom))
    Advice.new(symptom:, label: s[:label], severity:, tips: s[:tips], red_flag: s[:red_flag], urgent: !!urgent)
  end

  def recent(days = 7) = @logs.select { |l| l.logged_on >= Date.current - days }
  def severe_recently? = recent.any? { |l| l.severity >= 3 }

  # Which day after a dose symptoms cluster on (0 = dose day). Weekly medicines only.
  def by_day_after_dose
    return {} if @dose_logs.empty?
    doses = @dose_logs.map { |d| d.taken_at.to_date }.sort
    counts = Hash.new(0)
    @logs.each do |l|
      prior = doses.reverse.find { |d| d <= l.logged_on }
      next unless prior
      gap = (l.logged_on - prior).to_i
      counts[gap] += l.severity if gap <= 6
    end
    counts
  end

  def pattern_note
    c = by_day_after_dose
    return nil if c.values.sum < 4
    peak = c.max_by { |_, v| v }.first
    "Your symptoms peak #{peak.zero? ? 'on dose day' : "#{peak} day#{'s' if peak > 1} after your dose"}. Plan lighter, lower-fat meals for that window."
  end

  def top_symptoms(days = 30)
    recent(days).group_by(&:symptom).map { |k, ls| [k, ls.size, ls.map(&:severity).max] }.sort_by { |_, n, _| -n }
  end
end
