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
from rag.kb.index_collections import chroma_run_dir
from rag.vector_backend.base import VectorBackend

_PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
DEFAULT_PERSIST_DIR = os.path.join(_PROJECT_ROOT, "chroma_db")
EMBEDDING_MODEL = "intfloat/multilingual-e5-small"

_INDEX_META_FILE = "index_meta.json"
_MANIFEST_FILE = "index_manifest.json"
_COLLECTION_NAME = "chunks"


def _persist_dir() -> str:
    return os.environ.get("CHROMA_PERSIST_DIR", DEFAULT_PERSIST_DIR).strip() or DEFAULT_PERSIST_DIR


def embedding_signature() -> dict:
    """Identity of the vectors in the index."""
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
    from rag.kb.documents import scan_registry_documents

    return scan_registry_documents()


def _resolve_scan_path(entry: dict) -> tuple[str, bool]:
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
        self._scope_stores: dict[str, Chroma] = {}
        self._embeddings = CachedHuggingFaceEmbeddings(model_name=EMBEDDING_MODEL)

    def reset(self) -> None:
        self._scope_stores = {}

    def _scope_cache_key(self, tenant_id: str, domain_id: str, run_id: str) -> str:
        return f"{tenant_id}/{domain_id}/{run_id}"

    def _scope_persist_dir(self, tenant_id: str, domain_id: str, run_id: str) -> str:
        return chroma_run_dir(_persist_dir(), tenant_id, domain_id, run_id)

    def _meta_path(self, tenant_id: str, domain_id: str, run_id: str) -> str:
        return os.path.join(self._scope_persist_dir(tenant_id, domain_id, run_id), _INDEX_META_FILE)

    def _manifest_path(self, tenant_id: str, domain_id: str, run_id: str) -> str:
        return os.path.join(self._scope_persist_dir(tenant_id, domain_id, run_id), _MANIFEST_FILE)

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

    def _signature_matches(self, tenant_id: str, domain_id: str, run_id: str) -> bool:
        meta = self._read_json(self._meta_path(tenant_id, domain_id, run_id))
        return bool(meta) and meta.get("embedding") == embedding_signature()

    def _save_index_state(
        self,
        tenant_id: str,
        domain_id: str,
        run_id: str,
        manifest: dict[str, dict],
    ) -> None:
        self._write_json(self._meta_path(tenant_id, domain_id, run_id), {"embedding": embedding_signature()})
        self._write_json(
            self._manifest_path(tenant_id, domain_id, run_id),
            {key: {"sha1": st["sha1"]} for key, st in manifest.items()},
        )

    def open_scope(
        self,
        tenant_id: str,
        domain_id: str,
        *,
        run_id: str | None = None,
        for_write: bool = False,
    ) -> Chroma:
        resolved = self.resolve_run_id(tenant_id, domain_id, run_id=run_id, for_write=for_write)
        cache_key = self._scope_cache_key(tenant_id, domain_id, resolved)
        if cache_key in self._scope_stores:
            return self._scope_stores[cache_key]

        persist_dir = self._scope_persist_dir(tenant_id, domain_id, resolved)
        os.makedirs(persist_dir, exist_ok=True)
        store = Chroma(
            persist_directory=persist_dir,
            embedding_function=self._embeddings,
            collection_name=_COLLECTION_NAME,
        )
        self._scope_stores[cache_key] = store
        return store

    def load(self, *, force_reindex: bool = False) -> None:
        if force_reindex or os.environ.get("FORCE_RAG_REINDEX", "").lower() in ("1", "true", "yes"):
            runs_root = os.path.join(_persist_dir(), "runs")
            if os.path.isdir(runs_root):
                shutil.rmtree(runs_root, ignore_errors=True)
            self.reset()
        self.refresh()

    def refresh(self) -> dict:
        """Incrementally sync scoped indexes with the Postgres registry."""
        current = scan_kb_files()
        if not current:
            return {"mode": "full", "files": 0, "empty": True}

        by_scope: dict[tuple[str, str], dict[str, dict]] = {}
        for key, entry in current.items():
            tenant = str(entry.get("tenant") or "default")
            domain = str(entry.get("domain") or "default")
            by_scope.setdefault((tenant, domain), {})[key] = entry

        total_summary = {
            "mode": "incremental",
            "scopes": 0,
            "added": 0,
            "changed": 0,
            "removed": 0,
            "chunks_added": 0,
        }
        for (tenant, domain), scope_files in by_scope.items():
            run_id = self.resolve_run_id(tenant, domain, for_write=True)
            manifest = self._read_json(self._manifest_path(tenant, domain, run_id))
            persist_dir = self._scope_persist_dir(tenant, domain, run_id)
            has_data = os.path.isdir(persist_dir) and bool(os.listdir(persist_dir))

            if not has_data or manifest is None or not self._signature_matches(tenant, domain, run_id):
                store = self.open_scope(tenant, domain, run_id=run_id, for_write=True)
                for key, st in scope_files.items():
                    path, is_temp = _resolve_scan_path(st)
                    try:
                        chunks = split_file_documents(st["domain"], path, tenant_id=st["tenant"])
                        if chunks:
                            store.add_documents(chunks)
                            total_summary["chunks_added"] += len(chunks)
                    finally:
                        if is_temp:
                            os.remove(path)
                self._save_index_state(tenant, domain, run_id, scope_files)
                total_summary["scopes"] += 1
                continue

            store = self.open_scope(tenant, domain, run_id=run_id)
            scope_keys = set(scope_files)
            manifest_keys = {k for k in manifest if k.startswith(f"{tenant}/{domain}/")}
            added = [k for k in scope_files if k not in manifest]
            removed = [k for k in manifest_keys if k not in scope_keys]
            changed = [
                k for k in scope_files if k in manifest and manifest[k].get("sha1") != scope_files[k]["sha1"]
            ]

            for key in removed + changed:
                _, _, filename = key.split("/", 2)
                store._collection.delete(  # noqa: SLF001
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
                st = scope_files[key]
                path, is_temp = _resolve_scan_path(st)
                try:
                    chunks = split_file_documents(st["domain"], path, tenant_id=st["tenant"])
                    if chunks:
                        store.add_documents(chunks)
                        chunks_added += len(chunks)
                finally:
                    if is_temp:
                        os.remove(path)

            self._save_index_state(tenant, domain, run_id, scope_files)
            total_summary["scopes"] += 1
            total_summary["added"] += len(added)
            total_summary["changed"] += len(changed)
            total_summary["removed"] += len(removed)
            total_summary["chunks_added"] += chunks_added

        print(f"Incremental reindex: {total_summary}")
        return total_summary

    def delete_kb_file(
        self,
        tenant_id: str,
        domain_id: str,
        filename: str,
        *,
        run_id: str | None = None,
    ) -> None:
        store = self.open_scope(tenant_id, domain_id, run_id=run_id, for_write=True)
        store._collection.delete(  # noqa: SLF001
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
        run_id: str | None = None,
    ) -> int:
        store = self.open_scope(tenant_id, domain_id, run_id=run_id, for_write=True)
        name = filename or os.path.basename(path)
        self.delete_kb_file(tenant_id, domain_id, name, run_id=run_id)
        chunks = split_file_documents(domain_id, path, tenant_id=tenant_id)
        if chunks:
            store.add_documents(chunks)
        self.touch_manifest_entry(tenant_id, domain_id, path, name, run_id=run_id)
        return len(chunks)

    def touch_manifest_entry(
        self,
        tenant_id: str,
        domain_id: str,
        path: str,
        filename: str | None = None,
        *,
        run_id: str | None = None,
    ) -> None:
        resolved = self.resolve_run_id(tenant_id, domain_id, run_id=run_id, for_write=True)
        name = filename or os.path.basename(path)
        key = f"{tenant_id}/{domain_id}/{name}"
        manifest_path = self._manifest_path(tenant_id, domain_id, resolved)
        manifest = self._read_json(manifest_path) or {}
        sha = _file_sha1(path) if os.path.isfile(path) else ""
        manifest[key] = {"sha1": sha}
        self._write_json(manifest_path, manifest)
        meta_path = self._meta_path(tenant_id, domain_id, resolved)
        meta = self._read_json(meta_path) or {}
        meta["embedding"] = embedding_signature()
        self._write_json(meta_path, meta)

    def sync_manifest(self, manifest: dict[str, dict]) -> None:
        by_scope: dict[tuple[str, str], dict[str, dict]] = {}
        for key, entry in manifest.items():
            parts = key.split("/", 2)
            if len(parts) < 3:
                continue
            tenant, domain = parts[0], parts[1]
            by_scope.setdefault((tenant, domain), {})[key] = entry
        for (tenant, domain), scope_manifest in by_scope.items():
            run_id = self.resolve_run_id(tenant, domain)
            self._save_index_state(tenant, domain, run_id, scope_manifest)

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
        run_id = self.resolve_run_id(tenant_id, domain_id)
        store = self.open_scope(tenant_id, domain_id, run_id=run_id)
        return store.similarity_search(query, k=k, filter=self._filter(domain_id, tenant_id))

    def index_stats_for_domain(self, domain_id: str, tenant_id: str) -> list[dict]:
        run_id = self.resolve_run_id(tenant_id, domain_id)
        store = self.open_scope(tenant_id, domain_id, run_id=run_id)
        try:
            data = store._collection.get(  # noqa: SLF001
                where=self._filter(domain_id, tenant_id),
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
