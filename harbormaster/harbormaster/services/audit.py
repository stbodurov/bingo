"""Entry audit: recompute every line with the engine, compare with what was filed, and
turn the differences (and invalidated duties) into findings with recovery routes."""

from __future__ import annotations

from sqlalchemy import delete, select
from sqlalchemy.orm import Session

from ..db import Entry, Finding, Org, utcnow
from ..tariff import engine, recovery

# Chapter 99 heading prefixes → measure, for parsing filed duty lines.
CH99_PREFIX = (
    ("9903.01.01", "IEEPA_FENT"), ("9903.01.10", "IEEPA_FENT"), ("9903.01.2", "IEEPA_FENT"),
    ("9903.01.77", "IEEPA_BR"), ("9903.01.84", "IEEPA_IN"),
    ("9903.01.", "IEEPA_RECIP"), ("9903.02.", "IEEPA_RECIP"),
    ("9903.03.", "S122"),
    ("9903.88.", "S301_CN"), ("9903.91.", "S301_CN"),
    ("9903.80.", "S232_METAL"), ("9903.81.", "S232_METAL"), ("9903.85.", "S232_METAL"), ("9903.78.", "S232_METAL"),
    ("9903.94.", "S232_AUTO"), ("9903.74.", "S232_TRUCK"), ("9903.76.", "S232_WOOD"),
)


def measure_for_heading(heading: str) -> str | None:
    h = heading.strip()
    if h in ("9903.01.24", "9903.01.20"):
        return "IEEPA_FENT"
    for prefix, measure in CH99_PREFIX:
        if h.startswith(prefix):
            return measure
    return None


def recompute(entry: Entry) -> None:
    """Fill each line's `correct` breakdown and the entry totals from the engine."""
    total_by: dict[str, float] = {}
    paid_by: dict[str, float] = {}
    for line in entry.lines:
        sh = engine.Shipment(
            hts=line.hts, origin=line.origin, value=line.value, entry_date=entry.entry_date,
            loaded_date=entry.loaded_date, metal_share=line.metal_share,
            usmca=line.usmca, fta=line.usmca, mode=entry.mode, quantity=line.quantity or 1,
            flags=frozenset((line.sku.flags if line.sku else []) or []),
        )
        stack = engine.compute(sh)
        line.correct = stack.by_measure()
        line.explanation = (
            [{"type": "charge", "measure": c.measure, "rate": c.rate, "base": c.base, "amount": c.amount,
              "heading": c.heading, "note": c.note} for c in stack.charges]
            + [{"type": "exclusion", "measure": x.measure, "reason": x.reason} for x in stack.exclusions]
            + [{"type": "warning", "text": w} for w in stack.warnings]
        )
        for m, v in line.correct.items():
            total_by[m] = round(total_by.get(m, 0) + v, 2)
        for m, v in (line.paid or {}).items():
            paid_by[m] = round(paid_by.get(m, 0) + v, 2)
    fees = engine.entry_fees(entry.entered_value or sum(l.value for l in entry.lines), entry.mode)
    for f in fees:
        total_by[f.measure] = f.amount
    entry.correct_by_measure = total_by
    entry.paid_by_measure = {**paid_by, **{k: v for k, v in (entry.paid_by_measure or {}).items() if k in ("MPF", "HMF")}}
    entry.duty_correct = round(sum(v for k, v in total_by.items() if k not in ("MPF", "HMF")), 2)
    entry.audited_at = utcnow()


def audit(session: Session, org: Org, entry: Entry) -> list[Finding]:
    recompute(entry)
    session.execute(delete(Finding).where(Finding.entry_id == entry.id, Finding.status == "open"))
    kept = {(f.kind, f.measure) for f in session.scalars(select(Finding).where(Finding.entry_id == entry.id))}
    facts = recovery.EntryFacts(
        entry_number=entry.entry_number, entry_date=entry.entry_date, summary_date=entry.summary_date,
        liquidation_date=entry.liquidation_date, liquidation_status=entry.liquidation_status,
        paid=entry.paid_by_measure or {}, computed=entry.correct_by_measure or {}, cit_plaintiff=org.cit_plaintiff,
    )
    created = []
    for r in recovery.routes(facts):
        if (r.kind, r.measure) in kept:
            continue
        f = Finding(org_id=org.id, entry_id=entry.id, kind=r.kind, measure=r.measure, route=r.route,
                    title=r.title, detail=r.detail, authority=r.authority, amount=r.amount,
                    interest=r.interest, deadline=r.deadline, urgency=r.urgency)
        session.add(f)
        created.append(f)
    return created


def audit_all(session: Session, org: Org) -> int:
    entries = session.scalars(select(Entry).where(Entry.org_id == org.id)).all()
    n = 0
    for e in entries:
        n += len(audit(session, org, e))
    session.commit()
    return n
