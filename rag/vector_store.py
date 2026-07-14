# Vector store facade — delegates to pluggable backend (Chroma default, Qdrant optional).

import os

from rag.domains_config import normalize_domain_id
from rag.kb_discovery import DEFAULT_TENANT
from rag.rerank import rerank_documents, reranker_mode
from rag.rrf import reciprocal_rank_fusion
from rag.sparse_index import ensure_sparse_index, reset_sparse_index
from rag.vector_backend import get_vector_backend, reset_vector_backend
from rag.vector_backend.chroma_backend import DEFAULT_PERSIST_DIR

PERSIST_DIR = DEFAULT_PERSIST_DIR


def reset_vector_store():
    reset_vector_backend()
    reset_sparse_index()


def load_all_documents():
    from rag.indexing import load_kb_documents

    return load_kb_documents()


def create_vector_store():
    backend = get_vector_backend()
    backend.load(force_reindex=True)
    ensure_sparse_index(force_reindex=True)
    return backend


def load_vector_store(force_reindex: bool = False):
    backend = get_vector_backend()
    backend.load(force_reindex=force_reindex)
    ensure_sparse_index(force_reindex=force_reindex)
    return backend


def retrieval_mode() -> str:
    """vector | hybrid (BM25 dense + sparse + RRF)."""
    return (os.environ.get("RAG_RETRIEVAL_MODE") or "vector").strip().lower()


def _rrf_k() -> int:
    raw = (os.environ.get("RAG_RRF_K") or "60").strip()
    try:
        return max(1, int(raw))
    except ValueError:
        return 60


def _fetch_multiplier(k: int, *, hybrid: bool, rerank: str) -> int:
    if hybrid:
        return min(max(k * 3, k), 40)
    if rerank != "none":
        return min(max(k * 2, k), 20)
    return k


def _hybrid_search(
    query: str,
    *,
    domain_id: str,
    tenant_id: str,
    k: int,
    fetch_k: int,
    rerank: str,
):
    backend = get_vector_backend()
    sparse = ensure_sparse_index()
    dense_hits = backend.similarity_search(
        query,
        k=fetch_k,
        domain_id=domain_id,
        tenant_id=tenant_id,
    )
    sparse_hits = sparse.search(
        query,
        domain_id=domain_id,
        tenant_id=tenant_id,
        k=fetch_k,
    )
    fused = reciprocal_rank_fusion(
        dense_hits,
        sparse_hits,
        k=fetch_k,
        rrf_k=_rrf_k(),
    )
    if rerank != "none" and len(fused) > k:
        return rerank_documents(query, fused, k, mode=rerank)
    return fused[:k]


def search(query: str, domain_id: str, tenant_id: str = DEFAULT_TENANT, k: int = 8):
    domain_id = normalize_domain_id(domain_id)
    tenant_id = (tenant_id or DEFAULT_TENANT).strip().lower() or DEFAULT_TENANT
    backend = get_vector_backend()
    rerank = reranker_mode()
    hybrid = retrieval_mode() == "hybrid"

    if hybrid:
        fetch_k = _fetch_multiplier(k, hybrid=True, rerank=rerank)
        return _hybrid_search(
            query,
            domain_id=domain_id,
            tenant_id=tenant_id,
            k=k,
            fetch_k=fetch_k,
            rerank=rerank,
        )

    fetch_k = _fetch_multiplier(k, hybrid=False, rerank=rerank)
    results = backend.similarity_search(
        query,
        k=fetch_k,
        domain_id=domain_id,
        tenant_id=tenant_id,
    )
    if rerank != "none" and len(results) > k:
        return rerank_documents(query, results, k, mode=rerank)
    return results[:k]


def index_stats_for_domain(domain_id: str, tenant_id: str = DEFAULT_TENANT) -> list[dict]:
    """Chunk counts per source file for a domain (admin index status)."""
    domain_id = normalize_domain_id(domain_id)
    tenant_id = (tenant_id or DEFAULT_TENANT).strip().lower() or DEFAULT_TENANT
    backend = get_vector_backend()
    return backend.index_stats_for_domain(domain_id, tenant_id)
