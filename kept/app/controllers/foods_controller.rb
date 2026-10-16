class FoodsController < ApplicationController
  def create
    attrs = if params[:food_key].present?
      f = FoodCatalog.fetch(params[:food_key])
      { name: f.name, protein_g: f.protein_g, fiber_g: f.fiber_g, calories: f.calories }
    else
      params.require(:food_entry).permit(:name, :protein_g, :fiber_g, :calories)
    end
    entry = current_user.food_entries.new(attrs.merge(eaten_on: current_user.today))
    if entry.save
      respond_to do |format|
        format.turbo_stream { @o = overview; render :update }
        format.html { redirect_to nutrition_path, notice: "#{entry.name} added." }
      end
    else
      redirect_to nutrition_path, alert: entry.errors.full_messages.to_sentence
    end
  rescue KeyError
    redirect_to nutrition_path, alert: "Unknown food."
  end

  def destroy
    current_user.food_entries.find(params[:id]).destroy
    respond_to do |format|
      format.turbo_stream { @o = overview; render :update }
      format.html { redirect_to nutrition_path, status: :see_other }
    end
  end
end
