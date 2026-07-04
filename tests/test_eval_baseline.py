"""Validate eval/*.jsonl baseline files (CI gate without running RAG)."""

import json
from pathlib import Path

import pytest

_ROOT = Path(__file__).resolve().parents[1]
EVAL_FILES = list((_ROOT / "eval").glob("rag_*_baseline.jsonl"))


@pytest.mark.parametrize("path", EVAL_FILES, ids=lambda p: p.name)
def test_baseline_jsonl_structure(path: Path):
    assert path.is_file(), f"missing {path}"
    cases = []
    with path.open(encoding="utf-8") as f:
        for line_no, line in enumerate(f, 1):
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            case = json.loads(line)
            assert "question" in case and case["question"].strip(), f"{path.name}:{line_no} question"
            assert "domain_id" in case, f"{path.name}:{line_no} domain_id"
            cases.append(case)
    assert len(cases) >= 5, f"{path.name}: need at least 5 cases, got {len(cases)}"
    if path.name == "rag_default_en_baseline.jsonl":
        assert len(cases) >= 15, f"{path.name}: Phase A requires at least 15 EN cases, got {len(cases)}"
    if path.name == "rag_it_support_baseline.jsonl":
        assert len(cases) >= 10, f"{path.name}: IT template requires at least 10 cases, got {len(cases)}"
    if path.name == "rag_legal_faq_baseline.jsonl":
        assert len(cases) >= 10, f"{path.name}: Legal FAQ template requires at least 10 cases, got {len(cases)}"
    if path.name == "rag_adversarial_baseline.jsonl":
        assert len(cases) >= 20, f"{path.name}: adversarial pack requires at least 20 cases, got {len(cases)}"
        types = {c.get("adversarial_type") for c in cases}
        assert "wrong_number" in types, "adversarial pack needs wrong_number cases"
        assert "cross_domain" in types, "adversarial pack needs cross_domain cases"
        assert "prompt_injection" in types, "adversarial pack needs prompt_injection cases"
        for case in cases:
            assert case.get("adversarial_type"), "adversarial_type required"
    for case in cases:
        if case.get("expect_out_of_scope"):
            continue
        assert case.get("expect_context", True), "expect_context or expect_out_of_scope"
        if case.get("expect_contains"):
            assert isinstance(case["expect_contains"], list)
