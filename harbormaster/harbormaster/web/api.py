"""JSON API (v1): the landed-cost engine as a metered API, plus read access to an
importer's findings. Authenticated with `Authorization: Bearer hm_live_…`."""

from __future__ import annotations

from datetime import date

from fastapi import APIRouter, Depends, Header, HTTPException, Request
from pydantic import BaseModel, Field
from sqlalchemy import select
from sqlalchemy.orm import Session

from ..db import Org
from ..security import api_key_hash
from ..services import estimator as est, queries
from ..tariff import catalog, engine
from .deps import get_db, limiter

router = APIRouter(prefix="/api/v1", tags=["v1"])


def org_from_key(authorization: str | None = Header(None), db: Session = Depends(get_db)) -> Org:
    if not authorization or not authorization.lower().startswith("bearer "):
        raise HTTPException(401, "Send Authorization: Bearer <api key>")
    digest = api_key_hash(authorization.split(" ", 1)[1].strip())
    org = db.scalar(select(Org).where(Org.api_key_hash == digest))
    if org is None or not org.api_key_hash:
        raise HTTPException(401, "Unknown API key")
    return org


class StackIn(BaseModel):
    hts: str = Field(examples=["6110.20.20"])
    origin: str = Field(min_length=2, max_length=2, examples=["VN"])
    value: float = Field(gt=0, examples=[12000])
    entry_date: date | None = None
    loaded_date: date | None = None
    metal_share: float | None = Field(default=None, ge=0, le=1)
    usmca: bool = False
    fta: bool = False
    flags: list[str] = []
    mode: str = "ocean"
    quantity: float = 1
    freight: float = 0


@router.post("/stack")
def stack(body: StackIn, request: Request, org: Org = Depends(org_from_key)):
    limiter.check(request, "api", 600, 60)
    if catalog.lookup(body.hts) is None:
        raise HTTPException(422, f"HTS {body.hts} is outside the reference catalog")
    sh = engine.Shipment(hts=body.hts, origin=body.origin.upper(), value=body.value,
                         entry_date=body.entry_date or engine.today(), loaded_date=body.loaded_date,
                         metal_share=body.metal_share, usmca=body.usmca, fta=body.fta,
                         flags=frozenset(body.flags), mode=body.mode, quantity=body.quantity)
    lc = engine.landed_cost(sh, freight=body.freight)
    st = lc["stack"]
    return {
        "hts": body.hts, "origin": sh.origin_code, "entry_date": sh.entry_date.isoformat(),
        "duty": st.duty, "effective_rate": st.rate, "fees": lc["fee_total"], "landed_cost": lc["total"],
        "landed_per_unit": lc["per_unit"],
        "charges": [{"measure": c.measure, "label": c.label, "rate": c.rate, "base": c.base, "amount": c.amount,
                     "heading": c.heading, "status": engine.status(c.measure), "note": c.note, "authority": c.authority}
                    for c in st.charges],
        "fees_detail": [{"measure": f.measure, "amount": f.amount} for f in lc["fees"]],
        "not_applied": [{"measure": x.measure, "reason": x.reason} for x in st.exclusions],
        "warnings": st.warnings,
        "disclaimer": "Reference tariff data; verify against the HTSUS and CBP CSMS before filing.",
    }


class EstimateIn(BaseModel):
    annual_imports: float = Field(gt=0)
    mix: dict[str, float]
    metals_pct: float = 0
    ior: str = "self"
    export_pct: float = 0


@router.post("/estimate")
def estimate(body: EstimateIn, request: Request):
    limiter.check(request, "estimate", 60, 60)
    e = est.estimate(body.annual_imports, body.mix, body.metals_pct, body.ior, body.export_pct)
    return e.as_dict() | {"warnings": e.warnings, "assumptions": e.assumptions}


@router.get("/findings")
def findings(org: Org = Depends(org_from_key), db: Session = Depends(get_db)):
    return [{"entry": f.entry.entry_number, "kind": f.kind, "measure": f.measure, "route": f.route, "title": f.title,
             "amount": f.amount, "interest": f.interest, "deadline": f.deadline, "urgency": f.urgency,
             "authority": f.authority} for f in queries.open_findings(db, org)]


@router.get("/catalog")
def reference_catalog():
    return [{"hts": l.code, "description": l.description, "mfn": l.mfn, "s232": l.s232, "metal": l.metal}
            for l in catalog.LINES]
