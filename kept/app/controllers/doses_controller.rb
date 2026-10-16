class DosesController < ApplicationController
  before_action :require_medication!

  def index
    @o = overview
    @logs = current_user.dose_logs.includes(:medication).order(taken_at: :desc).limit(40)
  end

  def create
    med = current_user.medication
    unless med
      return redirect_to(edit_medication_path, alert: "You're marked as off medication. If you've restarted with your prescriber, update your medicine first.")
    end
    log = current_user.dose_logs.new(medication: med, dose_mg: med.dose_mg,
                                     taken_at: params.dig(:dose_log, :taken_at).presence || Time.current,
                                     site: params.dig(:dose_log, :site).presence, skipped: params.dig(:dose_log, :skipped) == "1",
                                     note: params.dig(:dose_log, :note))
    if log.save
      redirect_back fallback_location: today_path, notice: log.skipped ? "Marked as skipped." : "#{med.label} logged."
    else
      redirect_back fallback_location: doses_path, alert: log.errors.full_messages.to_sentence
    end
  end

  def destroy
    current_user.dose_logs.find(params[:id]).destroy
    redirect_to doses_path, notice: "Removed.", status: :see_other
  end
end
