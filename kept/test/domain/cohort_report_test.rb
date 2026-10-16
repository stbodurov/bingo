require "test_helper"

class CohortReportTest < ActiveSupport::TestCase
  setup { @employer = Employer.create!(name: "Acme", slug: "acme", access_code: "acme-code", covered_lives: 1000, pmpm: 0.85, monthly_drug_cost: 1000) }

  test "suppresses small cohorts so no one can be singled out" do
    5.times { build_member(employer: @employer) }
    report = CohortReport.new(@employer)
    assert report.suppressed?
    assert_nil report.persistence(3)
  end

  test "reports persistence, loss and spend for a large enough cohort" do
    12.times { |i| build_member(employer: @employer, weeks: 20) }
    report = CohortReport.new(@employer)
    refute report.suppressed?
    assert_equal 1.0, report.persistence(3)
    assert_nil report.persistence(12)
    assert_operator report.avg_loss_pct, :>, 8
    assert_equal 12, report.active_count
    assert_equal 12_000, report.monthly_drug_spend
    assert_equal 850, report.monthly_kept_cost
    assert_equal [["Zepbound", 12]], report.by_drug
  end
end
