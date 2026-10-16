class Medication < ApplicationRecord
  belongs_to :user
  has_many :dose_logs, dependent: :destroy

  validates :drug, inclusion: { in: DrugCatalog::DRUGS.keys }
  validates :dose_mg, numericality: { greater_than: 0, less_than_or_equal_to: 60 }
  validates :started_on, :dose_started_on, presence: true
  validates :dose_weekday, inclusion: { in: 0..6 }, allow_nil: true
  validates :monthly_cost, numericality: { greater_than_or_equal_to: 0, less_than: 5000 }, allow_nil: true
  validate :dose_on_ladder

  def info = DrugCatalog.fetch(drug)
  def label = "#{info.brand} #{dose_mg.to_f.to_s.sub(/\.0\z/, '')} mg"

  private

  def dose_on_ladder
    return unless DrugCatalog::DRUGS.key?(drug) && dose_mg
    errors.add(:dose_mg, "isn't a labelled dose for #{info.brand}") unless info.ladder.any? { |d| (d - dose_mg.to_f).abs < 0.001 }
  end
end
