"""Index-run scoped collection / persist path naming (blue/green vector indexes)."""

from __future__ import annotations

import os
import re

_SCOPE_RE = re.compile(r"[^a-z0-9_]+")


def run_suffix(run_id: str) -> str:
    """Short stable suffix from a UUID index run id."""
    return run_id.replace("-", "")[:12]


def sanitize_scope_part(value: str) -> str:
    raw = (value or "default").strip().lower()
    cleaned = _SCOPE_RE.sub("_", raw).strip("_")
    return cleaned or "default"


def collection_name(base: str, tenant_id: str, domain_id: str, run_id: str) -> str:
    """Remote collection name: {base}_{tenant}_{domain}_{run_suffix}."""
    t = sanitize_scope_part(tenant_id)
    d = sanitize_scope_part(domain_id)
    suffix = run_suffix(run_id)
    return f"{base}_{t}_{d}_{suffix}"


def chroma_run_dir(base_persist_dir: str, tenant_id: str, domain_id: str, run_id: str) -> str:
    """Per-run Chroma persist directory under the backend root."""
    t = sanitize_scope_part(tenant_id)
    d = sanitize_scope_part(domain_id)
    suffix = run_suffix(run_id)
    return os.path.join(base_persist_dir, "runs", f"{t}_{d}_{suffix}")


def sparse_run_dir(base_sparse_dir: str, tenant_id: str, domain_id: str, run_id: str) -> str:
    """Per-run BM25 persist directory."""
    t = sanitize_scope_part(tenant_id)
    d = sanitize_scope_part(domain_id)
    suffix = run_suffix(run_id)
    return os.path.join(base_sparse_dir, "runs", f"{t}_{d}_{suffix}")
