"""Read models for the workspace: dashboard metrics, recovery pipeline, deadlines, SKU
economics, broker scorecards. Shared by the web UI, the JSON API and the copilot's tools."""

from __future__ import annotations

from collections import defaultdict
from datetime import date, timedelta

from sqlalchemy import select
from sqlalchemy.orm import Session, selectinload

from ..db import Broker, Case, Entry, ExportRecord, Finding, Org, Sku
from ..tariff import drawback, engine, sourcing
from ..tariff.recovery import deadlines as entry_deadlines, EntryFacts

GROUPS = (  # display series for duty charts, fixed order
    ("MFN", "MFN", ("MFN",)),
    ("IEEPA", "IEEPA (refundable)", ("IEEPA_FENT", "IEEPA_RECIP", "IEEPA_BR", "IEEPA_IN")),
    ("S122", "Section 122 (contested)", ("S122",)),
    ("S301", "Section 301", ("S301_CN", "S301_FL", "S301_BR")),
    ("S232", "Section 232", ("S232_METAL", "S232_AUTO", "S232_TRUCK", "S232_WOOD", "S232_SEMI", "S232_PHARMA")),
)
GROUP_OF = {m: g for g, _, ms in GROUPS for m in ms}
ROUTE_LABELS = {
    "CAPE": "CAPE declaration", "PSC": "Post-summary correction", "PROTEST": "Protest",
    "CLASS": "CIT class / §1581(i)", "CIT": "CIT action", "PRESERVE": "Preserve rights",
    "DISCLOSURE": "Prior disclosure", "CLOSED": "Closed",
}


def entries(session: Session, org: Org) -> list[Entry]:
    return list(session.scalars(
        select(Entry).where(Entry.org_id == org.id).options(selectinload(Entry.lines), selectinload(Entry.broker))
        .order_by(Entry.entry_date.desc())))


def open_findings(session: Session, org: Org) -> list[Finding]:
    return list(session.scalars(
        select(Finding).where(Finding.org_id == org.id, Finding.status == "open")
        .options(selectinload(Finding.entry)).order_by(Finding.deadline)))


def dashboard(session: Session, org: Org) -> dict:
    today = engine.today()
    es = entries(session, org)
    fs = open_findings(session, org)
    ytd_start = date(today.year, 1, 1)
    paid_ytd = sum(e.duty_paid for e in es if e.entry_date >= ytd_start)
    value_ytd = sum(e.entered_value for e in es if e.entry_date >= ytd_start)
    paid_all = sum(e.duty_paid for e in es)
    actionable = [f for f in fs if f.kind in ("refund", "overpayment") and f.urgency != "closed" and f.route != "PRESERVE"]
    recoverable = sum(f.amount + f.interest for f in actionable)
    due30 = [f for f in actionable if f.deadline and 0 <= (f.deadline - today).days <= 30]
    exposure = sum(f.amount for f in fs if f.kind == "underpayment")
    contested = sum(f.amount for f in fs if f.kind == "contested")
    recovered = sum(c.recovered for c in session.scalars(select(Case).where(Case.org_id == org.id)))

    months: dict[str, dict[str, float]] = defaultdict(lambda: defaultdict(float))
    for e in es:
        key = e.entry_date.strftime("%Y-%m")
        for m, v in (e.paid_by_measure or {}).items():
            g = GROUP_OF.get(m)
            if g:
                months[key][g] += v
    month_keys = sorted(months)[-21:]
    series = [{"key": g, "label": label} for g, label, _ in GROUPS]
    chart = [{"month": k, "values": {g: round(months[k].get(g, 0), 2) for g, _, _ in GROUPS}} for k in month_keys]

    by_route: dict[str, dict] = {}
    for f in fs:
        if f.urgency == "closed":
            continue
        r = by_route.setdefault(f.route, {"route": f.route, "label": ROUTE_LABELS.get(f.route, f.route),
                                           "amount": 0.0, "count": 0, "next": None})
        r["amount"] += f.amount + f.interest
        r["count"] += 1
        if f.deadline and (r["next"] is None or f.deadline < r["next"]):
            r["next"] = f.deadline
    return {
        "paid_ytd": paid_ytd, "value_ytd": value_ytd, "rate_ytd": (100 * paid_ytd / value_ytd) if value_ytd else 0,
        "paid_all": paid_all, "recoverable": recoverable, "due30": due30, "due30_amount": sum(f.amount + f.interest for f in due30),
        "exposure": exposure, "contested": contested, "recovered": recovered, "series": series, "chart": chart,
        "by_route": sorted(by_route.values(), key=lambda r: -r["amount"]), "entries": len(es),
        "upcoming": [f for f in actionable if f.deadline and f.deadline >= today][:8],
    }


