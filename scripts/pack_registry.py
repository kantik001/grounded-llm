"""Load and validate the official template pack registry."""

from __future__ import annotations

import json
import os
from pathlib import Path
from typing import Any

try:
    import yaml
except ImportError as exc:  # pragma: no cover
    raise SystemExit("pack registry requires PyYAML: pip install pyyaml") from exc

ROOT = Path(os.environ.get("GROUNDED_LLM_ROOT", Path(__file__).resolve().parents[1])).resolve()
PACKS_DIR = ROOT / "packs"
REGISTRY_PATH = PACKS_DIR / "registry.yaml"


def load_registry(path: Path | None = None) -> dict[str, Any]:
    registry_path = path or REGISTRY_PATH
    if not registry_path.is_file():
        raise FileNotFoundError(f"Registry not found: {registry_path}")
    with registry_path.open(encoding="utf-8") as f:
        data = yaml.safe_load(f) or {}
    if not isinstance(data, dict):
        raise ValueError(f"Invalid registry: {registry_path}")
    return data


def build_registry_index(registry: dict[str, Any] | None = None) -> list[dict[str, Any]]:
    registry = registry or load_registry()
    entries = registry.get("packs") or []
    if not isinstance(entries, list):
        raise ValueError("registry.yaml: packs must be a list")
    return entries


def validate_registry(registry: dict[str, Any] | None = None) -> list[str]:
    """Return list of validation errors (empty = OK)."""
    errors: list[str] = []
    registry = registry or load_registry()
    entries = build_registry_index(registry)

    seen_ids: set[str] = set()
    for i, entry in enumerate(entries):
        if not isinstance(entry, dict):
            errors.append(f"packs[{i}]: must be a mapping")
            continue
        pack_id = (entry.get("id") or "").strip()
        if not pack_id:
            errors.append(f"packs[{i}]: missing id")
            continue
        if pack_id in seen_ids:
            errors.append(f"duplicate pack id: {pack_id}")
        seen_ids.add(pack_id)

        pack_dir = PACKS_DIR / pack_id
        for rel in ("pack.yaml", "eval.jsonl", "data"):
            target = pack_dir / rel if rel != "data" else pack_dir / "data"
            if not target.exists():
                errors.append(f"{pack_id}: missing {rel}")

        manifest_path = pack_dir / "pack.yaml"
        if manifest_path.is_file():
            with manifest_path.open(encoding="utf-8") as f:
                manifest = yaml.safe_load(f) or {}
            domain = (manifest.get("domain") or {}).get("id")
            if entry.get("domain_id") and domain and entry["domain_id"] != domain:
                errors.append(f"{pack_id}: registry domain_id {entry['domain_id']} != pack.yaml {domain}")
            eval_suite = (manifest.get("eval") or {}).get("suite")
            if entry.get("eval_suite") and eval_suite and entry["eval_suite"] != eval_suite:
                errors.append(f"{pack_id}: registry eval_suite mismatch")

        guide = entry.get("guide")
        if guide and not (ROOT / str(guide)).is_file():
            errors.append(f"{pack_id}: guide not found: {guide}")

        eval_baseline = ROOT / "eval" / f"rag_{entry.get('eval_suite', pack_id)}_baseline.jsonl"
        if entry.get("eval_suite") and not eval_baseline.is_file():
            errors.append(f"{pack_id}: eval baseline missing: {eval_baseline.relative_to(ROOT)}")

    for name in sorted(p.name for p in PACKS_DIR.iterdir() if p.is_dir() and (p / "pack.yaml").is_file()):
        if name not in seen_ids:
            errors.append(f"pack {name} has pack.yaml but is not listed in registry.yaml")

    return errors


def export_registry_json(registry: dict[str, Any] | None = None) -> str:
    registry = registry or load_registry()
    payload = {
        "version": registry.get("version", 1),
        "packs": build_registry_index(registry),
    }
    return json.dumps(payload, ensure_ascii=False, indent=2)
