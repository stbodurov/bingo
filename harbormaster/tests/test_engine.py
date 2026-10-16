from datetime import date

import pytest

from harbormaster.tariff import catalog, drawback, engine, sourcing, timeline
from harbormaster.tariff.engine import Shipment, compute

D = date


def stack(**kw):
    kw.setdefault("value", 10_000)
    return compute(Shipment(**kw))


def test_china_peak_april_2025_stacks_four_layers():
    st = stack(hts="6110.20.20", origin="CN", entry_date=D(2025, 4, 12))
    assert st.by_measure() == {"MFN": 1650.0, "IEEPA_FENT": 2000.0, "IEEPA_RECIP": 12500.0, "S301_CN": 750.0}
    assert st.rate == 169.0


def test_ieepa_ends_and_s122_then_forced_labor_replace_it():
    feb = stack(hts="6110.20.20", origin="CN", entry_date=D(2026, 2, 23))
    assert feb.amount("IEEPA_RECIP") and feb.amount("IEEPA_FENT")
    mar = stack(hts="6110.20.20", origin="CN", entry_date=D(2026, 3, 2))
    assert not mar.amount("IEEPA_RECIP") and mar.amount("S122") == 1000.0
    aug = stack(hts="6110.20.20", origin="CN", entry_date=D(2026, 8, 3))
    assert not aug.amount("S122") and aug.amount("S301_FL") == 1250.0


def test_eu_all_in_fifteen_percent_is_inclusive_of_mfn():
    st = stack(hts="6912.00.48", origin="IT", entry_date=D(2025, 9, 15))
    assert st.rate == pytest.approx(15.0)


def test_electronics_exemption_switches_reciprocal_off_but_not_fentanyl():
    st = stack(hts="8471.30.01", origin="CN", entry_date=D(2025, 6, 1))
    assert "IEEPA_RECIP" not in st.by_measure()
    assert st.amount("IEEPA_FENT") == 2000.0
    assert any("electronics" in x.reason for x in st.exclusions)


def test_in_transit_relief_for_country_rates():
    st = stack(hts="6110.20.20", origin="VN", entry_date=D(2025, 9, 1), loaded_date=D(2025, 8, 1))
    assert st.amount("IEEPA_RECIP") == 1000.0          # 10% baseline, not Vietnam's 20%
    st2 = stack(hts="6110.20.20", origin="VN", entry_date=D(2025, 9, 1), loaded_date=D(2025, 8, 20))
    assert st2.amount("IEEPA_RECIP") == 2000.0


def test_232_derivative_on_metal_content_then_full_value():
    before = stack(hts="7323.93.00", origin="CN", entry_date=D(2025, 10, 1), metal_share=0.6)
    assert before.amount("S232_METAL") == 3000.0        # 50% of $6,000 metal
    assert before.amount("IEEPA_RECIP") == 400.0        # 10% of the $4,000 non-metal part
    after = stack(hts="7323.93.00", origin="CN", entry_date=D(2026, 5, 1), metal_share=0.6)
    assert after.amount("S232_METAL") == 2500.0         # 25% derivative class on full value
    assert "S122" not in after.by_measure()             # non-stacking


def test_undeclared_metal_content_is_flagged():
    st = stack(hts="7323.93.00", origin="CN", entry_date=D(2025, 10, 1))
    assert any("Metal content not declared" in w for w in st.warnings)


def test_usmca_switches_off_mfn_border_duty_and_auto_parts_232():
    st = stack(hts="8708.29.50", origin="MX", entry_date=D(2025, 6, 1), usmca=True)
    assert st.duty == 0
    nonq = stack(hts="8708.29.50", origin="MX", entry_date=D(2025, 6, 1))
    assert nonq.amount("S232_AUTO") == 2500.0 and "IEEPA_FENT" not in nonq.by_measure()


def test_canada_border_duty_rises_to_35_percent():
    assert stack(hts="9403.60.80", origin="CA", entry_date=D(2025, 7, 1)).amount("IEEPA_FENT") == 2500.0
    assert stack(hts="9403.60.80", origin="CA", entry_date=D(2025, 8, 2)).amount("IEEPA_FENT") == 3500.0


