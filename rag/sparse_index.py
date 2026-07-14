"""BM25 sparse index for lexical retrieval (hybrid search with dense vectors)."""

from __future__ import annotations

import os
import pickle
import re
import shutil
from typing import Any

from langchain_core.documents import Document
from rank_bm25 import BM25Plus

from rag.indexing import split_kb_documents

_PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
_DEFAULT_DIR = os.path.join(_PROJECT_ROOT, "sparse_index")
_PERSIST_FILE = "bm25_index.pkl"
_INDEX_VERSION = 1

_sparse_index: "BM25SparseIndex | None" = None


def _tokenize(text: str) -> list[str]:
    return re.findall(r"[a-zа-яё0-9]+", (text or "").lower())


def _persist_path() -> str:
    base = (os.environ.get("SPARSE_INDEX_DIR") or _DEFAULT_DIR).strip() or _DEFAULT_DIR
    return os.path.join(base, _PERSIST_FILE)


class BM25SparseIndex:
    """In-memory BM25 indexes scoped by (tenant_id, domain_id)."""

    def __init__(self) -> None:
        self._chunks: list[Document] = []
        self._indexes: dict[tuple[str, str], BM25Plus] = {}
        self._scope_indices: dict[tuple[str, str], list[int]] = {}

    def is_ready(self) -> bool:
        return bool(self._indexes)

    def reset(self) -> None:
        self._chunks = []
        self._indexes = {}
        self._scope_indices = {}

    def build(self, chunks: list[Document] | None = None, *, persist: bool = True) -> None:
        chunks = chunks if chunks is not None else split_kb_documents()
        self._chunks = list(chunks)
        self._rebuild_indexes()
        if persist:
            self.save()

    def _rebuild_indexes(self) -> None:
        self._indexes = {}
        self._scope_indices = {}
        if not self._chunks:
            return

        by_scope: dict[tuple[str, str], list[int]] = {}
        for idx, doc in enumerate(self._chunks):
            meta = doc.metadata or {}
            scope = (
                str(meta.get("tenant_id") or "default").lower(),
                str(meta.get("domain_id") or "default").lower(),
            )
            by_scope.setdefault(scope, []).append(idx)

        for scope, indices in by_scope.items():
            corpus = [_tokenize(self._chunks[i].page_content) for i in indices]
            if not corpus or all(not row for row in corpus):
                continue
            self._indexes[scope] = BM25Plus(corpus)
            self._scope_indices[scope] = indices

    def save(self) -> None:
        path = _persist_path()
        os.makedirs(os.path.dirname(path), exist_ok=True)
        payload = {
            "version": _INDEX_VERSION,
            "chunks": [
                {"page_content": d.page_content, "metadata": dict(d.metadata or {})}
                for d in self._chunks
            ],
        }
        with open(path, "wb") as fh:
            pickle.dump(payload, fh, protocol=pickle.HIGHEST_PROTOCOL)

    def load(self) -> bool:
        path = _persist_path()
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
        self._rebuild_indexes()
        return self.is_ready()

    def clear_persisted(self) -> None:
        base = os.path.dirname(_persist_path())
        if os.path.isdir(base):
            shutil.rmtree(base, ignore_errors=True)
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
        if not q or k <= 0:
            return []

        scope = (tenant_id.strip().lower() or "default", domain_id.strip().lower() or "default")
        bm25 = self._indexes.get(scope)
        indices = self._scope_indices.get(scope)
        if bm25 is None or not indices:
            return []

        tokens = _tokenize(q)
        if not tokens:
            return []

        scores = bm25.get_scores(tokens)
        ranked = sorted(
            zip(indices, scores),
            key=lambda item: item[1],
            reverse=True,
        )
        out: list[Document] = []
        for idx, score in ranked[:k]:
            if score <= 0:
                continue
            out.append(self._chunks[idx])
        return out


def get_sparse_index() -> BM25SparseIndex:
    global _sparse_index
    if _sparse_index is None:
        _sparse_index = BM25SparseIndex()
    return _sparse_index


def reset_sparse_index() -> None:
    global _sparse_index
    if _sparse_index is not None:
        _sparse_index.reset()
    _sparse_index = None


def ensure_sparse_index(*, force_reindex: bool = False) -> BM25SparseIndex:
    """Load persisted BM25 index or rebuild from KB chunks."""
    idx = get_sparse_index()
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
