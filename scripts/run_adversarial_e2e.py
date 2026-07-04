#!/usr/bin/env python3
"""Adversarial E2E checks via POST /message (requires LLM_MOCK + RAG_MOCK on server)."""

from __future__ import annotations

import argparse
import json
import os
import sys
from pathlib import Path
from typing import Any, Dict, List

import requests

_ROOT = Path(__file__).resolve().parents[1]
E2E_PATH = _ROOT / "eval" / "rag_adversarial_e2e.jsonl"


def load_cases(path: Path) -> List[Dict[str, Any]]:
    cases = []
    with path.open(encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            cases.append(json.loads(line))
    return cases


def last_assistant(data: dict) -> dict | None:
    for msg in reversed(data.get("messages") or []):
        if msg.get("role") == "assistant":
            return msg
    return None


def check_case(case: dict, base_url: str, timeout: int) -> tuple[bool, str]:
    session_resp = requests.post(
        f"{base_url}/api/session",
        json={"domain_id": case.get("domain_id", "default")},
        timeout=timeout,
    )
    session_resp.raise_for_status()
    session_id = session_resp.json().get("session_id")
    if not session_id:
        return False, "no session_id"

    msg_resp = requests.post(
        f"{base_url}/api/message",
        json={
            "session_id": session_id,
            "domain_id": case.get("domain_id", "default"),
            "text": case["question"],
        },
        timeout=timeout,
    )
    msg_resp.raise_for_status()
    data = msg_resp.json()

    if case.get("expect_out_of_scope_answer"):
        if not data.get("success"):
            return True, data.get("error") or "out of scope"
        assistant = last_assistant(data)
        if not assistant:
            return True, "no assistant message"
        citations = assistant.get("citations") or []
        if len(citations) == 0:
            return True, "no citations"
        return False, "expected out-of-scope but got cited answer"

    if not data.get("success"):
        return False, data.get("error") or "success=false"

    assistant = last_assistant(data)
    if not assistant:
        return False, "no assistant message"

    text = (assistant.get("content") or "").lower()
    citations = assistant.get("citations") or []
    min_cites = case.get("expect_citations_min")
    if min_cites is not None and len(citations) < int(min_cites):
        return False, f"citations {len(citations)} < {min_cites}"

    for sub in case.get("expect_answer_contains") or []:
        if sub.lower() not in text:
            return False, f"missing answer substring: {sub}"

    for sub in case.get("expect_answer_not_contains") or []:
        if sub.lower() in text:
            return False, f"forbidden answer substring: {sub}"

    if case.get("expect_verify_pass") is True:
        if "could not verify" in text or text.startswith("⚠"):
            return False, "verify failed"

    return True, text[:120]


def main() -> int:
    parser = argparse.ArgumentParser(description="Adversarial E2E via /message")
    parser.add_argument("--base-url", default=os.environ.get("BASE_URL", "http://127.0.0.1:8080"))
    parser.add_argument("--cases", type=Path, default=E2E_PATH)
    parser.add_argument("--timeout", type=int, default=120)
    args = parser.parse_args()

    if not args.cases.is_file():
        print(f"Missing cases file: {args.cases}", file=sys.stderr)
        return 1

    cases = load_cases(args.cases)
    failed = 0
    for i, case in enumerate(cases):
        ok, detail = check_case(case, args.base_url.rstrip("/"), args.timeout)
        label = case.get("adversarial_type") or case.get("category") or f"case_{i}"
        if ok:
            print(f"[OK] {label}")
        else:
            failed += 1
            print(f"[FAIL] {label}: {detail}")

    if failed:
        print(f"Adversarial E2E FAILED: {failed}/{len(cases)}")
        return 1
    print(f"Adversarial E2E PASSED ({len(cases)} cases)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
