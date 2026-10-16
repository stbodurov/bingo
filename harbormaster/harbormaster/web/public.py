"""Public pages: landing + estimator + waitlist, login, and the waitlist admin."""

from __future__ import annotations

import csv
import io
import os
import re
from datetime import date

from fastapi import APIRouter, Depends, Form, HTTPException, Request
from fastapi.responses import HTMLResponse, RedirectResponse, Response
from sqlalchemy import func, select
from sqlalchemy.exc import IntegrityError
from sqlalchemy.orm import Session

from ..db import WaitlistSignup
from ..security import new_token, same, verify_password
from ..services import estimator as est, queries
from ..services.seed import DEMO_EMAIL
from ..tariff import engine
from .deps import csrf_protect, get_db, limiter, render, user_by_email

router = APIRouter(dependencies=[Depends(csrf_protect)])

EMAIL = re.compile(r"^[^@\s]{1,64}@[^@\s]{1,190}\.[a-z]{2,24}$", re.I)
BANDS = [("under-1m", "Under $1M", 600_000), ("1-5m", "$1M – $5M", 2_500_000), ("5-25m", "$5M – $25M", 12_000_000),
         ("25-100m", "$25M – $100M", 50_000_000), ("100m+", "$100M+", 150_000_000)]
ROLES = ["Founder / CEO", "Finance / CFO", "Operations / supply chain", "E-commerce / marketplace", "Customs broker / forwarder", "Other"]
CHANNELS = ["Shopify / DTC", "Amazon", "Wholesale", "Own retail stores", "B2B / industrial"]


def _float(v, default=0.0) -> float:
    try:
        return float(str(v).replace(",", "").replace("$", "").strip() or default)
    except ValueError:
        return default


def _stats(db: Session) -> dict:
    n = db.scalar(select(func.count(WaitlistSignup.id))) or 0
    return {"signups": n}


HERO = [  # (product, hts, origin, entry date, metal share, usmca)
    ("Cotton hoodie from China", "6110.20.20", "CN", date(2025, 4, 12), None, False),
    ("Same hoodie, entered today", "6110.20.20", "CN", None, None, False),
    ("Stainless skillet from India", "7323.93.00", "IN", date(2025, 9, 15), 0.75, False),
    ("Linen sofa from Vietnam", "9401.61.40", "VN", date(2026, 3, 10), None, False),
]


def hero_stacks() -> list[dict]:
    out = []
    for product, hts, origin, on, share, usmca in HERO:
        on = on or engine.today()
        st = engine.compute(engine.Shipment(hts=hts, origin=origin, value=10_000, entry_date=on, metal_share=share, usmca=usmca))
        refundable = sum(st.amount(m) for m in ("IEEPA_FENT", "IEEPA_RECIP", "IEEPA_BR", "IEEPA_IN"))
        contested = st.amount("S122")
        if refundable:
            note = f"${refundable:,.0f} of every $10,000 is IEEPA duty — refundable with interest."
        elif contested:
            note = f"${contested:,.0f} per $10,000 is Section 122 — held unlawful; protest to preserve it."
        else:
            note = "Every layer is in force today. This is the number your price has to carry."
        out.append({
            "product": product, "when": f"HTS {hts} · entered {on:%b %-d, %Y}", "rate": f"{st.rate:g}",
            "duty": f"${st.duty:,.0f} on $10,000", "note": note,
            "layers": [{"label": c.label, "rate": f"{c.rate:g}", "group": queries.GROUP_OF.get(c.measure, "MFN")}
                       for c in st.charges],
        })
    return out


@router.get("/", response_class=HTMLResponse)
def landing(request: Request, db: Session = Depends(get_db), ref: str | None = None):
    if ref:
        request.session["ref"] = ref[:12]
    preset = est.estimate(5_000_000, {"CN": 55, "VN": 20, "IN": 10, "MX": 15}, metals_pct=15, ior="self", export_pct=5)
    return render(request, "landing.html", est=preset, hero=hero_stacks(), groups=est.GROUP_LABELS, bands=BANDS, roles=ROLES,
                  channels=CHANNELS, stats=_stats(db), today=engine.today(),
                  default_mix={"CN": 55, "VN": 20, "IN": 10, "MX": 15, "EU": 0, "ASIA": 0, "OTHER": 0})


@router.post("/estimate", response_class=HTMLResponse)
async def estimate(request: Request):
    limiter.check(request, "estimate", 60, 60)
    form = await request.form()
    annual = _float(form.get("annual_imports"), 5_000_000)
    mix = {g: _float(form.get(f"mix_{g}"), 0) for g in est.ORIGIN_GROUPS}
    if sum(mix.values()) <= 0:
        mix = {"CN": 100}
    e = est.estimate(min(annual, 5e9), mix, metals_pct=_float(form.get("metals_pct")), ior=str(form.get("ior", "self")),
                     export_pct=_float(form.get("export_pct")))
    request.session["estimate"] = e.as_dict() | {"inputs": {"annual": annual, "mix": mix}}
    return render(request, "partials/estimate.html", est=e)


