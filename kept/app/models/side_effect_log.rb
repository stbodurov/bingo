class SideEffectLog < ApplicationRecord
  belongs_to :user
  validates :logged_on, presence: true
  validates :symptom, inclusion: { in: SideEffectCoach::SYMPTOMS.keys }
  validates :severity, inclusion: { in: 1..3 }
  validates :note, length: { maximum: 280 }
end
