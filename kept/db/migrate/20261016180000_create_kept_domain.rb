class CreateKeptDomain < ActiveRecord::Migration[8.1]
  def change
    create_table :employers do |t|
      t.string :name, null: false
      t.string :slug, null: false, index: { unique: true }
      t.integer :covered_lives, null: false, default: 0
      t.decimal :pmpm, precision: 6, scale: 2, null: false, default: 6
      t.decimal :monthly_drug_cost, precision: 8, scale: 2, null: false, default: 1000
      t.string :access_code, null: false, index: { unique: true }
      t.timestamps
    end

    change_table :users do |t|
      t.string :name, null: false, default: ""
      t.string :sex, null: false, default: "female"
      t.integer :birth_year
      t.decimal :height_cm, precision: 5, scale: 1
      t.string :units, null: false, default: "imperial"
      t.decimal :goal_weight_kg, precision: 5, scale: 1
      t.decimal :protein_factor, precision: 3, scale: 2, null: false, default: 1.4
      t.integer :meals_per_day, null: false, default: 3
      t.string :coverage, null: false, default: "commercial_covered"
      t.string :plan, null: false, default: "free"
      t.string :time_zone, null: false, default: "Eastern Time (US & Canada)"
      t.references :employer, foreign_key: true
      t.boolean :admin, null: false, default: false
      t.date :maintenance_started_on
    end

    create_table :medications do |t|
      t.references :user, null: false, foreign_key: true
      t.string :drug, null: false
      t.decimal :dose_mg, precision: 6, scale: 2, null: false
      t.date :started_on, null: false
      t.date :dose_started_on, null: false
      t.integer :dose_weekday
      t.decimal :monthly_cost, precision: 8, scale: 2
      t.boolean :active, null: false, default: true
      t.timestamps
    end

    create_table :dose_logs do |t|
      t.references :user, null: false, foreign_key: true
      t.references :medication, null: false, foreign_key: true
      t.datetime :taken_at, null: false
      t.decimal :dose_mg, precision: 6, scale: 2, null: false
      t.string :site
      t.boolean :skipped, null: false, default: false
      t.string :note
      t.timestamps
    end
    add_index :dose_logs, [:user_id, :taken_at]

    create_table :side_effect_logs do |t|
      t.references :user, null: false, foreign_key: true
      t.date :logged_on, null: false
      t.string :symptom, null: false
      t.integer :severity, null: false
      t.string :note
      t.timestamps
    end
    add_index :side_effect_logs, [:user_id, :logged_on]

    create_table :weight_entries do |t|
      t.references :user, null: false, foreign_key: true
      t.date :measured_on, null: false
      t.decimal :weight_kg, precision: 5, scale: 2, null: false
      t.decimal :body_fat_pct, precision: 4, scale: 1
      t.decimal :waist_cm, precision: 5, scale: 1
      t.timestamps
    end
    add_index :weight_entries, [:user_id, :measured_on], unique: true

    create_table :food_entries do |t|
      t.references :user, null: false, foreign_key: true
      t.date :eaten_on, null: false
      t.string :name, null: false
      t.decimal :protein_g, precision: 5, scale: 1, null: false, default: 0
      t.decimal :fiber_g, precision: 5, scale: 1, null: false, default: 0
      t.integer :calories
      t.timestamps
    end
    add_index :food_entries, [:user_id, :eaten_on]

    create_table :water_entries do |t|
      t.references :user, null: false, foreign_key: true
      t.date :drank_on, null: false
      t.integer :ml, null: false
      t.timestamps
    end
    add_index :water_entries, [:user_id, :drank_on]

    create_table :workouts do |t|
      t.references :user, null: false, foreign_key: true
      t.date :performed_on, null: false
      t.string :kind, null: false
      t.integer :minutes, null: false
      t.string :program_day
      t.timestamps
    end
    add_index :workouts, [:user_id, :performed_on]

    create_table :check_ins do |t|
      t.references :user, null: false, foreign_key: true
      t.date :checked_on, null: false
      t.integer :appetite, null: false
      t.integer :food_noise, null: false
      t.integer :energy, null: false
      t.integer :mood, null: false
      t.timestamps
    end
    add_index :check_ins, [:user_id, :checked_on], unique: true

    create_table :waitlist_signups do |t|
      t.string :email, null: false, index: { unique: true }
      t.string :first_name
      t.string :role, null: false, default: "member"
      t.string :drug
      t.string :stage
      t.string :company
      t.string :code, null: false, index: { unique: true }
      t.string :referred_by
      t.string :source
      t.timestamps
    end
  end
end
