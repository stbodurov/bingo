class ProfilesController < ApplicationController
  def edit
    @user = current_user
  end

  def update
    @user = current_user
    p = params.require(:user).permit(:name, :sex, :birth_year, :units, :height, :goal_weight, :protein_factor, :meals_per_day, :coverage, :time_zone)
    @user.assign_attributes(p.except(:height, :goal_weight))
    @user.height_cm = @user.imperial? ? (p[:height].to_f * 2.54).round(1) : p[:height].to_f if p[:height].present?
    @user.goal_weight_kg = to_kg(p[:goal_weight]) if p[:goal_weight].present?
    if @user.save
      redirect_to edit_profile_path, notice: "Profile saved."
    else
      render :edit, status: :unprocessable_entity
    end
  end

  # Kept Plus through work: the benefits team hands out a code; nothing about the member flows back
  # except de-identified, suppressed totals.
  def enroll
    employer = Employer.find_by(enrollment_code: params[:code].to_s.strip.upcase)
    if employer
      current_user.update!(employer:)
      redirect_to edit_profile_path, notice: "You're on Kept Plus through #{employer.name}."
    else
      redirect_to edit_profile_path, alert: "That code didn't match a Kept employer plan."
    end
  end

  def unenroll
    current_user.update!(employer: nil)
    redirect_to edit_profile_path, notice: "Removed. Your logs stay with you."
  end

  def destroy
    user = current_user
    unless user.authenticate(params[:password].to_s)
      return redirect_to edit_profile_path, alert: "That password didn't match, so nothing was deleted."
    end
    terminate_session
    user.destroy!
    redirect_to root_path, notice: "Your account and every log in it have been deleted."
  end
end
