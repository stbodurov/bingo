"""The importer workspace (behind login)."""

from __future__ import annotations

from datetime import date, timedelta

from fastapi import APIRouter, Depends, File, Form, HTTPException, Request, UploadFile
from fastapi.responses import HTMLResponse, RedirectResponse, Response
from sqlalchemy import select
from sqlalchemy.orm import selectinload

from ..ai import classify as ai_classify, copilot, llm
from ..db import Activity, Classification, Entry, Finding, Message, Sku
from ..security import new_api_key
from ..services import audit, cases, importer, queries
from ..tariff import catalog, engine, sourcing, timeline
from ..tariff.countries import COUNTRIES
from ..tariff.recovery import EntryFacts, deadlines as entry_deadlines
from . import charts
from .deps import Ctx, csrf_protect, current, limiter, render

router = APIRouter(prefix="/app", dependencies=[Depends(csrf_protect)])

NAV = [
    ("dashboard", "/app", "Overview"), ("recovery", "/app/recovery", "Recovery"),
    ("entries", "/app/entries", "Entries"), ("deadlines", "/app/deadlines", "Deadlines"),
    ("catalog", "/app/catalog", "Landed cost & margin"), ("sourcing", "/app/sourcing", "Sourcing"),
    ("calculator", "/app/calculator", "Duty calculator"), ("classify", "/app/classify", "Classification"),
    ("changes", "/app/changes", "Tariff changes"), ("drawback", "/app/drawback", "Drawback"),
    ("brokers", "/app/brokers", "Broker scorecard"), ("copilot", "/app/copilot", "Copilot"),
]


def page(request: Request, ctx: Ctx, name: str, active: str, **kw):
    return render(request, name, user=ctx.user, org=ctx.org, nav=NAV, active=active, ai_on=llm.enabled(), **kw)


def _chart_payload(d: dict):
    cats = [d_["month"] for d_ in d["chart"]]
    labels = [date(int(m[:4]), int(m[5:]), 1).strftime("%b %Y") for m in cats]
    short = [date(int(m[:4]), int(m[5:]), 1).strftime("%b %y") for m in cats]
    return charts.stacked_columns(short, labels, d["series"], [c["values"] for c in d["chart"]],
                                  title="Duty paid by month and measure")


@router.get("", response_class=HTMLResponse)
def dashboard(request: Request, ctx: Ctx = Depends(current)):
    d = queries.dashboard(ctx.db, ctx.org)
    acts = ctx.db.scalars(select(Activity).where(Activity.org_id == ctx.org.id).order_by(Activity.created_at.desc()).limit(6)).all()
    return page(request, ctx, "app/dashboard.html", "dashboard", d=d, chart=_chart_payload(d), activity=acts)


# ── Entries ───────────────────────────────────────────────────────────────────

@router.get("/entries", response_class=HTMLResponse)
def entries(request: Request, ctx: Ctx = Depends(current), origin: str = "", status: str = "", broker: str = "",
            q: str = "", flag: str = ""):
    es = queries.entries(ctx.db, ctx.org)
    fs = ctx.db.scalars(select(Finding).where(Finding.org_id == ctx.org.id, Finding.status.in_(("open", "in_case")))).all()
    by_entry: dict[int, list[Finding]] = {}
    for f in fs:
        by_entry.setdefault(f.entry_id, []).append(f)
    rows = []
    for e in es:
        if origin and e.origin != origin:
            continue
        if status and e.liquidation_status != status:
            continue
        if broker and (e.broker is None or e.broker.filer_code != broker):
            continue
        if q and q.lower() not in e.entry_number.lower() and not any(q.lower() in (l.description or "").lower() for l in e.lines):
            continue
        ef = by_entry.get(e.id, [])
        if flag == "errors" and not any(f.kind in ("overpayment", "underpayment") for f in ef):
            continue
        if flag == "refund" and not any(f.kind == "refund" for f in ef):
            continue
        rows.append((e, ef))
    origins = sorted({e.origin for e in es})
    brokers = sorted({e.broker.filer_code for e in es if e.broker})
    ctx_kw = dict(rows=rows[:400], total=len(rows), origins=origins, brokers=brokers,
                  f={"origin": origin, "status": status, "broker": broker, "q": q, "flag": flag})
    if request.headers.get("hx-request") and request.headers.get("hx-target") == "entries-table":
        return render(request, "app/partials/entries_table.html", **ctx_kw)
    return page(request, ctx, "app/entries.html", "entries", **ctx_kw)


