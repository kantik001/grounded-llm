"""Unit tests for config/domains.json."""

import os

import pytest

from rag.domains_config import get_domain, list_domains, normalize_domain_id


@pytest.fixture(autouse=True)
def domains_config_path(monkeypatch):
    root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
    monkeypatch.setenv("DOMAINS_CONFIG_PATH", os.path.join(root, "config", "domains.json"))
    import rag.domains_config as dc

    dc._CONFIG = None
    dc._CONFIG_MTIME = None


def test_normalize_domain_id_default():
    assert normalize_domain_id("default") == "default"
    assert normalize_domain_id(" Default ") == "default"


def test_normalize_domain_id_unknown():
    with pytest.raises(ValueError, match="Неизвестный"):
        normalize_domain_id("banana_xyz")


def test_list_domains_has_default():
    data = list_domains()
    assert data["default_domain"] == "default"
    assert "default" in data["domains"]
    assert get_domain("default").get("rag_enabled") is True
