"""The duty-stacking engine.

Given one shipment line — what it is, where it was made, what it is worth and when it
entered — compute every duty layer that applied on that date, and record every layer
that was considered and *not* applied together with the reason. The exclusions are as
important as the charges: most overpayments we find are a layer that should have been
switched off (an exemption, an in-transit rule, a USMCA claim, a 232 non-stacking rule).

Measure definitions encode the 2025–26 tariff regime as Harbormaster's reference data.
Each carries its legal authority and current status so refund logic can reason about
it. Production deployments load measure tables from the rules service; this module is
the executable specification those tables are tested against.
"""

from __future__ import annotations

import os
from dataclasses import dataclass, field
from datetime import date

from . import catalog
from .catalog import HTSLine
from .countries import EU, country

D = date

# ── Key dates ─────────────────────────────────────────────────────────────────
IEEPA_END = D(2026, 2, 24)            # collection stopped after Learning Resources v. Trump
SCOTUS_RULING = D(2026, 2, 20)
S122_START, S122_END = D(2026, 2, 24), D(2026, 7, 24)   # end is exclusive
S122_CIT_RULING = D(2026, 5, 7)
FL301_START = D(2026, 7, 24)
BR301_START = D(2026, 7, 22)
METAL_FULL_VALUE = D(2026, 4, 6)      # 232 metals move from metal content to full value


def today() -> date:
    """The 'as of' date for statuses and deadlines (override with HARBORMASTER_TODAY)."""
    override = os.environ.get("HARBORMASTER_TODAY")
    return date.fromisoformat(override) if override else date.today()


MPF_RATE, MPF_MIN, MPF_MAX = 0.003464, 33.58, 651.50
HMF_RATE = 0.00125

INVALIDATED, CONTESTED, IN_FORCE, FEE = "invalidated", "contested", "in_force", "fee"

MEASURES: dict[str, dict] = {
    "MFN": {"label": "MFN (column 1)", "authority": "HTSUS General Notes; column 1 general rate"},
    "IEEPA_FENT": {"label": "IEEPA border/fentanyl", "authority":
                   "IEEPA — EOs 14193/14194/14195 as amended; held unlawful in Learning Resources v. Trump (S. Ct., Feb 20 2026)"},
    "IEEPA_RECIP": {"label": "IEEPA reciprocal", "authority":
                    "IEEPA — EO 14257 as amended (incl. EO 14326); held unlawful Feb 20 2026"},
    "IEEPA_BR": {"label": "IEEPA Brazil", "authority": "IEEPA — EO 14323 (Brazil, +40%); held unlawful Feb 20 2026"},
    "IEEPA_IN": {"label": "IEEPA India (Russian oil)", "authority": "IEEPA — EO 14329 (India, +25%); held unlawful Feb 20 2026"},
    "S122": {"label": "Section 122 surcharge", "authority":
             "Trade Act of 1974 §122 — Proclamation 11012; held unlawful by the CIT May 7 2026 (relief limited to plaintiffs; on appeal)"},
    "S301_CN": {"label": "Section 301 China", "authority": "Trade Act of 1974 §301 — China Lists 1–4A and 2024 four-year review increases"},
    "S301_FL": {"label": "Section 301 forced labor", "authority": "Trade Act of 1974 §301 — forced-labor action, effective Jul 24 2026"},
    "S301_BR": {"label": "Section 301 Brazil", "authority": "Trade Act of 1974 §301 — Brazil action, effective Jul 22 2026"},
    "S232_METAL": {"label": "Section 232 steel/aluminum/copper", "authority": "Trade Expansion Act of 1962 §232 — Procls. 9705/9704 as amended, 10896, 10895, 10962; Apr & Jun 2026 amendments"},
    "S232_AUTO": {"label": "Section 232 autos & parts", "authority": "Trade Expansion Act of 1962 §232 — Proclamation 10908 as amended"},
    "S232_TRUCK": {"label": "Section 232 trucks", "authority": "Trade Expansion Act of 1962 §232 — medium/heavy trucks, effective Nov 1 2025"},
    "S232_WOOD": {"label": "Section 232 timber, lumber & furniture", "authority": "Trade Expansion Act of 1962 §232 — wood products, effective Oct 14 2025; increases delayed to Jan 1 2027"},
    "S232_SEMI": {"label": "Section 232 semiconductors", "authority": "Trade Expansion Act of 1962 §232 — covered advanced chips, effective Jan 15 2026"},
    "S232_PHARMA": {"label": "Section 232 pharmaceuticals", "authority": "Trade Expansion Act of 1962 §232 — patented pharmaceuticals, Proclamation of Apr 2 2026"},
    "MPF": {"label": "Merchandise processing fee", "authority": "19 CFR 24.23 (0.3464%, FY2026 min/max)"},
    "HMF": {"label": "Harbor maintenance fee", "authority": "19 CFR 24.24 (0.125%, ocean)"},
}


