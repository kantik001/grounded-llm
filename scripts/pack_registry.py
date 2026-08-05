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
PACK_SCHEMA_PATH = PACKS_DIR / "schemas" / "pack.schema.json"


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


def validate_pack_manifest(manifest: dict[str, Any], *, pack_id: str = "") -> list[str]:
    """Validate pack.yaml content against packs/schemas/pack.schema.json."""
    errors: list[str] = []
    label = pack_id or (manifest.get("pack") or "pack")
    try:
        from jsonschema import Draft202012Validator
    except ImportError:
        # Registry CLI still works without jsonschema; CI tests install it via tests/.
        if not isinstance(manifest.get("domain"), dict) or not (manifest.get("domain") or {}).get("id"):
            errors.append(f"{label}: domain.id required")
        if not (manifest.get("eval") or {}).get("suite"):
            errors.append(f"{label}: eval.suite required")
        return errors

    if not PACK_SCHEMA_PATH.is_file():
        errors.append(f"missing schema: {PACK_SCHEMA_PATH}")
        return errors
    schema = json.loads(PACK_SCHEMA_PATH.read_text(encoding="utf-8"))
    Draft202012Validator.check_schema(schema)
    for err in sorted(Draft202012Validator(schema).iter_errors(manifest), key=lambda e: list(e.path)):
        errors.append(f"{label}: {list(err.path)}: {err.message}")
    pack_name = (manifest.get("pack") or "").strip()
    if pack_id and pack_name and pack_name != pack_id:
        errors.append(f"{label}: pack field {pack_name!r} != folder id {pack_id!r}")
    return errors


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
            if not isinstance(manifest, dict):
                errors.append(f"{pack_id}: invalid pack.yaml")
            else:
                errors.extend(validate_pack_manifest(manifest, pack_id=pack_id))
                domain = (manifest.get("domain") or {}).get("id")
                if entry.get("domain_id") and domain and entry["domain_id"] != domain:
                    errors.append(
                        f"{pack_id}: registry domain_id {entry['domain_id']} != pack.yaml {domain}"
                    )
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
