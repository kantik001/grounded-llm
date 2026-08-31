"""Chroma vector backend (default reference implementation)."""

from __future__ import annotations

import hashlib
import json
import os
import shutil
from typing import Any

from langchain_chroma import Chroma

from rag.embedding_cache import CachedHuggingFaceEmbeddings, e5_prefixes_enabled
from rag.indexing import split_file_documents, split_kb_documents
from rag.vector_backend.base import VectorBackend

_PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
DEFAULT_PERSIST_DIR = os.path.join(_PROJECT_ROOT, "chroma_db")
EMBEDDING_MODEL = "intfloat/multilingual-e5-small"

_INDEX_META_FILE = "index_meta.json"
_MANIFEST_FILE = "index_manifest.json"


def _persist_dir() -> str:
    return os.environ.get("CHROMA_PERSIST_DIR", DEFAULT_PERSIST_DIR).strip() or DEFAULT_PERSIST_DIR


def embedding_signature() -> dict:
    """Identity of the vectors in the index. A mismatch (model swap, prefix
    flip) makes existing vectors incompatible with new queries, so load()
    rebuilds instead of silently mixing embedding spaces."""
    return {
        "model": EMBEDDING_MODEL,
        "e5_prefixes": e5_prefixes_enabled(EMBEDDING_MODEL),
        "schema": 1,
    }


def _file_sha1(path: str) -> str:
    h = hashlib.sha1()
    with open(path, "rb") as fh:
        for block in iter(lambda: fh.read(1 << 20), b""):
            h.update(block)
    return h.hexdigest()


def scan_kb_files() -> dict[str, dict]:
    """Current KB state from Postgres registry (compat alias for scan_registry_documents)."""
    from rag.kb.documents import scan_registry_documents

    return scan_registry_documents()


def _resolve_scan_path(entry: dict) -> tuple[str, bool]:
    """Return local path for a registry scan entry; materialize blob when needed."""
    path = entry.get("path") or ""
    if path and os.path.isfile(path):
        return path, False
    storage_key = entry.get("storage_key") or ""
    if not storage_key:
        raise FileNotFoundError(f"missing storage_key for {entry.get('filename')}")
    from rag.kb.documents import DocumentTarget, materialize_to_temp

    target = DocumentTarget(
        document_id=str(entry.get("document_id") or ""),
        version_id="",
        tenant_id=str(entry.get("tenant") or "default"),
        domain_id=str(entry.get("domain") or "default"),
        logical_key=str(entry.get("filename") or ""),
        content_sha256=str(entry.get("sha256") or entry.get("sha1") or ""),
        storage_key=storage_key,
    )
    return materialize_to_temp(target), True


