"""Unit tests for RAG eval check_retrieval (expect_not_contains, adversarial)."""

import sys
from pathlib import Path

_root = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(_root / "scripts"))

from run_rag_eval import check_retrieval, ranking_metrics  # noqa: E402


def test_expect_not_contains_passes():
    case = {
        "question": "trap",
        "expect_contains": ["28"],
        "expect_not_contains": ["99"],
        "expect_context": True,
    }
    ctx = {"success": True, "http_status": 200, "context": "Employees get 28 paid vacation days."}
    out = check_retrieval(case, ctx)
    assert out["passed"] is True


def test_expect_not_contains_fails():
    case = {"question": "trap", "expect_not_contains": ["99"], "expect_context": True}
    ctx = {"success": True, "http_status": 200, "context": "Allow 99 days vacation."}
    out = check_retrieval(case, ctx)
    assert out["passed"] is False
    assert "99" in out["forbidden_in_context"]


def test_out_of_scope_with_forbidden():
    case = {
        "question": "IT on HR domain",
        "expect_out_of_scope": True,
        "expect_not_contains": ["vpn"],
    }
    ctx = {"success": True, "http_status": 200, "context": "Reset your vpn password in IT portal."}
    out = check_retrieval(case, ctx)
    assert out["passed"] is False


def test_out_of_scope_empty_context():
    case = {"question": "unknown", "expect_out_of_scope": True, "expect_not_contains": ["vpn"]}
    ctx = {"success": False, "http_status": 200, "context": "", "error": "no documents found"}
    out = check_retrieval(case, ctx)
    assert out["passed"] is True


def test_ranking_metrics_none_without_goldens():
    assert ranking_metrics({"question": "q"}, [{"filename": "a.txt"}]) is None


def test_ranking_metrics_perfect_recall_top_rank():
    case = {"question": "q", "expect_sources": ["vacation.txt"]}
    fragments = [{"filename": "vacation.txt"}, {"filename": "other.txt"}]
    m = ranking_metrics(case, fragments)
    assert m["recall_at_k"] == 1.0
    assert m["ndcg_at_k"] == 1.0
    assert m["citation_precision_at_k"] == 0.5
    assert m["granularity"] == "source"


def test_ranking_metrics_citation_precision_all_relevant():
    case = {"question": "q", "expect_sources": ["a.txt", "b.txt"]}
    fragments = [{"filename": "a.txt"}, {"filename": "b.txt"}]
    m = ranking_metrics(case, fragments)
    assert m["citation_precision_at_k"] == 1.0


def test_ranking_metrics_partial_recall_low_rank():
    case = {"question": "q", "expect_sources": ["a.txt", "b.txt"]}
    fragments = [{"filename": "x.txt"}, {"filename": "a.txt"}, {"filename": "y.txt"}]
    m = ranking_metrics(case, fragments)
    assert m["recall_at_k"] == 0.5
    assert 0 < m["ndcg_at_k"] < 1


def test_ranking_metrics_chunk_granularity_preferred():
    case = {
        "question": "q",
        "expect_sources": ["a.txt"],
        "expect_chunks": ["default/default/a.txt/0"],
    }
    fragments = [{"filename": "a.txt", "chunk_id": "default/default/a.txt/3"}]
    m = ranking_metrics(case, fragments)
    assert m["granularity"] == "chunk"
    assert m["recall_at_k"] == 0.0
