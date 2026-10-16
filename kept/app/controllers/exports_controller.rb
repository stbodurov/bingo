require "csv"

# Your data, yours to take: every log as CSV (weights) or as one JSON document (everything).
class ExportsController < ApplicationController
  def show
    u = current_user
    respond_to do |format|
      format.csv do
        csv = CSV.generate do |rows|
          rows << %w[date weight_kg body_fat_pct waist_cm]
          u.weight_entries.order(:measured_on).each { |e| rows << [e.measured_on, e.weight_kg, e.body_fat_pct, e.waist_cm] }
        end
        send_data csv, filename: "kept-weights-#{u.today}.csv", type: "text/csv"
      end
      format.json do
        data = {
          exported_at: Time.current.iso8601,
          profile: u.slice(:name, :email_address, :sex, :birth_year, :height_cm, :units, :goal_weight_kg, :protein_factor, :meals_per_day, :coverage, :time_zone),
          medications: u.medications.order(:started_on).map { |m| m.slice(:drug, :dose_mg, :started_on, :dose_started_on, :active) },
          doses: u.dose_logs.order(:taken_at).map { |d| d.slice(:taken_at, :dose_mg, :site, :skipped, :note) },
          side_effects: u.side_effect_logs.order(:logged_on).map { |s| s.slice(:logged_on, :symptom, :severity, :note) },
          weights: u.weight_entries.order(:measured_on).map { |w| w.slice(:measured_on, :weight_kg, :body_fat_pct, :waist_cm) },
          food: u.food_entries.order(:eaten_on).map { |f| f.slice(:eaten_on, :name, :protein_g, :fiber_g, :calories) },
          water: u.water_entries.order(:drank_on).map { |w| w.slice(:drank_on, :ml) },
          workouts: u.workouts.order(:performed_on).map { |w| w.slice(:performed_on, :kind, :minutes, :program_day) },
          check_ins: u.check_ins.order(:checked_on).map { |c| c.slice(:checked_on, :appetite, :food_noise, :energy, :mood) }
        }
        send_data JSON.pretty_generate(data), filename: "kept-export-#{u.today}.json", type: "application/json"
      end
    end
  end
end
