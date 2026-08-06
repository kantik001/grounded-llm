#!/usr/bin/env python3
"""
Прогон eval-набора RAG: POST /rag/context и проверка метрик.
Режим по умолчанию — retrieval (без LLM). End-to-end чат — через POST /message (Web App / API).
"""

from __future__ import annotations

import argparse
import json
import math
import os
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Dict, List

import requests

_ROOT = Path(__file__).resolve().parents[1]
EVAL_DIR = _ROOT / "eval"
RESULTS_DIR = EVAL_DIR / "results"

SUITES: Dict[str, Path] = {}
# Keyword-heavy cases for BM25+RRF; skipped in --suite all unless RAG_RETRIEVAL_MODE=hybrid
HYBRID_ONLY_SUITES = frozenset({"hybrid"})


def _retrieval_mode() -> str:
    return (os.environ.get("RAG_RETRIEVAL_MODE") or "vector").strip().lower()


def suites_for_all() -> List[str]:
    names = list(get_suites().keys())
    if _retrieval_mode() != "hybrid":
        names = [n for n in names if n not in HYBRID_ONLY_SUITES]
    return names


def discover_suites() -> Dict[str, Path]:
    """Map suite name → JSONL path (rag_{suite}_baseline.jsonl)."""
    suites: Dict[str, Path] = {}
    if not EVAL_DIR.is_dir():
        return suites
    for path in sorted(EVAL_DIR.glob("rag_*_baseline.jsonl")):
        stem = path.stem
        if stem.startswith("rag_") and stem.endswith("_baseline"):
            name = stem[4:-9]
            if name:
                suites[name] = path
    return suites


def get_suites() -> Dict[str, Path]:
    global SUITES
    if not SUITES:
        SUITES = discover_suites()
    return SUITES


def _is_out_of_scope_error(error: str) -> bool:
    err = error.lower()
    return any(
        token in err
        for token in ("not found", "no information", "no documents", "нет")
    )


