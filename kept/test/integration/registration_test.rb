require "test_helper"

class RegistrationTest < ActionDispatch::IntegrationTest
  test "signs up, adds a medication and lands on Today" do
    Employer.create!(name: "Acme", slug: "acme", access_code: "acme-dash", enrollment_code: "ACME2026")
    post registration_path, params: { user: { name: "Sam", email_address: "sam@example.com", password: "a-long-password", sex: "male", birth_year: 1979, units: "metric" }, enrollment_code: "acme2026" }
    assert_redirected_to new_medication_path
    user = User.find_by!(email_address: "sam@example.com")
    assert_equal "Acme", user.employer.name
    get new_medication_path
    assert_response :success
  end

  test "rejects short passwords and unknown codes" do
    post registration_path, params: { user: { name: "Sam", email_address: "sam@example.com", password: "short" } }
    assert_response :unprocessable_entity
    post registration_path, params: { user: { name: "Sam", email_address: "sam2@example.com", password: "a-long-password" }, enrollment_code: "XYZ" }
    assert_response :unprocessable_entity
    refute User.exists?(email_address: "sam2@example.com")
  end
end
