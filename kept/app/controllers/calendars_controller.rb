# Dose reminders that work everywhere without an app install: a private calendar feed
# (iCalendar) with a recurring event and an alarm on dose day. The URL carries a random
# token because calendar apps can't sign in; members can rotate it from Settings.
class CalendarsController < ApplicationController
  allow_unauthenticated_access only: :show
  skip_forgery_protection only: :show

  def show
    user = User.find_by(calendar_token: params[:token].to_s)
    return head(:not_found) unless user && params[:token].to_s.length >= 20
    send_data DoseCalendar.new(user).to_ics, type: "text/calendar; charset=utf-8", disposition: "inline", filename: "kept-doses.ics"
  end

  def reset
    current_user.regenerate_calendar_token
    redirect_to edit_profile_path(anchor: "reminders"), notice: "New calendar link created. The old one no longer works."
  end
end