def _entry(ctx: Ctx, entry_id: int) -> Entry:
    e = ctx.db.scalar(select(Entry).where(Entry.org_id == ctx.org.id, Entry.id == entry_id)
                      .options(selectinload(Entry.lines), selectinload(Entry.broker)))
    if e is None:
        raise HTTPException(404)
    return e


@router.get("/entries/{entry_id}", response_class=HTMLResponse)
def entry_detail(request: Request, entry_id: int, ctx: Ctx = Depends(current)):
    e = _entry(ctx, entry_id)
    fs = ctx.db.scalars(select(Finding).where(Finding.entry_id == e.id).order_by(Finding.amount.desc())).all()
    dl = entry_deadlines(EntryFacts(e.entry_number, e.entry_date, e.summary_date, e.liquidation_date, e.liquidation_status))
    measures = sorted(set(e.paid_by_measure or {}) | set(e.correct_by_measure or {}),
                      key=lambda m: list(engine.MEASURES).index(m) if m in engine.MEASURES else 99)
    return page(request, ctx, "app/entry.html", "entries", e=e, findings=fs, dl=dl, measures=measures,
                status_of=engine.status)


@router.post("/entries/{entry_id}/audit")
def entry_reaudit(request: Request, entry_id: int, ctx: Ctx = Depends(current)):
    e = _entry(ctx, entry_id)
    audit.audit(ctx.db, ctx.org, e)
    ctx.db.commit()
    return RedirectResponse(f"/app/entries/{entry_id}", 303)


# ── Recovery ──────────────────────────────────────────────────────────────────

@router.get("/recovery", response_class=HTMLResponse)
def recovery(request: Request, ctx: Ctx = Depends(current)):
    fs = queries.open_findings(ctx.db, ctx.org)
    groups: dict[str, list[Finding]] = {}
    for f in fs:
        groups.setdefault(f.route, []).append(f)
    order = ["CAPE", "PROTEST", "PSC", "CLASS", "CIT", "PRESERVE", "DISCLOSURE", "CLOSED"]
    grouped = [(r, queries.ROUTE_LABELS.get(r, r), groups[r]) for r in order if r in groups]
    case_rows = ctx.db.scalars(select(cases.Case).where(cases.Case.org_id == ctx.org.id).order_by(cases.Case.created_at.desc())).all()
    d = queries.dashboard(ctx.db, ctx.org)
    bars = charts.single_bars([(r["label"], r["amount"]) for r in d["by_route"]], title="Open findings by route")
    return page(request, ctx, "app/recovery.html", "recovery", grouped=grouped, cases=case_rows, d=d, bars=bars)


@router.post("/cases")
async def create_case(request: Request, ctx: Ctx = Depends(current)):
    form = await request.form()
    ids = [int(x) for x in form.getlist("finding") if str(x).isdigit()]
    case = cases.create(ctx.db, ctx.org, ctx.user, ids)
    if case is None:
        return RedirectResponse("/app/recovery?empty=1", 303)
    return RedirectResponse(f"/app/cases/{case.id}", 303)


@router.get("/cases/{case_id}", response_class=HTMLResponse)
def case_detail(request: Request, case_id: int, ctx: Ctx = Depends(current)):
    case = cases.get(ctx.db, ctx.org, case_id)
    if case is None:
        raise HTTPException(404)
    return page(request, ctx, "app/case.html", "recovery", case=case, statuses=cases.STATUSES,
                fee=cases.fee_for(case, ctx.org.plan))


@router.post("/cases/{case_id}/status")
def case_status(request: Request, case_id: int, status: str = Form(...), recovered: str = Form(""),
                ctx: Ctx = Depends(current)):
    case = cases.get(ctx.db, ctx.org, case_id)
    if case is None:
        raise HTTPException(404)
    amount = float(recovered) if recovered.strip() else None
    cases.set_status(ctx.db, ctx.org, ctx.user, case, status, amount)
    return RedirectResponse(f"/app/cases/{case_id}", 303)


@router.get("/cases/{case_id}/cape.csv")
def case_cape(case_id: int, ctx: Ctx = Depends(current)):
    case = cases.get(ctx.db, ctx.org, case_id)
    if case is None:
        raise HTTPException(404)
    return Response(cases.cape_csv(case, ctx.org), media_type="text/csv",
                    headers={"Content-Disposition": f"attachment; filename=cape-case-{case_id}.csv"})


