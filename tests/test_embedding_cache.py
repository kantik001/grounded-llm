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
