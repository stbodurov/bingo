from datetime import date

from harbormaster.services import estimator
from harbormaster.tariff.recovery import EntryFacts, deadlines, routes

D = date
TODAY = D(2026, 10, 16)


def facts(**kw):
    base = dict(entry_number="HLB-1234567-8", entry_date=D(2025, 9, 1), summary_date=D(2025, 9, 10))
    base.update(kw)
    return EntryFacts(**base)


def test_deadlines_follow_the_statute():
    dl = deadlines(facts())
    assert dl.projected_liquidation == D(2026, 7, 21)       # summary + 314
    assert dl.psc_close == D(2026, 6, 28)                    # entry + 300, earlier than liq - 15
    liq = deadlines(facts(liquidation_status="liquidated", liquidation_date=D(2026, 7, 21)))
    assert liq.protest_close == D(2027, 1, 17) and liq.psc_close is None


def test_unliquidated_ieepa_goes_to_cape():
    r = routes(facts(entry_date=D(2026, 1, 5), summary_date=D(2026, 1, 14), paid={"IEEPA_RECIP": 1000}), TODAY)
    assert [x.route for x in r] == ["CAPE"] and r[0].interest > 0


def test_liquidated_inside_protest_window_needs_protest():
    e = facts(liquidation_status="liquidated", liquidation_date=D(2026, 6, 1), paid={"IEEPA_FENT": 500})
    r = routes(e, TODAY)
    assert r[0].route == "PROTEST" and r[0].deadline == D(2026, 11, 28)


def test_final_entries_route_to_class_or_phase3():
    e = facts(entry_date=D(2025, 3, 1), summary_date=D(2025, 3, 10), liquidation_status="liquidated",
              liquidation_date=D(2026, 1, 18), paid={"IEEPA_RECIP": 800})
    assert routes(e, TODAY)[0].route == "CLASS"
    e.cit_plaintiff = True
    assert routes(e, TODAY)[0].route == "CAPE"


def test_s122_split_between_contested_and_error():
    e = facts(entry_date=D(2026, 3, 2), summary_date=D(2026, 3, 10), paid={"S122": 300, "MFN": 100},
              computed={"S122": 100, "MFN": 100})
    r = {(x.kind, x.route): x.amount for x in routes(e, TODAY)}
    assert r[("contested", "PRESERVE")] == 100
    assert r[("overpayment", "PSC")] == 200


def test_underpayment_suggests_psc_then_disclosure():
    e = facts(entry_date=D(2026, 6, 1), summary_date=D(2026, 6, 9), paid={"S301_CN": 100}, computed={"S301_CN": 400})
    assert routes(e, TODAY)[0].route == "PSC"
    old = facts(liquidation_status="liquidated", liquidation_date=D(2026, 7, 21), paid={"S301_CN": 100},
                computed={"S301_CN": 400})
    assert routes(old, TODAY)[0].route == "DISCLOSURE"


def test_estimator_is_monotone_and_explains_itself():
    small = estimator.estimate(1_000_000, {"CN": 100})
    big = estimator.estimate(5_000_000, {"CN": 100})
    assert big.ieepa_paid > small.ieepa_paid > 0
    assert big.recoverable_high >= big.recoverable_low
    assert big.assumptions and big.current_rate > 0
    mx = estimator.estimate(1_000_000, {"MX": 100})
    assert mx.ieepa_paid < small.ieepa_paid
    ddp = estimator.estimate(1_000_000, {"CN": 100}, ior="supplier")
    assert any("importer of record" in w for w in ddp.warnings)
