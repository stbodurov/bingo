module ApplicationHelper
  def weight(kg, digits: 1)
    return "—" if kg.nil?
    current_user&.imperial? != false ? "#{number_with_precision(kg.to_f / 0.45359237, precision: digits)} lb" : "#{number_with_precision(kg, precision: digits)} kg"
  end

  def weight_value(kg)
    return nil if kg.nil?
    current_user&.imperial? ? (kg.to_f / 0.45359237).round(1) : kg.to_f.round(1)
  end

  def weight_unit = current_user&.imperial? ? "lb" : "kg"
  def height_value(cm) = cm && (current_user&.imperial? ? (cm.to_f / 2.54).round(1) : cm.to_f)
  def height_unit = current_user&.imperial? ? "in" : "cm"
  def mg(v) = "#{v.to_f.to_s.sub(/\.0\z/, '')} mg"
  def ml_to_oz(ml) = (ml.to_f / 29.5735).round

  def water_label(ml) = current_user&.imperial? ? "#{ml_to_oz(ml)} oz" : "#{(ml / 1000.0).round(1)} L"

  # Progress ring (single measure, so no legend needed); label sits in the middle.
  def ring(value, target, label:, sub:, size: 120, color: "var(--accent)")
    pct = target.to_f.zero? ? 0 : [value.to_f / target, 1].min
    r = (size - 14) / 2.0
    c = 2 * Math::PI * r
    tag.div(class: "ring", style: "width:#{size}px;height:#{size}px", role: "img", aria: { label: "#{label}: #{sub}" }) do
      tag.svg(viewBox: "0 0 #{size} #{size}", width: size, height: size) do
        safe_join([
          tag.circle(cx: size / 2, cy: size / 2, r: r, fill: "none", stroke: "var(--track)", "stroke-width": 12),
          tag.circle(cx: size / 2, cy: size / 2, r: r, fill: "none", stroke: color, "stroke-width": 12, "stroke-linecap": "round",
                     "stroke-dasharray": "#{(c * pct).round(1)} #{c.round(1)}", transform: "rotate(-90 #{size / 2} #{size / 2})")
        ])
      end + tag.div(class: "ring-label") { tag.strong(label) + tag.span(sub) }
    end
  end

  def status_badge(status, text)
    icon = { good: "✓", warning: "!", critical: "✕", info: "i" }[status]
    tag.span(class: "badge badge-#{status}") { tag.span(icon, aria: { hidden: true }) + " " + text }
  end

  def symptom_label(key) = SideEffectCoach.label(key)
  def nav_link(label, path, icon)
    tag.a(href: path, class: "tab", aria: { current: current_page?(path) ? "page" : nil }) { tag.span(icon, class: "tab-icon", aria: { hidden: true }) + tag.span(label) }
  end

  def pct(v, digits = 0) = v.nil? ? "—" : "#{number_with_precision(v * 100, precision: digits)}%"
end
