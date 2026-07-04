"""Tests for Python API readiness and internal auth."""

import os
import sys

_root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
sys.path.insert(0, _root)

from api.app import app  # noqa: E402


def test_health_public():
    client = app.test_client()
    resp = client.get("/health")
    assert resp.status_code == 200
    assert resp.get_json()["status"] == "healthy"


def test_ready_without_token_when_unconfigured():
    os.environ.pop("RAG_SERVICE_TOKEN", None)
    client = app.test_client()
    resp = client.get("/ready")
    assert resp.status_code == 200
    body = resp.get_json()
    assert body["status"] == "ready"
    assert body["checks"]["data"] == "ok"


def test_ready_rejects_wrong_token_when_configured():
    os.environ["RAG_SERVICE_TOKEN"] = "secret-token"
    try:
        client = app.test_client()
        resp = client.get("/ready")
        assert resp.status_code == 403
        resp2 = client.get("/ready", headers={"X-RAG-Service-Token": "secret-token"})
        assert resp2.status_code == 200
    finally:
        os.environ.pop("RAG_SERVICE_TOKEN", None)