def status(measure: str, as_of: date | None = None) -> str:
    """Legal status of a measure on a given day (drives refund eligibility)."""
    as_of = as_of or today()
    if measure.startswith("IEEPA"):
        return INVALIDATED if as_of >= SCOTUS_RULING else IN_FORCE
    if measure == "S122":
        return CONTESTED if as_of >= S122_CIT_RULING else IN_FORCE
    if measure in ("MPF", "HMF"):
        return FEE
    return IN_FORCE


# ── Inputs and outputs ────────────────────────────────────────────────────────

@dataclass
class Shipment:
    hts: str
    origin: str
    value: float
    entry_date: date
    loaded_date: date | None = None     # date loaded on the final vessel/conveyance
    metal_share: float | None = None    # share of value that is 232 metal (pre-Apr-2026 basis)
    us_content_share: float = 0.0       # US-origin content share (USMCA vehicles)
    usmca: bool = False                 # USMCA preference claimed and qualifying
    fta: bool = False                   # other FTA preference claimed
    flags: frozenset[str] = frozenset() # patented | annex_iii_company | advanced_chip
    mode: str = "ocean"
    quantity: float = 1.0

    @property
    def origin_code(self) -> str:
        return (self.origin or "").upper()


@dataclass
class Charge:
    measure: str
    rate: float          # percent
    base: float          # USD the rate applies to
    amount: float
    heading: str | None = None
    note: str = ""

    @property
    def label(self) -> str:
        return MEASURES[self.measure]["label"]

    @property
    def authority(self) -> str:
        return MEASURES[self.measure]["authority"]


@dataclass
class Exclusion:
    measure: str
    reason: str

    @property
    def label(self) -> str:
        return MEASURES[self.measure]["label"]


@dataclass
class Stack:
    shipment: Shipment
    line: HTSLine | None
    charges: list[Charge] = field(default_factory=list)
    exclusions: list[Exclusion] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)

    @property
    def duty(self) -> float:
        return round(sum(c.amount for c in self.charges), 2)

    @property
    def rate(self) -> float:
        v = self.shipment.value
        return round(100 * self.duty / v, 2) if v else 0.0

    def amount(self, measure: str) -> float:
        return round(sum(c.amount for c in self.charges if c.measure == measure), 2)

    def by_measure(self) -> dict[str, float]:
        out: dict[str, float] = {}
        for c in self.charges:
            out[c.measure] = round(out.get(c.measure, 0) + c.amount, 2)
        return out

    def by_status(self, as_of: date | None = None) -> dict[str, float]:
        out: dict[str, float] = {}
        for c in self.charges:
            s = status(c.measure, as_of)
            out[s] = round(out.get(s, 0) + c.amount, 2)
        return out


# ── Helpers ───────────────────────────────────────────────────────────────────

def _steps(on: date, steps: list[tuple[date, float]]) -> float | None:
    rate = None
    for since, r in steps:
        if since <= on:
            rate = r
    return rate


def _covered_by_232(line: HTSLine, on: date) -> str | None:
    """Return the 232 programme covering this line on a date, if any."""
    s = line.s232
    if s is None:
        return None
    starts = {
        "metal_primary": D(2018, 3, 23) if line.metal != "copper" else D(2025, 8, 1),
        "metal_derivative": line.derivative_since or D(2025, 3, 12),
        "auto": D(2025, 4, 3), "auto_part": D(2025, 5, 3), "truck": D(2025, 11, 1),
        "lumber": D(2025, 10, 14), "upholstered": D(2025, 10, 14), "cabinet": D(2025, 10, 14),
        "semi": D(2026, 1, 15), "pharma": D(2026, 7, 31),
    }
    return s if on >= starts[s] else None


