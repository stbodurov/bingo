class NutritionController < ApplicationController
  def show
    @o = overview
    @foods = FoodCatalog::FOODS.sort_by { |f| [f.featured ? 0 : 1, -f.protein_g] }
    # The member's own foods from the last 30 days, most recent first, for one-tap re-logging.
    @recent = current_user.food_entries.where(eaten_on: (@o.date - 30)..@o.date).order(created_at: :desc)
      .limit(200).to_a.uniq { |e| e.name.downcase }.first(8)
  end
end
