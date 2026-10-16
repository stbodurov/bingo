"""HTS classification assistant.

The model reads a product description and ranks candidate subheadings *from the
reference catalog* (so every candidate has a computable duty), citing the General Rules
of Interpretation it relied on and the facts it would need to be sure. Duty for each
candidate is then computed by the engine, so the user sees what the classification
decision is worth. A human confirms before a SKU's code changes.
"""

from __future__ import annotations

import re
from datetime import date

from ..tariff import catalog, engine
from . import llm

SYSTEM = """You are a licensed-customs-broker-grade HTS classification assistant for US imports.
Classify the product into the Harmonized Tariff Schedule of the United States using the General
Rules of Interpretation (GRI 1 first: headings and section/chapter notes; then GRI 3 for
composite goods; GRI 6 at subheading level).

You may only choose from the reference subheadings provided. If none fits, return the code
OUTSIDE_CATALOG as the top candidate and explain which heading you believe applies.

Rank up to three candidates. Confidence is your probability that the candidate is correct
(0 to 1; the candidates' confidences need not sum to 1). In `reasoning`, cite the decisive
facts and the GRI or note you relied on, in two or three sentences a customs reviewer could
check. List any facts that would change the answer in `missing_facts` (material composition by
weight, function, whether it is for household use, value per unit, and so on). Never invent
duty rates — they are computed separately."""


def _schema() -> dict:
    codes = [l.code for l in catalog.LINES] + ["OUTSIDE_CATALOG"]
    return {
        "type": "object",
        "properties": {
            "candidates": {"type": "array", "items": {
                "type": "object",
                "properties": {
                    "code": {"type": "string", "enum": codes},
                    "confidence": {"type": "number"},
                    "gri": {"type": "string"},
                    "reasoning": {"type": "string"},
                },
                "required": ["code", "confidence", "gri", "reasoning"],
                "additionalProperties": False,
            }},
            "missing_facts": {"type": "array", "items": {"type": "string"}},
            "summary": {"type": "string"},
        },
        "required": ["candidates", "missing_facts", "summary"],
        "additionalProperties": False,
    }


def _catalog_text() -> str:
    return "\n".join(f"{l.code}  {l.description}" for l in catalog.LINES)


MATERIAL_HINTS = {
    "material": ("cotton", "polyester", "steel", "stainless", "aluminum", "aluminium", "wood", "oak", "plastic",
                 "ceramic", "porcelain", "glass", "leather", "rubber", "copper", "brass", "linen", "bamboo"),
}


def _missing(desc: str) -> list[str]:
    d = desc.lower()
    out = []
    if not any(m in d for m in MATERIAL_HINTS["material"]):
        out.append("Material of the outer surface or principal component (by weight or value).")
    if not re.search(r"\d", d):
        out.append("Size, capacity or unit value — several subheadings split on these.")
    if any(w in d for w in ("set", "kit", "bundle")):
        out.append("Whether the set is put up for retail sale (GRI 3(b) essential character).")
    if any(w in d for w in ("electric", "battery", "usb", "led", "motor")):
        out.append("Power source and rated wattage.")
    return out


def classify(description: str, origin: str = "CN", on: date | None = None, value: float = 1000.0) -> dict:
    on = on or engine.today()
    origin = (origin or "CN").upper()
    result = None
    engine_name = "catalog"
    if llm.enabled():
        prompt = (f"Reference subheadings:\n{_catalog_text()}\n\nProduct description from the importer:\n"
                  f"\"\"\"\n{description.strip()[:4000]}\n\"\"\"\nCountry of origin: {origin}.")
        result = llm.structured(SYSTEM, prompt, _schema(), effort="high", max_tokens=6000)
        if result and result.get("candidates"):
            engine_name = "claude"
        else:
            result = None
    if result is None:
        hits = catalog.search(description, limit=3)
        top = hits[0][1] if hits else 1.0
        result = {
            "candidates": [{
                "code": line.code,
                "confidence": round(min(0.9, 0.35 + 0.55 * score / max(top, 1.0)) * (1 if i == 0 else 0.7), 2),
                "gri": "GRI 1 (heading text), GRI 6",
                "reasoning": f"Matched on: {', '.join(k for k in line.keywords if k in description.lower()) or 'description terms'}. "
                             f"Heading {line.heading} covers {line.description.lower()}.",
            } for i, (line, score) in enumerate(hits)],
            "missing_facts": _missing(description),
            "summary": "Keyword match against the reference catalog. Confirm with a broker before filing."
                       if hits else "No reference subheading matched; this product needs broker review.",
        }
    # Attach duty consequences from the engine.
    for c in result["candidates"]:
        line = catalog.lookup(c["code"]) if c["code"] != "OUTSIDE_CATALOG" else None
        c["description"] = line.description if line else "Outside the reference catalog"
        if line:
            st = engine.compute(engine.Shipment(hts=line.code, origin=origin, value=value, entry_date=on))
            c["duty_rate"] = st.rate
            c["measures"] = st.by_measure()
        else:
            c["duty_rate"] = None
            c["measures"] = {}
    rates = [c["duty_rate"] for c in result["candidates"] if c["duty_rate"] is not None]
    result["rate_spread"] = round(max(rates) - min(rates), 2) if len(rates) > 1 else 0.0
    result["engine"] = engine_name
    result["origin"] = origin
    return result