@router.get("/cases/{case_id}/packet", response_class=HTMLResponse)
def case_packet(request: Request, case_id: int, ctx: Ctx = Depends(current)):
    case = cases.get(ctx.db, ctx.org, case_id)
    if case is None:
        raise HTTPException(404)
    return render(request, "app/packet.html", case=case, org=ctx.org, user=ctx.user)


@router.post("/findings/{finding_id}/dismiss")
def dismiss(request: Request, finding_id: int, ctx: Ctx = Depends(current)):
    f = ctx.db.scalar(select(Finding).where(Finding.org_id == ctx.org.id, Finding.id == finding_id))
    if f is None:
        raise HTTPException(404)
    f.status = "dismissed"
    ctx.db.add(Activity(org_id=ctx.org.id, actor=ctx.user.name, action=f"Dismissed finding on {f.entry.entry_number}."))
    ctx.db.commit()
    if request.headers.get("hx-request"):
        return HTMLResponse("")
    return RedirectResponse("/app/recovery", 303)


# ── Deadlines ─────────────────────────────────────────────────────────────────

@router.get("/deadlines", response_class=HTMLResponse)
def deadlines(request: Request, ctx: Ctx = Depends(current)):
    items = queries.deadlines(ctx.db, ctx.org, 365)
    weeks: dict[date, list[dict]] = {}
    for it in items:
        monday = it["date"] - timedelta(days=it["date"].weekday())
        weeks.setdefault(monday, []).append(it)
    return page(request, ctx, "app/deadlines.html", "deadlines", weeks=sorted(weeks.items()), count=len(items))


@router.get("/deadlines.ics")
def deadlines_ics(ctx: Ctx = Depends(current)):
    lines = ["BEGIN:VCALENDAR", "VERSION:2.0", "PRODID:-//Harbormaster//Deadlines//EN", "CALSCALE:GREGORIAN"]
    for i, it in enumerate(queries.deadlines(ctx.db, ctx.org, 365)):
        d = it["date"].strftime("%Y%m%d")
        summary = f"{it['kind']}: {it['entry']} (${it['amount']:,.0f})".replace(",", "\\,")
        lines += ["BEGIN:VEVENT", f"UID:hm-{ctx.org.id}-{i}-{d}@harbormaster", f"DTSTART;VALUE=DATE:{d}",
                  f"SUMMARY:{summary}", f"DESCRIPTION:{it['title']}".replace(",", "\\,"), "END:VEVENT"]
    lines.append("END:VCALENDAR")
    return Response("\r\n".join(lines) + "\r\n", media_type="text/calendar",
                    headers={"Content-Disposition": "attachment; filename=harbormaster-deadlines.ics"})


# ── Catalog economics, calculator, sourcing ───────────────────────────────────

@router.get("/catalog", response_class=HTMLResponse)
def catalog_page(request: Request, ctx: Ctx = Depends(current), sort: str = "margin"):
    rows = queries.sku_economics(ctx.db, ctx.org)
    key = {"margin": lambda r: r["margin"], "duty": lambda r: -r["annual_duty"], "rate": lambda r: -r["rate"],
           "sku": lambda r: r["sku"].sku}.get(sort, lambda r: r["margin"])
    rows.sort(key=key)
    annual = sum(r["annual_duty"] for r in rows)
    below = [r for r in rows if r["below_target"]]
    gap = sum(max(0.0, r["hold_price"] - r["sku"].unit_price) * r["sku"].annual_units for r in below)
    return page(request, ctx, "app/catalog.html", "catalog", rows=rows, annual=annual, below=below, gap=gap, sort=sort)


@router.post("/skus/{sku_id}/price")
def sku_price(request: Request, sku_id: int, price: float = Form(...), ctx: Ctx = Depends(current)):
    s = ctx.db.scalar(select(Sku).where(Sku.org_id == ctx.org.id, Sku.id == sku_id))
    if s is None or price <= 0:
        raise HTTPException(404)
    old = s.unit_price
    s.unit_price = round(price, 2)
    ctx.db.add(Activity(org_id=ctx.org.id, actor=ctx.user.name, action=f"Repriced {s.sku} from ${old:,.2f} to ${s.unit_price:,.2f}."))
    ctx.db.commit()
    return RedirectResponse("/app/catalog", 303)


