"""Tests for incremental index bookkeeping (manifest scan + embedding signature).

These cover the pure bookkeeping helpers; the end-to-end incremental refresh
runs against a real Chroma store in the eval/e2e pipelines.
"""

from __future__ import annotations

from pathlib import Path

import pytest
from rag.vector_backend import chroma_backend as cb

_ROOT = Path(__file__).resolve().parents[1]


@pytest.fixture(autouse=True)
def _domains_config(monkeypatch):
    monkeypatch.setenv("DOMAINS_CONFIG_PATH", str(_ROOT / "config" / "domains.json"))


def _make_kb(tmp_path: Path) -> Path:
    kb = tmp_path / "acme" / "default"
    kb.mkdir(parents=True)
    (kb / "handbook.txt").write_text("Vacation is 28 days.", encoding="utf-8")
    (kb / "faq.txt").write_text("VPN reset: use the portal.", encoding="utf-8")
    return kb


def test_scan_kb_files_keys_and_hashes(monkeypatch, tmp_path):
    from rag.domains_config import reload_domains_config

    kb = _make_kb(tmp_path)
    monkeypatch.setenv("DATA_DIR", str(tmp_path))
    reload_domains_config()

    state = cb.scan_kb_files()
    assert set(state) == {"acme/default/handbook.txt", "acme/default/faq.txt"}
    entry = state["acme/default/handbook.txt"]
    assert entry["tenant"] == "acme"
    assert entry["domain"] == "default"
    assert len(entry["sha1"]) == 40

    # Content change must change the hash (drives changed-file detection).
    (kb / "handbook.txt").write_text("Vacation is 30 days.", encoding="utf-8")
    state2 = cb.scan_kb_files()
    assert state2["acme/default/handbook.txt"]["sha1"] != entry["sha1"]
    assert state2["acme/default/faq.txt"]["sha1"] == state["acme/default/faq.txt"]["sha1"]


def test_scan_kb_files_ignores_unsupported(monkeypatch, tmp_path):
    from rag.domains_config import reload_domains_config

    kb = _make_kb(tmp_path)
    (kb / "notes.exe").write_bytes(b"\x00")
    monkeypatch.setenv("DATA_DIR", str(tmp_path))
    reload_domains_config()

    state = cb.scan_kb_files()
    assert "acme/default/notes.exe" not in state


def test_embedding_signature_tracks_prefix_flag(monkeypatch):
    monkeypatch.delenv("RAG_E5_PREFIXES", raising=False)
    sig_on = cb.embedding_signature()
    assert sig_on["model"] == cb.EMBEDDING_MODEL
    assert sig_on["e5_prefixes"] is True

    monkeypatch.setenv("RAG_E5_PREFIXES", "0")
    sig_off = cb.embedding_signature()
    assert sig_off["e5_prefixes"] is False
    # Different signatures → load() must rebuild instead of mixing spaces.
    assert sig_on != sig_off
