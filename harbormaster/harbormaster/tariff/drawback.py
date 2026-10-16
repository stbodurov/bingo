"""Unused-merchandise substitution drawback (19 U.S.C. 1313(j)(2)).

Duty paid on imports can be recovered — up to 99% — when the same or commercially
interchangeable goods (same 8-digit HTS) are exported or destroyed within five years.
Most small importers that export at all never claim it; this module matches exports
to duty-paid imports and sizes the claim.

Eligible duty: MFN, Section 301 and Section 232 duties. IEEPA duties are excluded (the
executive orders barred drawback, and they are being refunded through CAPE anyway);
Section 122 drawback eligibility is unsettled and excluded until it is.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from datetime import date

DRAWBACK_ELIGIBLE = ("MFN", "S301_CN", "S301_FL", "S301_BR", "S232_METAL", "S232_AUTO",
                     "S232_TRUCK", "S232_WOOD", "S232_SEMI", "S232_PHARMA")
RATE = 0.99


@dataclass
class ImportLot:
    entry_number: str
    hts8: str
    entry_date: date
    quantity: float
    duty_by_measure: dict[str, float]

    @property
    def eligible_per_unit(self) -> float:
        if not self.quantity:
            return 0.0
        return sum(v for k, v in self.duty_by_measure.items() if k in DRAWBACK_ELIGIBLE) / self.quantity


@dataclass
class ExportEvent:
    ref: str
    hts8: str
    export_date: date
    quantity: float
    destination: str


@dataclass
class Match:
    export: ExportEvent
    lot: ImportLot
    quantity: float
    refund: float


@dataclass
class DrawbackClaim:
    matches: list[Match] = field(default_factory=list)
    unmatched_exports: list[tuple[ExportEvent, float]] = field(default_factory=list)

    @property
    def refund(self) -> float:
        return round(sum(m.refund for m in self.matches), 2)

    @property
    def units(self) -> float:
        return sum(m.quantity for m in self.matches)


def match(imports: list[ImportLot], exports: list[ExportEvent], today: date) -> DrawbackClaim:
    """Designate imports against exports, highest eligible duty first.

    Substitution lets the claimant pick which duty-paid imports to designate as long as
    they share the 8-digit HTS and were imported within five years before the export,
    so designating the highest-duty lots first maximises the claim. Lots older than five
    years *today* are skipped because the claim must be filed within five years of import.
    """
    claim = DrawbackClaim()
    remaining = {id(lot): lot.quantity for lot in imports}
    for ex in sorted(exports, key=lambda e: e.export_date):
        need = ex.quantity
        candidates = [
            lot for lot in imports
            if lot.hts8 == ex.hts8 and lot.entry_date <= ex.export_date
            and (ex.export_date - lot.entry_date).days <= 5 * 365
            and (today - lot.entry_date).days <= 5 * 365
            and remaining[id(lot)] > 0
        ]
        candidates.sort(key=lambda lot: -lot.eligible_per_unit)
        for lot in candidates:
            if need <= 0:
                break
            take = min(need, remaining[id(lot)])
            remaining[id(lot)] -= take
            need -= take
            claim.matches.append(Match(ex, lot, take, round(take * lot.eligible_per_unit * RATE, 2)))
        if need > 0:
            claim.unmatched_exports.append((ex, need))
    return claim
