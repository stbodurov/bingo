class Workout < ApplicationRecord
  KINDS = %w[strength walk cardio].freeze
  belongs_to :user
  validates :performed_on, presence: true
  validates :kind, inclusion: { in: KINDS }
  validates :minutes, numericality: { in: 1..300 }
  validates :program_day, inclusion: { in: %w[A B C] }, allow_nil: true
end
