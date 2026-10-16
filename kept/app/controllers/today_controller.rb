class TodayController < ApplicationController
  before_action :require_medication!

  def show
    @o = overview
    @check_in = @o.check_in_today || current_user.check_ins.new(checked_on: @o.date, appetite: 3, food_noise: 2, energy: 3, mood: 3)
  end
end
