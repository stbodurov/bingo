class AddEnrollmentCodeToEmployers < ActiveRecord::Migration[8.1]
  def change
    add_column :employers, :enrollment_code, :string
    add_index :employers, :enrollment_code, unique: true
  end
end
