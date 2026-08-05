"""Validate config/domains.json against config/schemas/domains.schema.json."""

from __future__ import annotations

import json
from pathlib import Path

import pytest

ROOT = Path(__file__).resolve().parents[1]
DOMAINS_PATH = ROOT / "config" / "domains.json"
SCHEMA_PATH = ROOT / "config" / "schemas" / "domains.schema.json"


def test_domains_schema_files_exist():
    assert DOMAINS_PATH.is_file(), f"missing {DOMAINS_PATH}"
    assert SCHEMA_PATH.is_file(), f"missing {SCHEMA_PATH}"


def test_domains_json_matches_schema():
    jsonschema = pytest.importorskip("jsonschema")
    Draft202012Validator = jsonschema.Draft202012Validator

    schema = json.loads(SCHEMA_PATH.read_text(encoding="utf-8"))
    data = json.loads(DOMAINS_PATH.read_text(encoding="utf-8"))
    Draft202012Validator.check_schema(schema)
    errors = sorted(Draft202012Validator(schema).iter_errors(data), key=lambda e: list(e.path))
    assert not errors, "\n".join(f"{list(e.path)}: {e.message}" for e in errors)


def test_default_domain_is_listed():
    data = json.loads(DOMAINS_PATH.read_text(encoding="utf-8"))
    default = data["default_domain"]
    assert default in data["domains"]


def test_domain_ids_are_stable_slugs():
    data = json.loads(DOMAINS_PATH.read_text(encoding="utf-8"))
    for domain_id in data["domains"]:
        assert domain_id == domain_id.lower()
        assert domain_id.replace("_", "").replace("-", "").isalnum()
