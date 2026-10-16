class FoodEntry < ApplicationRecord
  belongs_to :user
  validates :eaten_on, :name, presence: true
  validates :name, length: { maximum: 80 }
  validates :protein_g, numericality: { in: 0..200 }
  validates :fiber_g, numericality: { in: 0..100 }
end
