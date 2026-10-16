class NutritionController < ApplicationController
  def show
    @o = overview
    @foods = FoodCatalog::FOODS.sort_by { |f| -f.protein_g }
  end
end
