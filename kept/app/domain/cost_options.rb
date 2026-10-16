# Ways to pay for a GLP-1 in October 2026 and what each costs per month, so people don't
# stop treatment over price when a cheaper legitimate route exists. Figures are planning
# estimates from manufacturer and federal announcements; Kept tells people to confirm.
class CostOptions
  Option = Data.define(:key, :label, :monthly, :how, :eligible, :note)

  COVERAGE = {
    "commercial_covered" => "Employer / commercial insurance that covers it",
    "commercial_not_covered" => "Commercial insurance that doesn't cover it",
    "medicare" => "Medicare",
    "medicaid" => "Medicaid",
    "uninsured" => "No insurance / paying cash"
  }.freeze

  def initialize(drug_key:, dose_mg:, coverage:)
    @drug = DrugCatalog.fetch(drug_key)
    @dose = dose_mg.to_f
    @coverage = coverage
  end

  def cash_price
    prices = @drug.cash_prices
    return nil if prices.empty?
    prices.find { |d, _| (d - @dose).abs < 0.001 }&.last || prices.values.max
  end

  def options
    out = []
    if @coverage == "commercial_covered"
      out << Option.new(key: "savings_card", label: "Manufacturer savings card", monthly: 25,
                        how: "Activate the #{@drug.maker} savings card and give it to your pharmacy with your insurance.",
                        eligible: true, note: "Typical floor price with coverage; terms and annual caps apply.")
    end
    if @coverage == "medicare"
      out << Option.new(key: "medicare", label: "Medicare GLP-1 program", monthly: 50,
                        how: "Ask your prescriber to document eligibility (obesity with qualifying conditions) for the Medicare GLP-1 program.",
                        eligible: true, note: "Launched mid-2026 for eligible beneficiaries; not every enrollee qualifies.")
    end
    if cash_price
      out << Option.new(key: "direct", label: "#{@drug.maker} direct / TrumpRx cash price", monthly: cash_price,
                        how: "Buy direct from the manufacturer's pharmacy or via TrumpRx with a prescription.",
                        eligible: @coverage != "medicaid", note: "Self-pay; doesn't count toward your deductible.")
    end
    alt = cheaper_alternative
    if alt
      out << Option.new(key: "switch", label: "Ask about #{alt[:brand]}", monthly: alt[:price],
                        how: "An FDA-approved alternative at a lower cash price — switching is a prescriber decision.",
                        eligible: true, note: alt[:note])
    end
    if @coverage == "commercial_not_covered"
      out << Option.new(key: "appeal", label: "Appeal the coverage decision", monthly: nil,
                        how: "Kept drafts a letter of medical necessity from your logs for your prescriber to send.",
                        eligible: true, note: "Works best with documented BMI, related conditions and prior attempts.")
    end
    out.sort_by { |o| o.monthly || Float::INFINITY }
  end

  def best = options.select(&:eligible).min_by { |o| o.monthly || Float::INFINITY }

  def annual_saving(current_monthly)
    b = best
    return 0 unless b&.monthly && current_monthly
    [(current_monthly.to_f - b.monthly) * 12, 0].max.round
  end

  private

  def cheaper_alternative
    return nil if @drug.form == :pill
    if @drug.generic == "semaglutide"
      { brand: "the Wegovy pill", price: 149, note: "Same molecule taken daily; $149 for starter doses, $299 at higher doses." }
    elsif @drug.generic == "tirzepatide" && @dose > 2.5
      { brand: "Foundayo", price: 149, note: "Daily pill; $149–$349 depending on dose. Typically less weight loss than higher tirzepatide doses." }
    end
  end
end
