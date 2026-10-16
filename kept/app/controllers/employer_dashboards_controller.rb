class EmployerDashboardsController < ApplicationController
  allow_unauthenticated_access
  before_action :resume_session
  layout "marketing"
  rate_limit to: 10, within: 5.minutes, only: :access, with: -> { redirect_to employers_path, alert: "Too many attempts." }

  def show
    @employer = Employer.find_by!(slug: params[:slug])
    return render(:locked) unless allowed?
    @report = @employer.report
  end

  def access
    employer = Employer.find_by!(slug: params[:slug])
    if ActiveSupport::SecurityUtils.secure_compare(params[:code].to_s, employer.access_code)
      session[:employer_access] = employer.id
      redirect_to employer_dashboard_path(employer.slug)
    else
      redirect_to employer_dashboard_path(employer.slug), alert: "That access code didn't match."
    end
  end

  private

  def allowed? = session[:employer_access] == @employer.id || current_user&.admin?
end
