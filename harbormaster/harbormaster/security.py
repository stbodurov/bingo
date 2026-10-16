"""Password hashing (scrypt), CSRF tokens and API keys."""

from __future__ import annotations

import base64
import hashlib
import hmac
import secrets

_N, _R, _P = 2**14, 8, 1


def hash_password(password: str) -> str:
    salt = secrets.token_bytes(16)
    dk = hashlib.scrypt(password.encode(), salt=salt, n=_N, r=_R, p=_P, dklen=32)
    return "scrypt$" + base64.b64encode(salt).decode() + "$" + base64.b64encode(dk).decode()


def verify_password(password: str, stored: str) -> bool:
    try:
        scheme, salt_b64, dk_b64 = stored.split("$")
    except ValueError:
        return False
    if scheme != "scrypt":
        return False
    salt, expected = base64.b64decode(salt_b64), base64.b64decode(dk_b64)
    dk = hashlib.scrypt(password.encode(), salt=salt, n=_N, r=_R, p=_P, dklen=len(expected))
    return hmac.compare_digest(dk, expected)


def new_token(n: int = 24) -> str:
    return secrets.token_urlsafe(n)


def same(a: str | None, b: str | None) -> bool:
    return bool(a) and bool(b) and hmac.compare_digest(a, b)


def new_api_key() -> tuple[str, str]:
    key = "hm_live_" + secrets.token_urlsafe(24)
    return key, hashlib.sha256(key.encode()).hexdigest()


def api_key_hash(key: str) -> str:
    return hashlib.sha256(key.encode()).hexdigest()
