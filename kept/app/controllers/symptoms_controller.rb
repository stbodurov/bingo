class SymptomsController < ApplicationController
  def index
    @o = overview
    @logs = current_user.side_effect_logs.order(logged_on: :desc, id: :desc).limit(50)
    @advice = (s = params[:advice]).present? && SideEffectCoach::SYMPTOMS.key?(s) ? @o.coach.advice_for(s, params[:severity].to_i.clamp(1, 3)) : nil
  end

  def create
    log = current_user.side_effect_logs.new(params.require(:side_effect_log).permit(:symptom, :severity, :note))
    log.logged_on = current_user.today
    if log.save
      redirect_to symptoms_path(advice: log.symptom, severity: log.severity)
    else
      redirect_to symptoms_path, alert: log.errors.full_messages.to_sentence
    end
  end

  def destroy
    current_user.side_effect_logs.find(params[:id]).destroy
    redirect_to symptoms_path, status: :see_other
  end
end
