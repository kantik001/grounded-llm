# Vector store facade — delegates to pluggable backend (Chroma default, Qdrant optional).

from rag.domains_config import normalize_domain_id
from rag.kb_discovery import DEFAULT_TENANT
from rag.rerank import rerank_documents, reranker_mode
from rag.vector_backend import get_vector_backend, reset_vector_backend
from rag.vector_backend.chroma_backend import DEFAULT_PERSIST_DIR

PERSIST_DIR = DEFAULT_PERSIST_DIR


def reset_vector_store():
    reset_vector_backend()


def load_all_documents():
    from rag.vector_backend.chroma_backend import _load_all_documents

    return _load_all_documents()


def create_vector_store():
    backend = get_vector_backend()
    backend.load(force_reindex=True)
    return backend


def load_vector_store(force_reindex: bool = False):
    backend = get_vector_backend()
    backend.load(force_reindex=force_reindex)
    return backend


def _fetch_multiplier(k: int, rerank: str) -> int:
    if rerank != "none":
        return min(max(k * 2, k), 20)
    return k


def search(query: str, domain_id: str, tenant_id: str = DEFAULT_TENANT, k: int = 8):
    domain_id = normalize_domain_id(domain_id)
    tenant_id = (tenant_id or DEFAULT_TENANT).strip().lower() or DEFAULT_TENANT
    backend = get_vector_backend()
    rerank = reranker_mode()
    fetch_k = _fetch_multiplier(k, rerank)
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
