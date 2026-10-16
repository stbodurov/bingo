class CheckInsController < ApplicationController
  def create
    c = current_user.check_ins.find_or_initialize_by(checked_on: current_user.today)
    c.assign_attributes(params.require(:check_in).permit(:appetite, :food_noise, :energy, :mood))
    if c.save
      redirect_to today_path, notice: "Check-in saved."
    else
      redirect_to today_path, alert: c.errors.full_messages.to_sentence
    end
  end
end
