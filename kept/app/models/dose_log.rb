class DoseLog < ApplicationRecord
  belongs_to :user
  belongs_to :medication
  validates :taken_at, presence: true
  validates :dose_mg, numericality: { greater_than: 0 }
  validates :site, inclusion: { in: DrugCatalog::INJECTION_SITES }, allow_nil: true
  validates :note, length: { maximum: 280 }
end
