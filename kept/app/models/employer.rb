class Employer < ApplicationRecord
  has_many :users, dependent: :nullify
  normalizes :enrollment_code, with: ->(c) { c.strip.upcase }
  validates :name, :slug, :access_code, :enrollment_code, presence: true
  validates :slug, :access_code, :enrollment_code, uniqueness: true
  before_validation { self.enrollment_code ||= SecureRandom.alphanumeric(8).upcase }

  def report = CohortReport.new(self)
end
