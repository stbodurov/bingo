require "test_helper"

class MemberFlowTest < ActionDispatch::IntegrationTest
  setup do
    @user = build_member(email: "jo@example.com", weeks: 10)
    post session_path, params: { email_address: "jo@example.com", password: "long-enough-pw" }
    assert_redirected_to today_url
  end

  test "every screen renders for a member with history" do
    %w[/app /app/doses /app/symptoms /app/nutrition /app/strength /app/progress /app/plan /app/report /app/appeal
       /app/profile/edit /app/medication/edit].each do |path|
      get path
      assert_response :success, path
    end
  end

  test "logs a dose, a symptom, food, water, a workout, a weight and a check-in" do
    assert_difference -> { @user.dose_logs.count } do
      post doses_path, params: { dose_log: { taken_at: Time.current.iso8601, site: "thigh_left" } }
    end
    assert_difference -> { @user.side_effect_logs.count } do
      post symptoms_path, params: { side_effect_log: { symptom: "nausea", severity: 2 } }
    end
    assert_difference -> { @user.food_entries.count } do
      post foods_path(food_key: "shake"), as: :turbo_stream
    end
    assert_response :success
    assert_match "turbo-stream", response.body
    assert_difference -> { @user.water_entries.sum(:ml) }, 500 do
      post waters_path(ml: 500)
    end
    assert_difference -> { @user.workouts.count } do
      post workouts_path, params: { workout: { kind: "strength", minutes: 40, program_day: "A" } }
    end
    post weights_path, params: { weight_entry: { weight: 180 } }
    assert_in_delta 81.6, @user.weight_entries.find_by(measured_on: @user.today).weight_kg.to_f, 0.1
    post check_ins_path, params: { check_in: { appetite: 2, food_noise: 1, energy: 4, mood: 4 } }
    assert_equal 1, @user.check_ins.where(checked_on: @user.today).count
  end

  test "severe red-flag symptoms surface an urgent alert on Today" do
    post symptoms_path, params: { side_effect_log: { symptom: "abdominal_pain", severity: 3 } }
    get today_path
    assert_select ".alert-critical"
  end

  test "members can't touch someone else's logs" do
    other = build_member(email: "other@example.com")
    entry = other.weight_entries.first
    delete weight_path(entry)
    assert WeightEntry.exists?(entry.id)
  end

  test "exports everything as JSON and weights as CSV" do
    get export_path(format: :json)
    data = JSON.parse(response.body)
    assert_equal 11, data["weights"].size
    assert_equal "zepbound", data["medications"].first["drug"]
    get export_path(format: :csv)
    assert_equal "date,weight_kg,body_fat_pct,waist_cm", response.body.lines.first.strip
  end

  test "enrolls with an employer code and leaves again" do
    employer = Employer.create!(name: "Acme", slug: "acme", access_code: "acme-dash", enrollment_code: "ACME2026")
    post enroll_profile_path, params: { code: "acme2026" }
    assert_equal employer, @user.reload.employer
    assert @user.plus?
    delete enroll_profile_path
    assert_nil @user.reload.employer
    post enroll_profile_path, params: { code: "nope" }
    assert_equal "That code didn't match a Kept employer plan.", flash[:alert]
  end

  test "deleting the account needs the password and removes every log" do
    delete profile_path, params: { password: "wrong" }
    assert User.exists?(@user.id)
    delete profile_path, params: { password: "long-enough-pw" }
    refute User.exists?(@user.id)
    assert_equal 0, WeightEntry.where(user_id: @user.id).count
    assert_redirected_to root_path
  end
end
