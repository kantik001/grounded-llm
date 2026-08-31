"""Register connector-synced files into Postgres + blob store (SoT)."""

from __future__ import annotations

import mimetypes
from pathlib import Path

from rag.kb.documents import upsert_document


def register_synced_file(
    *,
    tenant_id: str,
    domain_id: str,
    path: Path,
    source: str,
    source_ref: dict | None = None,
    created_by: str = "",
) -> str:
    """Upsert one synced file into the KB registry. Returns document id."""
    data = path.read_bytes()
    mime, _ = mimetypes.guess_type(str(path))
    doc = upsert_document(
        tenant_id=tenant_id,
        domain_id=domain_id,
        logical_key=path.name,
        data=data,
        mime_type=mime or "application/octet-stream",
        source=source,
        source_ref=source_ref or {"path": str(path)},
        created_by=created_by,
    )
    return doc.id


def register_synced_tree(
    *,
    tenant_id: str,
    domain_id: str,
    target_dir: Path,
    source: str,
    suffixes: set[str] | None = None,
) -> list[str]:
    """Register all supported files under target_dir."""
    suffixes = suffixes or {".txt", ".pdf", ".docx"}
    ids: list[str] = []
    for path in sorted(target_dir.rglob("*")):
        if not path.is_file() or path.suffix.lower() not in suffixes:
            continue
        rel = path.relative_to(target_dir)
        doc = upsert_document(
            tenant_id=tenant_id,
            domain_id=domain_id,
            logical_key=str(rel).replace("\\", "/"),
            data=path.read_bytes(),
            source=source,
            source_ref={"relative": str(rel)},
        )
        ids.append(doc.id)
    return ids
