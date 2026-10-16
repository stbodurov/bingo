"""The landing-page estimator: from five answers, size what an importer is owed and
what its duty bill looks like now. Every number is derived from the engine, and every
assumption is returned alongside the result so the page can show its work."""

from __future__ import annotations

from dataclasses import dataclass, field
from datetime import date, timedelta
from functools import cache

from ..tariff import engine

ORIGIN_GROUPS = {   # group -> representative origin(s)
    "CN": ("CN",), "VN": ("VN",), "IN": ("IN",), "MX": ("MX", "CA"), "EU": ("DE", "IT", "JP"),
    "ASIA": ("TH", "ID", "KH", "BD", "MY"), "OTHER": ("GB", "TR", "BR", "KR"),
}
GROUP_LABELS = {"CN": "China", "VN": "Vietnam", "IN": "India", "MX": "Mexico & Canada",
                "EU": "EU & Japan", "ASIA": "Other Southeast & South Asia", "OTHER": "Rest of world"}
REPRESENTATIVE = ("6110.20.20", "6912.00.48", "3926.90.99", "9403.60.80", "8516.71.00")
IEEPA_WINDOW = (date(2025, 2, 4), engine.IEEPA_END)
S122_WINDOW = (engine.S122_START, engine.S122_END)
NON_EXEMPT_SHARE = 0.85    # share of value outside Annex II / electronics exemptions (assumption)
USMCA_SHARE = 0.6          # share of Mexico/Canada imports claiming USMCA (assumption)
ERROR_RATE = (0.01, 0.03)  # filing errors as a share of non-IEEPA duty (assumption range)


@dataclass
class Estimate:
    annual_imports: float
    ieepa_paid: float
    ieepa_interest: float
    ieepa_final_share: float     # share of IEEPA money on entries already final (needs CIT/class)
    s122_paid: float
    errors_low: float
    errors_high: float
    drawback: float
    current_rate: float           # effective duty rate today, %
    current_annual_duty: float
    by_group: list[dict] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)
    assumptions: list[str] = field(default_factory=list)

    @property
    def refund_total(self) -> float:
        return round(self.ieepa_paid + self.ieepa_interest, 0)

    @property
    def recoverable_low(self) -> float:
        return round(self.refund_total + self.errors_low + self.drawback * 0.5, 0)

    @property
    def recoverable_high(self) -> float:
        return round(self.refund_total + self.errors_high + self.drawback + self.s122_paid, 0)

    def as_dict(self) -> dict:
        return {
            "annual_imports": self.annual_imports, "ieepa_paid": round(self.ieepa_paid),
            "ieepa_interest": round(self.ieepa_interest), "refund_total": self.refund_total,
            "ieepa_final_share": round(self.ieepa_final_share, 3), "s122_paid": round(self.s122_paid),
            "errors_low": round(self.errors_low), "errors_high": round(self.errors_high),
            "drawback": round(self.drawback), "current_rate": self.current_rate,
            "current_annual_duty": round(self.current_annual_duty), "recoverable_low": self.recoverable_low,
            "recoverable_high": self.recoverable_high, "by_group": self.by_group,
        }


def _daily(start: date, end: date, step: int = 3):
    d = start
    while d < end:
        yield d
        d += timedelta(days=step)


@cache
def _avg_rates(origin: str, metal: bool) -> dict[str, float]:
    """Average IEEPA / s122 / other-duty rates (percent) over each window for a
    representative basket from one origin."""
    hts_list = ("7323.93.00",) if metal else REPRESENTATIVE
    usmca = origin in ("MX", "CA")
    out = {"ieepa": 0.0, "s122": 0.0, "other_ieepa_window": 0.0, "now": 0.0}
    n = 0
    for d in _daily(*IEEPA_WINDOW):
        for h in hts_list:
            st = engine.compute(engine.Shipment(hts=h, origin=origin, value=1000, entry_date=d, metal_share=0.7))
            ieepa = sum(st.amount(m) for m in ("IEEPA_FENT", "IEEPA_RECIP", "IEEPA_BR", "IEEPA_IN"))
            if usmca:
                ieepa *= (1 - USMCA_SHARE)
            out["ieepa"] += ieepa / 10
            out["other_ieepa_window"] += (st.duty - ieepa) / 10
            n += 1
    out["ieepa"] /= n
    out["other_ieepa_window"] /= n
    m = 0
    for d in _daily(*S122_WINDOW):
        for h in hts_list:
            st = engine.compute(engine.Shipment(hts=h, origin=origin, value=1000, entry_date=d, metal_share=0.7))
            out["s122"] += st.amount("S122") / 10 * ((1 - USMCA_SHARE) if usmca else 1)
            m += 1
    out["s122"] /= max(m, 1)
    now = engine.today()
    for h in hts_list:
        st = engine.compute(engine.Shipment(hts=h, origin=origin, value=1000, entry_date=now, metal_share=0.7,
                                            usmca=usmca))
        out["now"] += st.duty / 10 / len(hts_list)
    return out


