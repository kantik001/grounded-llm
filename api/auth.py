"""Internal auth helpers for Python RAG (HTTP + gRPC).

Shared-secret checks only — no Flask/gRPC types. Callers extract credentials
from headers or metadata, then pass a plain string here.
"""

from __future__ import annotations

import hmac
import os
import threading
from collections.abc import Iterable

# Floor for production shared secrets (brute-force / accidental weak values).
MIN_SECRET_LEN = 16

_lock = threading.Lock()
_rag_token: str = ""
_admin_secret: str = ""
_loaded: bool = False


def reload_secrets() -> None:
    """Re-read ``RAG_SERVICE_TOKEN`` / ``ADMIN_SECRET`` from the environment."""
    global _rag_token, _admin_secret, _loaded
    with _lock:
        _rag_token = (os.environ.get("RAG_SERVICE_TOKEN") or "").strip()
        _admin_secret = (os.environ.get("ADMIN_SECRET") or "").strip()
        _loaded = True


def _ensure_loaded() -> None:
    if not _loaded:
        reload_secrets()


def _expected_rag() -> str:
    _ensure_loaded()
    with _lock:
        return _rag_token


def _expected_admin() -> str:
    _ensure_loaded()
    with _lock:
        return _admin_secret


def _consteq(expected: str, provided: str) -> bool:
    """Constant-time compare; safe when lengths differ on Python 3.11."""
    a = expected.encode("utf-8")
    b = provided.encode("utf-8")
    if len(a) != len(b):
        hmac.compare_digest(a, a)
        return False
    return hmac.compare_digest(a, b)


def extract_bearer(authorization: str | None) -> str:
    """Parse ``Authorization: Bearer <token>``; empty if missing/malformed."""
    raw = (authorization or "").strip()
    if len(raw) >= 7 and raw[:7].lower() == "bearer ":
        return raw[7:].strip()
    return ""


def resolve_rag_token(
    *,
    header_token: str | None = None,
    authorization: str | None = None,
) -> str:
    """Prefer ``X-RAG-Service-Token``, else ``Authorization: Bearer``."""
    direct = (header_token or "").strip()
    if direct:
        return direct
    return extract_bearer(authorization)


def resolve_admin_secret(
    *,
    header_secret: str | None = None,
    authorization: str | None = None,
) -> str:
    """Prefer ``X-Admin-Secret``, else ``Authorization: Bearer``."""
    direct = (header_secret or "").strip()
    if direct:
        return direct
    return extract_bearer(authorization)


def rag_token_from_metadata(metadata: Iterable[tuple[str, str]]) -> str:
    """Extract RAG service token from gRPC invocation metadata."""
    header = ""
    authz = ""
    for key, value in metadata:
        k = key.lower()
        if k == "x-rag-service-token":
            header = value
        elif k == "authorization":
            authz = value
    return resolve_rag_token(header_token=header, authorization=authz)


def rag_service_token_ok(provided: str | None) -> bool:
    """Open when ``RAG_SERVICE_TOKEN`` unset (local/dev). Else require match."""
    expected = _expected_rag()
    if not expected:
        return True
    return _consteq(expected, (provided or "").strip())


def admin_secret_ok(provided: str | None) -> bool:
    """Admin routes require ``ADMIN_SECRET``; empty env → always deny."""
    expected = _expected_admin()
    if not expected:
        return False
    return _consteq(expected, (provided or "").strip())


def validate_production_secrets(*, min_len: int = MIN_SECRET_LEN) -> None:
    """Fail-fast when production secrets are missing or too short."""
    reload_secrets()
    rag = _expected_rag()
    admin = _expected_admin()
    problems: list[str] = []
    if not rag:
        problems.append("RAG_SERVICE_TOKEN must be set when GROUNDED_ENV=production")
    elif len(rag) < min_len:
        problems.append(
            f"RAG_SERVICE_TOKEN must be at least {min_len} characters in production"
        )
    if not admin:
        problems.append("ADMIN_SECRET must be set when GROUNDED_ENV=production")
    elif len(admin) < min_len:
        problems.append(
            f"ADMIN_SECRET must be at least {min_len} characters in production"
        )
    if problems:
        raise RuntimeError(
            "production safety check failed:\n  - " + "\n  - ".join(problems)
        )