def deadlines(session: Session, org: Org, horizon_days: int = 365) -> list[dict]:
    today = engine.today()
    out = []
    for f in open_findings(session, org):
        if f.deadline and today <= f.deadline <= today + timedelta(days=horizon_days) and f.route != "PRESERVE":
            out.append({"date": f.deadline, "kind": ROUTE_LABELS.get(f.route, f.route), "entry": f.entry.entry_number,
                        "entry_id": f.entry_id, "amount": f.amount + f.interest, "title": f.title, "finding": f})
    # Liquidations (which open protest windows) for entries with no finding yet.
    for e in entries(session, org):
        if e.liquidation_status != "liquidated":
            dl = entry_deadlines(EntryFacts(e.entry_number, e.entry_date, e.summary_date, e.liquidation_date,
                                            e.liquidation_status))
            if dl.projected_liquidation and today <= dl.projected_liquidation <= today + timedelta(days=min(horizon_days, 60)):
                out.append({"date": dl.projected_liquidation, "kind": "Scheduled liquidation", "entry": e.entry_number,
                            "entry_id": e.id, "amount": e.duty_paid, "title": "Liquidation starts the 180-day protest clock",
                            "finding": None})
    out.sort(key=lambda d: d["date"])
    return out


def sku_economics(session: Session, org: Org) -> list[dict]:
    today = engine.today()
    rows = []
    for s in session.scalars(select(Sku).where(Sku.org_id == org.id).order_by(Sku.sku)):
        units = 1000
        sh = engine.Shipment(hts=s.hts, origin=s.origin, value=s.unit_cost * units, entry_date=today,
                             metal_share=s.metal_share, usmca=s.usmca, fta=s.usmca, quantity=units,
                             flags=frozenset(s.flags or []))
        lc = engine.landed_cost(sh, freight=engine.country(s.origin).freight_per_cbm * s.cbm_per_unit * units)
        landed = lc["per_unit"]
        margin = 100 * (s.unit_price - landed) / s.unit_price if s.unit_price else 0
        rows.append({
            "sku": s, "landed": landed, "duty_unit": lc["duty_per_unit"], "rate": lc["effective_rate"],
            "margin": round(margin, 1), "annual_duty": round(lc["duty_per_unit"] * s.annual_units, 0),
            "hold_price": sourcing.price_to_hold_margin(landed, org.target_margin),
            "measures": lc["stack"].by_measure(), "below_target": margin < org.target_margin,
        })
    return rows


def broker_scorecards(session: Session, org: Org) -> list[dict]:
    out = []
    es = entries(session, org)
    findings = list(session.scalars(select(Finding).where(Finding.org_id == org.id)))
    by_entry = defaultdict(list)
    for f in findings:
        by_entry[f.entry_id].append(f)
    for b in session.scalars(select(Broker).where(Broker.org_id == org.id)):
        mine = [e for e in es if e.broker_id == b.id]
        if not mine:
            continue
        err_entries = [e for e in mine if any(f.kind in ("overpayment", "underpayment") for f in by_entry[e.id])]
        over = sum(f.amount for e in mine for f in by_entry[e.id] if f.kind == "overpayment")
        under = sum(f.amount for e in mine for f in by_entry[e.id] if f.kind == "underpayment")
        kinds = defaultdict(float)
        for e in mine:
            for line in e.lines:
                for m in set(line.paid or {}) | set(line.correct or {}):
                    d = (line.paid or {}).get(m, 0) - (line.correct or {}).get(m, 0)
                    if abs(d) > 0.5:
                        kinds[m] += d
            for m in ("MPF", "HMF"):
                d = (e.paid_by_measure or {}).get(m, 0) - (e.correct_by_measure or {}).get(m, 0)
                if abs(d) > 0.5:
                    kinds[m] += d
        out.append({"broker": b, "entries": len(mine), "error_entries": len(err_entries),
                    "error_rate": 100 * len(err_entries) / len(mine), "overpaid": over, "underpaid": under,
                    "value": sum(e.entered_value for e in mine), "by_measure": sorted(kinds.items(), key=lambda kv: -abs(kv[1]))})
    return out


def drawback_claim(session: Session, org: Org) -> drawback.DrawbackClaim:
    today = engine.today()
    lots = []
    for e in entries(session, org):
        for line in e.lines:
            if line.sku_id is None:
                continue
            lots.append(drawback.ImportLot(e.entry_number, line.hts[:10], e.entry_date, line.quantity,
                                           {m: v for m, v in (line.paid or {}).items()}))
    exports = []
    for x in session.scalars(select(ExportRecord).where(ExportRecord.org_id == org.id).options(selectinload(ExportRecord.sku))):
        exports.append(drawback.ExportEvent(x.ref, x.sku.hts[:10], x.export_date, x.quantity, x.destination))
    return drawback.match(lots, exports, today)


def measure_totals(es: list[Entry]) -> dict[str, float]:
    out: dict[str, float] = defaultdict(float)
    for e in es:
        for m, v in (e.paid_by_measure or {}).items():
            out[m] += v
    return dict(out)
