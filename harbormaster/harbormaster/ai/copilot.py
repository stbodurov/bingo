"""The copilot: questions in plain English, answered by calling the engine.

Claude plans which tools to call and writes the answer; every number in the answer
comes from a tool result. Without a model, a keyword router calls the same tools and
fills a template, so the copilot degrades to a smaller vocabulary rather than failing.
"""

from __future__ import annotations

import json
import re
from dataclasses import dataclass, field
from datetime import date
from typing import Any, Callable

from sqlalchemy.orm import Session

from ..db import Org
from ..services import queries
from ..tariff import catalog, engine, sourcing, timeline
from . import llm

MAX_TURNS = 6


def _tool(name: str, description: str, props: dict) -> dict:
    return {"name": name, "description": description, "strict": True, "input_schema": {
        "type": "object", "properties": props, "required": list(props), "additionalProperties": False}}


TOOLS = [
    _tool("recovery_summary", "Totals of money recoverable, contested, at risk and already recovered for this importer, "
          "broken down by recovery route (CAPE, PSC, protest, CIT/class).", {}),
    _tool("find_findings", "List open findings (refunds, overpayments, underpayments, contested duty) with entry, amount, "
          "route and deadline. Use empty strings for filters you do not need.", {
              "origin": {"type": "string", "description": "ISO-2 country of origin, or empty"},
              "route": {"type": "string", "description": "CAPE, PSC, PROTEST, CLASS, CIT, PRESERVE, DISCLOSURE or empty"},
              "kind": {"type": "string", "description": "refund, overpayment, underpayment, contested or empty"},
              "limit": {"type": "integer", "description": "max rows, 1-25"}}),
    _tool("upcoming_deadlines", "Filing deadlines (PSC windows, protest windows, CIT limits, liquidations) in the next N days.", {
        "days": {"type": "integer", "description": "horizon in days, 1-365"}}),
    _tool("landed_cost", "Compute the duty stack and landed cost per unit for an HTS code, origin and unit cost on a date.", {
        "hts": {"type": "string", "description": "HTS subheading, e.g. 6110.20.20"},
        "origin": {"type": "string", "description": "ISO-2 country of origin"},
        "unit_cost": {"type": "number", "description": "FOB unit cost in USD"},
        "units": {"type": "integer", "description": "units in the shipment"},
        "entry_date": {"type": "string", "description": "YYYY-MM-DD, or empty for today"}}),
    _tool("sku_margins", "Current landed cost, duty rate and gross margin for the importer's SKUs, worst margin first.", {
        "query": {"type": "string", "description": "filter by SKU or description text, or empty"},
        "limit": {"type": "integer", "description": "max rows, 1-25"}}),
    _tool("compare_origins", "Rank alternative countries of origin for one of the importer's SKUs by landed cost per unit.", {
        "sku": {"type": "string", "description": "the importer's SKU code"},
        "scenario": {"type": "string", "enum": list(sourcing.SCENARIOS), "description": "tariff scenario to layer on"}}),
    _tool("search_hts", "Search the reference tariff catalog by product words. Returns candidate subheadings.", {
        "query": {"type": "string", "description": "product description words"}}),
    _tool("tariff_changes", "Tariff measure changes in the 2025-26 timeline on or after a date, with their effect on this "
          "importer's SKUs.", {"since": {"type": "string", "description": "YYYY-MM-DD"}}),
]


