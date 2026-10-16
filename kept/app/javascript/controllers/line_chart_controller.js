import { Controller } from "@hotwired/stimulus"

// Weight chart: raw readings as dots, smoothed trend as a line, optional goal line,
// crosshair + tooltip on hover/touch. Rendered at the container's real width.
export default class extends Controller {
  static values = { points: Array, goal: Number, unit: String, factor: { type: Number, default: 1 } }

  connect() {
    this.tip = document.createElement("div")
    this.tip.className = "chart-tip"
    this.tip.hidden = true
    this.element.appendChild(this.tip)
    this.ro = new ResizeObserver(() => this.render())
    this.ro.observe(this.element)
  }
  disconnect() { this.ro?.disconnect() }

  render() {
    const pts = this.pointsValue
    if (!pts.length) return
    const W = Math.max(280, this.element.clientWidth), H = Math.round(Math.min(320, Math.max(200, W * 0.45)))
    const m = { t: 10, r: 10, b: 24, l: 44 }
    const f = this.factorValue
    const vals = pts.flatMap(p => [p.t, p.w].filter(v => v != null)).map(v => v * f)
    if (this.goalValue) vals.push(this.goalValue * f)
    let lo = Math.min(...vals), hi = Math.max(...vals)
    const pad = (hi - lo) * 0.08 || 1; lo -= pad; hi += pad
    const t0 = Date.parse(pts[0].d), t1 = Date.parse(pts[pts.length - 1].d)
    const x = d => m.l + (W - m.l - m.r) * ((Date.parse(d) - t0) / Math.max(1, t1 - t0))
    const y = v => m.t + (H - m.t - m.b) * (1 - (v * f - lo) / (hi - lo))
    const ns = "http://www.w3.org/2000/svg"
    const svg = document.createElementNS(ns, "svg")
    svg.setAttribute("viewBox", `0 0 ${W} ${H}`); svg.setAttribute("width", W); svg.setAttribute("height", H)
    svg.setAttribute("role", "img"); svg.setAttribute("aria-label", "Weight readings and trend")
    const el = (name, attrs, text) => { const e = document.createElementNS(ns, name); for (const k in attrs) e.setAttribute(k, attrs[k]); if (text) e.textContent = text; svg.appendChild(e); return e }
    const step = niceStep((hi - lo) / 4)
    for (let v = Math.ceil(lo / step) * step; v <= hi; v += step) {
      const yy = m.t + (H - m.t - m.b) * (1 - (v - lo) / (hi - lo))
      el("line", { class: "gridline", x1: m.l, x2: W - m.r, y1: yy, y2: yy })
      el("text", { class: "axis", x: m.l - 6, y: yy + 4, "text-anchor": "end" }, Math.round(v))
    }
    const months = new Set()
    pts.forEach(p => { const k = p.d.slice(0, 7); if (!months.has(k) && p.d.endsWith("-01")) { months.add(k); el("text", { class: "axis", x: x(p.d), y: H - 6, "text-anchor": "middle" }, new Date(p.d + "T12:00").toLocaleString("en-US", { month: "short" })) } })
    if (this.goalValue) {
      el("line", { class: "goal", x1: m.l, x2: W - m.r, y1: y(this.goalValue), y2: y(this.goalValue) })
      el("text", { class: "axis goal-label", x: W - m.r, y: y(this.goalValue) - 6, "text-anchor": "end" }, `Goal ${(this.goalValue * f).toFixed(0)} ${this.unitValue}`)
    }
    pts.forEach(p => { if (p.w != null) el("circle", { class: "raw", cx: x(p.d), cy: y(p.w), r: 2.2 }) })
    el("path", { class: "trend", d: pts.map((p, i) => `${i ? "L" : "M"}${x(p.d).toFixed(1)},${y(p.t).toFixed(1)}`).join("") })
    const cross = el("line", { class: "cross", y1: m.t, y2: H - m.b, x1: 0, x2: 0, visibility: "hidden" })
    const dot = el("circle", { class: "hover-dot", r: 5, cx: 0, cy: 0, visibility: "hidden" })
    const move = ev => {
      const box = svg.getBoundingClientRect()
      const px = ((ev.touches ? ev.touches[0].clientX : ev.clientX) - box.left) * (W / box.width)
      let best = pts[0], bd = Infinity
      for (const p of pts) { const d = Math.abs(x(p.d) - px); if (d < bd) { bd = d; best = p } }
      cross.setAttribute("visibility", "visible"); dot.setAttribute("visibility", "visible")
      cross.setAttribute("x1", x(best.d)); cross.setAttribute("x2", x(best.d))
      dot.setAttribute("cx", x(best.d)); dot.setAttribute("cy", y(best.t))
      const unit = this.unitValue
      this.tip.hidden = false
      this.tip.textContent = `${new Date(best.d + "T12:00").toLocaleDateString("en-US", { month: "short", day: "numeric" })} · trend ${(best.t * f).toFixed(1)} ${unit}${best.w != null ? ` · scale ${(best.w * f).toFixed(1)}` : ""}`
      this.tip.style.left = `${(x(best.d) / W) * box.width}px`; this.tip.style.top = `${(y(best.t) / H) * box.height}px`
    }
    svg.addEventListener("mousemove", move); svg.addEventListener("touchmove", move, { passive: true })
    svg.addEventListener("mouseleave", () => { this.tip.hidden = true; cross.setAttribute("visibility", "hidden"); dot.setAttribute("visibility", "hidden") })
    this.element.querySelector("svg")?.remove()
    this.element.insertBefore(svg, this.tip)
  }
}

function niceStep(raw) {
  const p = Math.pow(10, Math.floor(Math.log10(raw || 1)))
  for (const k of [1, 2, 2.5, 5, 10]) if (k * p >= raw) return k * p
  return 10 * p
}
