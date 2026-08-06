"""Tests for knowledge-base directory discovery (multi-tenant + legacy layouts)."""

from __future__ import annotations

import os
from pathlib import Path

import pytest

_ROOT = Path(__file__).resolve().parents[1]


@pytest.fixture(autouse=True)
def _domains_config(monkeypatch):
    monkeypatch.setenv("DOMAINS_CONFIG_PATH", str(_ROOT / "config" / "domains.json"))
    # Pin the KB root: other test modules import api.http.app, whose load_dotenv()
    # can inject a container-only DATA_DIR (e.g. /app/data) into this process.
    monkeypatch.setenv("DATA_DIR", str(_ROOT / "data"))


def test_discover_nested_default_and_it_support():
    from rag.domains_config import reload_domains_config
    from rag.kb_discovery import discover_kb_directories

    reload_domains_config()
    pairs = {(tenant_id, domain_id) for tenant_id, domain_id, _ in discover_kb_directories()}

    assert ("default", "default") in pairs
    assert ("default", "it_support") in pairs
    hr_paths = [p for t, d, p in discover_kb_directories() if t == "default" and d == "default"]
    assert len(hr_paths) == 1
    assert hr_paths[0].endswith(os.path.join("data", "default", "default"))


def test_discover_nested_paths_include_it_support_files():
    from rag.domains_config import reload_domains_config
    from rag.kb_discovery import discover_kb_directories

    reload_domains_config()
    it_paths = [p for t, d, p in discover_kb_directories() if t == "default" and d == "it_support"]
    assert len(it_paths) == 1
    assert it_paths[0].endswith(os.path.join("data", "default", "it_support"))


def test_data_dir_env_override(monkeypatch, tmp_path):
    from rag.kb_discovery import data_dir

    monkeypatch.setenv("DATA_DIR", str(tmp_path))
    assert data_dir() == str(tmp_path)
    monkeypatch.delenv("DATA_DIR")
    assert data_dir().endswith("data")


def test_discovery_honors_data_dir_env(monkeypatch, tmp_path):
    """Indexing must see the same tree the Go server uploads into (DATA_DIR)."""
    from rag.domains_config import reload_domains_config
    from rag.kb_discovery import discover_kb_directories

    kb = tmp_path / "acme" / "default"
    kb.mkdir(parents=True)
    (kb / "handbook.txt").write_text("policy text", encoding="utf-8")

    monkeypatch.setenv("DATA_DIR", str(tmp_path))
    reload_domains_config()
    pairs = {(t, d) for t, d, _ in discover_kb_directories()}
    assert pairs == {("acme", "default")}
