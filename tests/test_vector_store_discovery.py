"""Tests for KB registry manifest scan."""

from __future__ import annotations

from rag.kb.documents import DocumentTarget, scan_registry_documents


def test_scan_registry_documents(monkeypatch):
    targets = [
        DocumentTarget(
            document_id="d1",
            version_id="v1",
            tenant_id="acme",
            domain_id="default",
            logical_key="handbook.txt",
            content_sha256="abc123",
            storage_key="key1",
        )
    ]
    monkeypatch.setattr("rag.kb.documents.list_all_active_documents", lambda: targets)
    state = scan_registry_documents()
    assert list(state) == ["acme/default/handbook.txt"]
    assert state["acme/default/handbook.txt"]["storage_key"] == "key1"
