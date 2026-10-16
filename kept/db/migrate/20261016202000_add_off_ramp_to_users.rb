class AddOffRampToUsers < ActiveRecord::Migration[8.1]
  def change
    add_column :users, :maintenance_path, :string
    add_column :users, :maintenance_from_dose_mg, :decimal, precision: 6, scale: 2
    add_column :users, :maintenance_baseline_kg, :decimal, precision: 5, scale: 1
    add_column :medications, :stopped_on, :date
  end
end
