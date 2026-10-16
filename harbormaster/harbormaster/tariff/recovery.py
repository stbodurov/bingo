"""Refund routes and filing deadlines.

For every entry, decide which recovery path is open *today* for each kind of money the
importer may be owed (or may owe), and when that path closes. The rules mirror the
procedural law the money moves through:

* Post-summary correction (PSC) — unliquidated entries, within 300 days of entry and
  no later than 15 days before scheduled liquidation (19 CFR 101.9(b), CSMS guidance).
* Liquidation — by operation of law ~314 days after the entry summary unless extended.
* Protest — 180 days from liquidation (19 U.S.C. 1514(c)(3)); jurisdictional.
* CAPE — CBP's IEEPA refund process (Phase 1 from Apr 20 2026: unliquidated entries and
  entries within 80 days of liquidation; Phase 3 from Oct 6 2026: finally liquidated
  entries, plaintiffs in a CIT IEEPA case only; class certified Oct 8 2026).
* Court of International Trade — 28 U.S.C. 1581(i) actions, two years from accrual
  (28 U.S.C. 2636(i)).
* Drawback — five years from import (19 U.S.C. 1313).

Harbormaster drafts; a licensed customs broker or attorney files. Every route carries
the authority it relies on so a reviewer can check it.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from datetime import date, timedelta

from .engine import today as _today

IEEPA = ("IEEPA_FENT", "IEEPA_RECIP", "IEEPA_BR", "IEEPA_IN")
CAPE_PHASE1 = date(2026, 4, 20)
CAPE_PHASE3 = date(2026, 10, 6)
CLASS_CERTIFIED = date(2026, 10, 8)
INTEREST_RATE = 0.07   # planning estimate of the 19 U.S.C. 1505 overpayment rate


@dataclass
class EntryFacts:
    entry_number: str
    entry_date: date
    summary_date: date
    liquidation_date: date | None = None
    liquidation_status: str = "unliquidated"   # unliquidated | liquidated | extended | suspended
    paid: dict[str, float] = field(default_factory=dict)       # by measure, as filed
    computed: dict[str, float] = field(default_factory=dict)   # by measure, per engine
    cit_plaintiff: bool = False


@dataclass
class Route:
    kind: str           # refund | overpayment | underpayment | contested
    measure: str
    amount: float
    route: str          # CAPE | PSC | PROTEST | CIT | CLASS | PRESERVE | DISCLOSURE | CLOSED
    title: str
    detail: str
    authority: str
    deadline: date | None
    urgency: str        # now | soon | later | closed
    interest: float = 0.0

    @property
    def days_left(self) -> int | None:
        return None if self.deadline is None else (self.deadline - _today()).days


@dataclass
class Deadlines:
    psc_close: date | None
    projected_liquidation: date | None
    protest_close: date | None
    cit_close: date
    drawback_close: date


def deadlines(e: EntryFacts) -> Deadlines:
    liquidated = e.liquidation_status == "liquidated" and e.liquidation_date is not None
    proj = None if liquidated else (e.liquidation_date or e.summary_date + timedelta(days=314))
    psc = None
    if not liquidated:
        psc = e.entry_date + timedelta(days=300)
        if e.liquidation_status != "extended":
            psc = min(psc, proj - timedelta(days=15))
        else:
            psc = proj - timedelta(days=15)
    protest = e.liquidation_date + timedelta(days=180) if liquidated else None
    return Deadlines(
        psc_close=psc, projected_liquidation=proj, protest_close=protest,
        cit_close=_add_years(e.summary_date, 2), drawback_close=_add_years(e.entry_date, 5),
    )


def _add_years(d: date, n: int) -> date:
    try:
        return d.replace(year=d.year + n)
    except ValueError:
        return d.replace(year=d.year + n, day=28)


def _urgency(deadline: date | None, today: date) -> str:
    if deadline is None:
        return "later"
    left = (deadline - today).days
    if left < 0:
        return "closed"
    if left <= 30:
        return "now"
    if left <= 90:
        return "soon"
    return "later"


def _interest(amount: float, since: date, today: date) -> float:
    years = max(0, (today - since).days) / 365.25
    return round(amount * INTEREST_RATE * years, 2)


def routes(e: EntryFacts, today: date | None = None) -> list[Route]:
    """Every open (or just-closed) recovery path for one entry."""
    today = today or _today()
    out: list[Route] = []
    dl = deadlines(e)
    liquidated = e.liquidation_status == "liquidated" and e.liquidation_date is not None
    days_since_liq = (today - e.liquidation_date).days if liquidated else None

    # 1. IEEPA duties — invalidated; refundable with interest.
    ieepa = round(sum(e.paid.get(m, 0.0) for m in IEEPA), 2)
    if ieepa > 0:
        interest = _interest(ieepa, e.summary_date, today)
        if not liquidated or (days_since_liq is not None and days_since_liq <= 80):
            out.append(Route(
                "refund", "IEEPA", ieepa, "CAPE", "File a CAPE declaration",
                "Upload this entry in a CAPE declaration in the ACE portal. CBP removes the IEEPA "
                "Chapter 99 lines, recalculates duty and refunds with interest to the importer of record."
                + (" Liquidation is approaching — file before it posts." if not liquidated else ""),
                "CBP CAPE Phase 1 (Apr 20 2026); CIT order of Mar 4 2026",
                dl.protest_close if liquidated else dl.projected_liquidation,
                "now" if liquidated else _urgency(dl.projected_liquidation, today), interest))
        elif dl.protest_close and today <= dl.protest_close:
            out.append(Route(
                "refund", "IEEPA", ieepa, "PROTEST", "Protest the liquidation, then claim in CAPE",
                "The entry liquidated with IEEPA duty included. A protest under 19 U.S.C. 1514 keeps it "
                "from becoming final and preserves the CAPE refund.",
                "19 U.S.C. 1514(c)(3)", dl.protest_close, _urgency(dl.protest_close, today), interest))
        elif e.cit_plaintiff and today >= CAPE_PHASE3:
            out.append(Route(
                "refund", "IEEPA", ieepa, "CAPE", "Claim in CAPE Phase 3 (plaintiff entry)",
                "Finally liquidated, but you are a plaintiff in a CIT IEEPA case, which makes the entry "
                "eligible for CAPE Phase 3.", "CBP CAPE Phase 3 (Oct 6 2026)", dl.cit_close,
                _urgency(dl.cit_close, today), interest))
        elif today <= dl.cit_close:
            out.append(Route(
                "refund", "IEEPA", ieepa, "CLASS", "Join the class or file at the CIT",
                "Finally liquidated. CAPE Phase 3 currently requires CIT plaintiff status. The CIT certified "
                "an importer class on Oct 8 2026; confirm membership with counsel, or file a §1581(i) action "
                "before the two-year limit.", "28 U.S.C. 1581(i), 2636(i); CIT class order Oct 8 2026",
                dl.cit_close, _urgency(dl.cit_close, today), interest))
        else:
            out.append(Route("refund", "IEEPA", ieepa, "CLOSED", "Recovery window closed",
                             "Finally liquidated and outside the two-year CIT limit.", "28 U.S.C. 2636(i)",
                             dl.cit_close, "closed"))

    # 2. Section 122 — held unlawful for the plaintiffs only; preserve rights. Only the
    #    amount the statute would have required is "contested"; anything charged beyond
    #    that (on exempt goods) is an ordinary filing error, handled below.
    s122 = round(min(e.paid.get("S122", 0.0), e.computed.get("S122", 0.0)), 2)
    if s122 > 0:
        if liquidated and dl.protest_close and today <= dl.protest_close:
            out.append(Route(
                "contested", "S122", s122, "PROTEST", "Protest to preserve the Section 122 refund",
                "The CIT held the s122 surcharge unlawful (May 7 2026) but limited relief to the plaintiffs, "
                "and CAPE does not cover s122. Protest the liquidation so the entry stays open while the "
                "appeal runs.", "19 U.S.C. 1514; Oregon v. Trump (CIT May 7 2026)",
                dl.protest_close, _urgency(dl.protest_close, today)))
        elif not liquidated:
            out.append(Route(
                "contested", "S122", s122, "PRESERVE", "Track liquidation; protest when it posts",
                "Unliquidated. Nothing to file yet — Harbormaster opens a protest task the day this "
                "entry liquidates (180-day window).", "19 U.S.C. 1514(c)(3)",
                dl.projected_liquidation, "later"))
        elif today <= dl.cit_close:
            out.append(Route(
                "contested", "S122", s122, "CIT", "Consider a CIT action",
                "Protest window passed. A §1581(i) action within two years is the remaining path.",
                "28 U.S.C. 1581(i), 2636(i)", dl.cit_close, _urgency(dl.cit_close, today)))

    # 3. Filing errors — paid vs. what the law required.
    over, under = 0.0, 0.0
    detail_over, detail_under = [], []
    for m in sorted(set(e.paid) | set(e.computed)):
        if m in IEEPA or (m in ("MPF", "HMF") and m not in e.paid):
            continue   # fees absent from the filed data are unknown, not unpaid
        diff = round(e.paid.get(m, 0.0) - e.computed.get(m, 0.0), 2)
        if diff > 0.5:
            over += diff
            detail_over.append(f"{m} {diff:,.2f}")
        elif diff < -0.5:
            under += -diff
            detail_under.append(f"{m} {-diff:,.2f}")
    if over > 0:
        over = round(over, 2)
        if not liquidated and dl.psc_close and today <= dl.psc_close:
            out.append(Route("overpayment", "ERROR", over, "PSC", "Correct with a post-summary correction",
                             "Overpaid against the correct stack (" + ", ".join(detail_over) + "). A PSC fixes "
                             "the entry before liquidation and the difference is refunded at liquidation.",
                             "19 CFR 101.9(b); CSMS #43528998", dl.psc_close, _urgency(dl.psc_close, today)))
        elif liquidated and dl.protest_close and today <= dl.protest_close:
            out.append(Route("overpayment", "ERROR", over, "PROTEST", "Protest the overpayment",
                             "Overpaid (" + ", ".join(detail_over) + "). The entry has liquidated; a protest "
                             "is the remaining administrative remedy.", "19 U.S.C. 1514",
                             dl.protest_close, _urgency(dl.protest_close, today)))
        elif not liquidated:
            out.append(Route("overpayment", "ERROR", over, "PROTEST", "Protest after liquidation",
                             "The PSC window has closed. Protest within 180 days once the entry liquidates.",
                             "19 U.S.C. 1514", dl.projected_liquidation, "later"))
        else:
            out.append(Route("overpayment", "ERROR", over, "CLOSED", "Overpayment no longer recoverable",
                             "Liquidated and past the protest window.", "19 U.S.C. 1514", dl.protest_close, "closed"))
    if under > 0:
        under = round(under, 2)
        if not liquidated and dl.psc_close and today <= dl.psc_close:
            out.append(Route("underpayment", "ERROR", under, "PSC", "Pay the difference with a PSC",
                             "Underpaid (" + ", ".join(detail_under) + "). Correcting now avoids a penalty case.",
                             "19 CFR 101.9(b)", dl.psc_close, _urgency(dl.psc_close, today)))
        else:
            out.append(Route("underpayment", "ERROR", under, "DISCLOSURE", "Consider a prior disclosure",
                             "Underpaid (" + ", ".join(detail_under) + "). A prior disclosure before CBP opens an "
                             "inquiry caps penalties at interest for negligence.", "19 U.S.C. 1592(c)(4)",
                             None, "soon"))
    return out