def _calc_inputs(form) -> dict:
    def f(name, default=""):
        return (form.get(name) or default).strip()
    on = date.fromisoformat(f("entry_date")) if f("entry_date") else engine.today()
    loaded = date.fromisoformat(f("loaded_date")) if f("loaded_date") else None
    share = f("metal_share")
    return dict(hts=f("hts", "6110.20.20"), origin=f("origin", "CN").upper(), unit_cost=float(f("unit_cost", "10") or 10),
                units=max(1, int(float(f("units", "1000") or 1000))), entry_date=on, loaded_date=loaded,
                metal_share=float(share) / 100 if share else None, usmca=f("usmca") == "on",
                patented=f("patented") == "on", mode=f("mode", "ocean"), freight=float(f("freight", "0") or 0))


def _calc(inputs: dict) -> dict:
    sh = engine.Shipment(hts=inputs["hts"], origin=inputs["origin"], value=inputs["unit_cost"] * inputs["units"],
                         entry_date=inputs["entry_date"], loaded_date=inputs["loaded_date"],
                         metal_share=inputs["metal_share"], usmca=inputs["usmca"], fta=inputs["usmca"],
                         flags=frozenset({"patented"} if inputs["patented"] else ()), mode=inputs["mode"],
                         quantity=inputs["units"])
    return engine.landed_cost(sh, freight=inputs["freight"])


@router.get("/calculator", response_class=HTMLResponse)
def calculator(request: Request, ctx: Ctx = Depends(current), hts: str = "6110.20.20", origin: str = "CN"):
    inputs = _calc_inputs({"hts": hts, "origin": origin})
    return page(request, ctx, "app/calculator.html", "calculator", inputs=inputs, lc=_calc(inputs),
                lines=catalog.LINES, countries=COUNTRIES, status_of=engine.status)


@router.post("/calculator", response_class=HTMLResponse)
async def calculator_post(request: Request, ctx: Ctx = Depends(current)):
    form = await request.form()
    try:
        inputs = _calc_inputs(form)
    except ValueError:
        raise HTTPException(422, "Check the numbers and dates.")
    return render(request, "app/partials/stack.html", inputs=inputs, lc=_calc(inputs), status_of=engine.status)


@router.get("/sourcing", response_class=HTMLResponse)
def sourcing_page(request: Request, ctx: Ctx = Depends(current), sku: str = "", scenario: str = "none"):
    skus = ctx.db.scalars(select(Sku).where(Sku.org_id == ctx.org.id).order_by(Sku.sku)).all()
    chosen = next((s for s in skus if s.sku == sku), None) or next((s for s in skus if s.sku == "LP-DUT-5"), skus[0])
    result = _sourcing(chosen, scenario)
    kw = dict(skus=skus, chosen=chosen, scenario=scenario, scenarios=sourcing.SCENARIOS, **result)
    if request.headers.get("hx-request"):
        return render(request, "app/partials/sourcing_result.html", **kw)
    return page(request, ctx, "app/sourcing.html", "sourcing", **kw)


def _sourcing(s: Sku, scenario: str) -> dict:
    opts = sourcing.compare(s.hts, s.origin, s.unit_cost, 1000, s.cbm_per_unit, engine.today(),
                            metal_share=s.metal_share, usmca=s.usmca, scenario=scenario)
    series = [{"key": "cost", "label": "Unit cost"}, {"key": "freight", "label": "Freight"},
              {"key": "duty", "label": "Duty & fees"}, {"key": "scenario", "label": "Scenario add-on"}]
    vals = [{"cost": o.unit_cost, "freight": o.freight_per_unit, "duty": o.duty_per_unit + o.fees_per_unit,
             "scenario": o.scenario_per_unit} for o in opts]
    hi = next((i for i, o in enumerate(opts) if o.current), None)
    chart = charts.stacked_rows([o.name for o in opts], series, vals, highlight=hi, title="Landed cost per unit by origin")
    current_opt = opts[hi] if hi is not None else None
    best = opts[0]
    saving = (current_opt.landed_per_unit - best.landed_per_unit) * s.annual_units if current_opt else 0
    return {"options": opts, "chart": chart, "series": series, "current": current_opt, "best": best, "saving": saving}


# ── Classification ────────────────────────────────────────────────────────────

