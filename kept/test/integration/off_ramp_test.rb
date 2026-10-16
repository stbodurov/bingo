require "test_helper"

class OffRampTest < ActionDispatch::IntegrationTest
  setup do
    @user = build_member(email: "taper@example.com", drug: "wegovy", dose: 2.4, weeks: 30, start_kg: 110)
    post session_path, params: { email_address: "taper@example.com", password: "long-enough-pw" }
  end

  test "taper: start, step down the ladder, take the last dose, stay in maintenance" do
    post off_ramp_path(path: "taper")
    @user.reload
    assert_equal "taper", @user.maintenance_path
    assert_equal 2.4, @user.maintenance_from_dose_mg.to_f
    assert @user.maintenance_baseline_kg.present?
    get plan_path
    assert_select ".ramp-now", text: /week 1/
    post step_off_ramp_path
    assert_equal 1.7, @user.medication.reload.dose_mg.to_f

    # Thirteen weeks later the plan calls for stopping.
    travel 12.weeks do
      get plan_path
      assert_select "button", text: "I've taken my last dose"
      post stop_off_ramp_path
      assert @user.reload.off_medication?
      get today_path
      assert_select "h3", text: "Maintenance mode"
      %w[/app/doses /app/plan /app/report /app/progress].each { |p| get(p); assert_response :success, p }
      post doses_path, params: { dose_log: { taken_at: Time.current.iso8601 } }
      assert_redirected_to edit_medication_path
      get dose_calendar_path(@user.calendar_token, format: :ics)
      refute_includes response.body, "RRULE"
    end
  end

  test "the regain guardrail raises a critical alert" do
    post off_ramp_path(path: "stop")
    post stop_off_ramp_path
    base = @user.reload.maintenance_baseline_kg.to_f
    (1..40).each { |d| @user.weight_entries.create!(measured_on: Date.current + d, weight_kg: base * 1.04) }
    travel 40.days do
      get today_path
      assert_select ".alert-critical", text: /Regain guardrail/
    end
  end

  test "restarting the medicine ends the stop plan and restarts the schedule" do
    post off_ramp_path(path: "stop")
    post stop_off_ramp_path
    patch medication_path, params: { medication: { dose_mg: 0.25 } }
    med = @user.reload.medication
    assert med.active?
    assert_equal 0.25, med.dose_mg.to_f
    assert_equal Date.current, med.dose_started_on
    assert_nil @user.maintenance_path
  end

  test "rejects unknown paths and can end a plan" do
    post off_ramp_path(path: "yolo")
    assert_nil @user.reload.maintenance_path
    post off_ramp_path(path: "maintain")
    delete off_ramp_path
    assert_nil @user.reload.maintenance_path
  end
end
