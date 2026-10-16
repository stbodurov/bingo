class WeightEntry < ApplicationRecord
  belongs_to :user
  validates :measured_on, presence: true, uniqueness: { scope: :user_id }
  validates :weight_kg, numericality: { in: 30..350 }
  validates :body_fat_pct, numericality: { in: 3..70 }, allow_nil: true
  validates :waist_cm, numericality: { in: 40..250 }, allow_nil: true
end
