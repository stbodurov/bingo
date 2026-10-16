# The off-ramp in motion: start a maintain / taper / stop plan, record each step down the
# label ladder, record the last dose, or end the plan. Every step is the member reporting
# what they and their prescriber decided — Kept never changes a dose by itself.
class OffRampsController < ApplicationController
  before_action :require_medication!
  before_action :load_plan

  def create
    path = params[:path].to_s
    return redirect_to(plan_path, alert: "Pick one of the three paths.") unless MaintenancePlan::PATHS.include?(path)
    med = current_user.medication or return redirect_to(plan_path, alert: "You're already off medication.")
    current_user.update!(maintenance_path: path, maintenance_started_on: current_user.today,
                         maintenance_from_dose_mg: med.dose_mg, maintenance_baseline_kg: @o.trend.current_trend&.round(1))
    redirect_to plan_path(anchor: "off-ramp"), notice: "Plan started. Kept will watch your trend against today's #{helpers.weight(@o.trend.current_trend)}."
  end

  # Step down to the dose the plan calls for now (or the next labelled step).
  def step
    med = current_user.medication or return redirect_to(plan_path)
    target = @o.off_ramp&.step&.dose_mg.to_f
    return stop if target.zero?
    if (target - med.dose_mg.to_f).abs < 0.001
      return redirect_to(plan_path(anchor: "off-ramp"), notice: "You're already on #{med.label}.")
    end
    med.update!(dose_mg: target, dose_started_on: current_user.today)
    redirect_to plan_path(anchor: "off-ramp"), notice: "Logged the step down to #{med.label}."
  end

  # The last dose has been taken: maintenance mode from here, with the guardrail still on.
  def stop
    med = current_user.medication or return redirect_to(plan_path)
    med.update!(active: false, stopped_on: current_user.today)
    unless current_user.maintenance_path
      current_user.update!(maintenance_path: "stop", maintenance_started_on: current_user.today,
                           maintenance_from_dose_mg: med.dose_mg, maintenance_baseline_kg: @o.trend.current_trend&.round(1))
    end
    redirect_to today_path, notice: "Marked as off medication. Kept keeps watching your trend — keep weighing in 3× a week."
  end

  def destroy
    current_user.update!(maintenance_path: nil, maintenance_started_on: nil, maintenance_from_dose_mg: nil, maintenance_baseline_kg: nil)
    redirect_to plan_path(anchor: "off-ramp"), notice: "Plan ended.", status: :see_other
  end

  private

  def load_plan = @o = overview
end
