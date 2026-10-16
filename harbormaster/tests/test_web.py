import re
from types import SimpleNamespace

import pytest
from fastapi.testclient import TestClient
from sqlalchemy import select

from harbormaster.ai import classify, copilot, llm
from harbormaster.db import Case, Entry, Finding, Org
from harbormaster.services import importer
from harbormaster.services.seed import demo_api_key
from harbormaster.web.app import create_app

PAGES = ["/app", "/app/entries", "/app/recovery", "/app/deadlines", "/app/catalog", "/app/sourcing",
         "/app/calculator", "/app/classify", "/app/changes", "/app/drawback", "/app/brokers", "/app/copilot",
         "/app/upload", "/app/settings"]


@pytest.fixture()
def app(monkeypatch):
    monkeypatch.setenv("RATE_LIMIT", "off")
    monkeypatch.setenv("HARBORMASTER_AI", "off")
    return create_app("sqlite:///:memory:", secret="t" * 40)


@pytest.fixture()
def client(app):
    return TestClient(app)


def token(c, path="/"):
    return re.search(r'name="csrf-token" content="([^"]+)"', c.get(path).text).group(1)


def login(c):
    t = token(c, "/login")
    assert c.post("/demo", data={"csrf": t}, follow_redirects=False).status_code == 303
    return token(c, "/app")


def test_landing_estimate_and_waitlist(client):
    r = client.get("/")
    assert r.status_code == 200 and "Get early access" in r.text
    assert "script-src 'self'" in r.headers["content-security-policy"]
    t = token(client)
    est = client.post("/estimate", data={"csrf": t, "annual_imports": "3000000", "mix_CN": "70", "mix_VN": "30"})
    assert est.status_code == 200 and "Estimated recoverable" in est.text
    r = client.post("/waitlist", data={"csrf": t, "email": "cfo@example.com", "company": "Example Goods",
                                       "origins": ["CN", "VN"]}, follow_redirects=False)
    assert r.status_code == 303 and r.headers["location"].startswith("/w/")
    done = client.get(r.headers["location"])
    assert "Onboarding position" in done.text and "Example Goods" in done.text
    # duplicate email returns the same status page
    again = client.post("/waitlist", data={"csrf": t, "email": "cfo@example.com", "company": "X"}, follow_redirects=False)
    assert again.headers["location"] == r.headers["location"]
    bad = client.post("/waitlist", data={"csrf": t, "email": "nope", "company": "X"})
    assert bad.status_code == 422


def test_csrf_required(client):
    assert client.post("/estimate", data={"annual_imports": "1"}).status_code == 403
    assert client.post("/demo").status_code == 403


def test_workspace_requires_login(client):
    r = client.get("/app", follow_redirects=False)
    assert r.status_code == 303 and r.headers["location"].startswith("/login")


def test_bad_password(client):
    t = token(client, "/login")
    r = client.post("/login", data={"csrf": t, "email": "demo@larkspur.example", "password": "wrong"})
    assert r.status_code == 401


def test_all_workspace_pages_render(client):
    login(client)
    for p in PAGES:
        r = client.get(p)
        assert r.status_code == 200, p


def test_entry_detail_and_reaudit(app, client):
    t = login(client)
    with app.state.db() as s:
        e = s.scalar(select(Entry).join(Finding).where(Finding.kind == "overpayment").limit(1))
    r = client.get(f"/app/entries/{e.id}")
    assert r.status_code == 200 and e.entry_number in r.text and "Line-by-line stack" in r.text
    assert client.post(f"/app/entries/{e.id}/audit", data={"csrf": t}, follow_redirects=False).status_code == 303
    assert client.get("/app/entries/999999").status_code == 404


def test_create_case_cape_csv_packet_and_status(app, client):
    t = login(client)
    with app.state.db() as s:
        ids = [f.id for f in s.scalars(select(Finding).where(Finding.route == "PROTEST", Finding.status == "open")).all()[:3]]
    r = client.post("/app/cases", data={"csrf": t, "finding": [str(i) for i in ids]}, follow_redirects=False)
    assert r.status_code == 303
    url = r.headers["location"]
    assert client.get(url).status_code == 200
    assert "Statement of reasons" in client.get(url + "/packet").text
    assert client.get(url + "/cape.csv").text.startswith("Entry Number")
    client.post(url + "/status", data={"csrf": t, "status": "recovered", "recovered": "1000"})
    with app.state.db() as s:
        case = s.get(Case, int(url.rsplit("/", 1)[1]))
        assert case.status == "recovered" and case.recovered == 1000
    assert "Success-fee ledger" in client.get("/app/settings").text


