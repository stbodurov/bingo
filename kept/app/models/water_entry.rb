class WaterEntry < ApplicationRecord
  belongs_to :user
  validates :drank_on, presence: true
  validates :ml, numericality: { in: 50..3000 }
end
