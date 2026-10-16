class WaitlistSignup < ApplicationRecord
  ROLES = { "member" => "I take (or plan to take) a GLP-1", "clinician" => "I prescribe or care for patients on GLP-1s",
            "employer" => "I buy benefits for an employer or health plan" }.freeze
  STAGES = { "considering" => "Considering it", "starting" => "Just started (0–3 months)", "on" => "On it 3+ months",
             "stopping" => "Planning to stop or reduce", "stopped" => "Stopped" }.freeze

  normalizes :email, with: ->(e) { e.strip.downcase }
  validates :email, presence: true, uniqueness: true, format: { with: URI::MailTo::EMAIL_REGEXP }, length: { maximum: 200 }
  validates :role, inclusion: { in: ROLES.keys }
  validates :stage, inclusion: { in: STAGES.keys }, allow_blank: true
  validates :drug, inclusion: { in: DrugCatalog::DRUGS.keys + ["none"] }, allow_blank: true
  validates :first_name, :company, length: { maximum: 80 }
  before_validation { self.code ||= SecureRandom.alphanumeric(8).downcase }

  def referrals = WaitlistSignup.where(referred_by: code).count
  def position
    ordered = WaitlistSignup.order(:created_at).pluck(:id, :code)
    counts = WaitlistSignup.where.not(referred_by: nil).group(:referred_by).count
    ranked = ordered.each_with_index.sort_by { |(_, c), i| [i - 5 * counts.fetch(c, 0), i] }.map { |(row_id, _), _| row_id }
    ranked.index(id).to_i + 1
  end
end
