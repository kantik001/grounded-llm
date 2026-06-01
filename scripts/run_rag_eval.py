#!/usr/bin/env python3
"""
Прогон eval-набора RAG: POST /rag/context и проверка метрик.
Режим по умолчанию — retrieval (без LLM). End-to-end чат — через POST /message (Web App / API).
"""

from __future__ import annotations

import argparse
import json
import os
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Dict, List

import requests

_ROOT = Path(__file__).resolve().parents[1]
EVAL_DIR = _ROOT / "eval"
RESULTS_DIR = EVAL_DIR / "results"

SUITES = {
    "default": EVAL_DIR / "rag_default_baseline.jsonl",
}


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


def check_retrieval(case: Dict[str, Any], ctx: Dict[str, Any]) -> Dict[str, Any]:
    ok = ctx.get("success") is True and resp_status_ok(ctx)
    context_text = (ctx.get("context") or "").lower()
    fragments = ctx.get("fragments") or []

    if case.get("expect_out_of_scope"):
        # Допускаем soft fail от RAG или короткий контекст
        ok = ok or (not context_text.strip()) or "нет" in (ctx.get("error") or "").lower()
    else:
        if case.get("expect_context", True) and ok:
            ok = bool(context_text.strip()) or len(fragments) > 0

    missing = []
    for sub in case.get("expect_contains") or []:
        if sub.lower() not in context_text:
            missing.append(sub)

    return {
        "passed": ok and not missing,
        "retrieval_ok": ctx.get("success"),
        "missing_in_context": missing,
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
    for i, case in enumerate(cases):
        q = case["question"]
        domain_id = case.get("domain_id", "default")
        ctx = fetch_context(rag_url, q, domain_id, timeout)
        check = check_retrieval(case, ctx)
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
    return {
        "suite": suite_name,
        "total": total,
        "passed": passed,
        "pass_rate": round(passed / total, 3) if total else 0.0,
        "cases": results,
    }


def main() -> int:
    parser = argparse.ArgumentParser(description="RAG eval (retrieval)")
    parser.add_argument(
        "--suite",
        choices=["default", "all"],
        default="default",
        help="Набор вопросов",
    )
    parser.add_argument(
        "--rag-url",
        default=os.environ.get("PYTHON_RAG_URL", "http://localhost:5000/rag/context"),
    )
    parser.add_argument("--timeout", type=int, default=120)
    args = parser.parse_args()

    suites = list(SUITES.keys()) if args.suite == "all" else [args.suite]
    report = {
        "mode": "retrieval",
        "rag_url": args.rag_url,
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "suites": [],
    }
    exit_code = 0
    for name in suites:
        path = SUITES[name]
        if not path.is_file():
            print(f"Нет файла: {path}", file=sys.stderr)
            exit_code = 1
            continue
        summary = run_suite(name, path, args.rag_url, args.timeout)
        report["suites"].append(summary)
        print(f"[{name}] {summary['passed']}/{summary['total']} passed ({summary['pass_rate']})")
        if summary["passed"] < summary["total"]:
            exit_code = 1
            for c in summary["cases"]:
                if not c["check"]["passed"]:
                    print(f"  FAIL: {c['question'][:60]}… missing={c['check']['missing_in_context']}")

    RESULTS_DIR.mkdir(parents=True, exist_ok=True)
    stamp = datetime.now(timezone.utc).strftime("%Y%m%d_%H%M%S")
    out = RESULTS_DIR / f"{stamp}_{args.suite}.json"
    out.write_text(json.dumps(report, ensure_ascii=False, indent=2), encoding="utf-8")
    print(f"Отчёт: {out}")
    return exit_code


if __name__ == "__main__":
    sys.exit(main())
