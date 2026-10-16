class SessionsController < ApplicationController
  allow_unauthenticated_access only: %i[ new create demo ]
  rate_limit to: 20, within: 3.minutes, only: :demo, with: -> { redirect_to root_path, alert: "Try again later." }
  rate_limit to: 10, within: 3.minutes, only: :create, with: -> { redirect_to new_session_path, alert: "Try again later." }

  def new
  end

  def create
    if user = User.authenticate_by(params.permit(:email_address, :password))
      start_new_session_for user
      redirect_to after_authentication_url == root_url ? today_url : after_authentication_url
    else
      redirect_to new_session_path, alert: "Try another email address or password."
    end
  end

  def demo
    user = User.find_by(email_address: "demo@kept.example")
    return redirect_to(root_path, alert: "The demo isn't available right now.") unless user
    start_new_session_for user
    redirect_to today_path
  end

  def destroy
    terminate_session
    redirect_to root_path, status: :see_other
  end
end
