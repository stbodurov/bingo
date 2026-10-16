require "test_helper"

class WaitlistSignupTest < ActiveSupport::TestCase
  test "referrals move you up five places each" do
    early = (1..8).map { |i| WaitlistSignup.create!(email: "w#{i}@example.com", role: "member", created_at: i.minutes.ago - 1.hour) }
    late = WaitlistSignup.create!(email: "late@example.com", role: "member")
    assert_equal 9, late.position
    WaitlistSignup.create!(email: "friend@example.com", role: "member", referred_by: late.code)
    assert_equal 5, late.reload.position # 9th - 5 places, after the earlier signup it ties with
    assert_equal 1, late.referrals
    assert early.first.code.present?
  end

  test "normalises email and rejects duplicates" do
    WaitlistSignup.create!(email: " Jo@Example.com ", role: "member")
    dup = WaitlistSignup.new(email: "jo@example.com", role: "member")
    refute dup.valid?
  end
end
