"""Tests for knowledge-base directory discovery (multi-tenant + legacy layouts)."""

from __future__ import annotations

import os
from pathlib import Path

import pytest

_ROOT = Path(__file__).resolve().parents[1]


@pytest.fixture(autouse=True)
def _domains_config(monkeypatch):
    monkeypatch.setenv("DOMAINS_CONFIG_PATH", str(_ROOT / "config" / "domains.json"))


def test_discover_legacy_default_and_nested_it_support():
    from rag.domains_config import reload_domains_config
    from rag.vector_store import discover_kb_directories

    reload_domains_config()
    pairs = {(tenant_id, domain_id) for tenant_id, domain_id, _ in discover_kb_directories()}

    assert ("default", "default") in pairs
    assert ("default", "it_support") in pairs


def test_discover_nested_paths_include_it_support_files():
    from rag.domains_config import reload_domains_config
    from rag.vector_store import discover_kb_directories

    reload_domains_config()
    it_paths = [p for t, d, p in discover_kb_directories() if t == "default" and d == "it_support"]
    assert len(it_paths) == 1
    assert it_paths[0].endswith(os.path.join("data", "default", "it_support"))
