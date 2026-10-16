"""Country-of-origin scenarios: what would this SKU cost, landed, from somewhere else?

Unit cost from an alternative origin is scaled by a planning labour/overhead index
relative to the current origin; freight from per-CBM lane assumptions; duty from the
engine on the scenario date, plus any hypothetical measure the user layers on (for
example the pending §301 excess-capacity action). Everything is an estimate meant to
rank options, and the UI says so.
"""

from __future__ import annotations

from dataclasses import dataclass
from datetime import date

from .countries import COUNTRIES, country
from .engine import Shipment, compute, entry_fees

DEFAULT_ALTERNATIVES = ("CN", "VN", "IN", "MX", "BD", "KH", "TH", "ID", "TR", "PT", "KR", "TW")


@dataclass
class Scenario:
    name: str
    surcharge: dict[str, float]   # origin -> extra ad valorem percent


SCENARIOS = {
    "none": Scenario("Current law", {}),
    "excess_capacity": Scenario("§301 excess-capacity action (+15% on CN, VN, TH, MY, ID, KH)",
                                {c: 15.0 for c in ("CN", "VN", "TH", "MY", "ID", "KH")}),
    "fl301_up": Scenario("Forced-labor §301 raised to 25% for the 12.5% tier",
                         {c: 12.5 for c in ("CN", "HK", "VN", "KH", "BD", "MY", "TH", "ID", "PK", "LK", "TR", "PH")}),
    "s122_back": Scenario("A new 10% global surcharge (s122-style)", {c: 10.0 for c in COUNTRIES if c not in ("CA", "MX")}),
}


@dataclass
class Option:
    origin: str
    name: str
    unit_cost: float
    freight_per_unit: float
    duty_per_unit: float
    fees_per_unit: float
    scenario_per_unit: float
    landed_per_unit: float
    effective_rate: float
    lead_time_days: int
    measures: dict[str, float]
    current: bool = False


def compare(hts: str, current_origin: str, unit_cost: float, units: int, cbm_per_unit: float,
            on: date, *, alternatives: tuple[str, ...] = DEFAULT_ALTERNATIVES, metal_share: float | None = None,
            usmca: bool = False, scenario: str = "none") -> list[Option]:
    base = country(current_origin)
    scen = SCENARIOS.get(scenario, SCENARIOS["none"])
    origins = [current_origin.upper()] + [a for a in alternatives if a != current_origin.upper()]
    out: list[Option] = []
    for code in origins:
        c = country(code)
        cost = unit_cost * (c.labor_index / base.labor_index)
        value = cost * units
        sh = Shipment(hts=hts, origin=code, value=value, entry_date=on, metal_share=metal_share,
                      usmca=usmca and code in ("CA", "MX"), fta=True, quantity=units)
        stack = compute(sh)
        fees = sum(f.amount for f in entry_fees(value, "truck" if code in ("CA", "MX") else "ocean"))
        extra = value * scen.surcharge.get(code, 0.0) / 100
        freight = c.freight_per_cbm * cbm_per_unit
        landed = cost + freight + (stack.duty + fees + extra) / units
        out.append(Option(
            origin=code, name=c.name, unit_cost=round(cost, 2), freight_per_unit=round(freight, 2),
            duty_per_unit=round(stack.duty / units, 2), fees_per_unit=round(fees / units, 3),
            scenario_per_unit=round(extra / units, 2), landed_per_unit=round(landed, 2),
            effective_rate=round(100 * (stack.duty + extra) / value, 1) if value else 0.0,
            lead_time_days=c.lead_time_days, measures=stack.by_measure(), current=(code == current_origin.upper()),
        ))
    out.sort(key=lambda o: o.landed_per_unit)
    return out


def price_to_hold_margin(landed: float, target_margin_pct: float) -> float:
    """Selling price that keeps a gross margin (on price) at the target."""
    m = min(max(target_margin_pct, 0.0), 95.0) / 100
    return round(landed / (1 - m), 2)
