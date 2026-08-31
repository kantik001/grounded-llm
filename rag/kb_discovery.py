"""Discover knowledge-base directories on disk (no Chroma / embedding deps)."""

from __future__ import annotations

import glob
import os
from typing import Iterator, Tuple

from rag.document_loaders import supported_extensions
from rag.domains_config import list_domains

_PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
_FALLBACK_DATA_DIR = os.path.join(_PROJECT_ROOT, "data")
DEFAULT_TENANT = os.environ.get("DEFAULT_TENANT_ID", "default")


def data_dir() -> str:
    """Knowledge-base root. Honors DATA_DIR env (same variable the Go server
    uses for uploads) so indexing and uploads always see the same tree;
    falls back to <repo>/data for local development."""
    env = (os.environ.get("DATA_DIR") or "").strip()
    return env or _FALLBACK_DATA_DIR


# Deprecated module-level constant (evaluated at import); prefer data_dir().
DATA_DIR = data_dir()


def kb_data_dir(tenant_id: str, domain_id: str) -> str:
    """Path to KB files for tenant/domain (matches Go tenant.KBDataDir)."""
    root = data_dir()
    tid = (tenant_id or DEFAULT_TENANT).strip().lower() or DEFAULT_TENANT
    nested = os.path.join(root, tid, domain_id)
    if os.path.isdir(nested):
        return nested
    if tid == DEFAULT_TENANT:
        legacy = os.path.join(root, domain_id)
        if os.path.isdir(legacy):
            return legacy
    return nested


def _has_kb_files(path: str) -> bool:
    if not os.path.isdir(path):
        return False
    for ext in supported_extensions():
        if glob.glob(os.path.join(path, f"*{ext}")):
            return True
    return False


def discover_kb_directories() -> Iterator[Tuple[str, str, str]]:
    """Yield (tenant_id, domain_id, directory_path).

    Layouts:
    - Multi-tenant (preferred): data/{tenant_id}/{domain_id}/*.{txt,pdf,docx}
    - Legacy: data/{domain_id}/*.{txt,pdf,docx} (default tenant only)

    When a folder name is both a legacy domain (e.g. ``default``) and a tenant
    with nested domains (e.g. ``default/it_support/``), both are indexed if they
    contain KB files.
    """
    root = data_dir()
    if not os.path.isdir(root):
        return
    domain_ids = set(list_domains().get("domains", {}).keys())

    for name in sorted(os.listdir(root)):
        path = os.path.join(root, name)
        if not os.path.isdir(path):
            continue
        if name in domain_ids and _has_kb_files(path):
            yield DEFAULT_TENANT, name, path
        for domain_id in sorted(os.listdir(path)):
            if domain_id not in domain_ids:
                continue
            dpath = os.path.join(path, domain_id)
            if os.path.isdir(dpath) and _has_kb_files(dpath):
                yield name, domain_id, dpath
