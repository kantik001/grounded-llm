"""Grounded conformance CLI — check API v1 compatibility."""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
CONFORMANCE_DIR = Path(__file__).resolve().parent


def _emit(args: argparse.Namespace, command: str, exit_code: int, **extra) -> int:
    if getattr(args, "json", False):
        payload = {"command": command, "passed": exit_code == 0, **extra}
        print(json.dumps(payload, indent=2))
    return exit_code


def _run_pytest(
    args: argparse.Namespace,
    command: str,
    target: str,
    extra_env: dict | None = None,
) -> int:
    env = os.environ.copy()
    if extra_env:
        env.update(extra_env)
    cmd = [
        sys.executable,
        "-m",
        "pytest",
        str(CONFORMANCE_DIR / target),
        "-q" if getattr(args, "json", False) else "-v",
        "--tb=short",
    ]
    if getattr(args, "json", False):
        result = subprocess.run(cmd, cwd=str(ROOT), env=env, capture_output=True, text=True)
        if result.stdout.strip():
            print(result.stdout.strip(), file=sys.stderr)
        exit_code = result.returncode
    else:
        print("+", " ".join(cmd))
        exit_code = subprocess.call(cmd, cwd=str(ROOT), env=env)
    return _emit(args, command, exit_code, target=target)


def cmd_spec(args: argparse.Namespace) -> int:
    return _run_pytest(args, "spec", "test_openapi_spec.py")


def cmd_http(args: argparse.Namespace) -> int:
    return _run_pytest(
        args,
        "http",
        "test_openapi_http.py",
        {"CONFORMANCE_BASE_URL": args.url.rstrip("/"), "CONFORMANCE_SKIP_HTTP": ""},
    )


def cmd_retrieval(args: argparse.Namespace) -> int:
    return _run_pytest(
        args,
        "retrieval",
        "test_golden_retrieval.py",
        {"CONFORMANCE_RAG_URL": args.rag_url.rstrip("/")},
    )


def cmd_check(args: argparse.Namespace) -> int:
    code = cmd_spec(args)
    if code != 0:
        return code
    if args.url:
        return cmd_http(args)
    if not getattr(args, "json", False):
        print("Skip HTTP checks (pass --url for full check)")
    return _emit(args, "check", 0, http="skipped")


def cmd_all(args: argparse.Namespace) -> int:
    code = cmd_check(args)
    if code != 0:
        return code
    if args.rag_url:
        return cmd_retrieval(args)
    return _emit(args, "all", 0, retrieval="skipped")


def _add_json_flag(parser: argparse.ArgumentParser) -> None:
    parser.add_argument(
        "--json",
        action="store_true",
        help="Print JSON result to stdout (quiet pytest)",
    )


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="grounded-conformance",
        description="Grounded LLM API v1 conformance checks",
    )
    sub = parser.add_subparsers(dest="command", required=True)

    p_spec = sub.add_parser("spec", help="Validate OpenAPI spec (offline)")
    _add_json_flag(p_spec)
    p_spec.set_defaults(func=cmd_spec)

    p_http = sub.add_parser("http", help="HTTP checks against running server")
    p_http.add_argument("--url", required=True, help="Base URL, e.g. http://127.0.0.1:8080")
    _add_json_flag(p_http)
    p_http.set_defaults(func=cmd_http)

    p_ret = sub.add_parser("retrieval", help="Golden retrieval eval via RAG URL")
    p_ret.add_argument(
        "--rag-url",
        required=True,
        help="RAG context URL, e.g. http://127.0.0.1:5000/rag/context",
    )
    _add_json_flag(p_ret)
    p_ret.set_defaults(func=cmd_retrieval)

    p_check = sub.add_parser("check", help="spec + optional http")
    p_check.add_argument("--url", default="", help="If set, run HTTP checks too")
    _add_json_flag(p_check)
    p_check.set_defaults(func=cmd_check)

    p_all = sub.add_parser("all", help="check + optional retrieval")
    p_all.add_argument("--url", default="", help="Server base URL")
    p_all.add_argument("--rag-url", default="", help="RAG context URL")
    _add_json_flag(p_all)
    p_all.set_defaults(func=cmd_all)

    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
