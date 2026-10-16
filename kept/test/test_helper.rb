ENV["RAILS_ENV"] ||= "test"
require_relative "../config/environment"
require "rails/test_help"
require_relative "test_helpers/session_test_helper"

module ActiveSupport
  class TestCase
    parallelize(workers: :number_of_processors)
    fixtures :all

    # A member on Zepbound with `weeks` of weekly doses and a steady weight loss.
    def build_member(email: "m#{SecureRandom.hex(4)}@example.com", drug: "zepbound", dose: 5, weeks: 10, start_kg: 100.0,
                     weekly_loss: 0.8, employer: nil, today: Date.current)
      user = User.create!(email_address: email, password: "long-enough-pw", name: "Test Member", birth_year: 1980,
                          height_cm: 170, employer:)
      start = today - weeks * 7
      med = user.medications.create!(drug:, dose_mg: dose, started_on: start, dose_started_on: start, dose_weekday: start.wday)
      (0..weeks).each do |w|
        day = start + w * 7
        user.dose_logs.create!(medication: med, taken_at: day.in_time_zone.change(hour: 9), dose_mg: dose, site: DrugCatalog::INJECTION_SITES[w % 6])
        user.weight_entries.create!(measured_on: day, weight_kg: (start_kg * (1 - weekly_loss / 100.0)**w).round(1))
      end
      user
    end
  end
end