def _final_share(today: date) -> float:
    """Share of the IEEPA window whose entries are already final (liquidated + 180 days)."""
    cutoff = today - timedelta(days=9 + 314 + 180)
    start, end = IEEPA_WINDOW
    if cutoff <= start:
        return 0.0
    return min(1.0, (min(cutoff, end) - start).days / (end - start).days)


def estimate(annual_imports: float, mix: dict[str, float], metals_pct: float = 0.0,
             ior: str = "self", export_pct: float = 0.0) -> Estimate:
    annual_imports = max(0.0, float(annual_imports))
    total = sum(max(0.0, v) for v in mix.values()) or 1.0
    mix = {k: max(0.0, v) / total for k, v in mix.items() if k in ORIGIN_GROUPS}
    metals = min(max(metals_pct, 0.0), 100.0) / 100
    ieepa_years = (IEEPA_WINDOW[1] - IEEPA_WINDOW[0]).days / 365.25
    s122_years = (S122_WINDOW[1] - S122_WINDOW[0]).days / 365.25
    ieepa = s122 = other = now = 0.0
    groups = []
    for g, share in mix.items():
        if share <= 0:
            continue
        origins = ORIGIN_GROUPS[g]
        rates = {k: 0.0 for k in ("ieepa", "s122", "other_ieepa_window", "now")}
        for o in origins:
            plain, metal = _avg_rates(o, False), _avg_rates(o, True)
            for k in rates:
                rates[k] += ((1 - metals) * plain[k] + metals * metal[k]) / len(origins)
        value = annual_imports * share
        g_ieepa = value * ieepa_years * rates["ieepa"] / 100 * NON_EXEMPT_SHARE
        ieepa += g_ieepa
        s122 += value * s122_years * rates["s122"] / 100 * NON_EXEMPT_SHARE
        other += value * ieepa_years * rates["other_ieepa_window"] / 100
        now += value * rates["now"] / 100
        groups.append({"group": g, "label": GROUP_LABELS[g], "share": round(share * 100),
                       "ieepa": round(g_ieepa), "rate_now": round(rates["now"], 1)})
    today = engine.today()
    interest = ieepa * 0.07 * max(0.0, (today - date(2025, 8, 15)).days / 365.25)
    drawback = other * min(max(export_pct, 0.0), 100.0) / 100 * 0.99 / ieepa_years * min(5.0, 1.0 + ieepa_years)
    est = Estimate(
        annual_imports=annual_imports, ieepa_paid=ieepa, ieepa_interest=interest,
        ieepa_final_share=_final_share(today), s122_paid=s122,
        errors_low=other * ERROR_RATE[0], errors_high=other * ERROR_RATE[1], drawback=drawback,
        current_rate=round(100 * now / annual_imports, 1) if annual_imports else 0.0,
        current_annual_duty=now, by_group=sorted(groups, key=lambda x: -x["ieepa"]),
    )
    if ior == "supplier":
        est.warnings.append("CBP pays refunds only to the importer of record. If your supplier or forwarder "
                            "imported on a DDP basis, the refund goes to them — Harbormaster drafts the "
                            "pass-through request and reconciles it against your invoices.")
    elif ior == "mixed":
        est.warnings.append("Part of your volume arrived DDP. Refunds on those entries are paid to whoever was "
                            "importer of record; we separate the two in the audit.")
    if est.ieepa_final_share > 0:
        est.warnings.append(f"About {est.ieepa_final_share:.0%} of the IEEPA period is on entries that are already "
                            "final. CAPE currently requires CIT plaintiff status for those; the two-year limit "
                            "for a CIT action starts running from the earliest payments in Feb 2025.")
    est.assumptions = [
        f"IEEPA period {IEEPA_WINDOW[0]:%b %-d, %Y} – {IEEPA_WINDOW[1]:%b %-d, %Y}; Section 122 period "
        f"{S122_WINDOW[0]:%b %-d} – {S122_WINDOW[1] - timedelta(days=1):%b %-d, %Y}.",
        "Rates come from Harbormaster's tariff engine for a representative basket per origin "
        "(apparel, ceramics, plastics, wood furniture, small appliances; stainless kitchenware for metals).",
        f"{NON_EXEMPT_SHARE:.0%} of value assumed outside Annex II and electronics exemptions; "
        f"{USMCA_SHARE:.0%} of Mexico/Canada volume assumed USMCA-qualifying.",
        f"Filing errors: {ERROR_RATE[0]:.0%}–{ERROR_RATE[1]:.0%} of non-IEEPA duty (range, not a promise).",
        "Interest at ~7% simple from mid-2025; CBP pays statutory interest on refunds.",
        "Steady import volume across the period. Upload your ACE ES-003 report for an entry-by-entry number.",
    ]
    return est
