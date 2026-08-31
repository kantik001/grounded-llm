"""Tests for KB document registry and blob store."""

from __future__ import annotations

import os
import tempfile
import uuid

import pytest


@pytest.fixture
def blob_env(monkeypatch, tmp_path):
    monkeypatch.setenv("KB_BLOB_BACKEND", "local")
    monkeypatch.setenv("KB_BLOB_DIR", str(tmp_path / "blobs"))
    from rag.storage import blob_store

    blob_store.reset_blob_store()
    yield tmp_path
    blob_store.reset_blob_store()


def test_local_blob_store_roundtrip(blob_env):
    from rag.storage.blob_store import get_blob_store

    store = get_blob_store()
    key = store.build_key("default", "hr", str(uuid.uuid4()), 1, "abc123", "txt")
    store.put(key, b"hello world", content_type="text/plain")
    assert store.get(key) == b"hello world"
    store.delete(key)
    assert not (blob_env / "blobs" / key.replace("/", os.sep)).exists()


def test_build_versioned_key():
    from rag.storage.blob_store import build_versioned_key

    key = build_versioned_key("t1", "d1", "doc-id", 2, "deadbeef", "pdf")
    assert key == "tenants/t1/domains/d1/docs/doc-id/v2/deadbeef.pdf"
