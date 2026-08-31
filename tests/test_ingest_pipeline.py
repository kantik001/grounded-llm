"""Tests for ingest pipeline helpers (no DB/Redis)."""

from __future__ import annotations

import os
import tempfile

from rag.ingest.pipeline import discover_files, staging_path
from rag.ingest.store import IngestJob


def test_staging_path_sanitizes_file_key():
    path = staging_path("tenant/hr/doc.pdf")
    assert "tenant__hr__doc.pdf" in path


def test_discover_files_empty_domain(tmp_path, monkeypatch):
    monkeypatch.setenv("DATA_DIR", str(tmp_path))
    job = IngestJob(
        id=1,
        status="queued",
        tenant_id="default",
        domain_id="missing",
        source="admin",
        actor="",
        mode="incremental",
        files=[],
    )
    assert discover_files(job) == []


def test_discover_files_explicit_list(tmp_path, monkeypatch):
    root = tmp_path / "default" / "hr"
    root.mkdir(parents=True)
    doc = root / "note.txt"
    doc.write_text("hello", encoding="utf-8")
    monkeypatch.setenv("DATA_DIR", str(tmp_path))

    job = IngestJob(
        id=2,
        status="queued",
        tenant_id="default",
        domain_id="hr",
        source="admin",
        actor="",
        mode="incremental",
        files=["note.txt"],
    )
    targets = discover_files(job)
    assert len(targets) == 1
    assert targets[0].filename == "note.txt"
    assert os.path.isfile(targets[0].path)
