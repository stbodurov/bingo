require "csv"

# Operator console: waitlist segments and export, employer accounts, and product health.
# Admin users only; never shows individual health data.
class AdminController < ApplicationController
  before_action :require_admin!

  def show
    @signups = WaitlistSignup.order(created_at: :desc)
    @by_role = WaitlistSignup.group(:role).count
    @by_drug = WaitlistSignup.where.not(drug: [nil, ""]).group(:drug).count.sort_by { |_, n| -n }
    @by_stage = WaitlistSignup.where.not(stage: [nil, ""]).group(:stage).count
    @referred = WaitlistSignup.where.not(referred_by: nil).count
    @employers = Employer.left_joins(:users).group(:id).select("employers.*, COUNT(users.id) AS members_count").order(:name)
    week_ago = 7.days.ago
    @metrics = {
      members: User.where(admin: false).count,
      plus: User.where(admin: false).where("plan = 'plus' OR employer_id IS NOT NULL").count,
      active_7d: DoseLog.where("taken_at >= ?", week_ago).distinct.count(:user_id),
      logging_7d: WeightEntry.where("measured_on >= ?", week_ago.to_date).distinct.count(:user_id),
      off_ramp: User.where.not(maintenance_path: nil).count
    }
    respond_to do |format|
      format.html
      format.csv do
        csv = CSV.generate do |rows|
          rows << %w[created_at email first_name role drug stage company code referred_by referrals source]
          counts = WaitlistSignup.where.not(referred_by: nil).group(:referred_by).count
          @signups.each do |s|
            rows << [s.created_at.iso8601, s.email, s.first_name, s.role, s.drug, s.stage, s.company, s.code, s.referred_by, counts.fetch(s.code, 0), s.source]
          end
        end
        send_data csv, filename: "kept-waitlist-#{Date.current}.csv", type: "text/csv"
      end
    end
  end

  def create_employer
    e = Employer.new(params.require(:employer).permit(:name, :slug, :covered_lives, :pmpm, :monthly_drug_cost))
    e.access_code = SecureRandom.alphanumeric(16).downcase
    if e.save
      redirect_to admin_path(anchor: "employers"), notice: "#{e.name} created. Dashboard code: #{e.access_code} · enrollment code: #{e.enrollment_code}"
    else
      redirect_to admin_path(anchor: "employers"), alert: e.errors.full_messages.to_sentence
    end
  end

  private

  def require_admin!
    head :not_found unless current_user&.admin?
  end
end