class ChromaBackend(VectorBackend):
    def __init__(self) -> None:
        self._store: Chroma | None = None
        self._embeddings = CachedHuggingFaceEmbeddings(model_name=EMBEDDING_MODEL)

    def reset(self) -> None:
        self._store = None

    # --- index metadata -------------------------------------------------

    def _meta_path(self) -> str:
        return os.path.join(_persist_dir(), _INDEX_META_FILE)

    def _manifest_path(self) -> str:
        return os.path.join(_persist_dir(), _MANIFEST_FILE)

    def _read_json(self, path: str) -> dict | None:
        try:
            with open(path, encoding="utf-8") as fh:
                data = json.load(fh)
            return data if isinstance(data, dict) else None
        except (OSError, ValueError):
            return None

    def _write_json(self, path: str, payload: dict) -> None:
        os.makedirs(os.path.dirname(path), exist_ok=True)
        tmp = path + ".tmp"
        with open(tmp, "w", encoding="utf-8") as fh:
            json.dump(payload, fh, ensure_ascii=False, indent=1)
        os.replace(tmp, path)

    def _signature_matches(self) -> bool:
        meta = self._read_json(self._meta_path())
        return bool(meta) and meta.get("embedding") == embedding_signature()

    def _save_index_state(self, manifest: dict[str, dict]) -> None:
        self._write_json(self._meta_path(), {"embedding": embedding_signature()})
        self._write_json(
            self._manifest_path(),
            {key: {"sha1": st["sha1"]} for key, st in manifest.items()},
        )

    # --- build / load ----------------------------------------------------

    def _create_store(self) -> Chroma | None:
        print("Creating vector store (Chroma)...")
        docs = split_kb_documents()
        if not docs:
            print("No documents to index.")
            return None
        print(f"Chunks: {len(docs)}")
        persist_dir = _persist_dir()
        store = Chroma.from_documents(docs, self._embeddings, persist_directory=persist_dir)
        self._save_index_state(scan_kb_files())
        print(f"Vector store saved to {persist_dir}")
        return store

    def load(self, *, force_reindex: bool = False) -> None:
        if self._store is not None and not force_reindex:
            return

        force = force_reindex or os.environ.get("FORCE_RAG_REINDEX", "").lower() in (
            "1",
            "true",
            "yes",
        )
        persist_dir = _persist_dir()

        has_data = os.path.isdir(persist_dir) and bool(os.listdir(persist_dir))
        if has_data and not force and not self._signature_matches():
            print(
                "Embedding signature changed (model or e5 prefixes) — "
                "rebuilding vector store to avoid mixing embedding spaces."
            )
            force = True

        if force and os.path.isdir(persist_dir):
            print("Reindex: removing old chroma_db")
            shutil.rmtree(persist_dir, ignore_errors=True)
            has_data = False

        if has_data:
            self._store = Chroma(persist_directory=persist_dir, embedding_function=self._embeddings)
        else:
            self._store = self._create_store()

    # --- incremental update ----------------------------------------------

    def refresh(self) -> dict:
        """Incrementally sync the index with files on disk.

        Diffs the persisted manifest against the current KB tree and only
        re-embeds added/changed files (deleting stale chunks by metadata),
        instead of a full rebuild. Falls back to a full rebuild when there
        is no usable index/manifest yet.
        """
        persist_dir = _persist_dir()
        has_data = os.path.isdir(persist_dir) and bool(os.listdir(persist_dir))
        manifest = self._read_json(self._manifest_path())

        if not has_data or manifest is None or not self._signature_matches():
            self._store = None
            self.load(force_reindex=True)
            current = scan_kb_files()
            return {"mode": "full", "files": len(current), "empty": self._store is None}

        self.load()
        if self._store is None:
            return {"mode": "full", "files": 0, "empty": True}

        current = scan_kb_files()
        added = [k for k in current if k not in manifest]
        removed = [k for k in manifest if k not in current]
        changed = [
            k for k in current if k in manifest and manifest[k].get("sha1") != current[k]["sha1"]
        ]

        for key in removed + changed:
            tenant, domain, filename = key.split("/", 2)
            self._store._collection.delete(  # noqa: SLF001
                where={
                    "$and": [
                        {"tenant_id": tenant},
                        {"domain_id": domain},
                        {"filename": filename},
                    ]
                }
            )

        chunks_added = 0
        for key in added + changed:
            st = current[key]
            path, is_temp = _resolve_scan_path(st)
            try:
                chunks = split_file_documents(st["domain"], path, tenant_id=st["tenant"])
                if chunks:
                    self._store.add_documents(chunks)
                    chunks_added += len(chunks)
            finally:
                if is_temp:
                    os.remove(path)

        self._save_index_state(current)
        summary = {
            "mode": "incremental",
            "added": len(added),
            "changed": len(changed),
            "removed": len(removed),
            "chunks_added": chunks_added,
        }
        print(f"Incremental reindex: {summary}")
        return summary

    def delete_kb_file(self, tenant_id: str, domain_id: str, filename: str) -> None:
        self.load()
        if self._store is None:
            return
        self._store._collection.delete(  # noqa: SLF001
            where={
                "$and": [
                    {"tenant_id": tenant_id},
                    {"domain_id": domain_id},
                    {"filename": filename},
                ]
            }
        )

    def upsert_kb_file(
        self,
        tenant_id: str,
        domain_id: str,
        path: str,
        *,
        filename: str | None = None,
    ) -> int:
        """Re-embed one file and update manifest entry."""
        self.load()
        if self._store is None:
            self._store = self._create_store()
        if self._store is None:
            return 0
        name = filename or os.path.basename(path)
        self.delete_kb_file(tenant_id, domain_id, name)
        chunks = split_file_documents(domain_id, path, tenant_id=tenant_id)
        if chunks:
            self._store.add_documents(chunks)
        self.touch_manifest_entry(tenant_id, domain_id, path, name)
        return len(chunks)

    def touch_manifest_entry(
        self,
        tenant_id: str,
        domain_id: str,
        path: str,
        filename: str | None = None,
    ) -> None:
        name = filename or os.path.basename(path)
        key = f"{tenant_id}/{domain_id}/{name}"
        manifest = self._read_json(self._manifest_path()) or {}
        sha = ""
        if os.path.isfile(path):
            sha = _file_sha1(path)
        manifest[key] = {"sha1": sha}
        self._write_json(self._manifest_path(), manifest)
        meta = self._read_json(self._meta_path()) or {}
        if "embedding" not in meta:
            meta["embedding"] = embedding_signature()
        self._write_json(self._meta_path(), meta)

    def sync_manifest(self, manifest: dict[str, dict]) -> None:
        self._save_index_state(manifest)

    # --- search -----------------------------------------------------------

    def _filter(self, domain_id: str, tenant_id: str) -> dict:
        return {"$and": [{"domain_id": domain_id}, {"tenant_id": tenant_id}]}

    def similarity_search(
        self,
        query: str,
        *,
        k: int,
        domain_id: str,
        tenant_id: str,
    ) -> list[Any]:
        self.load()
        if self._store is None:
            return []
        return self._store.similarity_search(query, k=k, filter=self._filter(domain_id, tenant_id))

    def index_stats_for_domain(self, domain_id: str, tenant_id: str) -> list[dict]:
        self.load()
        if self._store is None:
            return []
        try:
            data = self._store._collection.get(  # noqa: SLF001
                where=self._filter(domain_id, tenant_id),
                include=["metadatas"],
            )
        except Exception:
            try:
                data = self._store._collection.get(
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