def _metal_share(sh: Shipment, line: HTSLine, warnings: list[str]) -> float:
    if line.s232 == "metal_primary":
        return 1.0
    if sh.metal_share is None:
        warnings.append("Metal content not declared — CBP assesses 232 on the full value when content is not reported.")
        return 1.0
    return max(0.0, min(1.0, sh.metal_share))


def _inclusive(cap: float, mfn: float) -> float:
    """Rate for 'all-in' deals (EU/JP/CH-style 15% inclusive of MFN)."""
    return max(0.0, cap - mfn)


# ── Reciprocal schedule ───────────────────────────────────────────────────────

AUG7 = D(2025, 8, 7)
RECIP_AUG7 = {  # country-specific rates from Aug 7 2025 (EO 14326); 'incl' = inclusive of MFN
    "VN": 20, "TW": 20, "TH": 19, "MY": 19, "ID": 19, "PH": 19, "KH": 19, "BD": 20, "LK": 20,
    "PK": 19, "IN": 25, "BR": 10, "GB": 10, "AU": 10, "SG": 10, "TR": 15, "ZA": 30, "CO": 10,
    "KR": 15, "IL": 15,
}
INCLUSIVE_15 = set(EU) | {"JP"}


def reciprocal_rate(origin: str, on: date, mfn: float) -> tuple[float, str]:
    if origin in ("CA", "MX"):
        return 0.0, "Canada and Mexico were outside the reciprocal regime (border measures applied instead)."
    if on < D(2025, 4, 5) or on >= IEEPA_END:
        return 0.0, ""
    if origin in ("CN", "HK"):
        r = _steps(on, [(D(2025, 4, 5), 10), (D(2025, 4, 9), 84), (D(2025, 4, 10), 125), (D(2025, 5, 14), 10)])
        return float(r or 0), ""
    if on < AUG7:
        return 10.0, ""
    if origin in ("CH", "LI"):
        return (39.0, "") if on < D(2025, 11, 14) else (_inclusive(15, mfn), "Swiss framework: 15% inclusive of MFN.")
    if origin in INCLUSIVE_15:
        return _inclusive(15, mfn), "15% all-in rate inclusive of MFN."
    return float(RECIP_AUG7.get(origin, 10)), ""


# ── Main entry point ──────────────────────────────────────────────────────────

