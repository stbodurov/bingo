import { Controller } from "@hotwired/stimulus"

// Daily protein bars against the target line; bars under target are outlined, not filled.
export default class extends Controller {
  static values = { days: Array, target: Number, unit: { type: String, default: "g" } }
  connect() { this.ro = new ResizeObserver(() => this.render()); this.ro.observe(this.element) }
  disconnect() { this.ro?.disconnect() }
  render() {
    const days = this.daysValue, T = this.targetValue
    const W = Math.max(260, this.element.clientWidth), H = 170, m = { t: 8, r: 6, b: 22, l: 34 }
    const max = Math.max(T * 1.25, ...days.map(d => d[1]))
    const slot = (W - m.l - m.r) / days.length, bw = Math.min(24, slot * 0.66)
    const y = v => m.t + (H - m.t - m.b) * (1 - v / max)
    const ns = "http://www.w3.org/2000/svg", svg = document.createElementNS(ns, "svg")
    svg.setAttribute("viewBox", `0 0 ${W} ${H}`); svg.setAttribute("width", W); svg.setAttribute("height", H)
    svg.setAttribute("role", "img"); svg.setAttribute("aria-label", `Protein per day versus target of ${T} ${this.unitValue}`)
    const el = (n, a, t) => { const e = document.createElementNS(ns, n); for (const k in a) e.setAttribute(k, a[k]); if (t != null) e.textContent = t; svg.appendChild(e); return e }
    ;[0, Math.round(max / 2), Math.round(max)].forEach(v => { el("line", { class: "gridline", x1: m.l, x2: W - m.r, y1: y(v), y2: y(v) }); el("text", { class: "axis", x: m.l - 5, y: y(v) + 4, "text-anchor": "end" }, v) })
    days.forEach(([d, v], i) => {
      const cx = m.l + slot * i + slot / 2, h = Math.max(0, H - m.b - y(v)), x0 = cx - bw / 2, top = y(v), r = Math.min(4, bw / 2, h)
      const p = el("path", { class: `colbar${v < T ? " under" : ""}`, d: h > 0 ? `M${x0},${H - m.b} V${top + r} Q${x0},${top} ${x0 + r},${top} H${x0 + bw - r} Q${x0 + bw},${top} ${x0 + bw},${top + r} V${H - m.b} Z` : "" })
      const label = new Date(d + "T12:00").toLocaleDateString("en-US", { weekday: "short", month: "short", day: "numeric" })
      el("title", {}, `${label}: ${Math.round(v)} ${this.unitValue}`); p.appendChild(svg.lastChild)
      if (days.length <= 14 && (i % 2 === 0 || days.length <= 7)) el("text", { class: "axis", x: cx, y: H - 6, "text-anchor": "middle" }, new Date(d + "T12:00").toLocaleDateString("en-US", { weekday: "narrow" }))
    })
    el("line", { class: "target", x1: m.l, x2: W - m.r, y1: y(T), y2: y(T) })
    this.element.querySelector("svg")?.remove()
    this.element.appendChild(svg)
  }
}
