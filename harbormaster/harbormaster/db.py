"""Persistence: SQLAlchemy 2 models for importers, their catalog, entries, findings and
recovery cases, plus the public waitlist."""

from __future__ import annotations

import json
import os
from datetime import date, datetime, timezone

from sqlalchemy import (JSON, Boolean, Date, DateTime, Float, ForeignKey, Integer, String, Text,
                        UniqueConstraint, create_engine, event)
from sqlalchemy.orm import DeclarativeBase, Mapped, Session, mapped_column, relationship, sessionmaker


def utcnow() -> datetime:
    return datetime.now(timezone.utc).replace(tzinfo=None)


class Base(DeclarativeBase):
    type_annotation_map = {dict: JSON, list: JSON}


class Org(Base):
    __tablename__ = "orgs"
    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(String(200))
    slug: Mapped[str] = mapped_column(String(80), unique=True)
    plan: Mapped[str] = mapped_column(String(20), default="growth")   # free | growth | scale
    ior_last4: Mapped[str] = mapped_column(String(4), default="")
    cit_plaintiff: Mapped[bool] = mapped_column(Boolean, default=False)
    target_margin: Mapped[float] = mapped_column(Float, default=55.0)
    api_key_hash: Mapped[str] = mapped_column(String(64), default="")
    created_at: Mapped[datetime] = mapped_column(DateTime, default=utcnow)

    users: Mapped[list[User]] = relationship(back_populates="org")


class User(Base):
    __tablename__ = "users"
    id: Mapped[int] = mapped_column(primary_key=True)
    org_id: Mapped[int] = mapped_column(ForeignKey("orgs.id"))
    email: Mapped[str] = mapped_column(String(200), unique=True)
    name: Mapped[str] = mapped_column(String(120))
    role: Mapped[str] = mapped_column(String(20), default="admin")   # admin | analyst | broker
    password_hash: Mapped[str] = mapped_column(String(200))
    created_at: Mapped[datetime] = mapped_column(DateTime, default=utcnow)

    org: Mapped[Org] = relationship(back_populates="users")


class Broker(Base):
    __tablename__ = "brokers"
    id: Mapped[int] = mapped_column(primary_key=True)
    org_id: Mapped[int] = mapped_column(ForeignKey("orgs.id"), index=True)
    name: Mapped[str] = mapped_column(String(120))
    filer_code: Mapped[str] = mapped_column(String(3))


class Supplier(Base):
    __tablename__ = "suppliers"
    id: Mapped[int] = mapped_column(primary_key=True)
    org_id: Mapped[int] = mapped_column(ForeignKey("orgs.id"), index=True)
    name: Mapped[str] = mapped_column(String(160))
    country: Mapped[str] = mapped_column(String(2))
    incoterm: Mapped[str] = mapped_column(String(3), default="FOB")


class Sku(Base):
    __tablename__ = "skus"
    __table_args__ = (UniqueConstraint("org_id", "sku"),)
    id: Mapped[int] = mapped_column(primary_key=True)
    org_id: Mapped[int] = mapped_column(ForeignKey("orgs.id"), index=True)
    sku: Mapped[str] = mapped_column(String(60))
    description: Mapped[str] = mapped_column(String(300))
    hts: Mapped[str] = mapped_column(String(12))
    origin: Mapped[str] = mapped_column(String(2))
    supplier_id: Mapped[int | None] = mapped_column(ForeignKey("suppliers.id"), nullable=True)
    unit_cost: Mapped[float] = mapped_column(Float)          # FOB, USD
    unit_price: Mapped[float] = mapped_column(Float)         # selling price
    cbm_per_unit: Mapped[float] = mapped_column(Float, default=0.01)
    annual_units: Mapped[int] = mapped_column(Integer, default=0)
    metal_share: Mapped[float | None] = mapped_column(Float, nullable=True)
    usmca: Mapped[bool] = mapped_column(Boolean, default=False)
    flags: Mapped[list] = mapped_column(JSON, default=list)
    classified_by: Mapped[str] = mapped_column(String(20), default="broker")  # broker | harbormaster | ai+review

    supplier: Mapped[Supplier | None] = relationship()