def compute(sh: Shipment) -> Stack:
    line = catalog.lookup(sh.hts)
    st = Stack(shipment=sh, line=line)
    if line is None:
        st.warnings.append(f"HTS {sh.hts} is outside the reference catalog; only fees were computed.")
        return st
    origin, on, value = sh.origin_code, sh.entry_date, float(sh.value)
    add, skip, warn = st.charges.append, st.exclusions.append, st.warnings.append
    if line.adcvd_risk:
        warn(f"AD/CVD check: {line.adcvd_risk}")

    # MFN / FTA preference
    if (sh.usmca and origin in ("CA", "MX")) or (sh.fta and country(origin).fta):
        skip(Exclusion("MFN", f"{country(origin).fta or 'FTA'} preference claimed — MFN rate is free."))
    elif line.mfn:
        add(Charge("MFN", line.mfn, value, value * line.mfn / 100))

    cov = _covered_by_232(line, on)
    share = _metal_share(sh, line, st.warnings) if (cov in ("metal_primary", "metal_derivative")) else 0.0
    # The value a "residual" surcharge (reciprocal / s122) applies to when 232 metal
    # duty is assessed on metal content only: the non-metal part of the value.
    residual = value
    metal_on_content = cov == "metal_derivative" and on < METAL_FULL_VALUE
    if metal_on_content:
        residual = value * (1 - share)

    # ── Section 232 ───────────────────────────────────────────────────────────
    s232_charged = False
    if cov in ("metal_primary", "metal_derivative"):
        rate, base, note = _metal_232(line, origin, on, value, share, cov)
        if rate:
            add(Charge("S232_METAL", rate, base, base * rate / 100,
                       {"steel": "9903.80.01", "aluminum": "9903.85.01", "copper": "9903.78.01"}.get(line.metal), note))
            s232_charged = True
        elif note:
            skip(Exclusion("S232_METAL", note))
    elif cov in ("auto", "auto_part"):
        rate, base, note = _auto_232(line, sh, origin, on, value)
        if rate:
            add(Charge("S232_AUTO", rate, base, base * rate / 100, "9903.94.01" if cov == "auto" else "9903.94.05", note))
            s232_charged = True
        else:
            skip(Exclusion("S232_AUTO", note))
    elif cov == "truck":
        base = value * (1 - sh.us_content_share) if (sh.usmca and origin in ("CA", "MX")) else value
        note = "USMCA-qualifying: duty on non-US content only." if base != value else ""
        add(Charge("S232_TRUCK", 25.0, base, base * 0.25, None, note))
        s232_charged = True
    elif cov in ("lumber", "upholstered", "cabinet"):
        rate, note = _wood_232(line, cov, origin, on)
        add(Charge("S232_WOOD", rate, value, value * rate / 100, None, note))
        s232_charged = True
    elif cov == "semi":
        if "advanced_chip" in sh.flags:
            add(Charge("S232_SEMI", 25.0, value, value * 0.25, None, "Covered advanced computing chip."))
            s232_charged = True
        else:
            skip(Exclusion("S232_SEMI", "Not a covered advanced computing chip."))
    elif line.s232 == "pharma":
        start = D(2026, 7, 31) if "annex_iii_company" in sh.flags else D(2026, 9, 29)
        if on < start:
            pass
        elif "patented" not in sh.flags:
            skip(Exclusion("S232_PHARMA", "Generic/biosimilar — not subject to the pharmaceutical 232 at this time."))
        else:
            if origin in INCLUSIVE_15 or origin in ("KR", "CH", "LI"):
                rate, note = _inclusive(15, line.mfn), "Partner rate: 15% inclusive."
            elif origin == "GB":
                rate, note = 10.0, "UK pharmaceutical arrangement (verify current rate)."
            else:
                rate, note = 100.0, "Patented pharmaceutical."
            add(Charge("S232_PHARMA", rate, value, value * rate / 100, None, note))
            s232_charged = True
    full_232 = s232_charged and not metal_on_content

    # ── IEEPA border / fentanyl ───────────────────────────────────────────────
    if on < IEEPA_END:
        if origin in ("CN", "HK"):
            r = _steps(on, [(D(2025, 2, 4), 10), (D(2025, 3, 4), 20), (D(2025, 11, 10), 10)])
            if r:
                add(Charge("IEEPA_FENT", r, value, value * r / 100, "9903.01.24"))
        elif origin in ("CA", "MX"):
            r = _steps(on, [(D(2025, 3, 4), 25)] + ([(D(2025, 8, 1), 35)] if origin == "CA" else []))
            if r:
                if sh.usmca and on >= D(2025, 3, 7):
                    skip(Exclusion("IEEPA_FENT", "USMCA-qualifying goods were exempt from 7 Mar 2025."))
                elif s232_charged:
                    skip(Exclusion("IEEPA_FENT", "Goods subject to Section 232 were not also charged the Canada/Mexico border duty (EO 14289 non-stacking)."))
                else:
                    add(Charge("IEEPA_FENT", r, value, value * r / 100, "9903.01.10" if origin == "CA" else "9903.01.01"))

    # ── IEEPA reciprocal ──────────────────────────────────────────────────────
    rate, rnote = reciprocal_rate(origin, on, line.mfn)
    if rate or rnote:
        reason = _surcharge_exempt(line, on, full_232)
        if reason:
            skip(Exclusion("IEEPA_RECIP", reason))
        elif rate == 0 and rnote:
            skip(Exclusion("IEEPA_RECIP", rnote))
        else:
            transit = _recip_transit(sh, on, origin, line.mfn)
            if transit is not None:
                t_rate, t_note = transit
                if t_rate == 0:
                    skip(Exclusion("IEEPA_RECIP", t_note))
                else:
                    add(Charge("IEEPA_RECIP", t_rate, residual, residual * t_rate / 100, "9903.01.25", t_note))
            else:
                note = rnote + (" Applied to non-metal value." if metal_on_content else "")
                add(Charge("IEEPA_RECIP", rate, residual, residual * rate / 100,
                           "9903.01.25" if rate == 10 else None, note.strip()))

    # Brazil and India IEEPA add-ons
    if origin == "BR" and D(2025, 8, 6) <= on < IEEPA_END:
        if line.recip_exempt == "ag_nov2025" and on >= D(2025, 11, 20):
            skip(Exclusion("IEEPA_BR", "Agricultural products exempted from the Brazil duty on 20 Nov 2025."))
        elif full_232:
            skip(Exclusion("IEEPA_BR", "Goods subject to Section 232 are outside the Brazil IEEPA duty."))
        else:
            add(Charge("IEEPA_BR", 40.0, residual, residual * 0.40, "9903.01.77"))
    if origin == "IN" and D(2025, 8, 27) <= on < IEEPA_END:
        reason = _surcharge_exempt(line, on, full_232)
        if reason:
            skip(Exclusion("IEEPA_IN", reason))
        else:
            add(Charge("IEEPA_IN", 25.0, residual, residual * 0.25, "9903.01.84"))

    # ── Section 122 ───────────────────────────────────────────────────────────
    if S122_START <= on < S122_END:
        reason = _surcharge_exempt(line, on, full_232)
        if not reason and sh.usmca and origin in ("CA", "MX"):
            reason = "USMCA-qualifying goods were exempt from the Section 122 surcharge."
        if reason:
            skip(Exclusion("S122", reason))
        else:
            add(Charge("S122", 10.0, residual, residual * 0.10, None,
                       "Applied to non-metal value." if metal_on_content else ""))

    # ── Section 301 ───────────────────────────────────────────────────────────
    if origin == "CN":
        step = line.cn301_rate(on)
        if step:
            add(Charge("S301_CN", step.rate, value, value * step.rate / 100, step.heading))
        else:
            skip(Exclusion("S301_CN", "Not on Section 301 Lists 1–4A (or List 4B, never imposed)."))
    if on >= FL301_START:
        tier = FL301_TIERS.get(origin)
        if tier:
            if sh.loaded_date and sh.loaded_date < FL301_START and on < D(2026, 7, 28):
                skip(Exclusion("S301_FL", "In transit: loaded before 24 Jul 2026 and entered before 28 Jul 2026."))
            else:
                note = "USMCA treatment under review — confirm against the USTR annex." if (sh.usmca and origin in ("CA", "MX")) else ""
                add(Charge("S301_FL", tier, value, value * tier / 100, None, note))
    if origin == "BR" and on >= BR301_START:
        add(Charge("S301_BR", 25.0, value, value * 0.25))

    for c in st.charges:
        c.amount = round(c.amount, 2)
        c.base = round(c.base, 2)
        c.rate = round(c.rate, 3)
    return st


