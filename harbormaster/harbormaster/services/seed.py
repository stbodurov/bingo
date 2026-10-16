"""Demo workspace: Larkspur & Pine Home Co., a fictional DTC + wholesale home-goods brand
importing ~$6M a year from ten countries through two brokers.

The ledger is generated from the engine itself — every line is filed with the correct
stack — and then realistic broker mistakes are injected at known rates, so the audit
has something true to find. Nothing here describes a real company."""

from __future__ import annotations

import hashlib
import random
from datetime import date, datetime, timedelta

from sqlalchemy import select
from sqlalchemy.orm import Session

from ..db import (Activity, Broker, Case, Entry, EntryLine, ExportRecord, Finding, Org, Sku, Supplier,
                  User)
from ..security import hash_password
from ..tariff import engine
from . import audit

DEMO_EMAIL = "demo@larkspur.example"
DEMO_PASSWORD = "harbor-demo"

SUPPLIERS = [
    ("Ningbo Hearthware Co.", "CN", "FOB"), ("Binh Duong Living Furniture", "VN", "FOB"),
    ("Moradabad Metalcraft", "IN", "FOB"), ("Monterrey Casa Industrial", "MX", "FCA"),
    ("Barcelos Ceramica", "PT", "FOB"), ("Denizli Linen & Glass", "TR", "FOB"),
    ("Bandung Sleep Works", "ID", "FOB"), ("Dhaka Knit Partners", "BD", "FOB"),
    ("Huila Coffee Exporters", "CO", "FOB"), ("Faenza Dolci", "IT", "FCA"),
]

