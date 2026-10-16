"""Application factory."""

from __future__ import annotations

import html
import logging
import os
import secrets
from pathlib import Path

from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse, JSONResponse, Response
from fastapi.staticfiles import StaticFiles
from starlette.exceptions import HTTPException as StarletteHTTPException
from starlette.middleware.sessions import SessionMiddleware

from ..db import make_engine, session_factory
from ..services.seed import demo_api_key, seed_demo
from . import api, public, workspace
from .deps import HERE, render

CSP = ("default-src 'self'; script-src 'self'; style-src 'self' https://fonts.googleapis.com; style-src-attr 'unsafe-inline'; "
       "font-src 'self' https://fonts.gstatic.com; img-src 'self' data:; connect-src 'self'; "
       "frame-ancestors 'none'; base-uri 'self'; form-action 'self'")


def _secret() -> str:
    if os.environ.get("HARBORMASTER_SECRET"):
        return os.environ["HARBORMASTER_SECRET"]
    path = Path(os.environ.get("HARBORMASTER_DATA", "data")) / "session.key"
    path.parent.mkdir(parents=True, exist_ok=True)
    if not path.exists():
        path.write_text(secrets.token_urlsafe(48))
        path.chmod(0o600)
    return path.read_text().strip()


def create_app(db_url: str | None = None, *, seed: bool = True, secret: str | None = None) -> FastAPI:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s: %(message)s")
    engine = make_engine(db_url)
    app = FastAPI(title="Harbormaster API", version="1.0", docs_url="/api/docs", redoc_url=None,
                  openapi_url="/api/openapi.json")
    app.state.db = session_factory(engine)
    app.state.rate_limit_off = os.environ.get("RATE_LIMIT", "on") == "off"
    if seed:
        with app.state.db() as s:
            org = seed_demo(s)
            if not org.api_key_hash:
                org.api_key_hash = demo_api_key()[1]
                s.commit()

    @app.middleware("http")
    async def headers(request: Request, call_next):
        resp: Response = await call_next(request)
        resp.headers.setdefault("Content-Security-Policy", CSP)
        resp.headers["X-Content-Type-Options"] = "nosniff"
        resp.headers["Referrer-Policy"] = "same-origin"
        resp.headers["X-Frame-Options"] = "DENY"
        if request.url.path.startswith("/app"):
            resp.headers["Cache-Control"] = "no-store"
        return resp

    app.add_middleware(SessionMiddleware, secret_key=secret or _secret(), session_cookie="hm_session",
                       same_site="lax", max_age=14 * 24 * 3600)
    app.mount("/static", StaticFiles(directory=str(HERE / "static")), name="static")
    app.include_router(public.router)
    app.include_router(workspace.router)
    app.include_router(api.router)

    @app.exception_handler(StarletteHTTPException)
    async def http_error(request: Request, exc: StarletteHTTPException):
        if exc.status_code in (301, 302, 303, 307) or (exc.headers or {}).get("HX-Redirect"):
            return Response(status_code=exc.status_code, headers=exc.headers)
        if request.url.path.startswith("/api/"):
            return JSONResponse({"error": exc.detail}, status_code=exc.status_code, headers=exc.headers)
        if request.headers.get("hx-request"):
            return HTMLResponse(f'<p class="form-error" role="alert">{html.escape(str(exc.detail))}</p>',
                                status_code=exc.status_code)
        return render(request, "error.html", status_code=exc.status_code, code=exc.status_code, detail=exc.detail)

    @app.get("/healthz")
    def healthz():
        return {"ok": True}

    return app
