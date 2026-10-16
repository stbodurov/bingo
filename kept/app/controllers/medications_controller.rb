class MedicationsController < ApplicationController
  def new
    @medication = current_user.medications.new(drug: "zepbound", dose_mg: 2.5, started_on: current_user.today,
                                               dose_started_on: current_user.today, dose_weekday: current_user.today.wday)
  end

  def create
    @medication = current_user.medications.new(medication_params)
    @medication.dose_started_on ||= @medication.started_on
    @medication.dose_weekday = nil if @medication.drug.present? && DrugCatalog.fetch(@medication.drug).frequency == :daily
    if save_profile && @medication.save
      redirect_to today_path, notice: "#{@medication.label} is set up. Log your first weight to start the trend."
    else
      render :new, status: :unprocessable_entity
    end
  end

  def edit
    @medication = current_user.last_medication or return redirect_to(new_medication_path)
  end

  def update
    @medication = current_user.last_medication or return redirect_to(new_medication_path)
    restarting = !@medication.active?
    changed_dose = medication_params[:dose_mg].present? && medication_params[:dose_mg].to_f != @medication.dose_mg.to_f
    attrs = medication_params
    attrs = attrs.merge(dose_started_on: current_user.today) if changed_dose || restarting
    attrs = attrs.merge(active: true, stopped_on: nil) if restarting
    if @medication.update(attrs)
      current_user.update!(maintenance_path: nil, maintenance_started_on: nil) if restarting && current_user.maintenance_path == "stop"
      notice = restarting ? "Welcome back on #{@medication.label}. Your schedule restarts today." : (changed_dose ? "Dose updated to #{@medication.label}." : "Saved.")
      redirect_to doses_path, notice:
    else
      render :edit, status: :unprocessable_entity
    end
  end

  private

  def medication_params = params.require(:medication).permit(:drug, :dose_mg, :started_on, :dose_started_on, :dose_weekday, :monthly_cost)

  def save_profile
    p = params.fetch(:profile, {}).permit(:height, :weight, :goal_weight, :coverage)
    return true if p.blank?
    u = current_user
    u.height_cm = u.imperial? ? (p[:height].to_f * 2.54).round(1) : p[:height].to_f if p[:height].present?
    u.goal_weight_kg = to_kg(p[:goal_weight]) if p[:goal_weight].present?
    u.coverage = p[:coverage] if p[:coverage].present?
    ok = u.save
    if ok && p[:weight].present?
      u.weight_entries.find_or_initialize_by(measured_on: u.today).update(weight_kg: to_kg(p[:weight]))
    end
    @medication.errors.add(:base, u.errors.full_messages.to_sentence) unless ok
    ok
  end
end
