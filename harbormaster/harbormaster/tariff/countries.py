"""Countries of origin Harbormaster models, with the facts the duty engine and the
sourcing simulator need. Freight and lead-time figures are planning assumptions
(door-to-door, 40' container equivalent per cubic metre), not quotes."""

from __future__ import annotations

from dataclasses import dataclass

EU = {
    "AT", "BE", "BG", "HR", "CY", "CZ", "DK", "EE", "FI", "FR", "DE", "GR", "HU", "IE",
    "IT", "LV", "LT", "LU", "MT", "NL", "PL", "PT", "RO", "SK", "SI", "ES", "SE",
}

# Free-trade agreements whose preference zeroes the MFN rate when the importer claims
# it and the goods qualify. USMCA is tracked separately because qualification also
# switches several 2025-26 measures off.
FTA_PARTNERS = {
    "CA": "USMCA", "MX": "USMCA", "KR": "KORUS", "AU": "AUSFTA", "SG": "USSFTA",
    "CL": "USCFTA", "IL": "USIFTA", "CO": "USCTPA", "PE": "USPTPA", "JO": "USJFTA",
    "MA": "USMFTA", "BH": "USBFTA", "OM": "USOFTA", "PA": "USPATPA",
}


@dataclass(frozen=True)
class Country:
    code: str
    name: str
    region: str
    freight_per_cbm: float  # USD, origin port to US warehouse
    lead_time_days: int
    labor_index: float      # unit-cost multiplier vs China for the same product (planning)

    @property
    def fta(self) -> str | None:
        return FTA_PARTNERS.get(self.code)

    @property
    def is_eu(self) -> bool:
        return self.code in EU


COUNTRIES: dict[str, Country] = {c.code: c for c in [
    Country("CN", "China", "Asia", 62, 38, 1.00),
    Country("HK", "Hong Kong", "Asia", 64, 36, 1.12),
    Country("VN", "Vietnam", "Asia", 68, 40, 1.04),
    Country("IN", "India", "Asia", 74, 46, 0.98),
    Country("BD", "Bangladesh", "Asia", 80, 52, 0.86),
    Country("KH", "Cambodia", "Asia", 78, 48, 0.93),
    Country("TH", "Thailand", "Asia", 70, 42, 1.08),
    Country("MY", "Malaysia", "Asia", 70, 41, 1.10),
    Country("ID", "Indonesia", "Asia", 76, 47, 0.97),
    Country("PH", "Philippines", "Asia", 74, 44, 1.02),
    Country("TW", "Taiwan", "Asia", 60, 34, 1.30),
    Country("KR", "South Korea", "Asia", 58, 30, 1.38),
    Country("JP", "Japan", "Asia", 60, 31, 1.45),
    Country("PK", "Pakistan", "Asia", 82, 50, 0.90),
    Country("LK", "Sri Lanka", "Asia", 82, 50, 0.95),
    Country("TR", "Türkiye", "Europe", 72, 32, 1.12),
    Country("DE", "Germany", "Europe", 56, 24, 1.62),
    Country("IT", "Italy", "Europe", 58, 26, 1.55),
    Country("PL", "Poland", "Europe", 60, 27, 1.28),
    Country("PT", "Portugal", "Europe", 58, 24, 1.30),
    Country("ES", "Spain", "Europe", 57, 24, 1.40),
    Country("FR", "France", "Europe", 57, 24, 1.58),
    Country("NL", "Netherlands", "Europe", 54, 22, 1.60),
    Country("GB", "United Kingdom", "Europe", 55, 20, 1.55),
    Country("CH", "Switzerland", "Europe", 60, 24, 1.85),
    Country("MX", "Mexico", "Americas", 34, 9, 1.15),
    Country("CA", "Canada", "Americas", 28, 6, 1.42),
    Country("BR", "Brazil", "Americas", 66, 28, 1.10),
    Country("CO", "Colombia", "Americas", 48, 16, 1.06),
    Country("AU", "Australia", "Oceania", 70, 30, 1.60),
    Country("SG", "Singapore", "Asia", 62, 34, 1.50),
    Country("ZA", "South Africa", "Africa", 78, 38, 1.05),
]}


def country(code: str) -> Country:
    code = (code or "").upper().strip()
    if code in COUNTRIES:
        return COUNTRIES[code]
    return Country(code, code or "Unknown", "Other", 75, 40, 1.10)


def name(code: str) -> str:
    return country(code).name
