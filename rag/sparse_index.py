"""BM25 sparse index for lexical retrieval (hybrid search with dense vectors)."""

from __future__ import annotations

import heapq
import os
import pickle
import re

from langchain_core.documents import Document
from rank_bm25 import BM25Plus

from rag.indexing import split_kb_documents
from rag.kb.index_collections import sparse_run_dir
from rag.kb.index_runs import resolve_run_id

_PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
_DEFAULT_DIR = os.path.join(_PROJECT_ROOT, "sparse_index")
_PERSIST_FILE = "bm25_index.pkl"
_INDEX_VERSION = 2

_sparse_indexes: dict[str, BM25SparseIndex] = {}


def _tokenize(text: str) -> list[str]:
    return re.findall(r"[a-zа-яё0-9]+", (text or "").lower())


def _base_sparse_dir() -> str:
    return (os.environ.get("SPARSE_INDEX_DIR") or _DEFAULT_DIR).strip() or _DEFAULT_DIR


def _scope_key(tenant_id: str, domain_id: str, run_id: str) -> str:
    return f"{tenant_id}/{domain_id}/{run_id}"


def _persist_path(tenant_id: str, domain_id: str, run_id: str) -> str:
    run_dir = sparse_run_dir(_base_sparse_dir(), tenant_id, domain_id, run_id)
    return os.path.join(run_dir, _PERSIST_FILE)


def _filter_chunks_for_scope(
    chunks: list[Document],
    tenant_id: str,
    domain_id: str,
) -> list[Document]:
    tenant = (tenant_id or "default").strip().lower()
    domain = (domain_id or "default").strip().lower()
    out: list[Document] = []
    for doc in chunks:
        meta = doc.metadata or {}
        t = str(meta.get("tenant_id") or "default").strip().lower()
        d = str(meta.get("domain_id") or "default").strip().lower()
        if t == tenant and d == domain:
            out.append(doc)
    return out


class BM25SparseIndex:
    """In-memory BM25 index for one tenant/domain index run."""

    def __init__(self, *, tenant_id: str, domain_id: str, run_id: str) -> None:
        self.tenant_id = tenant_id
        self.domain_id = domain_id
        self.run_id = run_id
        self._chunks: list[Document] = []
        self._bm25: BM25Plus | None = None
        self._indices: list[int] = []

    def is_ready(self) -> bool:
        return self._bm25 is not None and bool(self._indices)

    def reset(self) -> None:
        self._chunks = []
        self._bm25 = None
        self._indices = []

    def build(self, chunks: list[Document] | None = None, *, persist: bool = True) -> None:
        if chunks is None:
            all_chunks = split_kb_documents()
            chunks = _filter_chunks_for_scope(all_chunks, self.tenant_id, self.domain_id)
        self._chunks = list(chunks)
        self._rebuild_index()
        if persist:
            self.save()

    def _rebuild_index(self) -> None:
        self._bm25 = None
        self._indices = []
        if not self._chunks:
            return
        corpus = [_tokenize(doc.page_content) for doc in self._chunks]
        if not corpus or all(not row for row in corpus):
            return
        self._bm25 = BM25Plus(corpus)
        self._indices = list(range(len(self._chunks)))

    def save(self) -> None:
        path = _persist_path(self.tenant_id, self.domain_id, self.run_id)
        os.makedirs(os.path.dirname(path), exist_ok=True)
        payload = {
            "version": _INDEX_VERSION,
            "tenant_id": self.tenant_id,
            "domain_id": self.domain_id,
            "run_id": self.run_id,
            "chunks": [
                {"page_content": d.page_content, "metadata": dict(d.metadata or {})}
                for d in self._chunks
            ],
        }
        with open(path, "wb") as fh:
            pickle.dump(payload, fh, protocol=pickle.HIGHEST_PROTOCOL)

    def load(self) -> bool:
        path = _persist_path(self.tenant_id, self.domain_id, self.run_id)
        if not os.path.isfile(path):
            return False
        try:
            with open(path, "rb") as fh:
                payload = pickle.load(fh)
        except (OSError, pickle.UnpicklingError):
            return False
        if payload.get("version") != _INDEX_VERSION:
            return False

        self._chunks = [
            Document(page_content=row["page_content"], metadata=row.get("metadata") or {})
            for row in payload.get("chunks") or []
        ]
        self._rebuild_index()
        return self.is_ready()

    def clear_persisted(self) -> None:
        try:
            os.remove(_persist_path(self.tenant_id, self.domain_id, self.run_id))
        except OSError:
            pass
        self.reset()

    def search(
        self,
        query: str,
        *,
        domain_id: str,
        tenant_id: str,
        k: int,
    ) -> list[Document]:
        q = (query or "").strip()
        if not q or k <= 0 or not self.is_ready() or self._bm25 is None:
            return []

        tokens = _tokenize(q)
        if not tokens:
            return []

        scores = self._bm25.get_scores(tokens)
        ranked = heapq.nlargest(k, zip(self._indices, scores), key=lambda item: item[1])
        out: list[Document] = []
        for idx, score in ranked:
            if score <= 0:
                continue
            out.append(self._chunks[idx])
        return out


def get_sparse_index(tenant_id: str, domain_id: str, run_id: str) -> BM25SparseIndex:
    key = _scope_key(tenant_id, domain_id, run_id)
    if key not in _sparse_indexes:
        _sparse_indexes[key] = BM25SparseIndex(tenant_id=tenant_id, domain_id=domain_id, run_id=run_id)
    return _sparse_indexes[key]


def reset_sparse_index() -> None:
    global _sparse_indexes
    for idx in _sparse_indexes.values():
        idx.reset()
    _sparse_indexes = {}


def ensure_sparse_index(
    *,
    force_reindex: bool = False,
    tenant_id: str | None = None,
    domain_id: str | None = None,
    run_id: str | None = None,
) -> BM25SparseIndex:
    """Load or rebuild BM25 for tenant/domain active index run."""
    tenant = (tenant_id or "default").strip().lower() or "default"
    domain = (domain_id or "default").strip().lower() or "default"
    resolved_run = run_id or resolve_run_id(tenant, domain)

    idx = get_sparse_index(tenant, domain, resolved_run)
    force = force_reindex or os.environ.get("FORCE_RAG_REINDEX", "").lower() in (
        "1",
        "true",
        "yes",
    )
    if force:
        idx.clear_persisted()
        idx.build(persist=True)
        return idx
    if idx.is_ready():
        return idx
    if idx.load():
        return idx
    idx.build(persist=True)
    return idx
