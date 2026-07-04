"""Validate OpenAPI spec file (offline — no server required)."""

import json
from pathlib import Path

from openapi_spec_validator import validate

ROOT = Path(__file__).resolve().parents[1]
OPENAPI_PATH = ROOT / "server" / "openapi.v1.json"


def test_openapi_file_exists():
    assert OPENAPI_PATH.is_file()


def test_openapi_validates():
    with OPENAPI_PATH.open(encoding="utf-8") as f:
        spec = json.load(f)
    validate(spec)
    assert spec["info"]["title"] == "Grounded LLM API"
    assert spec["openapi"].startswith("3.0")


def test_openapi_documents_v1_paths():
    with OPENAPI_PATH.open(encoding="utf-8") as f:
        spec = json.load(f)
    paths = spec.get("paths", {})
    assert "/health" in paths or any("health" in p for p in paths)
    assert any("session" in p for p in paths)
    assert any("message" in p for p in paths)
