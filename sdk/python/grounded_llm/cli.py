"""Command-line interface for Grounded LLM."""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
from pathlib import Path

from grounded_llm import GroundedClient
from grounded_llm.exceptions import GroundedAPIError


def _repo_root() -> Path:
    return Path(__file__).resolve().parents[3]


def cmd_health(args: argparse.Namespace) -> int:
    client = GroundedClient(args.base_url, api_key=args.api_key, tenant_id=args.tenant)
    print(json.dumps(client.health(), indent=2))
    return 0


def cmd_domains(args: argparse.Namespace) -> int:
    client = GroundedClient(args.base_url, api_key=args.api_key, tenant_id=args.tenant, locale=args.locale)
    print(json.dumps(client.list_domains(), indent=2))
    return 0


def cmd_chat(args: argparse.Namespace) -> int:
    client = GroundedClient(args.base_url, api_key=args.api_key, tenant_id=args.tenant, locale=args.locale)
    try:
        if args.stream:
            sid = args.session or client.create_session(domain_id=args.domain)
            print(f"session_id={sid}\n", file=sys.stderr)
            for delta in client.stream_message_deltas(args.text, session_id=sid, domain_id=args.domain):
                print(delta, end="", flush=True)
            print()
            return 0
        result = client.chat(args.text, domain_id=args.domain, session_id=args.session)
        assistant = result.last_assistant_message
        if assistant:
            print(assistant.get("content", ""))
            citations = assistant.get("citations") or []
            if citations and args.show_sources:
                print("\n--- Sources ---", file=sys.stderr)
                for c in citations:
                    print(f"  - {c.get('filename', c)}", file=sys.stderr)
        print(json.dumps(result.raw, indent=2) if args.json else "", end="")
        if args.json:
            print()
        print(f"session_id={result.session_id}", file=sys.stderr)
        return 0
    except GroundedAPIError as exc:
        print(f"Error: {exc}", file=sys.stderr)
        return 1


def cmd_eval(args: argparse.Namespace) -> int:
    root = _repo_root()
    script = root / "scripts" / "run_rag_eval.py"
    if not script.is_file():
        print("eval requires repo checkout with scripts/run_rag_eval.py", file=sys.stderr)
        return 1
    cmd = [sys.executable, str(script), "--suite", args.suite]
    if args.url:
        env = os.environ.copy()
        env["PYTHON_RAG_URL"] = args.url
        return subprocess.call(cmd, cwd=str(root), env=env)
    return subprocess.call(cmd, cwd=str(root))


def cmd_pack(args: argparse.Namespace) -> int:
    root = _repo_root()
    script = root / "scripts" / "init_pack.py"
    if not script.is_file():
        print("pack requires repo checkout with scripts/init_pack.py", file=sys.stderr)
        return 1
    cmd = [sys.executable, str(script), args.action]
    if args.action == "install":
        cmd.append(args.name)
        if args.tenant:
            cmd.extend(["--tenant", args.tenant])
        if args.force:
            cmd.append("--force")
    elif args.action == "new":
        cmd.extend([args.name, args.domain_id])
    return subprocess.call(cmd, cwd=str(root))


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(prog="grounded-llm", description="Grounded LLM API client and tools")
    p.add_argument("--base-url", default=os.environ.get("GROUNDED_BASE_URL", "http://localhost:8080"))
    p.add_argument("--api-key", default=os.environ.get("GROUNDED_API_KEY") or os.environ.get("X_API_KEY"))
    p.add_argument("--tenant", default=os.environ.get("GROUNDED_TENANT_ID", "default"))
    p.add_argument("--locale", default=os.environ.get("GROUNDED_LOCALE", "en"))

    sub = p.add_subparsers(dest="command", required=True)

    sub.add_parser("health", help="GET /health").set_defaults(func=cmd_health)
    sub.add_parser("domains", help="GET /api/domains").set_defaults(func=cmd_domains)

    chat = sub.add_parser("chat", help="Send a chat message")
    chat.add_argument("text", help="User question")
    chat.add_argument("--domain", default="default")
    chat.add_argument("--session", help="Existing session_id")
    chat.add_argument("--stream", action="store_true", help="SSE streaming output")
    chat.add_argument("--json", action="store_true", help="Print full JSON response")
    chat.add_argument("--show-sources", action="store_true", default=True)
    chat.add_argument("--no-show-sources", action="store_false", dest="show_sources")
    chat.set_defaults(func=cmd_chat)

    ev = sub.add_parser("eval", help="Run retrieval eval (requires repo scripts/)")
    ev.add_argument("--suite", default="default_en")
    ev.add_argument("--url", help="PYTHON_RAG_URL override")
    ev.set_defaults(func=cmd_eval)

    pack = sub.add_parser("pack", help="Template pack tools (requires repo scripts/)")
    pack_sub = pack.add_subparsers(dest="action", required=True)
    pl = pack_sub.add_parser("list")
    pl.set_defaults(func=cmd_pack, name="")
    pi = pack_sub.add_parser("install")
    pi.add_argument("name", help="Pack name, e.g. hr or it_support")
    pi.add_argument("--tenant", default="default")
    pi.add_argument("--force", action="store_true")
    pi.set_defaults(func=cmd_pack)
    pn = pack_sub.add_parser("new")
    pn.add_argument("name")
    pn.add_argument("domain_id")
    pn.set_defaults(func=cmd_pack)

    return p


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    return int(args.func(args))


if __name__ == "__main__":
    raise SystemExit(main())
