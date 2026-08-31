"""Tests for ingest pipeline helpers (no DB/Redis)."""

from __future__ import annotations

from rag.ingest.pipeline import discover_targets, staging_path
from rag.ingest.store import IngestJob
from rag.kb.documents import DocumentTarget


def test_staging_path_sanitizes_file_key():
    path = staging_path("tenant/hr/doc.pdf")
    assert "tenant__hr__doc.pdf" in path


def test_discover_targets_empty(monkeypatch):
    monkeypatch.setattr(
        "rag.ingest.pipeline.discover_document_targets",
        lambda tenant_id, domain_id, files=None: [],
    )
    job = IngestJob(
        id=1,
        status="queued",
        tenant_id="default",
        domain_id="missing",
        source="admin",
        actor="",
        mode="incremental",
        files=[],
    )
    assert discover_targets(job) == []


def test_discover_targets_explicit_list(monkeypatch):
    def fake_discover(tenant_id, domain_id, files=None):
        assert tenant_id == "default"
        assert domain_id == "hr"
        assert files == ["note.txt"]
        return [
            DocumentTarget(
                document_id="doc-1",
                version_id="ver-1",
                tenant_id=tenant_id,
                domain_id=domain_id,
                logical_key="note.txt",
                content_sha256="abc",
                storage_key="tenants/default/domains/hr/docs/doc-1/v1/abc.txt",
            )
        ]

    monkeypatch.setattr("rag.ingest.pipeline.discover_document_targets", fake_discover)
    job = IngestJob(
        id=2,
        status="queued",
        tenant_id="default",
        domain_id="hr",
        source="admin",
        actor="",
        mode="incremental",
        files=["note.txt"],
    )
    targets = discover_targets(job)
    assert len(targets) == 1
    assert targets[0].logical_key == "note.txt"


def test_load_kb_documents_materializes(tmp_path, monkeypatch):
    path = tmp_path / "doc.txt"
    path.write_text("hello policy", encoding="utf-8")
    target = DocumentTarget(
        document_id="d1",
        version_id="v1",
        tenant_id="default",
        domain_id="default",
        logical_key="doc.txt",
        content_sha256="abc",
        storage_key="key",
    )

    monkeypatch.setattr("rag.kb.documents.list_all_active_documents", lambda: [target])
    monkeypatch.setattr("rag.kb.documents.materialize_to_temp", lambda _t: str(path))
    from rag.indexing import load_kb_documents

    docs = load_kb_documents()
    assert len(docs) == 1
    assert "hello" in docs[0].page_content


def test_load_kb_documents_empty(monkeypatch):
    monkeypatch.setattr("rag.kb.documents.list_all_active_documents", lambda: [])
    from rag.indexing import load_kb_documents, split_kb_documents

    assert load_kb_documents() == []
    assert split_kb_documents() == []


def test_split_file_documents_assigns_chunk_id(tmp_path):
    from rag.indexing import split_file_documents

    path = tmp_path / "policy.txt"
    path.write_text(("Vacation policy paragraph. " * 40), encoding="utf-8")
    chunks = split_file_documents("default", str(path), tenant_id="default")
    assert len(chunks) >= 1
    assert chunks[0].metadata.get("chunk_id", "").startswith("default/default/policy.txt/")


def test_document_key_prefers_chunk_id():
    from langchain_core.documents import Document

    from rag.indexing import document_key

    doc = Document(page_content="x", metadata={"chunk_id": "t/d/f/0"})
    assert document_key(doc) == "t/d/f/0"

    fallback = Document(
        page_content="hello",
        metadata={"tenant_id": "acme", "domain_id": "hr", "filename": "a.txt", "page": 1},
    )
    key = document_key(fallback)
    assert key.startswith("acme/hr/a.txt/1:")
    assert document_key(Document(page_content="", metadata={})).startswith("default/default/unknown/")


def test_supported_filenames():
    from rag.document_loaders import is_supported_filename, supported_extensions

    assert ".txt" in supported_extensions()
    assert is_supported_filename("policy.pdf")
    assert not is_supported_filename("virus.exe")
