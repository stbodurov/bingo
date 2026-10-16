class ProgressController < ApplicationController
  def show
    @o = overview
    @entries = current_user.weight_entries.order(measured_on: :desc).limit(21)
  end
end
