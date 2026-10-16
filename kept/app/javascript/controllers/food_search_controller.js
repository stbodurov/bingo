import { Controller } from "@hotwired/stimulus"

// Filters the quick-add foods as you type. With no query only favourites show.
export default class extends Controller {
  static targets = ["input", "item", "status", "empty"]

  filter() {
    const q = this.inputTarget.value.trim().toLowerCase()
    let shown = 0
    this.itemTargets.forEach((el) => {
      const match = q ? el.dataset.name.includes(q) : el.dataset.featured === "true"
      el.hidden = !match
      if (match) shown++
    })
    this.emptyTarget.hidden = shown > 0
    this.statusTarget.textContent = q ? `${shown} ${shown === 1 ? "food" : "foods"} match “${this.inputTarget.value.trim()}”.` : "Showing favourites. Type to search everything."
  }

  clear() {
    this.inputTarget.value = ""
    this.filter()
  }
}
