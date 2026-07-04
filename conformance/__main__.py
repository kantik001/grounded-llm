"""Grounded conformance CLI — check API v1 compatibility."""

from __future__ import annotations

import argparse
import os
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
CONFORMANCE_DIR = Path(__file__).resolve().parent


def _run_pytest(target: str, extra_env: dict | None = None) -> int:
    env = os.environ.copy()
    if extra_env:
        env.update(extra_env)
    cmd = [
        sys.executable,
        "-m",
        "pytest",
        str(CONFORMANCE_DIR / target),
        "-v",
        "--tb=short",
    ]
    print("+", " ".join(cmd))
    return subprocess.call(cmd, cwd=str(ROOT), env=env)


def cmd_spec(_: argparse.Namespace) -> int:
    return _run_pytest("test_openapi_spec.py")


def cmd_http(args: argparse.Namespace) -> int:
    return _run_pytest(
        "test_openapi_http.py",
        {"CONFORMANCE_BASE_URL": args.url.rstrip("/"), "CONFORMANCE_SKIP_HTTP": ""},
    )


def cmd_retrieval(args: argparse.Namespace) -> int:
    return _run_pytest(
        "test_golden_retrieval.py",
        {"CONFORMANCE_RAG_URL": args.rag_url.rstrip("/")},
    )


def cmd_check(args: argparse.Namespace) -> int:
    code = cmd_spec(args)
    if code != 0:
        return code
    if args.url:
        return cmd_http(args)
    print("Skip HTTP checks (pass --url for full check)")
    return 0


def cmd_all(args: argparse.Namespace) -> int:
    code = cmd_check(args)
    if code != 0:
        return code
    if args.rag_url:
        return cmd_retrieval(args)
    return 0


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="grounded-conformance",
        description="Grounded LLM API v1 conformance checks",
    )
    sub = parser.add_subparsers(dest="command", required=True)

    p_spec = sub.add_parser("spec", help="Validate OpenAPI spec (offline)")
    p_spec.set_defaults(func=cmd_spec)

    p_http = sub.add_parser("http", help="HTTP checks against running server")
    p_http.add_argument("--url", required=True, help="Base URL, e.g. http://127.0.0.1:8080")
    p_http.set_defaults(func=cmd_http)

    p_ret = sub.add_parser("retrieval", help="Golden retrieval eval via RAG URL")
    p_ret.add_argument(
        "--rag-url",
        required=True,
        help="RAG context URL, e.g. http://127.0.0.1:5000/rag/context",
    )
    p_ret.set_defaults(func=cmd_retrieval)

    p_check = sub.add_parser("check", help="spec + optional http")
    p_check.add_argument("--url", default="", help="If set, run HTTP checks too")
    p_check.set_defaults(func=cmd_check)

    p_all = sub.add_parser("all", help="check + optional retrieval")
    p_all.add_argument("--url", default="", help="Server base URL")
    p_all.add_argument("--rag-url", default="", help="RAG context URL")
    p_all.set_defaults(func=cmd_all)

    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
