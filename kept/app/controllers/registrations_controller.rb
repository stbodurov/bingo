class RegistrationsController < ApplicationController
  allow_unauthenticated_access
  rate_limit to: 5, within: 10.minutes, only: :create, with: -> { redirect_to new_registration_path, alert: "Too many sign-ups from here — try later." }

  def new
    @user = User.new
  end

  def create
    @user = User.new(params.require(:user).permit(:name, :email_address, :password, :sex, :birth_year, :units))
    if (code = params[:enrollment_code].to_s.strip).present?
      @user.employer = Employer.find_by(enrollment_code: code.upcase)
      @user.errors.add(:base, "That work code didn't match an employer plan.") unless @user.employer
    end
    if @user.errors.none? && @user.save
      start_new_session_for @user
      redirect_to new_medication_path, notice: "Welcome to Kept, #{@user.first_name}."
    else
      render :new, status: :unprocessable_entity
    end
  end
end