@dataclass
class Workspace:
    session: Session
    org: Org
    calls: list[dict] = field(default_factory=list)

    def run(self, name: str, args: dict) -> Any:
        fn: Callable = getattr(self, "t_" + name, None)
        if fn is None:
            return {"error": f"unknown tool {name}"}
        try:
            result = fn(**args)
        except (TypeError, ValueError) as e:
            result = {"error": str(e)}
        self.calls.append({"tool": name, "args": args})
        return result

    # ── tools ─────────────────────────────────────────────────────────────────
    def t_recovery_summary(self) -> dict:
        d = queries.dashboard(self.session, self.org)
        return {"recoverable_now": round(d["recoverable"], 2), "due_in_30_days": round(d["due30_amount"], 2),
                "contested_s122": round(d["contested"], 2), "underpayment_exposure": round(d["exposure"], 2),
                "already_recovered": round(d["recovered"], 2),
                "by_route": [{"route": r["label"], "amount": round(r["amount"], 2), "findings": r["count"],
                              "next_deadline": r["next"].isoformat() if r["next"] else None} for r in d["by_route"]]}

    def t_find_findings(self, origin: str = "", route: str = "", kind: str = "", limit: int = 10) -> list[dict]:
        rows = []
        for f in queries.open_findings(self.session, self.org):
            if origin and f.entry.origin != origin.upper():
                continue
            if route and f.route != route.upper():
                continue
            if kind and f.kind != kind.lower():
                continue
            rows.append({"entry": f.entry.entry_number, "origin": f.entry.origin, "kind": f.kind, "route": f.route,
                         "title": f.title, "amount": round(f.amount + f.interest, 2),
                         "deadline": f.deadline.isoformat() if f.deadline else None, "urgency": f.urgency})
        rows.sort(key=lambda r: (r["deadline"] or "9999", -r["amount"]))
        return rows[: max(1, min(int(limit or 10), 25))]

    def t_upcoming_deadlines(self, days: int = 60) -> list[dict]:
        return [{"date": d["date"].isoformat(), "what": d["kind"], "entry": d["entry"], "amount": round(d["amount"], 2),
                 "note": d["title"]} for d in queries.deadlines(self.session, self.org, max(1, min(int(days), 365)))][:30]

    def t_landed_cost(self, hts: str, origin: str, unit_cost: float, units: int = 1000, entry_date: str = "") -> dict:
        on = date.fromisoformat(entry_date) if entry_date else engine.today()
        units = max(1, int(units or 1000))
        sh = engine.Shipment(hts=hts, origin=origin, value=float(unit_cost) * units, entry_date=on, quantity=units)
        lc = engine.landed_cost(sh, freight=engine.country(origin).freight_per_cbm * 0.01 * units)
        st = lc["stack"]
        return {"hts": hts, "origin": origin.upper(), "date": on.isoformat(), "duty_rate_pct": st.rate,
                "duty_per_unit": lc["duty_per_unit"], "landed_per_unit": lc["per_unit"],
                "layers": [{"measure": c.label, "rate_pct": c.rate, "amount": c.amount} for c in st.charges],
                "not_applied": [{"measure": x.label, "why": x.reason} for x in st.exclusions],
                "warnings": st.warnings}

    def t_sku_margins(self, query: str = "", limit: int = 10) -> list[dict]:
        rows = queries.sku_economics(self.session, self.org)
        q = (query or "").lower()
        rows = [r for r in rows if not q or q in r["sku"].sku.lower() or q in r["sku"].description.lower()]
        rows.sort(key=lambda r: r["margin"])
        return [{"sku": r["sku"].sku, "description": r["sku"].description, "origin": r["sku"].origin,
                 "duty_rate_pct": r["rate"], "landed": r["landed"], "price": r["sku"].unit_price,
                 "margin_pct": r["margin"], "price_to_hold_target": r["hold_price"],
                 "annual_duty": r["annual_duty"]} for r in rows[: max(1, min(int(limit or 10), 25))]]

    def t_compare_origins(self, sku: str, scenario: str = "none") -> dict:
        from sqlalchemy import select
        from ..db import Sku
        s = self.session.scalar(select(Sku).where(Sku.org_id == self.org.id, Sku.sku == sku.upper()))
        if s is None:
            return {"error": f"no SKU {sku}"}
        opts = sourcing.compare(s.hts, s.origin, s.unit_cost, 1000, s.cbm_per_unit, engine.today(),
                                metal_share=s.metal_share, usmca=s.usmca, scenario=scenario)
        return {"sku": s.sku, "current_origin": s.origin, "scenario": sourcing.SCENARIOS[scenario].name,
                "options": [{"origin": o.name, "landed_per_unit": o.landed_per_unit, "duty_rate_pct": o.effective_rate,
                             "lead_time_days": o.lead_time_days, "current": o.current} for o in opts[:8]]}

    def t_search_hts(self, query: str) -> list[dict]:
        return [{"hts": l.code, "description": l.description, "mfn_pct": l.mfn} for l, _ in catalog.search(query, 6)]

    def t_tariff_changes(self, since: str) -> list[dict]:
        d0 = date.fromisoformat(since)
        skus = [{"sku": r["sku"].sku, "description": r["sku"].description, "hts": r["sku"].hts, "origin": r["sku"].origin,
                 "unit_cost": r["sku"].unit_cost, "annual_units": r["sku"].annual_units,
                 "metal_share": r["sku"].metal_share, "usmca": r["sku"].usmca}
                for r in queries.sku_economics(self.session, self.org)]
        out = []
        for ev in timeline.EVENTS:
            if ev.on < d0:
                continue
            imp = timeline.impact(ev, skus)
            out.append({"date": ev.on.isoformat(), "change": ev.title, "skus_affected": len(imp),
                        "annual_duty_change": round(sum(i.annual_delta for i in imp), 2)})
        return out[:20]


