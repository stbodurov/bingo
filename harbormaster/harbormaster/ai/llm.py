"""Thin wrapper around the Claude API.

Harbormaster never depends on the model for a number: duty, deadlines and refund
routes come from the deterministic engine. The model is used for language — reading a
product description, explaining a stack, answering questions by calling the engine as
tools. Every caller has a deterministic fallback, so the product works with no API key
(HARBORMASTER_AI=off forces that mode).
"""

from __future__ import annotations

import json
import logging
import os
from typing import Any

log = logging.getLogger("harbormaster.ai")

MODEL = os.environ.get("HARBORMASTER_MODEL", "claude-opus-5-5")
FALLBACK_BETA = "server-side-fallback-2026-07-01"


def enabled() -> bool:
    if os.environ.get("HARBORMASTER_AI", "").lower() in ("off", "0", "false"):
        return False
    return bool(os.environ.get("ANTHROPIC_API_KEY") or os.environ.get("ANTHROPIC_AUTH_TOKEN"))


_client = None


def client():
    global _client
    if _client is None:
        import anthropic
        _client = anthropic.Anthropic(timeout=60.0, max_retries=2)
    return _client


def create(**kwargs) -> Any | None:
    """messages.create with refusal fallbacks opted in. Returns None on any API failure
    so callers drop to their deterministic path."""
    import anthropic
    try:
        resp = client().beta.messages.create(
            model=MODEL, betas=[FALLBACK_BETA], fallbacks="default", **kwargs)
    except anthropic.APIStatusError as e:
        log.warning("Claude API error %s: %s", e.status_code, e.message)
        return None
    except anthropic.APIConnectionError as e:
        log.warning("Claude API unreachable: %s", e)
        return None
    if resp.stop_reason == "refusal":
        log.info("Claude declined the request (%s)", getattr(resp.stop_details, "category", None))
        return None
    return resp


def text_of(resp) -> str:
    return "".join(b.text for b in resp.content if b.type == "text").strip()


def structured(system: str, user: str, schema: dict, *, effort: str = "medium", max_tokens: int = 4000) -> dict | None:
    """One call that must return JSON matching `schema` (structured outputs)."""
    if not enabled():
        return None
    resp = create(
        max_tokens=max_tokens, system=system, messages=[{"role": "user", "content": user}],
        output_config={"effort": effort, "format": {"type": "json_schema", "schema": schema}},
    )
    if resp is None or resp.stop_reason == "max_tokens":
        return None
    try:
        return json.loads(text_of(resp))
    except (json.JSONDecodeError, StopIteration):
        log.warning("Structured output did not parse")
        return None
