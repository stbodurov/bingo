class PlanController < ApplicationController
  before_action :require_medication!

  def show
    @o = overview
    @maintenance = @o.maintenance
    @cost = @o.cost
  end
end
