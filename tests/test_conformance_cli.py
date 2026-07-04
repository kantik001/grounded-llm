"""Smoke tests for conformance CLI."""

import json
import subprocess
import sys


def test_conformance_spec_json():
    proc = subprocess.run(
        [sys.executable, "-m", "conformance", "spec", "--json"],
        capture_output=True,
        text=True,
    )
    assert proc.returncode == 0, proc.stdout + proc.stderr
    data = json.loads(proc.stdout.strip())
    assert data["command"] == "spec"
    assert data["passed"] is True
