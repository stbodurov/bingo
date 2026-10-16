"""CSV import of entry-summary line data (the shape of an ACE ES-003 export or a broker's
7501 line report) and the matching sample export."""

from __future__ import annotations

import csv
import io
from dataclasses import dataclass, field
from datetime import date

from sqlalchemy import select
from sqlalchemy.orm import Session

from ..db import Broker, Entry, EntryLine, Org, Sku
from ..tariff import catalog, engine
from . import audit

MEASURE_COLS = [m for m in engine.MEASURES if m not in ("MPF", "HMF")]
COLUMNS = (["entry_number", "entry_date", "summary_date", "liquidation_status", "liquidation_date", "mode", "port",
            "origin", "loaded_date", "broker", "line_no", "sku", "hts", "description", "quantity", "value",
            "metal_share", "usmca", "mpf", "hmf", "ch99"] + [f"paid_{m.lower()}" for m in MEASURE_COLS])


@dataclass
class ImportResult:
    entries_created: int = 0
    entries_updated: int = 0
    lines: int = 0
    findings: int = 0
    errors: list[str] = field(default_factory=list)


def _date(v: str) -> date | None:
    v = (v or "").strip()
    if not v:
        return None
    for fmt in ("%Y-%m-%d", "%m/%d/%Y", "%m/%d/%y"):
        try:
            from datetime import datetime
            return datetime.strptime(v, fmt).date()
        except ValueError:
            continue
    raise ValueError(f"unrecognised date {v!r}")


def _num(v: str) -> float:
    v = (v or "").replace(",", "").replace("$", "").strip()
    return float(v) if v else 0.0


def _ch99(v: str) -> dict[str, float]:
    out: dict[str, float] = {}
    for part in (v or "").replace(",", ";").split(";"):
        if "=" not in part:
            continue
        heading, amount = part.split("=", 1)
        m = audit.measure_for_heading(heading)
        if m:
            out[m] = round(out.get(m, 0) + _num(amount), 2)
    return out


def import_csv(session: Session, org: Org, data: bytes, max_rows: int = 50_000) -> ImportResult:
    res = ImportResult()
    text = data.decode("utf-8-sig", errors="replace")
    reader = csv.DictReader(io.StringIO(text))
    if not reader.fieldnames:
        res.errors.append("The file is empty.")
        return res
    reader.fieldnames = [f.strip().lower() for f in reader.fieldnames]
    missing = {"entry_number", "entry_date", "hts", "origin", "value"} - set(reader.fieldnames)
    if missing:
        res.errors.append("Missing required columns: " + ", ".join(sorted(missing)))
        return res
    brokers = {b.filer_code: b for b in session.scalars(select(Broker).where(Broker.org_id == org.id))}
    skus = {s.sku: s for s in session.scalars(select(Sku).where(Sku.org_id == org.id))}
    touched: dict[str, Entry] = {}
    for i, row in enumerate(reader, start=2):
        if i > max_rows + 1:
            res.errors.append(f"Stopped after {max_rows} rows.")
            break
        try:
            number = row["entry_number"].strip().upper()
            hts = row["hts"].strip()
            if catalog.lookup(hts) is None:
                res.errors.append(f"Row {i}: HTS {hts} is outside the reference catalog — fees only.")
            entry = touched.get(number)
            if entry is None:
                entry = session.scalar(select(Entry).where(Entry.org_id == org.id, Entry.entry_number == number))
                if entry is None:
                    entry = Entry(org_id=org.id, entry_number=number)
                    session.add(entry)
                    res.entries_created += 1
                else:
                    entry.lines.clear()
                    res.entries_updated += 1
                touched[number] = entry
                entry.entry_date = _date(row["entry_date"])
                entry.summary_date = _date(row.get("summary_date", "")) or entry.entry_date
                entry.liquidation_status = (row.get("liquidation_status") or "unliquidated").strip().lower()
                entry.liquidation_date = _date(row.get("liquidation_date", ""))
                entry.mode = (row.get("mode") or "ocean").strip().lower()
                entry.port = (row.get("port") or "").strip()[:4]
                entry.origin = row["origin"].strip().upper()[:2]
                entry.loaded_date = _date(row.get("loaded_date", ""))
                code = (row.get("broker") or number[:3]).strip().upper()[:3]
                b = brokers.get(code)
                if b is None:
                    b = Broker(org_id=org.id, name=f"Broker {code}", filer_code=code)
                    session.add(b)
                    session.flush()
                    brokers[code] = b
                entry.broker_id = b.id
                fees = {}
                if row.get("mpf"):
                    fees["MPF"] = _num(row["mpf"])
                if row.get("hmf"):
                    fees["HMF"] = _num(row["hmf"])
                entry.paid_by_measure = fees
            paid = _ch99(row.get("ch99", ""))
            for m in MEASURE_COLS:
                v = row.get(f"paid_{m.lower()}")
                if v:
                    paid[m] = round(paid.get(m, 0) + _num(v), 2)
            sku = skus.get((row.get("sku") or "").strip().upper())
            share = row.get("metal_share", "").strip()
            line = EntryLine(
                line_no=int(_num(row.get("line_no", "")) or len(entry.lines) + 1), hts=hts,
                origin=row["origin"].strip().upper()[:2], description=(row.get("description") or "")[:300],
                quantity=_num(row.get("quantity", "")), value=_num(row["value"]),
                metal_share=float(share) if share else (sku.metal_share if sku else None),
                usmca=(row.get("usmca") or "").strip().lower() in ("1", "y", "yes", "true", "s+"),
                paid=paid, sku_id=sku.id if sku else None,
            )
            line.sku = sku
            entry.lines.append(line)
            res.lines += 1
        except (ValueError, KeyError) as e:
            res.errors.append(f"Row {i}: {e}")
    session.flush()
    for entry in touched.values():
        entry.entered_value = round(sum(l.value for l in entry.lines), 2)
        entry.duty_paid = round(sum(sum((l.paid or {}).values()) for l in entry.lines), 2)
        entry.fees_paid = round(sum((entry.paid_by_measure or {}).values()), 2)
        res.findings += len(audit.audit(session, org, entry))
    session.commit()
    return res


def sample_csv(session: Session, org: Org, limit: int = 40) -> str:
    buf = io.StringIO()
    w = csv.writer(buf)
    w.writerow(COLUMNS)
    entries = session.scalars(select(Entry).where(Entry.org_id == org.id).order_by(Entry.entry_date).limit(limit))
    for e in entries:
        for line in e.lines:
            fees = e.paid_by_measure or {}
            w.writerow([e.entry_number, e.entry_date, e.summary_date, e.liquidation_status, e.liquidation_date or "",
                        e.mode, e.port, e.origin, e.loaded_date or "", e.broker.filer_code if e.broker else "",
                        line.line_no, line.sku.sku if line.sku else "", line.hts, line.description, line.quantity,
                        line.value, "" if line.metal_share is None else line.metal_share, "Y" if line.usmca else "",
                        fees.get("MPF", "") if line.line_no == 1 else "", fees.get("HMF", "") if line.line_no == 1 else "", ""]
                       + [(line.paid or {}).get(m, "") for m in MEASURE_COLS])
    return buf.getvalue()
