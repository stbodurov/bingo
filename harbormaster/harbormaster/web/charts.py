"""Server-rendered SVG charts.

Marks follow one spec everywhere: bars ≤ 24 px thick, a 2 px surface-coloured gap
between stacked segments, 4 px rounded data-ends anchored to the baseline, hairline
grid, recessive axes, text in text tokens (never series colours). Every mark carries a
`data-tip` the page's tooltip script reads, and a <title> for assistive tech. Colours
come from CSS custom properties (--s1 … --s5) so light and dark modes swap in CSS.
"""

from __future__ import annotations

import html
import math

from markupsafe import Markup

SERIES_VARS = ("var(--s1)", "var(--s2)", "var(--s3)", "var(--s4)", "var(--s5)", "var(--s6)")


def _nice_max(v: float) -> tuple[float, float]:
    if v <= 0:
        return 1.0, 0.25
    exp = 10 ** math.floor(math.log10(v))
    for m in (1, 1.5, 2, 2.5, 3, 4, 5, 6, 8, 10):
        top = m * exp
        if top >= v:
            break
    step = top / 4
    return top, step


def _money_axis(v: float) -> str:
    if v >= 1e6:
        return f"${v / 1e6:.1f}M"
    if v >= 1e3:
        return f"${v / 1e3:.0f}k"
    return f"${v:.0f}"


def _rounded_top(x: float, y: float, w: float, h: float, r: float = 4.0) -> str:
    r = min(r, w / 2, h)
    return (f"M{x:.1f},{y + h:.1f} V{y + r:.1f} Q{x:.1f},{y:.1f} {x + r:.1f},{y:.1f} "
            f"H{x + w - r:.1f} Q{x + w:.1f},{y:.1f} {x + w:.1f},{y + r:.1f} V{y + h:.1f} Z")


def _rounded_right(x: float, y: float, w: float, h: float, r: float = 4.0) -> str:
    r = min(r, h / 2, w)
    return (f"M{x:.1f},{y:.1f} H{x + w - r:.1f} Q{x + w:.1f},{y:.1f} {x + w:.1f},{y + r:.1f} "
            f"V{y + h - r:.1f} Q{x + w:.1f},{y + h:.1f} {x + w - r:.1f},{y + h:.1f} H{x:.1f} Z")


def stacked_columns(categories: list[str], labels: list[str], series: list[dict], values: list[dict[str, float]],
                    *, height: int = 330, value_fmt=_money_axis, title: str = "") -> Markup:
    """Vertical stacked columns. `series` = [{key, label}], values[i][key] = number."""
    n = max(1, len(categories))
    left, right, top, bottom = 56, 8, 12, 28
    width = 720
    plot_w, plot_h = width - left - right, height - top - bottom
    totals = [sum(v.get(s["key"], 0) for s in series) for v in values]
    vmax, step = _nice_max(max(totals or [0]))
    slot = plot_w / n
    bar_w = min(24.0, slot * 0.62)
    parts = [f'<svg class="chart" viewBox="0 0 {width} {height}" role="img" aria-label="{html.escape(title)}" '
             f'preserveAspectRatio="xMinYMin meet">']
    k = 0.0
    while k <= vmax + 1e-9:
        y = top + plot_h - plot_h * k / vmax
        parts.append(f'<line class="grid" x1="{left}" x2="{width - right}" y1="{y:.1f}" y2="{y:.1f}"/>')
        parts.append(f'<text class="axis" x="{left - 8}" y="{y + 4:.1f}" text-anchor="end">{value_fmt(k)}</text>')
        k += step
    every = max(1, math.ceil(n / 12))
    for i, cat in enumerate(categories):
        cx = left + slot * i + slot / 2
        x = cx - bar_w / 2
        y_cursor = top + plot_h
        present = [s for s in series if values[i].get(s["key"], 0) > 0]
        for j, s in enumerate(present):
            v = values[i][s["key"]]
            h = plot_h * v / vmax
            y = y_cursor - h
            color = SERIES_VARS[series.index(s) % len(SERIES_VARS)]
            tip = html.escape(f"{labels[i]} · {s['label']}: ${v:,.0f}")
            gap = 2 if j < len(present) - 1 else 0
            seg_h = max(0.0, h - gap)
            if j == len(present) - 1:
                d = _rounded_top(x, y, bar_w, max(seg_h, 0.5))
                parts.append(f'<path d="{d}" fill="{color}" data-tip="{tip}"><title>{tip}</title></path>')
            else:
                parts.append(f'<rect x="{x:.1f}" y="{y + gap:.1f}" width="{bar_w:.1f}" height="{seg_h:.1f}" '
                             f'fill="{color}" data-tip="{tip}"><title>{tip}</title></rect>')
            y_cursor = y
        # wide invisible hit target with the column total
        tip = html.escape(f"{labels[i]} · total ${totals[i]:,.0f}")
        parts.append(f'<rect class="hit" x="{cx - slot / 2:.1f}" y="{top}" width="{slot:.1f}" height="{plot_h}" '
                     f'data-tip="{tip}" data-col="{i}"/>')
        if i % every == 0:
            parts.append(f'<text class="axis" x="{cx:.1f}" y="{height - 8}" text-anchor="middle">{html.escape(cat)}</text>')
    parts.append(f'<line class="baseline" x1="{left}" x2="{width - right}" y1="{top + plot_h}" y2="{top + plot_h}"/>')
    parts.append("</svg>")
    return Markup("".join(parts))


