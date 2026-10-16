class ApplicationController < ActionController::Base
  include Authentication
  allow_browser versions: :modern
  stale_when_importmap_changes

  helper_method :current_user, :overview

  private

  def current_user = Current.session&.user
  def overview = @overview ||= Overview.new(current_user)

  def require_medication!
    redirect_to new_medication_path, notice: "First, tell Kept which medicine you take." unless current_user.last_medication
  end

  def to_kg(value)
    v = value.to_s.strip
    return nil if v.empty?
    current_user.imperial? ? (v.to_f * 0.45359237).round(2) : v.to_f
  end
end
