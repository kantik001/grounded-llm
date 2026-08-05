"""Validate eval/*.jsonl baseline files (CI gate without running RAG)."""

from __future__ import annotations

import json
from pathlib import Path

import pytest

_ROOT = Path(__file__).resolve().parents[1]
EVAL_DIR = _ROOT / "eval"
BASELINE_FILES = sorted(EVAL_DIR.glob("rag_*_baseline.jsonl"))
E2E_FILE = EVAL_DIR / "rag_adversarial_e2e.jsonl"
BASELINE_SCHEMA = EVAL_DIR / "schemas" / "baseline_case.schema.json"
E2E_SCHEMA = EVAL_DIR / "schemas" / "adversarial_e2e_case.schema.json"


def _load_cases(path: Path) -> list[tuple[int, dict]]:
    out: list[tuple[int, dict]] = []
    with path.open(encoding="utf-8") as f:
        for line_no, line in enumerate(f, 1):
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            out.append((line_no, json.loads(line)))
    return out


def _validate_cases(schema_path: Path, cases: list[tuple[int, dict]], label: str) -> None:
    jsonschema = pytest.importorskip("jsonschema")
    Draft202012Validator = jsonschema.Draft202012Validator
    schema = json.loads(schema_path.read_text(encoding="utf-8"))
    Draft202012Validator.check_schema(schema)
    validator = Draft202012Validator(schema)
    errors = []
    for line_no, case in cases:
        for err in validator.iter_errors(case):
            errors.append(f"{label}:{line_no} {list(err.path)}: {err.message}")
    assert not errors, "\n".join(errors)


@pytest.mark.parametrize("path", BASELINE_FILES, ids=lambda p: p.name)
def test_baseline_jsonl_structure(path: Path):
    assert path.is_file(), f"missing {path}"
    cases = _load_cases(path)
    assert len(cases) >= 5, f"{path.name}: need at least 5 cases, got {len(cases)}"
    _validate_cases(BASELINE_SCHEMA, cases, path.name)

    if path.name == "rag_default_en_baseline.jsonl":
        assert len(cases) >= 15, f"{path.name}: Phase A requires at least 15 EN cases, got {len(cases)}"
    if path.name == "rag_it_support_baseline.jsonl":
        assert len(cases) >= 10, f"{path.name}: IT template requires at least 10 cases, got {len(cases)}"
    if path.name == "rag_legal_faq_baseline.jsonl":
        assert len(cases) >= 10, f"{path.name}: Legal FAQ template requires at least 10 cases, got {len(cases)}"
    if path.name == "rag_adversarial_baseline.jsonl":
        assert len(cases) >= 25, f"{path.name}: adversarial pack requires at least 25 cases, got {len(cases)}"
        types = {c.get("adversarial_type") for _, c in cases}
        assert "wrong_number" in types, "adversarial pack needs wrong_number cases"
        assert "cross_domain" in types, "adversarial pack needs cross_domain cases"
        assert "prompt_injection" in types, "adversarial pack needs prompt_injection cases"
        for line_no, case in cases:
            assert case.get("adversarial_type"), f"{path.name}:{line_no} adversarial_type required"

    for line_no, case in cases:
        if case.get("expect_out_of_scope"):
            continue
        # In-scope retrieval cases must ask for context (default true) and usually substrings.
        if case.get("expect_context") is False and not case.get("expect_contains"):
            pytest.fail(f"{path.name}:{line_no}: expect_context=false without expect_out_of_scope")


def test_adversarial_e2e_jsonl_structure():
    assert E2E_FILE.is_file()
    cases = _load_cases(E2E_FILE)
    assert len(cases) >= 5, f"{E2E_FILE.name}: need at least 5 cases"
    _validate_cases(E2E_SCHEMA, cases, E2E_FILE.name)


def test_default_ru_suite_is_documented_legacy():
    """RU suite remains for RU KB docs; primary HR gate is default_en."""
    ru = EVAL_DIR / "rag_default_baseline.jsonl"
    en = EVAL_DIR / "rag_default_en_baseline.jsonl"
    assert ru.is_file() and en.is_file()
    assert len(_load_cases(en)) >= len(_load_cases(ru))
