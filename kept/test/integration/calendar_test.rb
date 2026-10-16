require "test_helper"

class CalendarTest < ActionDispatch::IntegrationTest
  test "private feed serves a recurring dose event with an alarm" do
    user = build_member(weeks: 2)
    get dose_calendar_path(user.calendar_token, format: :ics)
    assert_response :success
    assert_equal "text/calendar", response.media_type
    body = response.body
    assert_includes body, "RRULE:FREQ=WEEKLY"
    assert_includes body, "BEGIN:VALARM"
    assert_includes body, "TZID=America/New_York"
    assert body.split("\r\n").all? { |l| l.bytesize <= 75 }, "lines must be folded"
    assert_includes body, "Dose check-in with your prescriber"
  end

  test "unknown or rotated tokens are refused" do
    user = build_member(weeks: 2)
    old = user.calendar_token
    get dose_calendar_path("x" * 24, format: :ics)
    assert_response :not_found
    user.regenerate_calendar_token
    get dose_calendar_path(old, format: :ics)
    assert_response :not_found
  end
end
