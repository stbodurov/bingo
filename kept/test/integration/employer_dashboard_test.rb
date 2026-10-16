require "test_helper"

class EmployerDashboardTest < ActionDispatch::IntegrationTest
  setup { @employer = Employer.create!(name: "Acme", slug: "acme", access_code: "acme-dash", covered_lives: 500, pmpm: 0.85, monthly_drug_cost: 1000) }

  test "locked without the access code" do
    get employer_dashboard_path("acme")
    assert_select "input[name=code]"
    post employer_access_path("acme"), params: { code: "wrong" }
    follow_redirect!
    assert_select "input[name=code]"
  end

  test "shows suppressed numbers for a small cohort and real ones for a large cohort" do
    3.times { build_member(employer: @employer) }
    post employer_access_path("acme"), params: { code: "acme-dash" }
    follow_redirect!
    assert_response :success
    assert_match(/fewer than 10/i, response.body)
    refute_match "Test Member", response.body

    9.times { build_member(employer: @employer) }
    get employer_dashboard_path("acme")
    assert_match "Persistence", response.body
    refute_match "Test Member", response.body
  end
end
