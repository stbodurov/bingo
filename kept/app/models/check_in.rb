class CheckIn < ApplicationRecord
  belongs_to :user
  validates :checked_on, presence: true, uniqueness: { scope: :user_id }
  validates :appetite, :food_noise, :energy, :mood, inclusion: { in: 1..5 }
end