# sku, description, hts, origin, unit cost, price, cbm/unit, annual units, metal share, usmca, flags
SKUS = [
    ("LP-MUG-12", "Stoneware mug, 12 oz, reactive glaze", "6912.00.48", "PT", 2.10, 18.0, 0.0016, 62000, None, False),
    ("LP-DIN-16", "Porcelain dinnerware set, 16 pieces", "6911.10.80", "CN", 14.50, 89.0, 0.045, 9000, None, False),
    ("LP-GLS-4", "Recycled-glass tumblers, set of 4", "7013.37.20", "TR", 4.20, 28.0, 0.006, 21000, None, False),
    ("LP-SKL-10", "Tri-ply stainless skillet, 10 in", "7323.93.00", "IN", 11.80, 79.0, 0.012, 18000, 0.75, False),
    ("LP-POT-8", "Stainless stockpot, 8 qt", "7323.93.00", "CN", 13.40, 85.0, 0.03, 9500, 0.8, False),
    ("LP-DUT-5", "Enamelled cast-aluminium Dutch oven, 5 qt", "7615.10.71", "CN", 17.90, 129.0, 0.028, 11000, 0.9, False),
    ("LP-BTL-20", "Insulated steel bottle, 20 oz", "7323.93.00", "CN", 3.10, 32.0, 0.0028, 48000, 0.6, False),
    ("LP-BLD-1", "Countertop blender, 1.2 L", "8509.40.00", "CN", 21.00, 119.0, 0.03, 7000, None, False),
    ("LP-KTL-1", "Gooseneck electric kettle", "8516.71.00", "CN", 12.30, 69.0, 0.012, 16000, None, False),
    ("LP-FAN-36", "Bladeless tower fan, 36 in", "8414.51.30", "CN", 24.00, 139.0, 0.06, 5200, None, False),
    ("LP-LMP-D", "Rattan desk lamp, LED", "9405.21.80", "VN", 9.80, 64.0, 0.02, 12000, None, False),
    ("LP-PND-L", "Woven pendant light", "9405.11.40", "CN", 15.20, 98.0, 0.05, 6000, None, False),
    ("LP-SOF-3", "Linen three-seat sofa", "9401.61.40", "VN", 310.00, 1690.0, 1.6, 2100, None, False),
    ("LP-ARM-1", "Oak frame lounge armchair", "9401.61.40", "MX", 142.00, 790.0, 0.9, 1800, None, True),
    ("LP-VAN-36", "Bathroom vanity, 36 in, solid wood", "9403.40.90", "VN", 188.00, 1150.0, 0.7, 1300, None, False),
    ("LP-TBL-72", "Oak dining table, 72 in", "9403.60.80", "VN", 265.00, 1480.0, 1.2, 1500, None, False),
    ("LP-SHF-5", "Powder-coated steel shelving, 5 tier", "9403.20.00", "CN", 38.00, 199.0, 0.09, 4200, 0.85, False),
    ("LP-HNG-6", "Brass cabinet hinges, 6 pack", "8302.42.30", "CN", 2.40, 19.0, 0.0009, 26000, 0.95, False),
    ("LP-MAT-Q", "Hybrid foam mattress, queen", "9404.21.00", "ID", 118.00, 899.0, 0.35, 3800, None, False),
    ("LP-SHT-Q", "Percale cotton sheet set, queen", "6302.31.90", "IN", 16.50, 129.0, 0.008, 22000, None, False),
    ("LP-DUV-Q", "Washed linen duvet cover, queen", "6302.31.90", "PT", 27.00, 189.0, 0.009, 9000, None, False),
    ("LP-TEE-1", "Organic cotton logo tee", "6109.10.00", "BD", 3.60, 34.0, 0.0012, 15000, None, False),
    ("LP-HDY-1", "Brushed cotton hoodie", "6110.20.20", "CN", 9.20, 78.0, 0.003, 8000, None, False),
    ("LP-TOT-1", "Heavy canvas market tote", "4202.92.31", "IN", 3.90, 36.0, 0.002, 19000, None, False),
    ("LP-CBD-L", "Acacia end-grain cutting board", "4421.99.98", "VN", 6.40, 58.0, 0.006, 14000, None, False),
    ("LP-BOX-M", "Corrugated mailer boxes (per 100)", "4819.10.00", "MX", 31.00, 31.0, 0.2, 1400, None, True),
    ("LP-MLR-P", "Recycled poly mailers (per 500)", "3923.21.00", "CN", 22.00, 22.0, 0.05, 900, None, False),
    ("LP-CNT-10", "Glass-lid storage containers, 10 pc", "3924.10.40", "CN", 7.70, 49.0, 0.01, 17000, None, False),
    ("LP-SPK-1", "Walnut Bluetooth speaker", "8518.22.00", "CN", 19.00, 129.0, 0.008, 6200, None, False),
    ("LP-CHG-65", "65 W USB-C charger", "8504.40.95", "CN", 6.10, 45.0, 0.0008, 21000, None, False),
    ("LP-CBL-2", "Braided USB-C cable, 2 m", "8544.42.90", "CN", 1.20, 19.0, 0.0002, 52000, 0.25, False),
    ("LP-PBK-10", "10,000 mAh power bank", "8507.60.00", "CN", 7.80, 49.0, 0.0007, 16000, None, False),
    ("LP-POS-T", "Retail POS tablet (store fixtures)", "8471.30.01", "CN", 142.00, 142.0, 0.01, 260, None, False),
    ("LP-BLB-4", "Filament LED bulbs, 4 pack", "8539.52.00", "CN", 3.30, 24.0, 0.0015, 24000, None, False),
    ("LP-GRL-2", "Portable two-burner gas grill", "7321.11.10", "MX", 64.00, 349.0, 0.15, 2600, 0.7, True),
    ("LP-CHC-B", "Dark chocolate gift bars", "1806.32.90", "IT", 1.40, 9.0, 0.0004, 30000, None, False),
    ("LP-COF-B", "Single-origin coffee beans, 1 kg", "0901.21.00", "CO", 9.50, 34.0, 0.0015, 12000, None, False),
    ("LP-SOP-3", "Olive-oil bar soap, 3 pack", "3401.11.50", "IN", 2.20, 18.0, 0.0006, 20000, None, False),
    ("LP-CHR-T", "Upholstered task chair, steel base", "9401.71.00", "CN", 58.00, 329.0, 0.25, 2400, 0.4, False),
]

