"""Recovery cases: bundle findings into one filing and produce what the filer needs."""

from __future__ import annotations

import csv
import io
from datetime import datetime

from sqlalchemy import select
from sqlalchemy.orm import Session, selectinload

from ..db import Activity, Case, Finding, Org, User, utcnow
from .queries import ROUTE_LABELS

STATUSES = ("draft", "ready", "filed", "recovered", "rejected")


def create(session: Session, org: Org, user: User, finding_ids: list[int]) -> Case | None:
    findings = list(session.scalars(select(Finding).where(
        Finding.org_id == org.id, Finding.id.in_(finding_ids), Finding.status == "open")))
    if not findings:
        return None
    routes = {f.route for f in findings}
    route = findings[0].route if len(routes) == 1 else "MIXED"
    label = ROUTE_LABELS.get(route, "Mixed filing")
    n = len(list(session.scalars(select(Case.id).where(Case.org_id == org.id)))) + 1
    case = Case(org_id=org.id, route=route, title=f"{label} #{n} — {len(findings)} entries",
                amount=round(sum(f.amount + f.interest for f in findings), 2))
    session.add(case)
    session.flush()
    for f in findings:
        f.status = "in_case"
        f.case_id = case.id
    session.add(Activity(org_id=org.id, actor=user.name, action=f"Created {case.title} ({len(findings)} findings)."))
    session.commit()
    return case


def get(session: Session, org: Org, case_id: int) -> Case | None:
    return session.scalar(select(Case).where(Case.org_id == org.id, Case.id == case_id)
                          .options(selectinload(Case.findings).selectinload(Finding.entry)))


def set_status(session: Session, org: Org, user: User, case: Case, status: str, recovered: float | None) -> None:
    if status not in STATUSES:
        return
    case.status = status
    if status == "filed" and case.filed_at is None:
        case.filed_at = utcnow()
    if status == "recovered":
        case.recovered = case.amount if recovered is None else max(0.0, recovered)
        for f in case.findings:
            f.status = "recovered"
    if status == "rejected":
        for f in case.findings:
            f.status = "open"
            f.case_id = None
    session.add(Activity(org_id=org.id, actor=user.name, action=f"Marked {case.title} as {status}."))
    session.commit()


def cape_csv(case: Case, org: Org) -> str:
    """Entry list in the shape CAPE uploads take: one entry number per row with the
    importer reference. The refund estimate column is for the importer's own records."""
    buf = io.StringIO()
    w = csv.writer(buf)
    w.writerow(["Entry Number", "Importer of Record (last 4)", "Entry Date", "Liquidation Status",
                "IEEPA Duty Paid (estimate)", "Interest (estimate)"])
    for f in sorted(case.findings, key=lambda f: f.entry.entry_number):
        e = f.entry
        w.writerow([e.entry_number.replace("-", ""), org.ior_last4, e.entry_date.isoformat(), e.liquidation_status,
                    f"{f.amount:.2f}", f"{f.interest:.2f}"])
    return buf.getvalue()


def fee_for(case: Case, plan: str) -> float:
    """Success fee Harbormaster bills when money lands (see docs/04)."""
    rate = {"free": 0.15, "growth": 0.10, "scale": 0.07}.get(plan, 0.10)
    return round(case.recovered * rate, 2)


def billing_ledger(session: Session, org: Org) -> list[dict]:
    rows = []
    for c in session.scalars(select(Case).where(Case.org_id == org.id).order_by(Case.created_at)):
        if c.recovered:
            rows.append({"case": c, "fee": fee_for(c, org.plan), "date": c.filed_at or c.created_at or datetime.now()})
    return rows