# Forced-labor §301 tiers (percent). The USTR annex lists ~60 economies; tier
# assignments here are Harbormaster's reference mapping and are re-synced from the
# annex by the rules service.
FL301_TIERS: dict[str, float] = {
    **{c: 12.5 for c in ("CN", "HK", "VN", "KH", "BD", "MY", "TH", "ID", "PK", "LK", "TR", "PH")},
    **{c: 10.0 for c in ("IN", "MX", "CA", "JP", "KR", "TW", "GB", "CH", "BR", "CO", "ZA", "AU", "SG")},
    **{c: 10.0 for c in EU},
}


def _surcharge_exempt(line: HTSLine, on: date, full_232: bool) -> str | None:
    if line.recip_exempt and (line.recip_exempt_since is None or on >= line.recip_exempt_since):
        return {
            "annex_ii": "Listed in Annex II (pharma, semiconductors, energy, critical minerals, lumber/copper) — exempt.",
            "electronics": "Covered by the 11 Apr 2025 electronics exemption (retroactive to 5 Apr 2025).",
            "ag_nov2025": "Agricultural product exempted from 13 Nov 2025.",
        }[line.recip_exempt]
    if full_232:
        return "Goods subject to Section 232 are not also charged the surcharge (non-stacking)."
    return None


def _recip_transit(sh: Shipment, on: date, origin: str, mfn: float) -> tuple[float, str] | None:
    """In-transit relief. Returns (rate, note) when it changes the outcome."""
    if not sh.loaded_date:
        return None
    if sh.loaded_date < D(2025, 4, 5) and on < D(2025, 5, 28):
        return 0.0, "In transit: loaded before 5 Apr 2025 and entered before 28 May 2025 — exempt."
    if origin in ("CN", "HK"):
        return None
    if sh.loaded_date < AUG7 <= on < D(2025, 10, 5):
        return 10.0, "In transit: loaded before 7 Aug 2025, entered before 5 Oct 2025 — 10% baseline instead of country rate."
    return None