LANES = {  # origin -> (supplier index, broker index, mode, port)
    "CN": (0, 0, "ocean", "2704"), "VN": (1, 0, "ocean", "2704"), "IN": (2, 1, "ocean", "1001"),
    "MX": (3, 1, "truck", "2304"), "PT": (4, 1, "ocean", "1001"), "TR": (5, 1, "ocean", "1001"),
    "ID": (6, 0, "ocean", "2704"), "BD": (7, 1, "ocean", "1001"), "CO": (8, 1, "air", "5206"),
    "IT": (9, 1, "air", "4701"),
}


def _entry_number(rng: random.Random, filer: str) -> str:
    digits = f"{rng.randrange(10**6, 10**7)}"
    check = sum(int(d) for d in digits) % 10
    return f"{filer}-{digits}-{check}"


def seed_demo(session: Session) -> Org:
    existing = session.scalar(select(Org).where(Org.slug == "larkspur"))
    if existing:
        return existing
    rng = random.Random(1407)
    today = engine.today()
    org = Org(name="Larkspur & Pine Home Co.", slug="larkspur", plan="growth", ior_last4="7731", target_margin=58.0)
    session.add(org)
    session.flush()
    session.add_all([
        User(org_id=org.id, email=DEMO_EMAIL, name="Dana Whitfield", role="admin", password_hash=hash_password(DEMO_PASSWORD)),
        User(org_id=org.id, email="ops@larkspur.example", name="Rafael Okafor", role="analyst", password_hash=hash_password(DEMO_PASSWORD)),
    ])
    brokers = [Broker(org_id=org.id, name="Harborline Customs Brokerage", filer_code="HLB"),
               Broker(org_id=org.id, name="Meridian Trade Services", filer_code="MTS")]
    suppliers = [Supplier(org_id=org.id, name=n, country=c, incoterm=i) for n, c, i in SUPPLIERS]
    session.add_all(brokers + suppliers)
    session.flush()
    skus: dict[str, Sku] = {}
    for sku, desc, hts, origin, cost, price, cbm, units, metal, usmca in SKUS:
        sup = suppliers[LANES[origin][0]]
        flags = []
        s = Sku(org_id=org.id, sku=sku, description=desc, hts=hts, origin=origin, supplier_id=sup.id,
                unit_cost=cost, unit_price=price, cbm_per_unit=cbm, annual_units=units, metal_share=metal,
                usmca=usmca, flags=flags, classified_by="broker")
        session.add(s)
        skus[sku] = s
    session.flush()

    by_origin: dict[str, list[Sku]] = {}
    for s in skus.values():
        by_origin.setdefault(s.origin, []).append(s)

    start = date(2025, 1, 6)
    entries: list[Entry] = []
    for origin, lane_skus in by_origin.items():
        sup_i, broker_i, mode, port = LANES[origin]
        broker = brokers[broker_i]
        annual_value = sum(s.unit_cost * s.annual_units for s in lane_skus)
        per_year = 26 if annual_value > 400_000 else 12 if annual_value > 120_000 else 6
        step = 365 // per_year
        d = start + timedelta(days=rng.randrange(0, step))
        while d <= today - timedelta(days=5):
            entries.append(_make_entry(session, rng, org, broker, origin, mode, port, lane_skus, d, per_year))
            d += timedelta(days=step + rng.randrange(-3, 4))
    session.flush()
    _inject_errors(rng, entries)
    for e in entries:
        audit.recompute(e)
    session.flush()
    audit.audit_all(session, org)
    _history(session, org, rng)
    _exports(session, org, skus, rng, today)
    session.add_all([
        Activity(org_id=org.id, actor="Harbormaster", action="Imported 21 months of ACE entry summaries (ES-003)."),
        Activity(org_id=org.id, actor="Harbormaster", action="Audited every line against the 2025–26 tariff stack."),
        Activity(org_id=org.id, actor="Dana Whitfield", action="Filed CAPE declaration #1 (Phase 1)."),
    ])
    session.commit()
    return org


