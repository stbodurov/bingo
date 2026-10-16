class User < ApplicationRecord
  has_secure_password
  has_many :sessions, dependent: :destroy
  belongs_to :employer, optional: true
  has_many :medications, dependent: :destroy
  has_many :dose_logs, dependent: :destroy
  has_many :side_effect_logs, dependent: :destroy
  has_many :weight_entries, dependent: :destroy
  has_many :food_entries, dependent: :destroy
  has_many :water_entries, dependent: :destroy
  has_many :workouts, dependent: :destroy
  has_many :check_ins, dependent: :destroy

  normalizes :email_address, with: ->(e) { e.strip.downcase }

  validates :email_address, presence: true, uniqueness: true, format: { with: URI::MailTo::EMAIL_REGEXP }
  validates :password, length: { minimum: 10 }, allow_nil: true
  validates :sex, inclusion: { in: %w[female male] }
  validates :units, inclusion: { in: %w[imperial metric] }
  validates :coverage, inclusion: { in: CostOptions::COVERAGE.keys }
  validates :protein_factor, numericality: { in: 1.0..2.0 }
  validates :meals_per_day, numericality: { in: 2..6 }
  validates :height_cm, numericality: { in: 120..230 }, allow_nil: true
  validates :goal_weight_kg, numericality: { in: 35..250 }, allow_nil: true
  validates :birth_year, numericality: { in: 1920..(Date.current.year - 18) }, allow_nil: true
  validates :name, length: { maximum: 60 }

  has_secure_token :calendar_token

  def medication = medications.where(active: true).order(:started_on).last
  def first_name = name.presence&.split&.first || "there"
  def imperial? = units == "imperial"
  def plus? = plan == "plus" || employer.present?
  def age = birth_year && Date.current.year - birth_year
  def today = Time.current.in_time_zone(time_zone).to_date
  def current_weight_kg = weight_entries.order(:measured_on).last&.weight_kg&.to_f
end