@router.post("/waitlist")
async def join(request: Request, db: Session = Depends(get_db), email: str = Form(...), company: str = Form(...),
               role: str = Form(""), imports_band: str = Form(""), website: str = Form("")):
    limiter.check(request, "waitlist", 8, 600)
    if website:   # honeypot
        return RedirectResponse("/", 303)
    form = await request.form()
    email = email.strip().lower()
    if not EMAIL.match(email) or not company.strip():
        return render(request, "partials/waitlist_error.html", status_code=422,
                      message="Enter a work email and your company name.")
    existing = db.scalar(select(WaitlistSignup).where(WaitlistSignup.email == email))
    if existing:
        return _done(request, existing)
    signup = WaitlistSignup(
        email=email, company=company.strip()[:200], role=role[:60] if role in ROLES else "Other",
        imports_band=imports_band if imports_band in {b[0] for b in BANDS} else "",
        top_origins=[o for o in form.getlist("origins") if o in est.GROUP_LABELS][:7],
        sells_on=[c for c in form.getlist("channels") if c in CHANNELS][:5],
        estimate=request.session.get("estimate", {}), code=new_token(6)[:8],
        referred_by=request.session.get("ref"), source=(request.headers.get("referer") or "")[:60],
    )
    db.add(signup)
    try:
        db.commit()
    except IntegrityError:
        db.rollback()
        signup = db.scalar(select(WaitlistSignup).where(WaitlistSignup.email == email))
    return _done(request, signup)


def _done(request: Request, signup: WaitlistSignup):
    url = f"/w/{signup.code}"
    if request.headers.get("hx-request"):
        return Response(status_code=204, headers={"HX-Redirect": url})
    return RedirectResponse(url, 303)


def _priority(db: Session, signup: WaitlistSignup) -> dict:
    """Onboarding order: money with the nearest deadline first, then signup order."""
    rows = db.scalars(select(WaitlistSignup).order_by(WaitlistSignup.created_at)).all()

    def score(s: WaitlistSignup) -> float:
        e = s.estimate or {}
        urgent = e.get("refund_total", 0) * (0.5 + e.get("ieepa_final_share", 0))
        return urgent + 25_000 * sum(1 for r in rows if r.referred_by == s.code)

    ranked = sorted(rows, key=lambda s: (-score(s), s.created_at))
    pos = next((i for i, s in enumerate(ranked, 1) if s.id == signup.id), len(ranked))
    return {"position": pos, "of": len(ranked), "referrals": sum(1 for r in rows if r.referred_by == signup.code)}


@router.get("/w/{code}", response_class=HTMLResponse)
def waitlist_status(request: Request, code: str, db: Session = Depends(get_db)):
    signup = db.scalar(select(WaitlistSignup).where(WaitlistSignup.code == code))
    if signup is None:
        raise HTTPException(404)
    share = str(request.base_url).rstrip("/") + f"/?ref={signup.code}"
    return render(request, "waitlist_done.html", s=signup, p=_priority(db, signup), share=share)


@router.get("/login", response_class=HTMLResponse)
def login_page(request: Request, next: str = "/app"):
    return render(request, "login.html", next=next if next.startswith("/") else "/app", demo_email=DEMO_EMAIL)


@router.post("/login")
def login(request: Request, db: Session = Depends(get_db), email: str = Form(...), password: str = Form(...),
          next: str = Form("/app")):
    limiter.check(request, "login", 10, 300)
    user = user_by_email(db, email)
    if user is None or not verify_password(password, user.password_hash):
        return render(request, "login.html", status_code=401, error="That email and password don't match.",
                      next=next, demo_email=DEMO_EMAIL)
    request.session.clear()
    request.session["uid"] = user.id
    return RedirectResponse(next if next.startswith("/") and not next.startswith("//") else "/app", 303)


@router.post("/demo")
def demo(request: Request, db: Session = Depends(get_db)):
    user = user_by_email(db, DEMO_EMAIL)
    if user is None:
        raise HTTPException(404)
    request.session.clear()
    request.session["uid"] = user.id
    return RedirectResponse("/app", 303)


@router.post("/logout")
def logout(request: Request):
    request.session.clear()
    return RedirectResponse("/", 303)


def _admin(token: str | None) -> None:
    if not same(token, os.environ.get("HARBORMASTER_ADMIN_TOKEN", "dev-admin")):
        raise HTTPException(404)


@router.get("/admin/waitlist", response_class=HTMLResponse)
def admin_waitlist(request: Request, token: str | None = None, db: Session = Depends(get_db)):
    _admin(token)
    rows = db.scalars(select(WaitlistSignup).order_by(WaitlistSignup.created_at.desc())).all()
    total = sum((r.estimate or {}).get("refund_total", 0) for r in rows)
    return render(request, "admin_waitlist.html", rows=rows, total=total, token=token)


@router.get("/admin/waitlist.csv")
def admin_waitlist_csv(token: str | None = None, db: Session = Depends(get_db)):
    _admin(token)
    buf = io.StringIO()
    w = csv.writer(buf)
    w.writerow(["created_at", "email", "company", "role", "imports_band", "origins", "channels", "refund_estimate", "code", "referred_by"])
    for r in db.scalars(select(WaitlistSignup).order_by(WaitlistSignup.created_at)):
        w.writerow([r.created_at.isoformat(), r.email, r.company, r.role, r.imports_band, " ".join(r.top_origins or []),
                    "; ".join(r.sells_on or []), (r.estimate or {}).get("refund_total", ""), r.code, r.referred_by or ""])
    return Response(buf.getvalue(), media_type="text/csv",
                    headers={"Content-Disposition": "attachment; filename=harbormaster-waitlist.csv"})