@router.get("/classify", response_class=HTMLResponse)
def classify_page(request: Request, ctx: Ctx = Depends(current)):
    history = ctx.db.scalars(select(Classification).where(Classification.org_id == ctx.org.id)
                             .order_by(Classification.created_at.desc()).limit(10)).all()
    skus = ctx.db.scalars(select(Sku).where(Sku.org_id == ctx.org.id).order_by(Sku.sku)).all()
    return page(request, ctx, "app/classify.html", "classify", history=history, skus=skus, countries=COUNTRIES)


@router.post("/classify", response_class=HTMLResponse)
def classify_post(request: Request, description: str = Form(...), origin: str = Form("CN"), sku: str = Form(""),
                  ctx: Ctx = Depends(current)):
    limiter.check(request, "classify", 20, 60)
    description = description.strip()[:4000]
    if len(description) < 3:
        raise HTTPException(422, "Describe the product.")
    result = ai_classify.classify(description, origin)
    rec = Classification(org_id=ctx.org.id, description=description, origin=origin.upper()[:2], result=result,
                         engine=result["engine"])
    ctx.db.add(rec)
    ctx.db.commit()
    skus = ctx.db.scalars(select(Sku).where(Sku.org_id == ctx.org.id).order_by(Sku.sku)).all()
    return render(request, "app/partials/classify_result.html", rec=rec, r=result, skus=skus, preselect=sku)


@router.post("/classify/{rec_id}/apply")
def classify_apply(request: Request, rec_id: int, code: str = Form(...), sku_id: int = Form(...),
                   ctx: Ctx = Depends(current)):
    rec = ctx.db.scalar(select(Classification).where(Classification.org_id == ctx.org.id, Classification.id == rec_id))
    s = ctx.db.scalar(select(Sku).where(Sku.org_id == ctx.org.id, Sku.id == sku_id))
    if rec is None or s is None or catalog.lookup(code) is None:
        raise HTTPException(404)
    old = s.hts
    s.hts = code
    s.classified_by = "ai+review" if rec.engine == "claude" else "harbormaster"
    rec.chosen = code
    rec.reviewed_by = ctx.user.name
    ctx.db.add(Activity(org_id=ctx.org.id, actor=ctx.user.name, action=f"Reclassified {s.sku} from {old} to {code}."))
    ctx.db.commit()
    return RedirectResponse("/app/catalog", 303)


# ── Tariff changes, drawback, brokers ─────────────────────────────────────────

def _sku_dicts(ctx: Ctx) -> list[dict]:
    return [{"sku": s.sku, "description": s.description, "hts": s.hts, "origin": s.origin, "unit_cost": s.unit_cost,
             "annual_units": s.annual_units, "metal_share": s.metal_share, "usmca": s.usmca, "flags": s.flags or []}
            for s in ctx.db.scalars(select(Sku).where(Sku.org_id == ctx.org.id))]


@router.get("/changes", response_class=HTMLResponse)
def changes(request: Request, ctx: Ctx = Depends(current)):
    skus = _sku_dicts(ctx)
    events = [(ev, timeline.impact(ev, skus)) for ev in reversed(timeline.EVENTS)]
    return page(request, ctx, "app/changes.html", "changes", events=events, today_=engine.today())


@router.get("/drawback", response_class=HTMLResponse)
def drawback_page(request: Request, ctx: Ctx = Depends(current)):
    claim = queries.drawback_claim(ctx.db, ctx.org)
    by_sku: dict[str, dict] = {}
    for m in claim.matches:
        row = by_sku.setdefault(m.export.hts8 + m.export.ref[:5], {"hts": m.export.hts8, "units": 0, "refund": 0.0, "exports": set()})
        row["units"] += m.quantity
        row["refund"] += m.refund
        row["exports"].add(m.export.ref)
    skus = {s.hts: s for s in ctx.db.scalars(select(Sku).where(Sku.org_id == ctx.org.id))}
    return page(request, ctx, "app/drawback.html", "drawback", claim=claim, rows=sorted(by_sku.values(), key=lambda r: -r["refund"]),
                skus=skus)


@router.get("/brokers", response_class=HTMLResponse)
def brokers(request: Request, ctx: Ctx = Depends(current)):
    return page(request, ctx, "app/brokers.html", "brokers", cards=queries.broker_scorecards(ctx.db, ctx.org))


# ── Copilot ───────────────────────────────────────────────────────────────────

