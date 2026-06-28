"""Discover knowledge-base directories on disk (no Chroma / embedding deps)."""

from __future__ import annotations

import glob
import os
from typing import Iterator, Tuple

from rag.document_loaders import supported_extensions
from rag.domains_config import list_domains

_PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
DATA_DIR = os.path.join(_PROJECT_ROOT, "data")
DEFAULT_TENANT = os.environ.get("DEFAULT_TENANT_ID", "default")


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
    - Legacy: data/{domain_id}/*.{txt,pdf,docx}
    - Multi-tenant: data/{tenant_id}/{domain_id}/*.{txt,pdf,docx}

    When a folder name is both a legacy domain (e.g. ``default``) and a tenant
    with nested domains (e.g. ``default/it_support/``), both are indexed.
    """
    if not os.path.isdir(DATA_DIR):
        return
    domain_ids = set(list_domains().get("domains", {}).keys())

    for name in sorted(os.listdir(DATA_DIR)):
        path = os.path.join(DATA_DIR, name)
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