def stacked_rows(rows: list[str], series: list[dict], values: list[dict[str, float]], *,
                 value_fmt=lambda v: f"${v:,.2f}", highlight: int | None = None, title: str = "") -> Markup:
    """Horizontal stacked bars, one row per category (e.g. landed cost by origin)."""
    n = max(1, len(rows))
    row_h, bar_h, label_w, right = 34, 20, 150, 90
    width, height = 720, n * row_h + 26
    plot_w = width - label_w - right
    totals = [sum(v.get(s["key"], 0) for s in series) for v in values]
    vmax, step = _nice_max(max(totals or [0]))
    parts = [f'<svg class="chart" viewBox="0 0 {width} {height}" role="img" aria-label="{html.escape(title)}">']
    k = 0.0
    while k <= vmax + 1e-9:
        x = label_w + plot_w * k / vmax
        parts.append(f'<line class="grid" x1="{x:.1f}" x2="{x:.1f}" y1="0" y2="{n * row_h}"/>')
        parts.append(f'<text class="axis" x="{x:.1f}" y="{height - 6}" text-anchor="middle">{value_fmt(k)}</text>')
        k += step
    for i, label in enumerate(rows):
        y = i * row_h + (row_h - bar_h) / 2
        cls = "rowlabel strong" if i == highlight else "rowlabel"
        parts.append(f'<text class="{cls}" x="{label_w - 10}" y="{y + bar_h / 2 + 4:.1f}" text-anchor="end">{html.escape(label)}</text>')
        x_cursor = label_w
        present = [s for s in series if values[i].get(s["key"], 0) > 0]
        for j, s in enumerate(present):
            v = values[i][s["key"]]
            w = plot_w * v / vmax
            color = SERIES_VARS[series.index(s) % len(SERIES_VARS)]
            tip = html.escape(f"{label} · {s['label']}: {value_fmt(v)}")
            gap = 2 if j < len(present) - 1 else 0
            seg_w = max(0.0, w - gap)
            if j == len(present) - 1:
                parts.append(f'<path d="{_rounded_right(x_cursor, y, max(seg_w, 0.5), bar_h)}" fill="{color}" data-tip="{tip}"><title>{tip}</title></path>')
            else:
                parts.append(f'<rect x="{x_cursor:.1f}" y="{y:.1f}" width="{seg_w:.1f}" height="{bar_h}" fill="{color}" data-tip="{tip}"><title>{tip}</title></rect>')
            x_cursor += w
        parts.append(f'<text class="value" x="{x_cursor + 8:.1f}" y="{y + bar_h / 2 + 4:.1f}">{value_fmt(totals[i])}</text>')
    parts.append("</svg>")
    return Markup("".join(parts))


def single_bars(items: list[tuple[str, float]], *, value_fmt=_money_axis, title: str = "") -> Markup:
    """One-series horizontal bars (no legend needed)."""
    return stacked_rows([i[0] for i in items], [{"key": "v", "label": title or "Amount"}],
                        [{"v": i[1]} for i in items], value_fmt=value_fmt, title=title)
