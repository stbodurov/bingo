require "test_helper"

class CostOptionsTest < ActiveSupport::TestCase
  test "covered members see the savings card first" do
    c = CostOptions.new(drug_key: "zepbound", dose_mg: 10, coverage: "commercial_covered")
    assert_equal "savings_card", c.best.key
    assert_equal 25, c.best.monthly
  end

  test "uncovered members see cash, a cheaper alternative and an appeal" do
    c = CostOptions.new(drug_key: "zepbound", dose_mg: 10, coverage: "commercial_not_covered")
    keys = c.options.map(&:key)
    assert_equal %w[switch direct appeal], keys
    assert_equal 449, c.cash_price
    assert_equal (449 - 149) * 12, c.annual_saving(449)
  end

  test "Medicaid members are not pointed at cash prices" do
    c = CostOptions.new(drug_key: "wegovy", dose_mg: 2.4, coverage: "medicaid")
    refute c.options.find { |o| o.key == "direct" }.eligible
  end

  test "Medicare members see the $50 program" do
    c = CostOptions.new(drug_key: "wegovy", dose_mg: 2.4, coverage: "medicare")
    assert_equal 50, c.options.find { |o| o.key == "medicare" }.monthly
  end
end
