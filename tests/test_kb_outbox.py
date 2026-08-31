"""Tests for KB ingest outbox (auto-enqueue)."""

from __future__ import annotations

import uuid

import pytest


@pytest.fixture
def registry_env(monkeypatch, tmp_path):
    monkeypatch.setenv("KB_BLOB_BACKEND", "local")
    monkeypatch.setenv("KB_BLOB_DIR", str(tmp_path / "blobs"))
    monkeypatch.setenv("DATABASE_URL", "postgresql://test:test@127.0.0.1:5432/test")
    from rag.storage import blob_store

    blob_store.reset_blob_store()
    yield
    blob_store.reset_blob_store()


def test_auto_ingest_disabled_by_default(monkeypatch):
    monkeypatch.delenv("KB_AUTO_INGEST", raising=False)
    from rag.kb.outbox import auto_ingest_enabled

    assert auto_ingest_enabled() is False


def test_auto_ingest_enabled_truthy(monkeypatch):
    monkeypatch.setenv("KB_AUTO_INGEST", "1")
    from rag.kb.outbox import auto_ingest_enabled

    assert auto_ingest_enabled() is True


def test_enqueue_outbox_tx_skips_when_disabled(monkeypatch):
    monkeypatch.setenv("KB_AUTO_INGEST", "0")

    class FakeCursor:
        def execute(self, *_a, **_k):
            raise AssertionError("should not write outbox when disabled")

    from rag.kb.outbox import enqueue_outbox_tx

    enqueue_outbox_tx(
        FakeCursor(),
        tenant_id="default",
        domain_id="hr",
        document_id=str(uuid.uuid4()),
        version_id=str(uuid.uuid4()),
        logical_key="policy.txt",
        content_sha256="abc",
        source="upload",
    )


def test_enqueue_outbox_tx_inserts_when_enabled(monkeypatch):
    monkeypatch.setenv("KB_AUTO_INGEST", "1")
    calls: list[tuple] = []

    class FakeCursor:
        def execute(self, sql, params):
            calls.append((sql, params))

    from rag.kb.outbox import enqueue_outbox_tx

    doc_id = str(uuid.uuid4())
    ver_id = str(uuid.uuid4())
    enqueue_outbox_tx(
        FakeCursor(),
        tenant_id="default",
        domain_id="hr",
        document_id=doc_id,
        version_id=ver_id,
        logical_key="policy.txt",
        content_sha256="deadbeef",
        source="google_drive",
    )
    assert len(calls) == 1
    assert "kb_ingest_outbox" in calls[0][0]
    assert calls[0][1][2] == doc_id


def test_flush_outbox_calls_ingest_http(monkeypatch):
    monkeypatch.setenv("KB_AUTO_INGEST", "1")
    monkeypatch.setenv("GROUNDED_SERVER_URL", "http://127.0.0.1:8080")

    from rag.kb import outbox as outbox_mod

    monkeypatch.setattr(outbox_mod, "_claim_pending", lambda _t, _d: ([1, 2], ["a.txt", "b.txt"], "google_drive"))
    monkeypatch.setattr(outbox_mod, "_finish_outbox", lambda *_a, **_k: None)

    captured: dict = {}

    def fake_trigger(**kwargs):
        captured.update(kwargs)
        return {"job_id": 99, "already_running": False}

    monkeypatch.setattr(outbox_mod, "trigger_ingest_http", fake_trigger)

    result = outbox_mod.flush_outbox(tenant_id="default", domain_id="hr")
    assert result.flushed == 2
    assert result.job_id == 99
    assert captured["files"] == ["a.txt", "b.txt"]
    assert captured["source"] == "google_drive"
