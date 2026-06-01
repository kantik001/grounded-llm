# Chroma vector store: documents under data/{domain_id}/*.{txt,pdf,docx}
import glob
import os

from langchain_chroma import Chroma
from langchain_huggingface import HuggingFaceEmbeddings
from langchain_text_splitters import RecursiveCharacterTextSplitter

from rag.document_loaders import load_file, supported_extensions
from rag.domains_config import list_domains, normalize_domain_id

_PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
DATA_DIR = os.path.join(_PROJECT_ROOT, "data")
PERSIST_DIR = os.path.join(_PROJECT_ROOT, "chroma_db")

_vector_store = None


def reset_vector_store():
    global _vector_store
    _vector_store = None


def load_all_documents():
    all_docs = []
    domains = list_domains().get("domains", {}).keys()

    for domain_id in domains:
        domain_dir = os.path.join(DATA_DIR, domain_id)
        if not os.path.isdir(domain_dir):
            continue
        for ext in supported_extensions():
            for file_path in glob.glob(os.path.join(domain_dir, f"*{ext}")):
                all_docs.extend(load_file(domain_id, file_path))

    return all_docs


def create_vector_store():
    print("Создаю векторную базу...")
    documents = load_all_documents()
    if not documents:
        print("Нет документов для индексации.")
        return None
    text_splitter = RecursiveCharacterTextSplitter(chunk_size=500, chunk_overlap=50)
    docs = text_splitter.split_documents(documents)
    print(f"Фрагментов: {len(docs)}")
    embeddings = HuggingFaceEmbeddings(model_name="intfloat/multilingual-e5-small")
    store = Chroma.from_documents(docs, embeddings, persist_directory=PERSIST_DIR)
    print(f"База сохранена в {PERSIST_DIR}")
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

        print("FORCE_RAG_REINDEX: удаляю старую chroma_db")
        shutil.rmtree(PERSIST_DIR, ignore_errors=True)

    if os.path.exists(PERSIST_DIR) and os.listdir(PERSIST_DIR):
        _vector_store = Chroma(persist_directory=PERSIST_DIR, embedding_function=embeddings)
    else:
        _vector_store = create_vector_store()
    return _vector_store


def search(query: str, domain_id: str, k: int = 8):
    domain_id = normalize_domain_id(domain_id)
    store = load_vector_store()
    if store is None:
        return []
    return store.similarity_search(
        query,
        k=k,
        filter={"domain_id": domain_id},
    )


def index_stats_for_domain(domain_id: str) -> list[dict]:
    """Chunk counts per source file for a domain (admin index status)."""
    domain_id = normalize_domain_id(domain_id)
    store = load_vector_store()
    if store is None:
        return []
    try:
        data = store._collection.get(  # noqa: SLF001 — Chroma admin introspection
            where={"domain_id": domain_id},
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
