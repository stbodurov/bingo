# Quick-add foods ranked by protein, chosen because they're easy to eat with a reduced
# appetite. Values per typical serving (USDA FoodData Central, rounded).
module FoodCatalog
  Food = Data.define(:key, :name, :serving, :protein_g, :fiber_g, :calories)

  FOODS = [
    Food.new(key: "shake", name: "Protein shake", serving: "1 bottle (11 oz)", protein_g: 30, fiber_g: 1, calories: 160),
    Food.new(key: "greek_yogurt", name: "Greek yogurt, nonfat", serving: "1 cup", protein_g: 23, fiber_g: 0, calories: 130),
    Food.new(key: "cottage", name: "Cottage cheese, low-fat", serving: "1 cup", protein_g: 28, fiber_g: 0, calories: 180),
    Food.new(key: "chicken", name: "Chicken breast", serving: "4 oz cooked", protein_g: 35, fiber_g: 0, calories: 185),
    Food.new(key: "salmon", name: "Salmon", serving: "4 oz cooked", protein_g: 25, fiber_g: 0, calories: 235),
    Food.new(key: "tuna", name: "Tuna, canned", serving: "1 can (5 oz)", protein_g: 27, fiber_g: 0, calories: 130),
    Food.new(key: "eggs", name: "Eggs", serving: "2 large", protein_g: 12, fiber_g: 0, calories: 140),
    Food.new(key: "egg_whites", name: "Egg whites", serving: "1 cup", protein_g: 26, fiber_g: 0, calories: 125),
    Food.new(key: "tofu", name: "Firm tofu", serving: "½ block (7 oz)", protein_g: 20, fiber_g: 2, calories: 180),
    Food.new(key: "lentils", name: "Lentils", serving: "1 cup cooked", protein_g: 18, fiber_g: 16, calories: 230),
    Food.new(key: "beans", name: "Black beans", serving: "1 cup cooked", protein_g: 15, fiber_g: 15, calories: 225),
    Food.new(key: "edamame", name: "Edamame", serving: "1 cup shelled", protein_g: 18, fiber_g: 8, calories: 190),
    Food.new(key: "turkey", name: "Turkey, deli", serving: "3 oz", protein_g: 18, fiber_g: 0, calories: 90),
    Food.new(key: "jerky", name: "Beef jerky", serving: "1 oz", protein_g: 9, fiber_g: 0, calories: 80),
    Food.new(key: "cheese_stick", name: "String cheese", serving: "1 stick", protein_g: 7, fiber_g: 0, calories: 80),
    Food.new(key: "bone_broth", name: "Bone broth", serving: "1 cup", protein_g: 10, fiber_g: 0, calories: 45),
    Food.new(key: "chia", name: "Chia seeds", serving: "2 tbsp", protein_g: 5, fiber_g: 10, calories: 140),
    Food.new(key: "berries", name: "Raspberries", serving: "1 cup", protein_g: 1, fiber_g: 8, calories: 65),
    Food.new(key: "oats", name: "Oatmeal", serving: "1 cup cooked", protein_g: 6, fiber_g: 4, calories: 165),
    Food.new(key: "broccoli", name: "Broccoli", serving: "1 cup cooked", protein_g: 4, fiber_g: 5, calories: 55)
  ].freeze

  def self.fetch(key) = FOODS.find { |f| f.key == key } || raise(KeyError, key)
end