class Entry(Base):
    __tablename__ = "entries"
    __table_args__ = (UniqueConstraint("org_id", "entry_number"),)
    id: Mapped[int] = mapped_column(primary_key=True)
    org_id: Mapped[int] = mapped_column(ForeignKey("orgs.id"), index=True)
    entry_number: Mapped[str] = mapped_column(String(16))
    broker_id: Mapped[int | None] = mapped_column(ForeignKey("brokers.id"), nullable=True)
    entry_type: Mapped[str] = mapped_column(String(2), default="01")
    mode: Mapped[str] = mapped_column(String(10), default="ocean")
    port: Mapped[str] = mapped_column(String(4), default="2704")
    origin: Mapped[str] = mapped_column(String(2))
    loaded_date: Mapped[date | None] = mapped_column(Date, nullable=True)
    entry_date: Mapped[date] = mapped_column(Date, index=True)
    summary_date: Mapped[date] = mapped_column(Date)
    liquidation_status: Mapped[str] = mapped_column(String(14), default="unliquidated")
    liquidation_date: Mapped[date | None] = mapped_column(Date, nullable=True)
    entered_value: Mapped[float] = mapped_column(Float, default=0)
    duty_paid: Mapped[float] = mapped_column(Float, default=0)
    fees_paid: Mapped[float] = mapped_column(Float, default=0)
    duty_correct: Mapped[float] = mapped_column(Float, default=0)
    paid_by_measure: Mapped[dict] = mapped_column(JSON, default=dict)
    correct_by_measure: Mapped[dict] = mapped_column(JSON, default=dict)
    audited_at: Mapped[datetime | None] = mapped_column(DateTime, nullable=True)

    lines: Mapped[list[EntryLine]] = relationship(back_populates="entry", cascade="all, delete-orphan",
                                                  order_by="EntryLine.line_no")
    broker: Mapped[Broker | None] = relationship()


class EntryLine(Base):
    __tablename__ = "entry_lines"
    id: Mapped[int] = mapped_column(primary_key=True)
    entry_id: Mapped[int] = mapped_column(ForeignKey("entries.id"), index=True)
    line_no: Mapped[int] = mapped_column(Integer)
    sku_id: Mapped[int | None] = mapped_column(ForeignKey("skus.id"), nullable=True)
    hts: Mapped[str] = mapped_column(String(12))
    origin: Mapped[str] = mapped_column(String(2))
    description: Mapped[str] = mapped_column(String(300), default="")
    quantity: Mapped[float] = mapped_column(Float, default=0)
    value: Mapped[float] = mapped_column(Float)
    metal_share: Mapped[float | None] = mapped_column(Float, nullable=True)
    usmca: Mapped[bool] = mapped_column(Boolean, default=False)
    paid: Mapped[dict] = mapped_column(JSON, default=dict)        # measure -> USD as filed
    correct: Mapped[dict] = mapped_column(JSON, default=dict)     # measure -> USD per engine
    explanation: Mapped[list] = mapped_column(JSON, default=list) # exclusions & warnings

    entry: Mapped[Entry] = relationship(back_populates="lines")
    sku: Mapped[Sku | None] = relationship()


class Finding(Base):
    """A dollar amount Harbormaster believes is recoverable or at risk, with the route."""
    __tablename__ = "findings"
    id: Mapped[int] = mapped_column(primary_key=True)
    org_id: Mapped[int] = mapped_column(ForeignKey("orgs.id"), index=True)
    entry_id: Mapped[int] = mapped_column(ForeignKey("entries.id"), index=True)
    kind: Mapped[str] = mapped_column(String(14))      # refund | overpayment | underpayment | contested
    measure: Mapped[str] = mapped_column(String(14))
    route: Mapped[str] = mapped_column(String(12))
    title: Mapped[str] = mapped_column(String(200))
    detail: Mapped[str] = mapped_column(Text)
    authority: Mapped[str] = mapped_column(String(300))
    amount: Mapped[float] = mapped_column(Float)
    interest: Mapped[float] = mapped_column(Float, default=0)
    deadline: Mapped[date | None] = mapped_column(Date, nullable=True)
    urgency: Mapped[str] = mapped_column(String(8))
    status: Mapped[str] = mapped_column(String(12), default="open")   # open | in_case | recovered | dismissed
    case_id: Mapped[int | None] = mapped_column(ForeignKey("cases.id"), nullable=True)

    entry: Mapped[Entry] = relationship()


class Case(Base):
    """A recovery filing that bundles findings (one CAPE declaration, one protest, …)."""
    __tablename__ = "cases"
    id: Mapped[int] = mapped_column(primary_key=True)
    org_id: Mapped[int] = mapped_column(ForeignKey("orgs.id"), index=True)
    route: Mapped[str] = mapped_column(String(12))
    title: Mapped[str] = mapped_column(String(200))
    status: Mapped[str] = mapped_column(String(14), default="draft")  # draft | ready | filed | recovered | rejected
    amount: Mapped[float] = mapped_column(Float, default=0)
    recovered: Mapped[float] = mapped_column(Float, default=0)
    created_at: Mapped[datetime] = mapped_column(DateTime, default=utcnow)
    filed_at: Mapped[datetime | None] = mapped_column(DateTime, nullable=True)
    notes: Mapped[str] = mapped_column(Text, default="")

    findings: Mapped[list[Finding]] = relationship(foreign_keys=[Finding.case_id])


