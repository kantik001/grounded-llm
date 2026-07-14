"""Tests for shared KB indexing helpers."""

from langchain_core.documents import Document

from rag.indexing import document_key, split_kb_documents


def test_document_key_uses_chunk_id():
    doc = Document(page_content="hello", metadata={"chunk_id": "default/default/a.txt/0"})
    assert document_key(doc) == "default/default/a.txt/0"


def test_split_kb_documents_assigns_chunk_ids(monkeypatch):
    sample = Document(
        page_content="x" * 600,
        metadata={"filename": "policy.txt", "domain_id": "default", "tenant_id": "default"},
    )

    monkeypatch.setattr("rag.indexing.load_kb_documents", lambda: [sample])
    chunks = split_kb_documents()
    assert len(chunks) >= 2
    assert all(c.metadata.get("chunk_id") for c in chunks)