def _metal_232(line: HTSLine, origin: str, on: date, value: float, share: float, cov: str) -> tuple[float, float, str]:
    metal = line.metal
    if metal == "copper":
        if on < D(2025, 8, 1):
            return 0.0, 0.0, ""
        rate = 50.0
    else:
        base_rate = 25.0 if metal == "steel" else 10.0
        steps = [(D(2018, 3, 23), base_rate), (D(2025, 3, 12), 25.0), (D(2025, 6, 4), 50.0)]
        rate = _steps(on, steps) or 0.0
        if on < D(2025, 3, 12) and origin in ("CA", "MX", "AU"):
            return 0.0, 0.0, "Country exemption in force before 12 Mar 2025."
        if cov == "metal_derivative" and on < D(2025, 3, 12):
            return 0.0, 0.0, ""
        if origin == "GB" and on >= D(2025, 6, 4):
            rate = 25.0
    if on >= METAL_FULL_VALUE:
        if cov == "metal_primary":
            return 50.0 if origin != "GB" else 25.0, value, "Specified metal product (full value from 6 Apr 2026)."
        rate = 15.0 if line.s232_reduced else 25.0
        return rate, value, ("Reduced 15% class" if line.s232_reduced else "Derivative 25% class") + " — full value from 6 Apr 2026."
    base = value * share
    note = "" if cov == "metal_primary" else f"Assessed on declared metal content ({share:.0%} of value)."
    return rate, base, note


def _auto_232(line: HTSLine, sh: Shipment, origin: str, on: date, value: float) -> tuple[float, float, str]:
    is_part = line.s232 == "auto_part"
    if origin in ("CA", "MX") and sh.usmca:
        if is_part:
            return 0.0, 0.0, "USMCA-qualifying auto parts are exempt pending Commerce's non-US-content process."
        base = value * (1 - sh.us_content_share)
        return 25.0, base, "USMCA vehicle: duty on non-US content only."
    if origin == "JP" and on >= D(2025, 9, 16):
        return _inclusive(15, line.mfn), value, "US–Japan deal: 15% inclusive of MFN."
    if origin in EU and on >= D(2025, 8, 1):
        return _inclusive(15, line.mfn), value, "US–EU framework: 15% inclusive of MFN."
    if origin == "KR" and on >= D(2025, 11, 1):
        return _inclusive(15, line.mfn), value, "US–Korea deal: 15% inclusive of MFN."
    if origin == "GB" and on >= D(2025, 6, 30) and not is_part:
        return _inclusive(10, line.mfn), value, "UK quota rate: 10% inclusive of MFN."
    return 25.0, value, ""


def _wood_232(line: HTSLine, cov: str, origin: str, on: date) -> tuple[float, str]:
    if cov == "lumber":
        return 10.0, "Softwood timber and lumber."
    later = on >= D(2027, 1, 1)
    rate = {"upholstered": 30.0 if later else 25.0, "cabinet": 50.0 if later else 25.0}[cov]
    if origin == "GB":
        return 10.0, "UK rate for wood products."
    if origin in INCLUSIVE_15:
        return min(rate, _inclusive(15, line.mfn)), "Partner cap: 15% inclusive."
    return rate, "Increase to 30%/50% scheduled for 1 Jan 2027." if not later else ""


# ── Entry-level fees ──────────────────────────────────────────────────────────

def entry_fees(total_value: float, mode: str = "ocean") -> list[Charge]:
    mpf = min(MPF_MAX, max(MPF_MIN, total_value * MPF_RATE))
    fees = [Charge("MPF", MPF_RATE * 100, total_value, round(mpf, 2), "499", "Min $33.58 / max $651.50 per formal entry.")]
    if mode == "ocean":
        fees.append(Charge("HMF", HMF_RATE * 100, total_value, round(total_value * HMF_RATE, 2), "501"))
    return fees


def landed_cost(sh: Shipment, *, freight: float = 0.0, insurance: float = 0.0, other: float = 0.0) -> dict:
    """Duty-inclusive landed cost for a single-line shipment."""
    stack = compute(sh)
    fees = entry_fees(sh.value, sh.mode)
    fee_total = round(sum(f.amount for f in fees), 2)
    total = round(sh.value + freight + insurance + other + stack.duty + fee_total, 2)
    qty = sh.quantity or 1
    return {
        "stack": stack, "fees": fees, "fee_total": fee_total, "freight": freight,
        "insurance": insurance, "other": other, "total": total,
        "per_unit": round(total / qty, 4), "duty_per_unit": round(stack.duty / qty, 4),
        "effective_rate": stack.rate,
    }