def test_calculator_sourcing_classify_copilot(client):
    t = login(client)
    r = client.post("/app/calculator", data={"csrf": t, "hts": "7323.93.00", "origin": "CN", "unit_cost": "5",
                                             "units": "100", "entry_date": "2025-10-01", "metal_share": "60"})
    assert r.status_code == 200 and "Section 232" in r.text
    assert "Landed cost per unit by origin" in client.get("/app/sourcing?sku=LP-HDY-1&scenario=excess_capacity",
                                                         headers={"hx-request": "true"}).text
    r = client.post("/app/classify", data={"csrf": t, "description": "cotton hoodie pullover sweatshirt", "origin": "VN"})
    assert "6110.20.20" in r.text
    r = client.post("/app/copilot", data={"csrf": t, "question": "What can we recover right now?"})
    assert "recoverable now" in r.text
    r = client.post("/app/copilot", data={"csrf": t, "question": "Landed cost of 6404.11.90 from VN at $14?"})
    assert "6404.11.90" in r.text


def test_upload_roundtrip(app, client):
    t = login(client)
    sample = client.get("/app/upload/sample.csv").content
    r = client.post("/app/upload", data={"csrf": t}, files={"file": ("e.csv", sample, "text/csv")})
    assert r.status_code == 200 and "Import complete" in r.text
    new = b"entry_number,entry_date,hts,origin,value,paid_mfn,paid_ieepa_recip\nZZZ-0000001-1,2025-10-01,6110.20.20,VN,10000,1650,2000\n"
    r = client.post("/app/upload", data={"csrf": t}, files={"file": ("n.csv", new, "text/csv")})
    assert "Entries created" in r.text
    with app.state.db() as s:
        e = s.scalar(select(Entry).where(Entry.entry_number == "ZZZ-0000001-1"))
        assert e.correct_by_measure["IEEPA_RECIP"] == 2000.0
        assert {f.route for f in s.scalars(select(Finding).where(Finding.entry_id == e.id))} == {"CAPE"}


def test_importer_rejects_missing_columns(app):
    with app.state.db() as s:
        org = s.scalar(select(Org))
        res = importer.import_csv(s, org, b"foo,bar\n1,2\n")
        assert res.errors and "Missing required columns" in res.errors[0]


def test_api_requires_key_and_computes(client):
    body = {"hts": "6110.20.20", "origin": "CN", "value": 10000, "entry_date": "2025-04-12"}
    assert client.post("/api/v1/stack", json=body).status_code == 401
    key = demo_api_key()[0]
    r = client.post("/api/v1/stack", json=body, headers={"Authorization": f"Bearer {key}"})
    assert r.status_code == 200 and r.json()["effective_rate"] == 169.0
    assert any(c["status"] == "invalidated" for c in r.json()["charges"])
    assert client.get("/api/v1/findings", headers={"Authorization": f"Bearer {key}"}).json()
    assert client.post("/api/v1/estimate", json={"annual_imports": 1e6, "mix": {"CN": 100}}).json()["ieepa_paid"] > 0


def test_admin_waitlist_hidden_without_token(client):
    assert client.get("/admin/waitlist").status_code == 404
    assert client.get("/admin/waitlist?token=dev-admin").status_code == 200


# ── AI paths, with the Claude client stubbed ────────────────────────────────

def _resp(stop, content):
    return SimpleNamespace(stop_reason=stop, content=content, stop_details=None)


def test_classifier_uses_claude_and_attaches_engine_duty(monkeypatch):
    monkeypatch.setattr(llm, "enabled", lambda: True)
    payload = '{"candidates":[{"code":"7323.93.00","confidence":0.82,"gri":"GRI 1; GRI 3(b)","reasoning":"Steel body gives essential character."}],"missing_facts":["Lid material by weight"],"summary":"Stainless household article."}'
    seen = {}

    def fake_create(**kw):
        seen.update(kw)
        return _resp("end_turn", [SimpleNamespace(type="text", text=payload)])
    monkeypatch.setattr(llm, "create", fake_create)
    out = classify.classify("insulated steel bottle", "CN")
    assert out["engine"] == "claude" and out["candidates"][0]["duty_rate"] > 0
    assert seen["output_config"]["format"]["type"] == "json_schema"


def test_classifier_falls_back_on_refusal(monkeypatch):
    monkeypatch.setattr(llm, "enabled", lambda: True)
    monkeypatch.setattr(llm, "create", lambda **kw: None)
    out = classify.classify("leather handbag", "IT")
    assert out["engine"] == "catalog" and out["candidates"][0]["code"] == "4202.21.90"


def test_copilot_tool_loop(monkeypatch, app):
    monkeypatch.setattr(llm, "enabled", lambda: True)
    calls = []

    def fake_create(**kw):
        calls.append(kw)
        if len(calls) == 1:
            return _resp("tool_use", [SimpleNamespace(type="tool_use", id="t1", name="recovery_summary", input={})])
        tool_result = kw["messages"][-1]["content"][0]
        assert tool_result["type"] == "tool_result" and "recoverable_now" in tool_result["content"]
        return _resp("end_turn", [SimpleNamespace(type="text", text="You can recover **$1**.")])
    monkeypatch.setattr(llm, "create", fake_create)
    with app.state.db() as s:
        ws = copilot.Workspace(s, s.scalar(select(Org)))
        out = copilot.answer(ws, "what can we recover?")
    assert out["engine"] == "claude" and out["tools"][0]["tool"] == "recovery_summary"
    assert all(t["strict"] for t in calls[0]["tools"])