def _make_entry(session, rng, org, broker, origin, mode, port, lane_skus, d, per_year) -> Entry:
    n_lines = min(len(lane_skus), rng.choice([1, 2, 2, 3, 4]))
    chosen = rng.sample(lane_skus, n_lines)
    summary = d + timedelta(days=rng.choice([7, 8, 9, 10]))
    loaded = d - timedelta(days=engine.country(origin).lead_time_days - rng.randrange(0, 5)) if mode == "ocean" else d - timedelta(days=3)
    e = Entry(org_id=org.id, entry_number=_entry_number(rng, broker.filer_code), broker_id=broker.id, mode=mode,
              port=port, origin=origin, loaded_date=loaded, entry_date=d, summary_date=summary)
    scheduled = summary + timedelta(days=314)
    if rng.random() < 0.06:
        e.liquidation_status = "extended"
        e.liquidation_date = scheduled + timedelta(days=365)
    elif scheduled <= engine.today():
        e.liquidation_status = "liquidated"
        e.liquidation_date = scheduled + timedelta(days=rng.choice([0, 0, 7, 14]))
    total = 0.0
    for i, s in enumerate(chosen, start=1):
        # Each entry carries a sample of the lane's SKUs; scale so annual volumes hold.
        qty = max(10, int(s.annual_units / per_year * len(lane_skus) / n_lines * rng.uniform(0.6, 1.4)))
        value = round(qty * s.unit_cost * rng.uniform(0.97, 1.03), 2)
        total += value
        line = EntryLine(line_no=i, sku_id=s.id, hts=s.hts, origin=origin, description=s.description,
                         quantity=qty, value=value, metal_share=s.metal_share, usmca=s.usmca)
        line.sku = s
        e.lines.append(line)
    e.entered_value = round(total, 2)
    # File the correct stack first; mistakes are injected afterwards.
    for line in e.lines:
        st = engine.compute(engine.Shipment(hts=line.hts, origin=origin, value=line.value, entry_date=d,
                                            loaded_date=loaded, metal_share=line.metal_share, usmca=line.usmca,
                                            fta=line.usmca, mode=mode))
        line.paid = st.by_measure()
    fees = {f.measure: f.amount for f in engine.entry_fees(e.entered_value, mode)}
    e.paid_by_measure = fees
    session.add(e)
    return e


def _bump(line: EntryLine, measure: str, amount: float) -> None:
    paid = dict(line.paid or {})
    paid[measure] = round(paid.get(measure, 0.0) + amount, 2)
    line.paid = paid


