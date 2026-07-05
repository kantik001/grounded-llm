"""Tests for packs/registry.yaml validation."""

from __future__ import annotations

import sys
from pathlib import Path

import pytest

_ROOT = Path(__file__).resolve().parents[1]
_SCRIPTS = _ROOT / "scripts"
if str(_SCRIPTS) not in sys.path:
    sys.path.insert(0, str(_SCRIPTS))

import pack_registry  # noqa: E402


@pytest.fixture(autouse=True)
def _grounded_root(monkeypatch):
    monkeypatch.setenv("GROUNDED_LLM_ROOT", str(_ROOT))


def test_registry_loads():
    registry = pack_registry.load_registry()
    assert registry.get("version") == 1
    ids = [p["id"] for p in registry["packs"]]
    assert ids == ["hr", "it_support", "legal_faq"]


def test_registry_validate_no_errors():
    errors = pack_registry.validate_registry()
    assert errors == [], errors


def test_registry_export_json():
    payload = pack_registry.export_registry_json()
    assert '"id": "hr"' in payload
    assert '"eval_suite": "legal_faq"' in payload
