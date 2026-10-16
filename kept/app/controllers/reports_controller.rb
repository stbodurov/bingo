class ReportsController < ApplicationController
  before_action :require_medication!
  layout "print"

  def show
    @o = overview
  end

  def appeal
    @o = overview
  end
end
