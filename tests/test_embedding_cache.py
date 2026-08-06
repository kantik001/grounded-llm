"""Unit tests for Redis embedding cache helpers (no Redis required)."""

from __future__ import annotations

import rag.embedding_cache as ec


def test_embedding_key_format():
    key = ec._key("hello", "sentence-transformers/all-MiniLM-L6-v2")
    assert key.startswith("embedding:")
    assert key.endswith(":sentence-transformers/all-MiniLM-L6-v2")
    assert ec._key("hello", "m") == ec._key("hello", "m")
    assert ec._key("hello", "m") != ec._key("hello!", "m")


def test_cache_stats_shape():
    stats = ec.cache_stats()
    assert "hits" in stats and "misses" in stats


def test_e5_prefixes_auto_for_e5_models(monkeypatch):
    monkeypatch.delenv("RAG_E5_PREFIXES", raising=False)
    assert ec.e5_prefixes_enabled("intfloat/multilingual-e5-small") is True
    assert ec.e5_prefixes_enabled("sentence-transformers/all-MiniLM-L6-v2") is False


def test_e5_prefixes_env_override(monkeypatch):
    monkeypatch.setenv("RAG_E5_PREFIXES", "0")
    assert ec.e5_prefixes_enabled("intfloat/multilingual-e5-small") is False
    monkeypatch.setenv("RAG_E5_PREFIXES", "1")
    assert ec.e5_prefixes_enabled("all-MiniLM-L6-v2") is True
