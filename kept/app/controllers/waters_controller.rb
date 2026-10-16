class WatersController < ApplicationController
  def create
    current_user.water_entries.create!(drank_on: current_user.today, ml: params[:ml].to_i.clamp(50, 3000))
    respond_to do |format|
      format.turbo_stream { @o = overview; render "foods/update" }
      format.html { redirect_back fallback_location: nutrition_path }
    end
  end
end
