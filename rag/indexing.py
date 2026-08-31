"""Shared knowledge-base loading and chunking for vector + sparse indexes."""

from __future__ import annotations

import os
from typing import List

from langchain_core.documents import Document
from langchain_text_splitters import RecursiveCharacterTextSplitter

from rag.document_loaders import is_supported_filename, load_file

CHUNK_SIZE = 500
CHUNK_OVERLAP = 50


def load_kb_documents() -> List[Document]:
    """Load all KB documents from Postgres registry (materialized to temp files)."""
    from rag.kb.documents import list_all_active_documents, materialize_to_temp

    all_docs: List[Document] = []
    for target in list_all_active_documents():
        if not is_supported_filename(target.logical_key):
            continue
        path = materialize_to_temp(target)
        try:
            all_docs.extend(load_file(target.domain_id, path, tenant_id=target.tenant_id))
        finally:
            os.remove(path)
    return all_docs


def _make_splitter() -> RecursiveCharacterTextSplitter:
    return RecursiveCharacterTextSplitter(
        chunk_size=CHUNK_SIZE,
        chunk_overlap=CHUNK_OVERLAP,
    )


def split_kb_documents() -> List[Document]:
    """Load all KB files, split into chunks, assign stable chunk_id metadata."""
    all_docs = load_kb_documents()
    if not all_docs:
        return []

    chunks = _make_splitter().split_documents(all_docs)
    _assign_chunk_ids(chunks)
    return chunks


def split_file_documents(domain_id: str, file_path: str, tenant_id: str = "default") -> List[Document]:
    """Load and chunk a single KB file (incremental index updates).

    Chunk ids match a full rebuild: sequence numbers are counted per
    (tenant, domain, filename), so re-splitting one file in isolation
    yields identical chunk_id values.
    """
    docs = load_file(domain_id, file_path, tenant_id=tenant_id)
    chunks = _make_splitter().split_documents(docs)
    _assign_chunk_ids(chunks)
    return chunks


def _assign_chunk_ids(chunks: List[Document]) -> None:
    """Stable id: {tenant}/{domain}/{filename}/{seq} — shared by dense and sparse indexes."""
    counters: dict[tuple[str, str, str], int] = {}
    for doc in chunks:
        meta = doc.metadata or {}
        tenant = str(meta.get("tenant_id") or "default")
        domain = str(meta.get("domain_id") or "default")
        filename = str(meta.get("filename") or meta.get("source_file") or "unknown")
        key = (tenant, domain, filename)
        seq = counters.get(key, 0)
        counters[key] = seq + 1
        meta["chunk_id"] = f"{tenant}/{domain}/{filename}/{seq}"
        doc.metadata = meta


def document_key(doc: Document) -> str:
    """Lookup key for RRF fusion; prefers chunk_id from metadata."""
    meta = doc.metadata or {}
    chunk_id = meta.get("chunk_id")
    if chunk_id:
        return str(chunk_id)
    tenant = str(meta.get("tenant_id") or "default")
    domain = str(meta.get("domain_id") or "default")
    filename = str(meta.get("filename") or meta.get("source_file") or "unknown")
    page = meta.get("page", "")
    content = (doc.page_content or "")[:120]
    return f"{tenant}/{domain}/{filename}/{page}:{hash(content)}"