def _system(ws: Workspace) -> str:
    return (
        f"You are Harbormaster's trade copilot for {ws.org.name}, a US importer. Today is {engine.today():%B %-d, %Y}.\n"
        "Answer questions about their duties, refunds, deadlines, landed costs and sourcing by calling the tools; "
        "every dollar figure, rate and date in your answer must come from a tool result in this conversation. "
        "If a tool cannot answer, say what you would need. Be brief and concrete: lead with the answer, then the "
        "two or three facts behind it, then the next action. Use short bullet lists and **bold** for key figures. "
        "Refund routes are drafted by Harbormaster and filed by a licensed customs broker or attorney; say so when "
        "you recommend filing. Do not give legal conclusions beyond what the tool results state."
    )


def answer(ws: Workspace, question: str, history: list[dict] | None = None) -> dict:
    if llm.enabled():
        out = _answer_claude(ws, question, history or [])
        if out is not None:
            return out
    return _answer_rules(ws, question)


def _answer_claude(ws: Workspace, question: str, history: list[dict]) -> dict | None:
    messages: list[dict] = [{"role": m["role"], "content": m["content"]} for m in history[-8:]]
    messages.append({"role": "user", "content": question})
    for _ in range(MAX_TURNS):
        resp = llm.create(max_tokens=6000, system=_system(ws), tools=TOOLS, messages=messages,
                          output_config={"effort": "medium"})
        if resp is None:
            return None
        if resp.stop_reason != "tool_use":
            text = llm.text_of(resp)
            return {"text": text or "I couldn't produce an answer.", "engine": "claude", "tools": ws.calls}
        messages.append({"role": "assistant", "content": resp.content})
        results = []
        for block in resp.content:
            if block.type == "tool_use":
                result = ws.run(block.name, dict(block.input))
                results.append({"type": "tool_result", "tool_use_id": block.id,
                                "content": json.dumps(result, default=str)[:20000]})
        messages.append({"role": "user", "content": results})
    return {"text": "That needed more steps than I allow in one answer — try a narrower question.",
            "engine": "claude", "tools": ws.calls}


# ── Deterministic fallback ────────────────────────────────────────────────────

def _money(v: float) -> str:
    return f"${v:,.0f}"


