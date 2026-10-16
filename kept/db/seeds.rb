# Demo data: one fully-logged member (Jordan) and a fictional employer cohort (Northwind
# Logistics) large enough for the de-identified dashboard. All people are fictional.
require "securerandom"

# Production runs seeds on first boot (db:prepare). Set KEPT_SEED_DEMO=0 to start empty.
return puts("Skipping demo seed (KEPT_SEED_DEMO=0)") if ENV["KEPT_SEED_DEMO"] == "0"

rng = Random.new(2026)
today = Date.current

employer = Employer.find_or_create_by!(slug: "northwind") do |e|
  e.name = "Northwind Logistics"
  e.covered_lives = 4_200
  e.pmpm = 0.85
  e.monthly_drug_cost = 1_050
  e.access_code = "northwind-demo"
  e.enrollment_code = "NORTHWIND"
end

def log_member(user, drug:, start:, doses:, start_weight:, weekly_loss_pct:, protein_ratio:, strength_per_week:, rng:, today:, stop_after_weeks: nil, side_effects: true)
  info = DrugCatalog.fetch(drug)
  current = doses.last
  dose_idx = doses.index(current)
  med = user.medications.create!(drug:, dose_mg: current, started_on: start,
                                 dose_started_on: start + (dose_idx * info.step_weeks * 7), dose_weekday: start.wday,
                                 monthly_cost: info.cash_prices.values.max)
  weeks = ((today - start).to_i / 7)
  active_weeks = stop_after_weeks ? [weeks, stop_after_weeks].min : weeks
  sites = DrugCatalog::INJECTION_SITES
  zone = ActiveSupport::TimeZone[user.time_zone]
  now = Time.current
  rows = { dose_logs: [], weight_entries: [], food_entries: [], water_entries: [], workouts: [], side_effect_logs: [], check_ins: [] }
  add = ->(table, attrs) { rows[table] << attrs.merge(user_id: user.id, created_at: now, updated_at: now) }
  (0..active_weeks).each do |w|
    dose = doses[[w / info.step_weeks, doses.size - 1].min]
    if info.frequency == :weekly
      day = start + w * 7
      next if day > today
      next if rng.rand < 0.04
      add.(:dose_logs, medication_id: med.id, taken_at: zone.local(day.year, day.month, day.day, 7 + rng.rand(3), rng.rand(60)), dose_mg: dose,
                       site: info.form == :injection ? sites[w % sites.size] : nil, skipped: false, note: nil)
    else
      7.times do |d|
        day = start + w * 7 + d
        next if day > today || rng.rand < 0.06
        add.(:dose_logs, medication_id: med.id, taken_at: zone.local(day.year, day.month, day.day, 7, rng.rand(30)), dose_mg: dose, site: nil, skipped: false, note: nil)
      end
    end
  end
  med.update!(active: false) if stop_after_weeks && weeks > stop_after_weeks + 2
  weight = start_weight
  (start..today).each_with_index do |day, i|
    w = i / 7.0
    on_med = stop_after_weeks.nil? || w <= stop_after_weeks
    rate = on_med ? weekly_loss_pct * Math.exp(-w / 40.0) : -weekly_loss_pct * 0.45 * Math.exp(-(w - stop_after_weeks.to_i) / 30.0)
    weight *= (1 - rate / 100.0 / 7)
    next unless rng.rand < 0.55 || day == start || day == today
    noise = (rng.rand - 0.5) * 1.2
    # Smart-scale body fat every 4 weeks: ~78% of the loss is fat, the rest lean.
    fat_kg = start_weight * 0.41 - (start_weight - weight) * (0.74 + protein_ratio * 0.06)
    bf = (i % 28).zero? ? (fat_kg / weight * 100 + (rng.rand - 0.5) * 0.6).round(1) : nil
    add.(:weight_entries, measured_on: day, weight_kg: (weight + noise).round(1), body_fat_pct: bf, waist_cm: nil)
  end
  target = Targets.new(user, weight).protein_g # protein target at today's weight
  foods = FoodCatalog::FOODS
  (today - 20..today).each do |day|
    goal = target * protein_ratio * (0.8 + rng.rand * 0.4)
    got = 0
    while got < goal
      f = foods[rng.rand(foods.size)]
      add.(:food_entries, eaten_on: day, name: f.name, protein_g: f.protein_g, fiber_g: f.fiber_g, calories: f.calories)
      got += f.protein_g
    end
    rng.rand(3..6).times { add.(:water_entries, drank_on: day, ml: 500) }
  end
  day_keys = %w[A B C]
  k = 0
  (today - 42..today).each do |day|
    if rng.rand < strength_per_week / 7.0
      add.(:workouts, performed_on: day, kind: "strength", minutes: 35 + rng.rand(20), program_day: day_keys[k % 3])
      k += 1
    elsif rng.rand < 0.3
      add.(:workouts, performed_on: day, kind: "walk", minutes: 20 + rng.rand(30), program_day: nil)
    end
  end
  if side_effects
    [["nausea", 2], ["nausea", 1], ["constipation", 2], ["fatigue", 1], ["reflux", 1]].each_with_index do |(s, sev), j|
      add.(:side_effect_logs, logged_on: [start + 2 + j * 9, today].min, symptom: s, severity: sev, note: nil)
    end
    rows[:dose_logs].select { |dl| dl[:taken_at] >= today - 56 }.each do |dl|
      next unless rng.rand < 0.75
      d = dl[:taken_at].to_date + 1 + rng.rand(2)
      add.(:side_effect_logs, logged_on: d, symptom: %w[nausea nausea reflux constipation].sample(random: rng), severity: 1 + rng.rand(2), note: nil) if d <= today
    end
  end
  (0..7).each do |wk|
    d = today - wk * 7
    add.(:check_ins, checked_on: d, appetite: 2 + rng.rand(2), food_noise: 1 + rng.rand(2), energy: 3 + rng.rand(2), mood: 3 + rng.rand(2))
  end
  { dose_logs: DoseLog, weight_entries: WeightEntry, food_entries: FoodEntry, water_entries: WaterEntry, workouts: Workout,
    side_effect_logs: SideEffectLog, check_ins: CheckIn }.each { |k, model| model.insert_all(rows[k]) if rows[k].any? }
  med
