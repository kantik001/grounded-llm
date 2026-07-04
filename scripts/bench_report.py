"""Generate benchmark summary JSON from eval runner or latest results."""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
RESULTS_DIR = ROOT / "eval" / "results"
EVAL_SCRIPT = ROOT / "scripts" / "run_rag_eval.py"


def run_eval(suite: str, rag_url: str) -> dict:
    env = os.environ.copy()
    env["PYTHON_RAG_URL"] = rag_url
    proc = subprocess.run(
        [sys.executable, str(EVAL_SCRIPT), "--suite", suite],
        cwd=str(ROOT),
        env=env,
        capture_output=True,
        text=True,
    )
    if proc.returncode != 0:
        print(proc.stdout, proc.stderr, file=sys.stderr)
        raise SystemExit(f"eval failed for suite {suite}")
    # Parse last results file for this suite
    pattern = f"*_{suite}.json"
    files = sorted(RESULTS_DIR.glob(pattern), key=lambda p: p.stat().st_mtime)
    if not files:
        raise SystemExit(f"no results file matching {pattern}")
    report = json.loads(files[-1].read_text(encoding="utf-8"))
    suites = report.get("suites") or []
    if len(suites) == 1:
        s = suites[0]
        return {
            "passed": s["passed"],
            "total": s["total"],
            "pass_rate": s["pass_rate"],
        }
    return {"passed": 0, "total": 0, "pass_rate": 0}


def discover_suites() -> list[str]:
    suites = []
    for path in sorted((ROOT / "eval").glob("rag_*_baseline.jsonl")):
        stem = path.stem
        if stem.startswith("rag_") and stem.endswith("_baseline"):
            name = stem[4:-9]
            if name:
                suites.append(name)
    return suites


def main() -> int:
    parser = argparse.ArgumentParser(description="Grounded benchmark report")
    parser.add_argument("--suite", default="all", help="Eval suite name or 'all'")
    parser.add_argument(
        "--rag-url",
        default=os.environ.get("PYTHON_RAG_URL", "http://localhost:5000/rag/context"),
    )
    parser.add_argument(
        "--write",
        default="",
        help="Write JSON summary to path (default: eval/results/latest_bench.json)",
    )
    parser.add_argument("--version", default=os.environ.get("GROUNDED_VERSION", "dev"))
    args = parser.parse_args()

    suite_names = discover_suites() if args.suite == "all" else [args.suite]
    summary = {
        "reference_impl": "grounded-llm",
        "version": args.version,
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "rag_url": args.rag_url,
        "suites": {},
    }
    for name in suite_names:
        print(f"Running suite: {name}")
        summary["suites"][name] = run_eval(name, args.rag_url)

    out_path = Path(args.write) if args.write else RESULTS_DIR / "latest_bench.json"
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(summary, indent=2), encoding="utf-8")
    print(f"Wrote {out_path}")
    for name, stats in summary["suites"].items():
        print(f"  {name}: {stats['passed']}/{stats['total']} ({stats['pass_rate']})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
