require "test_helper"

class AdminTest < ActionDispatch::IntegrationTest
  setup do
    @admin = User.create!(email_address: "ops@example.com", password: "long-enough-pw", admin: true)
    WaitlistSignup.create!(email: "a@example.com", role: "member", drug: "zepbound", stage: "on")
    WaitlistSignup.create!(email: "b@example.com", role: "employer", company: "Acme")
  end

  test "hidden from members" do
    build_member(email: "m@example.com")
    post session_path, params: { email_address: "m@example.com", password: "long-enough-pw" }
    get admin_path
    assert_response :not_found
  end

  test "admins see segments, export the waitlist and create employers" do
    post session_path, params: { email_address: "ops@example.com", password: "long-enough-pw" }
    assert_redirected_to admin_url
    get admin_path
    assert_response :success
    assert_select ".stat .v", text: "2"
    get admin_path(format: :csv)
    lines = response.body.lines
    assert_equal 3, lines.size
    assert_match "a@example.com", response.body

    post admin_employers_path, params: { employer: { name: "Globex", slug: "globex", covered_lives: 2000, pmpm: 0.85, monthly_drug_cost: 1000 } }
    e = Employer.find_by!(slug: "globex")
    assert_equal 16, e.access_code.size
    assert e.enrollment_code.present?
    post admin_employers_path, params: { employer: { name: "Bad", slug: "Bad Slug!", covered_lives: 10 } }
    refute Employer.exists?(name: "Bad")
  end
end
