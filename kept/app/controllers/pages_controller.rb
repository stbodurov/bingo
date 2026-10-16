class PagesController < ApplicationController
  allow_unauthenticated_access
  before_action :resume_session
  layout "marketing"

  def home
    session[:ref] = params[:ref].to_s.first(12) if params[:ref].present?
    @signup = WaitlistSignup.new(role: "member")
    @count = WaitlistSignup.count
  end

  def employers
    @signup = WaitlistSignup.new(role: "employer")
    @report = Employer.find_by(slug: "northwind")&.report
  end

  def science; end
  def privacy; end
end
