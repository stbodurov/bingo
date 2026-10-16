require "test_helper"

class FoodCatalogTest < ActiveSupport::TestCase
  test "keys are unique and values are sane" do
    keys = FoodCatalog::FOODS.map(&:key)
    assert_equal keys.uniq.size, keys.size
    assert_operator FoodCatalog::FOODS.size, :>=, 100
    FoodCatalog::FOODS.each do |f|
      assert_operator f.protein_g, :<=, 60, f.name
      assert_operator f.calories, :>=, f.protein_g * 4 - 10, "#{f.name}: calories must cover protein"
    end
  end

  test "search covers names and categories, empty shows favourites" do
    assert_equal ["Shrimp"], FoodCatalog.search("shrimp").map(&:name)
    assert FoodCatalog.search("dairy").all? { |f| f.category == "Dairy" }
    assert_equal FoodCatalog.featured, FoodCatalog.search("")
    assert FoodCatalog.featured.all?(&:featured)
  end
end
