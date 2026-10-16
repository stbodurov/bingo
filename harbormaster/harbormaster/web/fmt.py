"""Template filters."""

from __future__ import annotations

import html
import re
from datetime import date, datetime

from markupsafe import Markup

from ..tariff import engine
from ..tariff.countries import name as country_name


def money(v, cents: bool = False) -> str:
    if v is None:
        return "—"
    v = float(v)
    sign = "−" if v < 0 else ""
    v = abs(v)
    if cents:
        return f"{sign}${v:,.2f}"
    return f"{sign}${v:,.0f}"


def compact(v) -> str:
    if v is None:
        return "—"
    v = float(v)
    a = abs(v)
    sign = "−" if v < 0 else ""
    if a >= 1e9:
        return f"{sign}${a / 1e9:.1f}B"
    if a >= 1e6:
        return f"{sign}${a / 1e6:.2f}M" if a < 1e7 else f"{sign}${a / 1e6:.1f}M"
    if a >= 1e4:
        return f"{sign}${a / 1e3:.0f}k"
    if a >= 1e3:
        return f"{sign}${a / 1e3:.1f}k"
    return f"{sign}${a:,.0f}"


def pct(v, digits: int = 1) -> str:
    if v is None:
        return "—"
    return f"{float(v):.{digits}f}%"


def nice_date(d) -> str:
    if d is None:
        return "—"
    if isinstance(d, str):
        d = date.fromisoformat(d)
    if isinstance(d, datetime):
        d = d.date()
    return d.strftime("%b %-d, %Y")


def short_date(d) -> str:
    if d is None:
        return "—"
    return d.strftime("%b %-d")


def days_left(d) -> int | None:
    if d is None:
        return None
    return (d - engine.today()).days


def until(d) -> str:
    n = days_left(d)
    if n is None:
        return ""
    if n < 0:
        return f"{-n} days ago"
    if n == 0:
        return "today"
    if n == 1:
        return "tomorrow"
    return f"in {n} days"


def month_label(key: str) -> str:
    y, m = key.split("-")
    return date(int(y), int(m), 1).strftime("%b %Y")


def measure_label(m: str) -> str:
    return engine.MEASURES.get(m, {}).get("label", m)


_BOLD = re.compile(r"\*\*(.+?)\*\*")


def markdown_lite(text: str) -> Markup:
    """Escape, then render **bold**, '- ' bullet lists and paragraphs."""
    out, items = [], []

    def flush():
        if items:
            out.append("<ul>" + "".join(f"<li>{i}</li>" for i in items) + "</ul>")
            items.clear()

    for raw in (text or "").splitlines():
        line = _BOLD.sub(r"<strong>\1</strong>", html.escape(raw.strip()))
        if not line:
            flush()
            continue
        if line.startswith(("- ", "• ", "* ")):
            items.append(line[2:])
        else:
            flush()
            out.append(f"<p>{line}</p>")
    flush()
    return Markup("".join(out))


def register(env) -> None:
    env.filters.update(money=money, compact=compact, pct=pct, date=nice_date, short_date=short_date,
                       until=until, month=month_label, measure=measure_label, md=markdown_lite,
                       country=country_name)
    env.globals.update(days_left=days_left, today=engine.today)
