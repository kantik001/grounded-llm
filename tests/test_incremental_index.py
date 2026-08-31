"""Tests for incremental index bookkeeping (manifest scan + embedding signature)."""

from __future__ import annotations

from pathlib import Path

import pytest
from rag.kb.documents import DocumentTarget
from rag.vector_backend import chroma_backend as cb

_ROOT = Path(__file__).resolve().parents[1]


@pytest.fixture(autouse=True)
def _domains_config(monkeypatch):
    monkeypatch.setenv("DOMAINS_CONFIG_PATH", str(_ROOT / "config" / "domains.json"))


def _sample_targets() -> list[DocumentTarget]:
    return [
        DocumentTarget(
            document_id="doc-1",
            version_id="ver-1",
            tenant_id="acme",
            domain_id="default",
            logical_key="handbook.txt",
            content_sha256="sha-handbook",
            storage_key="tenants/acme/domains/default/docs/doc-1/v1/sha-handbook.txt",
        ),
        DocumentTarget(
            document_id="doc-2",
            version_id="ver-2",
            tenant_id="acme",
            domain_id="default",
            logical_key="faq.txt",
            content_sha256="sha-faq",
            storage_key="tenants/acme/domains/default/docs/doc-2/v1/sha-faq.txt",
        ),
    ]


def test_scan_kb_files_keys_and_hashes(monkeypatch):
    monkeypatch.setattr("rag.kb.documents.list_all_active_documents", lambda: _sample_targets())

    state = cb.scan_kb_files()
    assert set(state) == {"acme/default/handbook.txt", "acme/default/faq.txt"}
    entry = state["acme/default/handbook.txt"]
    assert entry["tenant"] == "acme"
    assert entry["domain"] == "default"
    assert entry["sha1"] == "sha-handbook"
    assert entry["sha256"] == "sha-handbook"

    updated = _sample_targets()
    updated[0] = DocumentTarget(
        document_id="doc-1",
        version_id="ver-3",
        tenant_id="acme",
        domain_id="default",
        logical_key="handbook.txt",
        content_sha256="sha-handbook-v2",
        storage_key="tenants/acme/domains/default/docs/doc-1/v2/sha-handbook-v2.txt",
    )
    monkeypatch.setattr("rag.kb.documents.list_all_active_documents", lambda: updated)
    state2 = cb.scan_kb_files()
    assert state2["acme/default/handbook.txt"]["sha1"] != entry["sha1"]
    assert state2["acme/default/faq.txt"]["sha1"] == state["acme/default/faq.txt"]["sha1"]


def test_scan_kb_files_ignores_unsupported(monkeypatch):
    targets = _sample_targets() + [
        DocumentTarget(
            document_id="doc-3",
            version_id="ver-3",
            tenant_id="acme",
            domain_id="default",
            logical_key="notes.exe",
            content_sha256="sha-exe",
            storage_key="tenants/acme/domains/default/docs/doc-3/v1/sha-exe.exe",
        )
    ]
    monkeypatch.setattr("rag.kb.documents.list_all_active_documents", lambda: targets)

    state = cb.scan_kb_files()
    assert "acme/default/notes.exe" not in state


def test_embedding_signature_tracks_prefix_flag(monkeypatch):
    monkeypatch.delenv("RAG_E5_PREFIXES", raising=False)
    sig_on = cb.embedding_signature()
    assert sig_on["model"] == cb.EMBEDDING_MODEL
    assert sig_on["e5_prefixes"] is True

    monkeypatch.setenv("RAG_E5_PREFIXES", "0")
    sig_off = cb.embedding_signature()
    assert sig_off["e5_prefixes"] is False
    assert sig_on != sig_off
