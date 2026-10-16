class StrengthController < ApplicationController
  def show
    @o = overview
    @next_day = StrengthProgram.next_day(@o.workouts)
    @home = params[:mode] == "home"
  end
end
