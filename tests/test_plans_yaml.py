"""Validate config/plans.yaml scaffold."""

from pathlib import Path

import yaml


def test_plans_yaml_loads():
    path = Path(__file__).resolve().parents[1] / "config" / "plans.yaml"
    data = yaml.safe_load(path.read_text(encoding="utf-8"))
    assert data["version"] == 1
    assert "starter" in data["plans"]
    assert data["plans"]["business"]["quotas"]["domains"] == 10
