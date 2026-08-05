"""HTTP conformance tests against a running Go server."""

from __future__ import annotations

import json
import os
from pathlib import Path

import pytest
import requests

ROOT = Path(__file__).resolve().parents[1]
OPENAPI_PATH = ROOT / "server" / "openapi.v1.json"

BASE_URL = os.environ.get("CONFORMANCE_BASE_URL", "").rstrip("/")
SKIP = os.environ.get("CONFORMANCE_SKIP_HTTP", "") == "1" or not BASE_URL

pytestmark = pytest.mark.skipif(SKIP, reason="Set CONFORMANCE_BASE_URL to run HTTP conformance")


def _public_get_paths():
    """GET paths that explicitly set security: [] (opt out of global API auth)."""
    with OPENAPI_PATH.open(encoding="utf-8") as f:
        spec = json.load(f)
    out = []
    for path, methods in spec.get("paths", {}).items():
        get_op = methods.get("get")
        if not get_op:
            continue
        # Missing security inherits global ApiKey/Telegram — not a public probe target.
        if get_op.get("security") != []:
            continue
        out.append(path)
    return out


@pytest.mark.parametrize("path", _public_get_paths())
def test_public_get_returns_2xx(path):
    url = f"{BASE_URL}{path}"
    if "{" in path:
        pytest.skip(f"path params not auto-filled: {path}")
    resp = requests.get(url, timeout=15)
    # OpenAPI lists both "/" and "/api/v1" servers; some paths exist only under v1.
    if resp.status_code == 404 and not path.startswith("/api"):
        resp = requests.get(f"{BASE_URL}/api/v1{path}", timeout=15)
    assert 200 <= resp.status_code < 300, f"{path} returned {resp.status_code}: {resp.text[:200]}"


def test_health_contract():
    resp = requests.get(f"{BASE_URL}/health", timeout=10)
    assert resp.status_code == 200
    body = resp.json()
    assert body.get("status") in ("healthy", "degraded")


def test_ready_contract():
    resp = requests.get(f"{BASE_URL}/ready", timeout=10)
    assert resp.status_code in (200, 503)
    body = resp.json()
    assert "status" in body
    assert "checks" in body


def test_v1_openapi_json():
    resp = requests.get(f"{BASE_URL}/api/v1/openapi.json", timeout=10)
    assert resp.status_code == 200
    spec = resp.json()
    assert spec.get("openapi", "").startswith("3.0")
    assert "paths" in spec


def test_session_message_history_contract():
    """Core chat paths from OpenAPI (requires TELEGRAM_AUTH_DISABLED or valid auth)."""
    session = requests.post(
        f"{BASE_URL}/session",
        json={"domain_id": "default"},
        timeout=15,
    )
    assert session.status_code == 200, session.text[:300]
    sbody = session.json()
    assert sbody.get("success") is True
    assert isinstance(sbody.get("session_id"), str) and sbody["session_id"]
    session_id = sbody["session_id"]
    if "domain_id" in sbody:
        assert sbody["domain_id"] == "default"

    message = requests.post(
        f"{BASE_URL}/message",
        json={
            "session_id": session_id,
            "domain_id": "default",
            "text": "How many paid vacation days do employees get?",
        },
        timeout=60,
    )
    assert message.status_code == 200, message.text[:400]
    mbody = message.json()
    assert mbody.get("success") is True
    assert mbody.get("session_id") == session_id
    assert isinstance(mbody.get("messages"), list)
    assert len(mbody["messages"]) >= 1

    history = requests.get(
        f"{BASE_URL}/history",
        params={"session_id": session_id},
        timeout=15,
    )
    assert history.status_code == 200, history.text[:300]
    hbody = history.json()
    assert hbody.get("success") is True
    assert hbody.get("session_id") == session_id
    assert isinstance(hbody.get("messages"), list)
    assert len(hbody["messages"]) >= 1
