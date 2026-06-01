# Chroma vector store: documents under data/{tenant_id}/{domain_id}/*.{txt,pdf,docx}
import glob
import os
from typing import Iterator, Tuple

from langchain_chroma import Chroma
from langchain_huggingface import HuggingFaceEmbeddings
from langchain_text_splitters import RecursiveCharacterTextSplitter

from rag.document_loaders import load_file, supported_extensions
from rag.domains_config import list_domains, normalize_domain_id

_PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
DATA_DIR = os.path.join(_PROJECT_ROOT, "data")
PERSIST_DIR = os.path.join(_PROJECT_ROOT, "chroma_db")
DEFAULT_TENANT = os.environ.get("DEFAULT_TENANT_ID", "default")

_vector_store = None


def reset_vector_store():
    global _vector_store
    _vector_store = None


def _has_kb_files(path: str) -> bool:
    if not os.path.isdir(path):
        return False
    for ext in supported_extensions():
        if glob.glob(os.path.join(path, f"*{ext}")):
            return True
    return False


def discover_kb_directories() -> Iterator[Tuple[str, str, str]]:
    """Yield (tenant_id, domain_id, directory_path). Supports legacy data/{domain_id}/ layout."""
    if not os.path.isdir(DATA_DIR):
        return
    domain_ids = set(list_domains().get("domains", {}).keys())

    for name in sorted(os.listdir(DATA_DIR)):
        path = os.path.join(DATA_DIR, name)
        if not os.path.isdir(path):
            continue
        if name in domain_ids and _has_kb_files(path):
            yield DEFAULT_TENANT, name, path
            continue
        for domain_id in sorted(os.listdir(path)):
            dpath = os.path.join(path, domain_id)
            if os.path.isdir(dpath) and _has_kb_files(dpath):
                yield name, domain_id, dpath


def load_all_documents():
    all_docs = []
    for tenant_id, domain_id, domain_dir in discover_kb_directories():
        for ext in supported_extensions():
            for file_path in glob.glob(os.path.join(domain_dir, f"*{ext}")):
                all_docs.extend(load_file(domain_id, file_path, tenant_id=tenant_id))
    return all_docs


def create_vector_store():
    print("Creating vector store...")
    documents = load_all_documents()
    if not documents:
        print("No documents to index.")
        return None
    text_splitter = RecursiveCharacterTextSplitter(chunk_size=500, chunk_overlap=50)
    docs = text_splitter.split_documents(documents)
    print(f"Chunks: {len(docs)}")
    embeddings = HuggingFaceEmbeddings(model_name="intfloat/multilingual-e5-small")
    store = Chroma.from_documents(docs, embeddings, persist_directory=PERSIST_DIR)
    print(f"Vector store saved to {PERSIST_DIR}")
    return store


def load_vector_store(force_reindex: bool = False):
    global _vector_store
    if _vector_store is not None and not force_reindex:
        return _vector_store

    force = force_reindex or os.environ.get("FORCE_RAG_REINDEX", "").lower() in (
        "1",
        "true",
        "yes",
    )
    embeddings = HuggingFaceEmbeddings(model_name="intfloat/multilingual-e5-small")

    if force and os.path.isdir(PERSIST_DIR):
        import shutil

        print("FORCE_RAG_REINDEX: removing old chroma_db")
        shutil.rmtree(PERSIST_DIR, ignore_errors=True)

    if os.path.exists(PERSIST_DIR) and os.listdir(PERSIST_DIR):
        _vector_store = Chroma(persist_directory=PERSIST_DIR, embedding_function=embeddings)
    else:
        _vector_store = create_vector_store()
    return _vector_store


def search(query: str, domain_id: str, tenant_id: str = DEFAULT_TENANT, k: int = 8):
    domain_id = normalize_domain_id(domain_id)
    tenant_id = (tenant_id or DEFAULT_TENANT).strip().lower() or DEFAULT_TENANT
    store = load_vector_store()
    if store is None:
        return []
    return store.similarity_search(
        query,
        k=k,
        filter={"domain_id": domain_id, "tenant_id": tenant_id},
    )


def index_stats_for_domain(domain_id: str, tenant_id: str = DEFAULT_TENANT) -> list[dict]:
    """Chunk counts per source file for a domain (admin index status)."""
    domain_id = normalize_domain_id(domain_id)
    tenant_id = (tenant_id or DEFAULT_TENANT).strip().lower() or DEFAULT_TENANT
    store = load_vector_store()
    if store is None:
        return []
    try:
        data = store._collection.get(  # noqa: SLF001
            where={"$and": [{"domain_id": domain_id}, {"tenant_id": tenant_id}]},
            include=["metadatas"],
        )
    except Exception:
        try:
            data = store._collection.get(
                where={"domain_id": domain_id, "tenant_id": tenant_id},
                include=["metadatas"],
            )
        except Exception:
            return []
    counts: dict[str, int] = {}
    for meta in data.get("metadatas") or []:
        if not meta:
            continue
        fn = meta.get("filename") or meta.get("source_file") or "unknown"
        counts[fn] = counts.get(fn, 0) + 1
    return [{"filename": name, "chunks": n} for name, n in sorted(counts.items())]
