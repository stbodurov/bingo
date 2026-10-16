class WeightsController < ApplicationController
  def create
    p = params.require(:weight_entry).permit(:weight, :body_fat_pct, :waist, :measured_on)
    on = p[:measured_on].presence ? Date.parse(p[:measured_on]) : current_user.today
    e = current_user.weight_entries.find_or_initialize_by(measured_on: on)
    e.weight_kg = to_kg(p[:weight])
    e.body_fat_pct = p[:body_fat_pct].presence
    e.waist_cm = p[:waist].present? ? (current_user.imperial? ? (p[:waist].to_f * 2.54).round(1) : p[:waist]) : nil
    if e.save
      redirect_back fallback_location: progress_path, notice: "Weight saved."
    else
      redirect_back fallback_location: progress_path, alert: e.errors.full_messages.to_sentence
    end
  rescue Date::Error
    redirect_to progress_path, alert: "That date isn't valid."
  end

  def destroy
    current_user.weight_entries.find(params[:id]).destroy
    redirect_to progress_path, status: :see_other
  end
end
