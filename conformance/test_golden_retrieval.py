"""Golden retrieval conformance — optional, requires indexed Python RAG."""

import os
import subprocess
import sys
from pathlib import Path

import pytest

ROOT = Path(__file__).resolve().parents[1]
RAG_URL = os.environ.get("CONFORMANCE_RAG_URL", "")

pytestmark = pytest.mark.skipif(not RAG_URL, reason="Set CONFORMANCE_RAG_URL for golden retrieval")


def test_all_retrieval_suites_pass():
    env = os.environ.copy()
    env["PYTHON_RAG_URL"] = RAG_URL
    script = ROOT / "scripts" / "run_rag_eval.py"
    result = subprocess.run(
        [sys.executable, str(script), "--suite", "all"],
        cwd=str(ROOT),
        env=env,
        capture_output=True,
        text=True,
        timeout=600,
    )
    assert result.returncode == 0, result.stdout + result.stderr