end

unless User.exists?(email_address: "demo@kept.example")
  jordan = User.create!(email_address: "demo@kept.example", password: "kept-demo-2026", name: "Jordan Ellis", sex: "female",
                        birth_year: 1982, height_cm: 168, units: "imperial", goal_weight_kg: 77, protein_factor: 1.4,
                        meals_per_day: 3, coverage: "commercial_not_covered", plan: "plus")
  log_member(jordan, drug: "zepbound", start: today - 7 * 21, doses: [2.5, 5, 7.5, 10, 10, 10], start_weight: 103.4,
             weekly_loss_pct: 1.05, protein_ratio: 0.82, strength_per_week: 2.3, rng:, today:)
end

if employer.users.count < 60
  first = %w[Avery Blake Casey Dana Eden Finley Gray Harper Indy Jules Kai Logan Morgan Noel Oakley Parker Quinn Reese Sage Taylor]
  110.times do |i|
    drug = %w[zepbound zepbound wegovy wegovy wegovy_pill foundayo zepbound].sample(random: rng)
    info = DrugCatalog.fetch(drug)
    start = today - rng.rand(30..420)
    weeks = (today - start).to_i / 7
    steps = [weeks / info.step_weeks + 1, info.ladder.size].min
    doses = info.ladder.first(steps)
    stop = rng.rand < 0.42 ? rng.rand(6..40) : nil
    u = User.create!(email_address: "member#{i}@northwind.example", password: "seeded-#{SecureRandom.hex(12)}",
                     name: "#{first[i % first.size]} #{('A'..'Z').to_a[i % 26]}.", sex: %w[female male].sample(random: rng),
                     birth_year: rng.rand(1962..1998), height_cm: rng.rand(155..190), employer:, plan: "plus")
    log_member(u, drug:, start:, doses:, start_weight: rng.rand(85.0..130.0).round(1), weekly_loss_pct: rng.rand(0.5..1.3),
               protein_ratio: rng.rand(0.5..1.05), strength_per_week: rng.rand(0.0..3.5), rng:, today:,
               stop_after_weeks: stop, side_effects: rng.rand < 0.5)
  end
end

# No default admin password in production: the admin only exists if you set one.
admin_password = ENV["KEPT_ADMIN_PASSWORD"].presence || (Rails.env.production? ? nil : "kept-admin-2026")
if admin_password
  admin = User.find_or_initialize_by(email_address: "admin@kept.example")
  admin.update!(password: admin_password, name: "Kept Admin", admin: true) if admin.new_record?
end
puts "Seeded: #{User.count} users, #{DoseLog.count} doses, #{WeightEntry.count} weights"
