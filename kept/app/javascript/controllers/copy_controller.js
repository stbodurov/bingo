import { Controller } from "@hotwired/stimulus"

// Copies either a source field's value or a static text value; the clicked button says "Copied".
export default class extends Controller {
  static targets = ["source", "button"]
  static values = { text: String }

  async copy(event) {
    const text = this.hasTextValue && this.textValue ? this.textValue : this.sourceTarget.value
    const button = this.hasButtonTarget ? this.buttonTarget : event.currentTarget
    try {
      await navigator.clipboard.writeText(text)
      const before = button.textContent
      button.textContent = "Copied"
      setTimeout(() => { button.textContent = before }, 1600)
    } catch {
      if (this.hasSourceTarget) this.sourceTarget.select()
    }
  }
}
