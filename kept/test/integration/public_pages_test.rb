require "test_helper"

class PublicPagesTest < ActionDispatch::IntegrationTest
  test "marketing pages render with a CSP header" do
    %w[/ /employers /science /privacy /session/new /registration/new /passwords/new].each do |path|
      get path
      assert_response :success, path
      assert_match "default-src 'self'", response.headers["Content-Security-Policy"]
    end
  end

  test "the app requires a session" do
    %w[/app /app/doses /app/nutrition /app/progress /app/report /app/export.json].each do |path|
      get path
      assert_redirected_to new_session_path, path
    end
  end

  test "waitlist sign-up, referral and honeypot" do
    post waitlist_signups_path, params: { waitlist_signup: { email: "first@example.com", role: "member", drug: "zepbound", stage: "on" } }
    first = WaitlistSignup.find_by!(email: "first@example.com")
    assert_redirected_to waitlist_status_path(first.code)
    follow_redirect!
    assert_select "h1"

    get root_path(ref: first.code)
    post waitlist_signups_path, params: { waitlist_signup: { email: "friend@example.com", role: "member" } }
    assert_equal first.code, WaitlistSignup.find_by!(email: "friend@example.com").referred_by

    assert_no_difference -> { WaitlistSignup.count } do
      post waitlist_signups_path, params: { waitlist_signup: { email: "bot@example.com", role: "member", website: "spam" } }
    end

    post waitlist_signups_path, params: { waitlist_signup: { email: "not-an-email", role: "member" } }
    assert_response :unprocessable_entity
  end
end
