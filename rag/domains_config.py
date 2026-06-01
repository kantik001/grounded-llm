"""Load config/domains.json — knowledge-domain catalog for RAG."""

import json
import os
from typing import Any, Dict, Optional

_PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))

_CONFIG: Optional[Dict[str, Any]] = None
_CONFIG_MTIME: Optional[float] = None


def _config_path() -> str:
    env = os.environ.get("DOMAINS_CONFIG_PATH") or os.environ.get("CROPS_CONFIG_PATH")
    if env and os.path.isfile(env):
        return env
    for candidate in (
        os.path.join(_PROJECT_ROOT, "config", "domains.json"),
        "/config/domains.json",
        os.path.join(_PROJECT_ROOT, "config", "crops.json"),
        "/config/crops.json",
    ):
        if os.path.isfile(candidate):
            return candidate
    return os.path.join(_PROJECT_ROOT, "config", "domains.json")


def _normalize_catalog(raw: Dict[str, Any]) -> Dict[str, Any]:
    """Support legacy crops.json (default_crop / crops keys)."""
    if "domains" in raw:
        return raw
    domains = raw.get("crops", {})
    default = raw.get("default_crop") or raw.get("default_domain") or "default"
    out: Dict[str, Any] = {"default_domain": default, "domains": {}}
    for did, info in domains.items():
        entry = dict(info)
        if "name" not in entry and "name_ru" in entry:
            entry["name"] = entry["name_ru"]
        out["domains"][did] = entry
    return out


def load_domains_config() -> Dict[str, Any]:
    global _CONFIG, _CONFIG_MTIME
    path = _config_path()
    try:
        mtime = os.path.getmtime(path)
    except OSError:
        mtime = None
    if _CONFIG is not None and _CONFIG_MTIME == mtime:
        return _CONFIG
    with open(path, encoding="utf-8") as f:
        _CONFIG = _normalize_catalog(json.load(f))
    _CONFIG_MTIME = mtime
    return _CONFIG


def reload_domains_config() -> Dict[str, Any]:
    global _CONFIG, _CONFIG_MTIME
    _CONFIG = None
    _CONFIG_MTIME = None
    return load_domains_config()


def default_domain_id() -> str:
    return load_domains_config().get("default_domain", "default")


def normalize_domain_id(domain_id: Optional[str]) -> str:
    did = (domain_id or "").strip().lower() or default_domain_id()
    domains = load_domains_config().get("domains", {})
    if did not in domains:
        raise ValueError(f"Неизвестный домен: {domain_id}")
    return did


def get_domain(domain_id: str) -> Dict[str, Any]:
    did = normalize_domain_id(domain_id)
    return load_domains_config()["domains"][did]


def list_domains() -> Dict[str, Any]:
    cfg = load_domains_config()
    return {
        "default_domain": cfg.get("default_domain", "default"),
        "domains": cfg.get("domains", {}),
    }