def _inject_errors(rng: random.Random, entries: list[Entry]) -> None:
    """Broker mistakes we see in real audits, at plausible frequencies."""
    for e in entries:
        harborline = e.entry_number.startswith("HLB")
        for line in e.lines:
            sku = line.sku
            v = line.value
            # 1. 232 metal duty on full value instead of declared metal content (pre-Apr-2026).
            if sku.metal_share and sku.metal_share < 1 and e.entry_date < engine.METAL_FULL_VALUE \
                    and (line.paid or {}).get("S232_METAL") and harborline and rng.random() < 0.55:
                correct = line.paid["S232_METAL"]
                _bump(line, "S232_METAL", round(correct / sku.metal_share - correct, 2))
            # 2. Apparel filed at List 3 (25%) instead of List 4A (7.5%).
            if line.hts.startswith("61") and e.origin == "CN" and rng.random() < 0.5:
                _bump(line, "S301_CN", round(v * 0.175, 2))
            # 3. USMCA preference not claimed on qualifying Mexican goods (Meridian).
            if line.usmca and not harborline and rng.random() < 0.35:
                mfn = (engine.catalog.lookup(line.hts).mfn or 0) / 100
                if mfn:
                    _bump(line, "MFN", round(v * mfn, 2))
                if engine.S122_START <= e.entry_date < engine.S122_END:
                    _bump(line, "S122", round(v * 0.10, 2))
            # 4. Section 122 charged on goods covered by 232.
            if engine.S122_START <= e.entry_date < engine.S122_END and sku.hts.startswith(("9401.61", "9403.40")) \
                    and rng.random() < 0.6:
                _bump(line, "S122", round(v * 0.10, 2))
            # 5. China 301 battery increase missed (underpayment exposure).
            if sku.sku == "LP-PBK-10" and e.entry_date >= date(2026, 1, 1) and rng.random() < 0.7:
                _bump(line, "S301_CN", -round(v * 0.175, 2))
            # 6. Forced-labor 301 charged on in-transit goods.
            if date(2026, 7, 24) <= e.entry_date < date(2026, 7, 28) and e.loaded_date and e.loaded_date < date(2026, 7, 24):
                tier = engine.FL301_TIERS.get(e.origin, 0)
                _bump(line, "S301_FL", round(v * tier / 100, 2))
            # 7. Reciprocal charged on the electronics exemption (refunded via CAPE either way).
            if sku.sku == "LP-POS-T" and engine.D(2025, 4, 5) <= e.entry_date < engine.IEEPA_END:
                _bump(line, "IEEPA_RECIP", round(v * 0.10, 2))
        # 8. MPF not capped / HMF on truck entries.
        fees = dict(e.paid_by_measure or {})
        if e.entered_value * engine.MPF_RATE > engine.MPF_MAX and rng.random() < 0.4:
            fees["MPF"] = round(e.entered_value * engine.MPF_RATE, 2)
        if e.mode == "truck" and rng.random() < 0.5:
            fees["HMF"] = round(e.entered_value * engine.HMF_RATE, 2)
        e.paid_by_measure = fees
        e.duty_paid = round(sum(sum((l.paid or {}).values()) for l in e.lines), 2)
        e.fees_paid = round(sum(fees.values()), 2)


def _history(session: Session, org: Org, rng: random.Random) -> None:
    """A first CAPE filing already went through in May; most of its money is back."""
    findings = session.scalars(select(Finding).where(
        Finding.org_id == org.id, Finding.route == "CAPE", Finding.status == "open")).all()
    filed = [f for f in findings if f.entry.entry_date < date(2026, 1, 15)]
    if not filed:
        return
    case = Case(org_id=org.id, route="CAPE", title="CAPE declaration #1 — Phase 1", status="recovered",
                created_at=datetime(2026, 5, 4, 15, 20), filed_at=datetime(2026, 5, 6, 10, 2),
                notes="Accepted May 9; refunds received Jun 12 and Jun 26 (with interest).")
    session.add(case)
    session.flush()
    total = 0.0
    for f in filed:
        f.status = "recovered"
        f.case_id = case.id
        total += f.amount + f.interest
    case.amount = round(total, 2)
    case.recovered = round(total, 2)


def _exports(session: Session, org: Org, skus: dict[str, Sku], rng: random.Random, today: date) -> None:
    """Wholesale shipments to a Canadian retail partner — drawback-eligible."""
    for code, share in (("LP-SKL-10", 0.08), ("LP-SHT-Q", 0.06), ("LP-BTL-20", 0.05), ("LP-HDY-1", 0.1)):
        s = skus[code]
        d = date(2025, 3, 3)
        i = 0
        while d < today:
            qty = int(s.annual_units * share / 6 * rng.uniform(0.7, 1.3))
            session.add(ExportRecord(org_id=org.id, ref=f"CA-WS-{2025 + (d.year - 2025)}-{i:03d}", sku_id=s.id,
                                     export_date=d, quantity=qty, destination="CA"))
            d += timedelta(days=61)
            i += 1


def demo_api_key() -> tuple[str, str]:
    key = "hm_demo_" + hashlib.sha256(b"larkspur-demo").hexdigest()[:24]
    return key, hashlib.sha256(key.encode()).hexdigest()
