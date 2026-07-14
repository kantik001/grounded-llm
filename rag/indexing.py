"""Shared knowledge-base loading and chunking for vector + sparse indexes."""

from __future__ import annotations

import glob
import os
from typing import List

from langchain_core.documents import Document
from langchain_text_splitters import RecursiveCharacterTextSplitter

from rag.document_loaders import load_file, supported_extensions
from rag.kb_discovery import discover_kb_directories

CHUNK_SIZE = 500
CHUNK_OVERLAP = 50


def load_kb_documents() -> List[Document]:
    """Load all KB files without chunking (admin / legacy helpers)."""
    all_docs: List[Document] = []
    for tenant_id, domain_id, domain_dir in discover_kb_directories():
        for ext in supported_extensions():
            for file_path in glob.glob(os.path.join(domain_dir, f"*{ext}")):
                all_docs.extend(load_file(domain_id, file_path, tenant_id=tenant_id))
    return all_docs


def split_kb_documents() -> List[Document]:
    """Load all KB files, split into chunks, assign stable chunk_id metadata."""
    all_docs = load_kb_documents()
    if not all_docs:
        return []

    splitter = RecursiveCharacterTextSplitter(
        chunk_size=CHUNK_SIZE,
        chunk_overlap=CHUNK_OVERLAP,
    )
    chunks = splitter.split_documents(all_docs)
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
