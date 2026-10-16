# This file is auto-generated from the current state of the database. Instead
# of editing this file, please use the migrations feature of Active Record to
# incrementally modify your database, and then regenerate this schema definition.
#
# This file is the source Rails uses to define your schema when running `bin/rails
# db:schema:load`. When creating a new database, `bin/rails db:schema:load` tends to
# be faster and is potentially less error prone than running all of your
# migrations from scratch. Old migrations may fail to apply correctly if those
# migrations use external dependencies or application code.
#
# It's strongly recommended that you check this file into your version control system.

ActiveRecord::Schema[8.1].define(version: 2026_10_16_201000) do
  create_table "check_ins", force: :cascade do |t|
    t.integer "user_id", null: false
    t.date "checked_on", null: false
    t.integer "appetite", null: false
    t.integer "food_noise", null: false
    t.integer "energy", null: false
    t.integer "mood", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["user_id", "checked_on"], name: "index_check_ins_on_user_id_and_checked_on", unique: true
    t.index ["user_id"], name: "index_check_ins_on_user_id"
  end

  create_table "dose_logs", force: :cascade do |t|
    t.integer "user_id", null: false
    t.integer "medication_id", null: false
    t.datetime "taken_at", null: false
    t.decimal "dose_mg", precision: 6, scale: 2, null: false
    t.string "site"
    t.boolean "skipped", default: false, null: false
    t.string "note"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["medication_id"], name: "index_dose_logs_on_medication_id"
    t.index ["user_id", "taken_at"], name: "index_dose_logs_on_user_id_and_taken_at"
    t.index ["user_id"], name: "index_dose_logs_on_user_id"
  end

  create_table "employers", force: :cascade do |t|
    t.string "name", null: false
    t.string "slug", null: false
    t.integer "covered_lives", default: 0, null: false
    t.decimal "pmpm", precision: 6, scale: 2, default: "6.0", null: false
    t.decimal "monthly_drug_cost", precision: 8, scale: 2, default: "1000.0", null: false
    t.string "access_code", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "enrollment_code"
    t.index ["access_code"], name: "index_employers_on_access_code", unique: true
    t.index ["enrollment_code"], name: "index_employers_on_enrollment_code", unique: true
    t.index ["slug"], name: "index_employers_on_slug", unique: true
  end

  create_table "food_entries", force: :cascade do |t|
    t.integer "user_id", null: false
    t.date "eaten_on", null: false
    t.string "name", null: false
    t.decimal "protein_g", precision: 5, scale: 1, default: "0.0", null: false
    t.decimal "fiber_g", precision: 5, scale: 1, default: "0.0", null: false
    t.integer "calories"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["user_id", "eaten_on"], name: "index_food_entries_on_user_id_and_eaten_on"
    t.index ["user_id"], name: "index_food_entries_on_user_id"
  end

  create_table "medications", force: :cascade do |t|
    t.integer "user_id", null: false
    t.string "drug", null: false
    t.decimal "dose_mg", precision: 6, scale: 2, null: false
    t.date "started_on", null: false
    t.date "dose_started_on", null: false
    t.integer "dose_weekday"
    t.decimal "monthly_cost", precision: 8, scale: 2
    t.boolean "active", default: true, null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["user_id"], name: "index_medications_on_user_id"
  end

  create_table "sessions", force: :cascade do |t|
    t.integer "user_id", null: false
    t.string "ip_address"
    t.string "user_agent"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["user_id"], name: "index_sessions_on_user_id"
  end

  create_table "side_effect_logs", force: :cascade do |t|
    t.integer "user_id", null: false
    t.date "logged_on", null: false
    t.string "symptom", null: false
    t.integer "severity", null: false
    t.string "note"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["user_id", "logged_on"], name: "index_side_effect_logs_on_user_id_and_logged_on"
    t.index ["user_id"], name: "index_side_effect_logs_on_user_id"
  end

  create_table "users", force: :cascade do |t|
    t.string "email_address", null: false
    t.string "password_digest", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "name", default: "", null: false
    t.string "sex", default: "female", null: false
    t.integer "birth_year"
    t.decimal "height_cm", precision: 5, scale: 1
    t.string "units", default: "imperial", null: false
    t.decimal "goal_weight_kg", precision: 5, scale: 1
    t.decimal "protein_factor", precision: 3, scale: 2, default: "1.4", null: false
    t.integer "meals_per_day", default: 3, null: false
    t.string "coverage", default: "commercial_covered", null: false
    t.string "plan", default: "free", null: false
    t.string "time_zone", default: "Eastern Time (US & Canada)", null: false
    t.integer "employer_id"
    t.boolean "admin", default: false, null: false
    t.date "maintenance_started_on"
    t.string "calendar_token"
    t.index ["calendar_token"], name: "index_users_on_calendar_token", unique: true
    t.index ["email_address"], name: "index_users_on_email_address", unique: true
    t.index ["employer_id"], name: "index_users_on_employer_id"
  end

  create_table "waitlist_signups", force: :cascade do |t|
    t.string "email", null: false
    t.string "first_name"
    t.string "role", default: "member", null: false
    t.string "drug"
    t.string "stage"
    t.string "company"
    t.string "code", null: false
    t.string "referred_by"
    t.string "source"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["code"], name: "index_waitlist_signups_on_code", unique: true
    t.index ["email"], name: "index_waitlist_signups_on_email", unique: true
  end

  create_table "water_entries", force: :cascade do |t|
    t.integer "user_id", null: false
    t.date "drank_on", null: false
    t.integer "ml", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["user_id", "drank_on"], name: "index_water_entries_on_user_id_and_drank_on"
    t.index ["user_id"], name: "index_water_entries_on_user_id"
  end

  create_table "weight_entries", force: :cascade do |t|
    t.integer "user_id", null: false
    t.date "measured_on", null: false
    t.decimal "weight_kg", precision: 5, scale: 2, null: false
    t.decimal "body_fat_pct", precision: 4, scale: 1
    t.decimal "waist_cm", precision: 5, scale: 1
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["user_id", "measured_on"], name: "index_weight_entries_on_user_id_and_measured_on", unique: true
    t.index ["user_id"], name: "index_weight_entries_on_user_id"
  end

  create_table "workouts", force: :cascade do |t|
    t.integer "user_id", null: false
    t.date "performed_on", null: false
    t.string "kind", null: false
    t.integer "minutes", null: false
    t.string "program_day"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["user_id", "performed_on"], name: "index_workouts_on_user_id_and_performed_on"
    t.index ["user_id"], name: "index_workouts_on_user_id"
  end

  add_foreign_key "check_ins", "users"
  add_foreign_key "dose_logs", "medications"
  add_foreign_key "dose_logs", "users"
  add_foreign_key "food_entries", "users"
  add_foreign_key "medications", "users"
  add_foreign_key "sessions", "users"
  add_foreign_key "side_effect_logs", "users"
  add_foreign_key "users", "employers"
  add_foreign_key "water_entries", "users"
  add_foreign_key "weight_entries", "users"
  add_foreign_key "workouts", "users"
end
