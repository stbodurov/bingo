class WorkoutsController < ApplicationController
  def create
    w = current_user.workouts.new(params.require(:workout).permit(:kind, :minutes, :program_day))
    w.performed_on = current_user.today
    if w.save
      redirect_back fallback_location: strength_path, notice: w.kind == "strength" ? "Strength session logged — your muscle thanks you." : "Logged."
    else
      redirect_back fallback_location: strength_path, alert: w.errors.full_messages.to_sentence
    end
  end

  def destroy
    current_user.workouts.find(params[:id]).destroy
    redirect_to strength_path, status: :see_other
  end
end
