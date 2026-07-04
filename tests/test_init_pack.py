"""Tests for template pack installer (pack.yaml v1)."""

from __future__ import annotations

import json
import sys
from pathlib import Path

import pytest

_ROOT = Path(__file__).resolve().parents[1]
_SCRIPTS = _ROOT / "scripts"
if str(_SCRIPTS) not in sys.path:
    sys.path.insert(0, str(_SCRIPTS))

import pack_installer  # noqa: E402


@pytest.fixture(autouse=True)
def _grounded_root(monkeypatch):
    monkeypatch.setenv("GROUNDED_LLM_ROOT", str(_ROOT))


def test_list_official_packs():
    names = pack_installer.list_packs()
    assert "hr" in names
    assert "it_support" in names
    assert "legal_faq" in names


def test_load_hr_manifest_domain_id():
    _, manifest = pack_installer.load_pack_manifest("hr")
    assert pack_installer._domain_id(manifest) == "default"
    assert manifest["eval"]["suite"] == "default_en"


def test_data_target_dir_legacy_hr():
    path = pack_installer.data_target_dir("default", "default")
    assert path == _ROOT / "data" / "default"


def test_data_target_dir_nested_it():
    path = pack_installer.data_target_dir("default", "it_support")
    assert path == _ROOT / "data" / "default" / "it_support"


def test_install_it_support_dry_run():
    plan = pack_installer.install_pack("it_support", dry_run=True)
    assert plan["domain_id"] == "it_support"
    assert plan["eval_suite"] == "it_support"
    assert "it_support" in plan["data_dir"]


def test_scaffold_new_pack(tmp_path, monkeypatch):
    monkeypatch.setenv("GROUNDED_LLM_ROOT", str(tmp_path))
    packs = tmp_path / "packs"
    packs.mkdir()
    monkeypatch.setattr(pack_installer, "PACKS_DIR", packs)
    monkeypatch.setattr(pack_installer, "ROOT", tmp_path)

    pack_dir = pack_installer.scaffold_new_pack("legal_faq", domain_id="legal_faq")
    assert (pack_dir / "pack.yaml").is_file()
    assert (pack_dir / "data" / "sample_policy.txt").is_file()
    assert (pack_dir / "eval.jsonl").is_file()

    (tmp_path / "config" / "locales" / "en").mkdir(parents=True)
    for name in ("prompts.json", "onboarding.json", "few_shot.json"):
        p = tmp_path / "config" / "locales" / "en" / name
        p.write_text("{}" if name != "onboarding.json" else "{}", encoding="utf-8")
    (tmp_path / "config").mkdir(exist_ok=True)
    (tmp_path / "config" / "domains.json").write_text(
        json.dumps({"default_domain": "default", "domains": {}}),
        encoding="utf-8",
    )

    result = pack_installer.install_pack("legal_faq", tenant_id="default")
    assert result["domain_id"] == "legal_faq"
    assert (tmp_path / "data" / "default" / "legal_faq" / "sample_policy.txt").is_file()
    assert (tmp_path / "eval" / "rag_legal_faq_baseline.jsonl").is_file()