def load_cases(path: Path) -> List[Dict[str, Any]]:
    cases = []
    with path.open(encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            cases.append(json.loads(line))
    return cases


def fetch_context(rag_url: str, question: str, domain_id: str, timeout: int) -> Dict[str, Any]:
    resp = requests.post(
        rag_url,
        json={"question": question, "domain_id": domain_id},
        headers={"Content-Type": "application/json; charset=utf-8"},
        timeout=timeout,
    )
    try:
        body = resp.json()
    except Exception:
        body = {"success": False, "error": resp.text[:500]}
    return {"http_status": resp.status_code, **body}


def ranking_metrics(case: Dict[str, Any], fragments: List[Dict[str, Any]]) -> Dict[str, Any] | None:
    """Recall@k / nDCG@k against golden sources.

    A case opts in with "expect_sources": ["file.txt", ...] (filenames) or
    "expect_chunks": ["tenant/domain/file.txt/0", ...] (chunk ids, stricter).
    Relevance is binary; each golden item is counted once at its best rank.
    """
    by_chunk = bool(case.get("expect_chunks"))
    expected = case.get("expect_chunks") or case.get("expect_sources")
    if not expected:
        return None
    expected_set = {str(s).strip().lower() for s in expected if str(s).strip()}
    if not expected_set:
        return None

    key = "chunk_id" if by_chunk else "filename"
    ranked = [str(f.get(key) or "").strip().lower() for f in fragments]

    recall = len(expected_set & set(ranked)) / len(expected_set)

    dcg = 0.0
    seen: set[str] = set()
    for rank, item in enumerate(ranked):
        if item in expected_set and item not in seen:
            dcg += 1.0 / math.log2(rank + 2)
            seen.add(item)
    ideal = sum(1.0 / math.log2(r + 2) for r in range(min(len(expected_set), max(len(ranked), 1))))
    ndcg = dcg / ideal if ideal else 0.0

    # Citation precision@k: fraction of retrieved items that are relevant goldens.
    # Empty retrieval → 0 when goldens exist (cannot cite correctly with no hits).
    if ranked:
        relevant_hits = sum(1 for item in ranked if item in expected_set)
        citation_precision = relevant_hits / len(ranked)
    else:
        citation_precision = 0.0

    return {
        "recall_at_k": round(recall, 3),
        "ndcg_at_k": round(ndcg, 3),
        "citation_precision_at_k": round(citation_precision, 3),
        "k": len(ranked),
        "granularity": "chunk" if by_chunk else "source",
    }


def check_retrieval(case: Dict[str, Any], ctx: Dict[str, Any]) -> Dict[str, Any]:
    ok = ctx.get("success") is True and resp_status_ok(ctx)
    context_text = (ctx.get("context") or "").lower()
    fragments = ctx.get("fragments") or []

    missing = []
    for sub in case.get("expect_contains") or []:
        if sub.lower() not in context_text:
            missing.append(sub)

    forbidden = []
    for sub in case.get("expect_not_contains") or []:
        if sub.lower() in context_text:
            forbidden.append(sub)

    if case.get("expect_out_of_scope"):
        soft = (
            (not context_text.strip() and len(fragments) == 0)
            or _is_out_of_scope_error(ctx.get("error") or "")
        )
        if case.get("expect_not_contains"):
            ok = soft or not forbidden
        else:
            ok = soft or ok
    else:
        if case.get("expect_context", True) and ok:
            ok = bool(context_text.strip()) or len(fragments) > 0

    passed = ok and not missing and not forbidden

    return {
        "passed": passed,
        "retrieval_ok": ctx.get("success"),
        "missing_in_context": missing,
        "forbidden_in_context": forbidden,
        "fragment_count": len(fragments),
    }


def resp_status_ok(ctx: Dict[str, Any]) -> bool:
    return 200 <= int(ctx.get("http_status", 0)) < 300


def run_suite(
    suite_name: str,
    path: Path,
    rag_url: str,
    timeout: int,
) -> Dict[str, Any]:
    cases = load_cases(path)
    results = []
    passed = 0
    recalls: List[float] = []
    ndcgs: List[float] = []
    precisions: List[float] = []
    for i, case in enumerate(cases):
        q = case["question"]
        domain_id = case.get("domain_id", "default")
        ctx = fetch_context(rag_url, q, domain_id, timeout)
        check = check_retrieval(case, ctx)
        ranking = ranking_metrics(case, ctx.get("fragments") or [])
        if ranking:
            check["ranking"] = ranking
            recalls.append(ranking["recall_at_k"])
            ndcgs.append(ranking["ndcg_at_k"])
            precisions.append(ranking["citation_precision_at_k"])
        if check["passed"]:
            passed += 1
        results.append(
            {
                "index": i,
                "category": case.get("category"),
                "question": q,
                "domain_id": domain_id,
                "check": check,
                "rag_error": ctx.get("error"),
            }
        )
    total = len(cases)
    summary: Dict[str, Any] = {
        "suite": suite_name,
        "total": total,
        "passed": passed,
        "pass_rate": round(passed / total, 3) if total else 0.0,
        "cases": results,
    }
    if recalls:
        summary["ranking"] = {
            "cases_with_goldens": len(recalls),
            "mean_recall_at_k": round(sum(recalls) / len(recalls), 3),
            "mean_ndcg_at_k": round(sum(ndcgs) / len(ndcgs), 3),
            "mean_citation_precision_at_k": round(sum(precisions) / len(precisions), 3),
        }
    return summary


def main() -> int:
    parser = argparse.ArgumentParser(description="RAG eval (retrieval)")
    parser.add_argument(
        "--suite",
        default="default_en",
        help="Question suite name (see eval/rag_{suite}_baseline.jsonl)",
    )
    parser.add_argument(
        "--rag-url",
        default=os.environ.get("PYTHON_RAG_URL", "http://localhost:5000/rag/context"),
    )
    parser.add_argument("--timeout", type=int, default=120)
    args = parser.parse_args()

    suites_map = get_suites()
    if args.suite == "all":
        suites = suites_for_all()
    else:
        if args.suite not in suites_map:
            print(f"Unknown suite: {args.suite}. Available: {', '.join(sorted(suites_map))}", file=sys.stderr)
            return 1
        suites = [args.suite]
    report = {
        "mode": "retrieval",
        "rag_url": args.rag_url,
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "suites": [],
    }
    exit_code = 0
    for name in suites:
        path = suites_map[name]
        if not path.is_file():
            print(f"File not found: {path}", file=sys.stderr)
            exit_code = 1
            continue
        summary = run_suite(name, path, args.rag_url, args.timeout)
        report["suites"].append(summary)
        print(f"[{name}] {summary['passed']}/{summary['total']} passed ({summary['pass_rate']})")
        ranking = summary.get("ranking")
        if ranking:
            print(
                f"[{name}] ranking (n={ranking['cases_with_goldens']}): "
                f"Recall@k={ranking['mean_recall_at_k']} nDCG@k={ranking['mean_ndcg_at_k']} "
                f"CiteP@k={ranking['mean_citation_precision_at_k']}"
            )
            min_recall = float(os.environ.get("EVAL_MIN_RECALL", "0") or 0)
            min_ndcg = float(os.environ.get("EVAL_MIN_NDCG", "0") or 0)
            min_cite = float(os.environ.get("EVAL_MIN_CITATION_PRECISION", "0") or 0)
            if (
                ranking["mean_recall_at_k"] < min_recall
                or ranking["mean_ndcg_at_k"] < min_ndcg
                or ranking["mean_citation_precision_at_k"] < min_cite
            ):
                print(
                    f"  RANKING GATE FAILED: recall>={min_recall} ndcg>={min_ndcg} "
                    f"cite_precision>={min_cite}",
                    file=sys.stderr,
                )
                exit_code = 1
        if summary["passed"] < summary["total"]:
            exit_code = 1
            for c in summary["cases"]:
                if not c["check"]["passed"]:
                    miss = c["check"]["missing_in_context"]
                    forbid = c["check"].get("forbidden_in_context") or []
                    print(f"  FAIL: {c['question'][:60]}… missing={miss} forbidden={forbid}")

    RESULTS_DIR.mkdir(parents=True, exist_ok=True)
    stamp = datetime.now(timezone.utc).strftime("%Y%m%d_%H%M%S")
    out = RESULTS_DIR / f"{stamp}_{args.suite}.json"
    out.write_text(json.dumps(report, ensure_ascii=False, indent=2), encoding="utf-8")
    print(f"Report: {out}")
    return exit_code


if __name__ == "__main__":
    sys.exit(main())