def test_india_and_brazil_addons():
    india = stack(hts="6404.11.90", origin="IN", entry_date=D(2025, 9, 15))
    assert india.amount("IEEPA_RECIP") == 2500.0 and india.amount("IEEPA_IN") == 2500.0
    brazil = stack(hts="6403.99.90", origin="BR", entry_date=D(2025, 9, 15))
    assert brazil.amount("IEEPA_BR") == 4000.0
    coffee = stack(hts="0901.21.00", origin="BR", entry_date=D(2025, 12, 1))
    assert "IEEPA_BR" not in coffee.by_measure()


def test_wood_232_and_scheduled_increase():
    assert stack(hts="9403.40.90", origin="VN", entry_date=D(2025, 12, 1)).amount("S232_WOOD") == 2500.0
    assert stack(hts="9403.40.90", origin="VN", entry_date=D(2027, 1, 4)).amount("S232_WOOD") == 5000.0
    assert stack(hts="9401.61.40", origin="DE", entry_date=D(2025, 12, 1)).amount("S232_WOOD") == 1500.0


def test_patented_pharma_and_generics():
    pat = stack(hts="3004.90.92", origin="IN", entry_date=D(2026, 10, 1), flags=frozenset({"patented"}))
    assert pat.amount("S232_PHARMA") == 10_000.0
    gen = stack(hts="3004.90.92", origin="IN", entry_date=D(2026, 10, 1))
    assert "S232_PHARMA" not in gen.by_measure()


def test_china_301_schedule_steps():
    assert stack(hts="8507.60.00", origin="CN", entry_date=D(2025, 6, 1)).amount("S301_CN") == 750.0
    assert stack(hts="8507.60.00", origin="CN", entry_date=D(2026, 2, 1)).amount("S301_CN") == 2500.0


def test_forced_labor_in_transit_exemption():
    st = stack(hts="6110.20.20", origin="VN", entry_date=D(2026, 7, 26), loaded_date=D(2026, 7, 10))
    assert "S301_FL" not in st.by_measure()


def test_fees_capped():
    fees = {f.measure: f.amount for f in engine.entry_fees(1_000_000, "ocean")}
    assert fees["MPF"] == 651.50 and fees["HMF"] == 1250.0
    assert {f.measure: f.amount for f in engine.entry_fees(1_000, "truck")} == {"MPF": 33.58}


def test_unknown_hts_is_reported():
    st = stack(hts="9999.99.99", origin="CN", entry_date=D(2026, 1, 1))
    assert st.duty == 0 and st.warnings


def test_status_by_date():
    assert engine.status("IEEPA_RECIP", D(2026, 3, 1)) == "invalidated"
    assert engine.status("IEEPA_RECIP", D(2026, 1, 1)) == "in_force"
    assert engine.status("S122", D(2026, 6, 1)) == "contested"


def test_catalog_lookup_and_search():
    assert catalog.lookup("6110202079").code == "6110.20.20"
    assert catalog.search("stainless steel cookware pot")[0][0].code == "7323.93.00"


def test_drawback_designates_highest_duty_lots_first():
    lots = [drawback.ImportLot("A", "6302.31.90", D(2025, 3, 1), 100, {"MFN": 100, "IEEPA_RECIP": 500}),
            drawback.ImportLot("B", "6302.31.90", D(2025, 6, 1), 100, {"MFN": 300})]
    exports = [drawback.ExportEvent("X", "6302.31.90", D(2025, 9, 1), 150, "CA")]
    claim = drawback.match(lots, exports, D(2026, 10, 1))
    assert claim.matches[0].lot.entry_number == "B"
    assert claim.refund == pytest.approx(0.99 * (300 + 50 * 1.0))


def test_sourcing_ranks_and_scenarios_add_cost():
    base = sourcing.compare("6110.20.20", "CN", 9.0, 1000, 0.003, D(2026, 10, 1))
    assert base == sorted(base, key=lambda o: o.landed_per_unit)
    scen = sourcing.compare("6110.20.20", "CN", 9.0, 1000, 0.003, D(2026, 10, 1), scenario="excess_capacity")
    cn = next(o for o in scen if o.origin == "CN")
    assert cn.scenario_per_unit > 0
    assert sourcing.price_to_hold_margin(40, 50) == 80


def test_timeline_impact_matches_engine():
    ev = next(e for e in timeline.EVENTS if e.on == D(2025, 6, 4))
    imp = timeline.impact(ev, [{"sku": "P", "description": "coil", "hts": "7208.39.00", "origin": "IN",
                                "unit_cost": 10, "annual_units": 1000, "metal_share": 1.0}])
    assert imp and imp[0].after_rate > imp[0].before_rate
