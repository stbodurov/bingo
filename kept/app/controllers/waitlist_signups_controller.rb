class WaitlistSignupsController < ApplicationController
  allow_unauthenticated_access
  layout "marketing"
  rate_limit to: 8, within: 10.minutes, only: :create, with: -> { redirect_to root_path(anchor: "waitlist"), alert: "Too many tries — please wait a few minutes." }

  def create
    return redirect_to(root_path) if params.dig(:waitlist_signup, :website).present? # honeypot
    @signup = WaitlistSignup.find_by(email: signup_params[:email].to_s.strip.downcase) || WaitlistSignup.new(signup_params)
    @signup.referred_by ||= session[:ref] if @signup.new_record? && WaitlistSignup.exists?(code: session[:ref])
    @signup.source ||= request.referer.to_s.first(120)
    if @signup.save
      redirect_to waitlist_status_path(@signup.code)
    else
      @count = WaitlistSignup.count
      render(@signup.role == "employer" ? "pages/employers" : "pages/home", status: :unprocessable_entity)
    end
  end

  def show
    @signup = WaitlistSignup.find_by!(code: params[:code])
  end

  private

  def signup_params = params.require(:waitlist_signup).permit(:email, :first_name, :role, :drug, :stage, :company)
end
