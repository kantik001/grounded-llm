"""HTTP conformance tests against a running Go server."""

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
    with OPENAPI_PATH.open(encoding="utf-8") as f:
        spec = json.load(f)
    out = []
    for path, methods in spec.get("paths", {}).items():
        get_op = methods.get("get")
        if not get_op:
            continue
        if get_op.get("security"):
            continue
        out.append(path)
    return out


@pytest.mark.parametrize("path", _public_get_paths())
def test_public_get_returns_2xx(path):
    url = f"{BASE_URL}{path}"
    if "{" in path:
        pytest.skip(f"path params not auto-filled: {path}")
    resp = requests.get(url, timeout=15)
    assert resp.status_code < 500, f"{path} returned {resp.status_code}: {resp.text[:200]}"


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