SUGGESTIONS = [
    "What can we recover right now, and what's due this month?",
    "Which SKUs are below our target margin at today's duty rates?",
    "Would LP-DUT-5 be cheaper from Vietnam or India?",
    "Landed cost of 6404.11.90 from VN at $14?",
    "What changed in tariffs since January and what did it cost us?",
]


@router.get("/copilot", response_class=HTMLResponse)
def copilot_page(request: Request, ctx: Ctx = Depends(current)):
    msgs = ctx.db.scalars(select(Message).where(Message.org_id == ctx.org.id, Message.user_id == ctx.user.id)
                          .order_by(Message.created_at.desc()).limit(20)).all()[::-1]
    return page(request, ctx, "app/copilot.html", "copilot", messages=msgs, suggestions=SUGGESTIONS)


@router.post("/copilot", response_class=HTMLResponse)
def copilot_post(request: Request, question: str = Form(...), ctx: Ctx = Depends(current)):
    limiter.check(request, "copilot", 20, 60)
    question = question.strip()[:2000]
    if not question:
        return HTMLResponse("")
    history = [{"role": m.role, "content": m.content} for m in ctx.db.scalars(
        select(Message).where(Message.org_id == ctx.org.id, Message.user_id == ctx.user.id)
        .order_by(Message.created_at.desc()).limit(8)).all()[::-1]]
    ws = copilot.Workspace(ctx.db, ctx.org)
    out = copilot.answer(ws, question, history)
    q = Message(org_id=ctx.org.id, user_id=ctx.user.id, role="user", content=question)
    a = Message(org_id=ctx.org.id, user_id=ctx.user.id, role="assistant", content=out["text"],
                meta={"engine": out["engine"], "tools": out["tools"]})
    ctx.db.add_all([q, a])
    ctx.db.commit()
    return render(request, "app/partials/copilot_pair.html", q=q, a=a)


# ── Upload, settings, billing ─────────────────────────────────────────────────

@router.get("/upload", response_class=HTMLResponse)
def upload_page(request: Request, ctx: Ctx = Depends(current)):
    return page(request, ctx, "app/upload.html", "entries", columns=importer.COLUMNS, result=None)


@router.post("/upload", response_class=HTMLResponse)
async def upload(request: Request, file: UploadFile = File(...), ctx: Ctx = Depends(current)):
    data = await file.read(20 * 1024 * 1024 + 1)
    if len(data) > 20 * 1024 * 1024:
        raise HTTPException(413, "Files up to 20 MB.")
    result = importer.import_csv(ctx.db, ctx.org, data)
    ctx.db.add(Activity(org_id=ctx.org.id, actor=ctx.user.name,
                        action=f"Uploaded {file.filename}: {result.lines} lines, {result.findings} new findings."))
    ctx.db.commit()
    return page(request, ctx, "app/upload.html", "entries", columns=importer.COLUMNS, result=result)


@router.get("/upload/sample.csv")
def sample(ctx: Ctx = Depends(current)):
    return Response(importer.sample_csv(ctx.db, ctx.org), media_type="text/csv",
                    headers={"Content-Disposition": "attachment; filename=harbormaster-sample-entries.csv"})


@router.get("/settings", response_class=HTMLResponse)
def settings(request: Request, ctx: Ctx = Depends(current)):
    ledger = cases.billing_ledger(ctx.db, ctx.org)
    return page(request, ctx, "app/settings.html", "settings", ledger=ledger, new_key=request.session.pop("new_key", None),
                fees=sum(r["fee"] for r in ledger))


@router.post("/settings")
def settings_post(request: Request, target_margin: float = Form(...), cit_plaintiff: str = Form(""),
                  ctx: Ctx = Depends(current)):
    ctx.org.target_margin = max(0.0, min(95.0, target_margin))
    was = ctx.org.cit_plaintiff
    ctx.org.cit_plaintiff = cit_plaintiff == "on"
    ctx.db.commit()
    if was != ctx.org.cit_plaintiff:
        audit.audit_all(ctx.db, ctx.org)
    return RedirectResponse("/app/settings", 303)


@router.post("/settings/api-key")
def rotate_key(request: Request, ctx: Ctx = Depends(current)):
    key, digest = new_api_key()
    ctx.org.api_key_hash = digest
    ctx.db.add(Activity(org_id=ctx.org.id, actor=ctx.user.name, action="Rotated the API key."))
    ctx.db.commit()
    request.session["new_key"] = key
    return RedirectResponse("/app/settings", 303)
