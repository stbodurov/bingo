"""Request plumbing shared by the routers: DB sessions, templates, auth, CSRF, rate limits."""

from __future__ import annotations

import time
from collections import defaultdict, deque
from pathlib import Path
from typing import Iterator

from fastapi import Depends, HTTPException, Request
from fastapi.templating import Jinja2Templates
from sqlalchemy import select
from sqlalchemy.orm import Session

from ..db import Org, User
from ..security import new_token, same
from . import fmt

HERE = Path(__file__).parent
templates = Jinja2Templates(directory=str(HERE / "templates"))
fmt.register(templates.env)


def get_db(request: Request) -> Iterator[Session]:
    session = request.app.state.db()
    try:
        yield session
    finally:
        session.close()


def csrf_token(request: Request) -> str:
    tok = request.session.get("csrf")
    if not tok:
        tok = new_token()
        request.session["csrf"] = tok
    return tok


async def csrf_protect(request: Request) -> None:
    if request.method in ("GET", "HEAD", "OPTIONS"):
        return
    sent = request.headers.get("x-csrf-token")
    if not sent:
        ctype = request.headers.get("content-type", "")
        if "form" in ctype:
            form = await request.form()
            sent = form.get("csrf")
    if not same(sent, request.session.get("csrf")):
        raise HTTPException(403, "Invalid or missing CSRF token")


class Ctx:
    def __init__(self, user: User, org: Org, db: Session):
        self.user, self.org, self.db = user, org, db


def current(request: Request, db: Session = Depends(get_db)) -> Ctx:
    uid = request.session.get("uid")
    user = db.get(User, uid) if uid else None
    if user is None:
        target = "/login?next=" + request.url.path
        if request.headers.get("hx-request"):
            raise HTTPException(401, headers={"HX-Redirect": target})
        raise HTTPException(303, headers={"Location": target})
    return Ctx(user, db.get(Org, user.org_id), db)


def render(request: Request, name: str, status_code: int = 200, **ctx):
    ctx.setdefault("active", "")
    return templates.TemplateResponse(request, name, {"csrf": csrf_token(request), **ctx}, status_code=status_code)


class RateLimiter:
    """Sliding-window limiter keyed by client IP and bucket name."""

    def __init__(self):
        self.hits: dict[tuple[str, str], deque] = defaultdict(deque)

    def check(self, request: Request, bucket: str, limit: int, window: float) -> None:
        if request.app.state.rate_limit_off:
            return
        ip = request.client.host if request.client else "?"
        q = self.hits[(bucket, ip)]
        now = time.monotonic()
        while q and now - q[0] > window:
            q.popleft()
        if len(q) >= limit:
            raise HTTPException(429, "Too many requests — try again shortly.")
        q.append(now)


limiter = RateLimiter()


def user_by_email(db: Session, email: str) -> User | None:
    return db.scalar(select(User).where(User.email == email.strip().lower()))
