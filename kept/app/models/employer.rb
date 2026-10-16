class Employer < ApplicationRecord
  has_many :users, dependent: :nullify
  normalizes :enrollment_code, with: ->(c) { c.strip.upcase }
  validates :name, :slug, :access_code, :enrollment_code, presence: true
  validates :slug, :access_code, :enrollment_code, uniqueness: true
  validates :slug, format: { with: /\A[a-z0-9-]{2,40}\z/ }
  validates :covered_lives, numericality: { greater_than_or_equal_to: 0 }, allow_nil: true
  before_validation { self.enrollment_code ||= SecureRandom.alphanumeric(8).upcase }

  def report = CohortReport.new(self)
end