class ExportRecord(Base):
    __tablename__ = "exports"
    id: Mapped[int] = mapped_column(primary_key=True)
    org_id: Mapped[int] = mapped_column(ForeignKey("orgs.id"), index=True)
    ref: Mapped[str] = mapped_column(String(40))
    sku_id: Mapped[int] = mapped_column(ForeignKey("skus.id"))
    export_date: Mapped[date] = mapped_column(Date)
    quantity: Mapped[float] = mapped_column(Float)
    destination: Mapped[str] = mapped_column(String(2))

    sku: Mapped[Sku] = relationship()


class Classification(Base):
    __tablename__ = "classifications"
    id: Mapped[int] = mapped_column(primary_key=True)
    org_id: Mapped[int] = mapped_column(ForeignKey("orgs.id"), index=True)
    description: Mapped[str] = mapped_column(Text)
    origin: Mapped[str] = mapped_column(String(2), default="CN")
    result: Mapped[dict] = mapped_column(JSON, default=dict)
    engine: Mapped[str] = mapped_column(String(20))      # claude | catalog
    chosen: Mapped[str | None] = mapped_column(String(12), nullable=True)
    reviewed_by: Mapped[str | None] = mapped_column(String(120), nullable=True)
    created_at: Mapped[datetime] = mapped_column(DateTime, default=utcnow)


class Message(Base):
    __tablename__ = "copilot_messages"
    id: Mapped[int] = mapped_column(primary_key=True)
    org_id: Mapped[int] = mapped_column(ForeignKey("orgs.id"), index=True)
    user_id: Mapped[int] = mapped_column(ForeignKey("users.id"))
    role: Mapped[str] = mapped_column(String(10))
    content: Mapped[str] = mapped_column(Text)
    meta: Mapped[dict] = mapped_column(JSON, default=dict)
    created_at: Mapped[datetime] = mapped_column(DateTime, default=utcnow)


class Activity(Base):
    __tablename__ = "activity"
    id: Mapped[int] = mapped_column(primary_key=True)
    org_id: Mapped[int] = mapped_column(ForeignKey("orgs.id"), index=True)
    actor: Mapped[str] = mapped_column(String(120))
    action: Mapped[str] = mapped_column(String(300))
    created_at: Mapped[datetime] = mapped_column(DateTime, default=utcnow)


class WaitlistSignup(Base):
    __tablename__ = "waitlist"
    id: Mapped[int] = mapped_column(primary_key=True)
    email: Mapped[str] = mapped_column(String(200), unique=True)
    company: Mapped[str] = mapped_column(String(200))
    role: Mapped[str] = mapped_column(String(60))
    imports_band: Mapped[str] = mapped_column(String(20))
    top_origins: Mapped[list] = mapped_column(JSON, default=list)
    sells_on: Mapped[list] = mapped_column(JSON, default=list)
    estimate: Mapped[dict] = mapped_column(JSON, default=dict)
    code: Mapped[str] = mapped_column(String(12), unique=True)
    referred_by: Mapped[str | None] = mapped_column(String(12), nullable=True)
    source: Mapped[str] = mapped_column(String(60), default="")
    created_at: Mapped[datetime] = mapped_column(DateTime, default=utcnow)


def make_engine(url: str | None = None):
    url = url or os.environ.get("HARBORMASTER_DB", "sqlite:///data/harbormaster.db")
    if url.startswith("sqlite:///") and not url.startswith("sqlite:///:memory:"):
        os.makedirs(os.path.dirname(url.removeprefix("sqlite:///")) or ".", exist_ok=True)
    kwargs = {"json_serializer": lambda o: json.dumps(o, default=str)}
    if url.startswith("sqlite"):
        kwargs["connect_args"] = {"check_same_thread": False}
        if ":memory:" in url:
            from sqlalchemy.pool import StaticPool
            kwargs["poolclass"] = StaticPool
    engine = create_engine(url, **kwargs)
    if url.startswith("sqlite"):
        @event.listens_for(engine, "connect")
        def _pragmas(conn, _):  # noqa: ANN001
            cur = conn.cursor()
            cur.execute("PRAGMA foreign_keys=ON")
            if ":memory:" not in url:
                cur.execute("PRAGMA journal_mode=WAL")
            cur.close()
    Base.metadata.create_all(engine)
    return engine


def session_factory(engine) -> sessionmaker[Session]:
    return sessionmaker(engine, expire_on_commit=False)