def _answer_rules(ws: Workspace, q: str) -> dict:
    ql = q.lower()
    lines: list[str] = []
    hts = re.search(r"\b(\d{4})\.?(\d{2})\.?(\d{2})\b", q)
    origin = next((c for c in ("CN", "VN", "IN", "MX", "CA", "BD", "TH", "ID", "KH", "TR", "PT", "IT", "DE", "JP", "KR", "TW")
                   if re.search(rf"\b{c}\b", q) or engine.country(c).name.lower() in ql), None)
    sku = re.search(r"\b(LP-[A-Z0-9-]+)\b", q.upper())

    if sku and any(w in ql for w in ("move", "source", "origin", "instead", "switch", "cheaper", "compare")):
        scen = "excess_capacity" if "excess" in ql else "none"
        r = ws.run("compare_origins", {"sku": sku.group(1), "scenario": scen})
        if "error" in r:
            lines.append(r["error"])
        else:
            lines.append(f"**{r['sku']}** landed cost by origin ({r['scenario']}):")
            for o in r["options"][:6]:
                lines.append(f"- {o['origin']}: **${o['landed_per_unit']:,.2f}**/unit, {o['duty_rate_pct']}% duty, "
                             f"{o['lead_time_days']} days{' — current' if o['current'] else ''}")
    elif hts:
        code = f"{hts.group(1)}.{hts.group(2)}.{hts.group(3)}"
        cost = re.search(r"\$\s?(\d+(?:\.\d+)?)", q)
        r = ws.run("landed_cost", {"hts": code, "origin": origin or "CN", "unit_cost": float(cost.group(1)) if cost else 10.0,
                                   "units": 1000, "entry_date": ""})
        lines.append(f"**{code} from {engine.country(r['origin']).name}**: {r['duty_rate_pct']}% total duty today, "
                     f"**${r['landed_per_unit']:,.2f}** landed per unit.")
        for layer in r["layers"]:
            lines.append(f"- {layer['measure']}: {layer['rate_pct']}%")
        for x in r["not_applied"][:3]:
            lines.append(f"- Not applied — {x['measure']}: {x['why']}")
    elif any(w in ql for w in ("deadline", "due", "expire", "clock", "urgent", "this week", "this month")) \
            and not any(w in ql for w in ("recover", "refund", "owed", "get back")):
        days = 30 if "month" in ql or "urgent" in ql else 7 if "week" in ql else 60
        rows = ws.run("upcoming_deadlines", {"days": days})
        if not rows:
            lines.append(f"Nothing is due in the next {days} days.")
        else:
            lines.append(f"**{len(rows)} deadlines** in the next {days} days:")
            for d in [r for r in rows if r["amount"] > 0][:8]:
                lines.append(f"- {d['date']}: {d['what']} — entry {d['entry']} ({_money(d['amount'])})")
    elif any(w in ql for w in ("margin", "price", "pricing", "landed", "sku", "profit")):
        rows = ws.run("sku_margins", {"query": "", "limit": 6})
        lines.append(f"Lowest-margin SKUs at today's tariff stack (target {ws.org.target_margin:.0f}%):")
        for r in rows:
            lines.append(f"- **{r['sku']}** {r['description']}: {r['margin_pct']}% margin, {r['duty_rate_pct']}% duty; "
                         f"price to hold target **${r['price_to_hold_target']:,.2f}** (now ${r['price']:,.2f})")
    elif any(w in ql for w in ("changed", "change", "new tariff", "timeline", "what happened")):
        rows = ws.run("tariff_changes", {"since": "2026-01-01"})
        lines.append("Tariff changes since January 2026 and what they did to your catalog:")
        for r in rows:
            if r["skus_affected"]:
                lines.append(f"- {r['date']} {r['change']}: {r['skus_affected']} SKUs, "
                             f"{'+' if r['annual_duty_change'] >= 0 else ''}{_money(r['annual_duty_change'])}/yr")
    else:
        r = ws.run("recovery_summary", {})
        lines.append(f"You have **{_money(r['recoverable_now'])}** recoverable now, "
                     f"**{_money(r['due_in_30_days'])}** of it with a deadline in the next 30 days.")
        for route in r["by_route"][:5]:
            nd = f", next deadline {route['next_deadline']}" if route["next_deadline"] else ""
            lines.append(f"- {route['route']}: {_money(route['amount'])} across {route['findings']} findings{nd}")
        if r["contested_s122"]:
            lines.append(f"- Section 122 (contested, preserve rights): {_money(r['contested_s122'])}")
        if r["underpayment_exposure"]:
            lines.append(f"- Underpayment exposure to correct: {_money(r['underpayment_exposure'])}")
        soon = [d for d in ws.run("upcoming_deadlines", {"days": 30}) if d["amount"] > 0 and d["what"] != "Scheduled liquidation"]
        if soon:
            lines.append("")
            lines.append(f"**Due in the next 30 days ({len(soon)}):**")
            for d in soon[:6]:
                lines.append(f"- {d['date']}: {d['what']} — entry {d['entry']} ({_money(d['amount'])})")
        lines.append("")
        lines.append("Open **Recovery** to bundle these into filings. Harbormaster drafts; your broker files.")
    return {"text": "\n".join(lines), "engine": "rules", "tools": ws.calls}
