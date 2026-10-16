"""The 2025–26 tariff timeline, and what each change did to a given catalog.

Each event is a date on which the engine's output changes. Impact is measured by
running the engine for every SKU the day before and the day of the event — so the
numbers always agree with the calculator, and adding a measure to the engine
automatically makes its impact visible here.
"""

from __future__ import annotations

from dataclasses import dataclass
from datetime import date, timedelta

from .engine import Shipment, compute

D = date


@dataclass(frozen=True)
class Event:
    on: date
    measure: str
    title: str
    summary: str
    countries: tuple[str, ...] = ()   # empty = all origins


EVENTS: tuple[Event, ...] = (
    Event(D(2025, 2, 4), "IEEPA_FENT", "10% IEEPA duty on China",
          "First border/fentanyl duty on products of China and Hong Kong.", ("CN", "HK")),
    Event(D(2025, 3, 4), "IEEPA_FENT", "Canada & Mexico 25%; China raised to 20%",
          "Border duty on Canada and Mexico; USMCA-qualifying goods exempted from Mar 7.", ("CA", "MX", "CN", "HK")),
    Event(D(2025, 3, 12), "S232_METAL", "Steel & aluminum 232 restored to 25% for all countries",
          "Country exemptions and quotas ended; derivative products charged on metal content."),
    Event(D(2025, 4, 3), "S232_AUTO", "25% on imported passenger vehicles", "Auto parts followed on May 3."),
    Event(D(2025, 4, 5), "IEEPA_RECIP", "10% reciprocal baseline on most countries",
          "Annex II goods exempt; electronics exempted on Apr 11 retroactively."),
    Event(D(2025, 4, 10), "IEEPA_RECIP", "China reciprocal rate reaches 125%",
          "Escalation from 34% to 84% to 125% in a week.", ("CN", "HK")),
    Event(D(2025, 5, 14), "IEEPA_RECIP", "China reciprocal cut to 10% (Geneva)",
          "90-day truce, later extended.", ("CN", "HK")),
    Event(D(2025, 6, 4), "S232_METAL", "Steel & aluminum 232 doubled to 50%", "UK held at 25%."),
    Event(D(2025, 8, 1), "S232_METAL", "50% on copper semi-finished products and derivatives", ""),
    Event(D(2025, 8, 7), "IEEPA_RECIP", "Country-specific reciprocal rates take effect",
          "EU/Japan 15% all-in; Vietnam 20%; India 25%; Switzerland 39%; most of SE Asia 19%."),
    Event(D(2025, 8, 18), "S232_METAL", "407 product codes added to the 232 derivative list",
          "Furniture fittings, appliances, tools and more now carry 232 on their metal content."),
    Event(D(2025, 8, 27), "IEEPA_IN", "India +25% (Russian oil)", "India all-in surcharge reaches 50%.", ("IN",)),
    Event(D(2025, 10, 14), "S232_WOOD", "232 on lumber, upholstered furniture and cabinets",
          "10% lumber; 25% upholstered wooden seating and kitchen cabinets/vanities."),
    Event(D(2025, 11, 1), "S232_TRUCK", "25% on medium and heavy trucks", ""),
    Event(D(2025, 11, 10), "IEEPA_FENT", "China border duty cut to 10%", "Part of the US–China arrangement.", ("CN", "HK")),
    Event(D(2025, 11, 13), "IEEPA_RECIP", "Coffee, bananas, tea and other foods exempted", ""),
    Event(D(2026, 1, 1), "S301_CN", "China 301 increases: batteries 25%, medical gloves 100%", "", ("CN",)),
    Event(D(2026, 2, 24), "S122", "IEEPA duties end; 10% Section 122 surcharge begins",
          "After Learning Resources v. Trump (Feb 20), IEEPA collection stopped and a 150-day s122 surcharge replaced it."),
    Event(D(2026, 4, 6), "S232_METAL", "232 metals assessed on full value",
          "50% metal products, 25% derivatives, 15% reduced class for ag equipment, residential HVAC and some machinery."),
    Event(D(2026, 4, 20), "CAPE", "CBP opens CAPE Phase 1 for IEEPA refunds",
          "Unliquidated entries and entries within 80 days of liquidation."),
    Event(D(2026, 5, 7), "S122", "CIT holds the Section 122 surcharge unlawful",
          "Relief limited to the plaintiffs; collection continued pending appeal."),
    Event(D(2026, 7, 22), "S301_BR", "25% Section 301 duty on Brazil", "", ("BR",)),
    Event(D(2026, 7, 24), "S301_FL", "Section 122 expires; forced-labor §301 begins",
          "10% or 12.5% on ~60 economies, stacking on MFN, 232 and China 301."),
    Event(D(2026, 7, 31), "S232_PHARMA", "100% 232 on patented pharmaceuticals (Annex III companies)",
          "Other companies from Sep 29; EU/Japan/Korea/Switzerland 15%."),
    Event(D(2026, 10, 6), "CAPE", "CAPE Phase 3 opens for finally liquidated entries",
          "Plaintiffs in CIT IEEPA cases only; a class was certified on Oct 8."),
    Event(D(2027, 1, 1), "S232_WOOD", "Scheduled: upholstered 30%, cabinets 50%", "Delayed from Jan 1 2026."),
)


@dataclass
class SkuImpact:
    sku: str
    description: str
    origin: str
    before_rate: float
    after_rate: float
    annual_value: float

    @property
    def delta_rate(self) -> float:
        return round(self.after_rate - self.before_rate, 2)

    @property
    def annual_delta(self) -> float:
        return round(self.annual_value * self.delta_rate / 100, 2)


def impact(event: Event, skus: list[dict]) -> list[SkuImpact]:
    """skus: dicts with sku, description, hts, origin, unit_cost, annual_units, metal_share, usmca."""
    out = []
    for s in skus:
        if event.countries and s["origin"] not in event.countries:
            continue
        value = float(s["unit_cost"]) * 1000
        kw = dict(hts=s["hts"], origin=s["origin"], value=value, metal_share=s.get("metal_share"),
                  usmca=bool(s.get("usmca")), flags=frozenset(s.get("flags", ())))
        before = compute(Shipment(entry_date=event.on - timedelta(days=1), **kw)).rate
        after = compute(Shipment(entry_date=event.on, **kw)).rate
        if abs(after - before) >= 0.01:
            out.append(SkuImpact(s["sku"], s["description"], s["origin"], before, after,
                                 float(s["unit_cost"]) * float(s.get("annual_units", 0))))
    out.sort(key=lambda i: -abs(i.annual_delta))
    return out
