"""Install and scaffold Grounded LLM template packs from packs/*/pack.yaml."""

from __future__ import annotations

import json
import os
import re
import shutil
from pathlib import Path
from typing import Any

try:
    import yaml
except ImportError as exc:  # pragma: no cover
    raise SystemExit("init-pack requires PyYAML: pip install pyyaml") from exc

ROOT = Path(os.environ.get("GROUNDED_LLM_ROOT", Path(__file__).resolve().parents[1])).resolve()
PACKS_DIR = ROOT / "packs"
DOMAIN_ID_RE = re.compile(r"^[a-z][a-z0-9_]*$")


def list_packs() -> list[str]:
    if not PACKS_DIR.is_dir():
        return []
    names = []
    for path in sorted(PACKS_DIR.iterdir()):
        if path.is_dir() and (path / "pack.yaml").is_file():
            names.append(path.name)
    return names


def load_pack_manifest(pack_name: str) -> tuple[Path, dict[str, Any]]:
    pack_dir = PACKS_DIR / pack_name
    manifest_path = pack_dir / "pack.yaml"
    if not manifest_path.is_file():
        raise FileNotFoundError(f"Pack not found: {pack_name} (missing {manifest_path})")
    with manifest_path.open(encoding="utf-8") as f:
        data = yaml.safe_load(f) or {}
    if not isinstance(data, dict):
        raise ValueError(f"Invalid pack manifest: {manifest_path}")
    return pack_dir, data


def _domain_id(manifest: dict[str, Any]) -> str:
    domain = manifest.get("domain") or {}
    domain_id = (domain.get("id") or manifest.get("id") or "").strip()
    if not domain_id or not DOMAIN_ID_RE.match(domain_id):
        raise ValueError("pack.yaml: domain.id must be a lowercase slug")
    return domain_id


def _locale(manifest: dict[str, Any]) -> str:
    loc = (manifest.get("locale") or "en").strip().lower()
    return loc if loc.startswith("ru") else "en"


def data_target_dir(tenant_id: str, domain_id: str) -> Path:
    """Resolve KB directory (legacy flat layout for tenant+domain both default)."""
    tenant_id = tenant_id.strip() or "default"
    if tenant_id == "default" and domain_id == "default":
        return ROOT / "data" / "default"
    return ROOT / "data" / tenant_id / domain_id


def _read_json(path: Path) -> dict | list:
    with path.open(encoding="utf-8") as f:
        return json.load(f)


def _write_json(path: Path, data: dict | list) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, indent=2)
        f.write("\n")


def merge_domains_json(manifest: dict[str, Any], *, force: bool = False) -> None:
    path = ROOT / "config" / "domains.json"
    cfg = _read_json(path)
    domains = cfg.setdefault("domains", {})
    domain_id = _domain_id(manifest)
    domain_cfg = manifest.get("domain") or {}
    entry = {
        "name": domain_cfg.get("name") or domain_id,
        "names": domain_cfg.get("names") or {"en": domain_id},
        "emoji": domain_cfg.get("emoji") or "📚",
        "rag_enabled": bool(domain_cfg.get("rag_enabled", True)),
        "rag_k": int(domain_cfg.get("rag_k", 8)),
    }
    if domain_id in domains and not force:
        return
    domains[domain_id] = entry
    _write_json(path, cfg)


def merge_locale_bundle(manifest: dict[str, Any], *, force: bool = False) -> None:
    locale = _locale(manifest)
    domain_id = _domain_id(manifest)
    locale_root = ROOT / "config" / "locales" / locale

    prompts_path = locale_root / "prompts.json"
    prompts = _read_json(prompts_path)
    if force or domain_id not in prompts:
        p = manifest.get("prompts") or {}
        prompts[domain_id] = {
            "rag_system": p.get("rag_system", ""),
            "rag_task_intro": p.get("rag_task_intro", ""),
        }
        _write_json(prompts_path, prompts)

    onboarding_path = locale_root / "onboarding.json"
    onboarding = _read_json(onboarding_path)
    if force or domain_id not in onboarding:
        onboarding[domain_id] = list(manifest.get("onboarding") or [])
        _write_json(onboarding_path, onboarding)

    few_shot_path = locale_root / "few_shot.json"
    few_shot = _read_json(few_shot_path)
    if force or domain_id not in few_shot:
        fs = manifest.get("few_shot") or {}
        few_shot[domain_id] = {"general": fs.get("general", "")}
        _write_json(few_shot_path, few_shot)


def copy_pack_data(pack_dir: Path, tenant_id: str, domain_id: str) -> Path:
    src = pack_dir / "data"
    if not src.is_dir():
        raise FileNotFoundError(f"Pack data/ missing: {src}")
    dest = data_target_dir(tenant_id, domain_id)
    dest.mkdir(parents=True, exist_ok=True)
    for item in src.iterdir():
        if item.is_file():
            shutil.copy2(item, dest / item.name)
    return dest


def copy_pack_eval(pack_dir: Path, manifest: dict[str, Any]) -> Path:
    eval_cfg = manifest.get("eval") or {}
    suite = (eval_cfg.get("suite") or _domain_id(manifest)).strip()
    src = pack_dir / "eval.jsonl"
    if not src.is_file():
        raise FileNotFoundError(f"Pack eval.jsonl missing: {src}")
    dest = ROOT / "eval" / f"rag_{suite}_baseline.jsonl"
    dest.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(src, dest)
    return dest


def install_pack(
    pack_name: str,
    *,
    tenant_id: str = "default",
    force: bool = False,
    dry_run: bool = False,
) -> dict[str, Any]:
    pack_dir, manifest = load_pack_manifest(pack_name)
    domain_id = _domain_id(manifest)
    eval_cfg = manifest.get("eval") or {}
    suite = (eval_cfg.get("suite") or domain_id).strip()

    plan = {
        "pack": pack_name,
        "domain_id": domain_id,
        "tenant_id": tenant_id,
        "data_dir": str(data_target_dir(tenant_id, domain_id)),
        "eval_suite": suite,
        "locale": _locale(manifest),
    }
    if dry_run:
        return plan

    merge_domains_json(manifest, force=force)
    merge_locale_bundle(manifest, force=force)
    copy_pack_data(pack_dir, tenant_id, domain_id)
    eval_path = copy_pack_eval(pack_dir, manifest)
    plan["eval_path"] = str(eval_path)
    return plan


def scaffold_new_pack(
    pack_name: str,
    *,
    domain_id: str | None = None,
    locale: str = "en",
    from_pack: str | None = None,
) -> Path:
    if not DOMAIN_ID_RE.match(pack_name):
        raise ValueError("pack name must be a lowercase slug")
    pack_dir = PACKS_DIR / pack_name
    if pack_dir.exists():
        raise FileExistsError(f"Pack already exists: {pack_dir}")

    if from_pack:
        src_dir, src_manifest = load_pack_manifest(from_pack)
        domain_id = domain_id or _domain_id(src_manifest)
        manifest = dict(src_manifest)
        manifest["pack"] = pack_name
        manifest["version"] = "0.1.0"
        manifest["description"] = f"Custom pack scaffolded from {from_pack}"
        if "domain" not in manifest:
            manifest["domain"] = {}
        manifest["domain"]["id"] = domain_id
        manifest["locale"] = locale
        eval_cfg = dict(manifest.get("eval") or {})
        eval_cfg["suite"] = domain_id if domain_id != "default" else f"{pack_name}_en"
        manifest["eval"] = eval_cfg
    else:
        domain_id = domain_id or pack_name
        manifest = {
            "pack": pack_name,
            "version": "0.1.0",
            "description": f"Template pack {pack_name}",
            "domain": {
                "id": domain_id,
                "emoji": "📚",
                "name": pack_name.replace("_", " ").title(),
                "names": {"en": pack_name.replace("_", " ").title()},
                "rag_enabled": True,
                "rag_k": 8,
            },
            "locale": locale,
            "prompts": {
                "rag_system": (
                    "You are an assistant for internal organizational documents. "
                    "Answer only from the provided context."
                ),
                "rag_task_intro": "Answer strictly based on the documents in the context.",
            },
            "onboarding": [
                "What is covered in this knowledge base?",
            ],
            "few_shot": {
                "general": (
                    'Sample question: "What is in the knowledge base?"\n'
                    "Sample answer: See the uploaded policy documents for details."
                ),
            },
            "eval": {"suite": domain_id},
        }

    pack_dir.mkdir(parents=True)
    (pack_dir / "data").mkdir()
    sample_doc = pack_dir / "data" / "sample_policy.txt"
    sample_doc.write_text(
        f"Sample policy for pack {pack_name} (demo).\n"
        "Replace this file with your organization's documents.\n",
        encoding="utf-8",
    )

    eval_lines = [
        {
            "domain_id": domain_id,
            "question": "What is covered in this knowledge base?",
            "expect_contains": ["policy"],
            "expect_context": True,
            "category": "policy",
        },
        {
            "domain_id": domain_id,
            "question": "What is the CEO salary on Jupiter in 3000?",
            "expect_out_of_scope": True,
            "expect_contains": [],
            "category": "edge",
        },
    ]
    with (pack_dir / "eval.jsonl").open("w", encoding="utf-8") as f:
        for row in eval_lines:
            f.write(json.dumps(row, ensure_ascii=False) + "\n")

    with (pack_dir / "pack.yaml").open("w", encoding="utf-8") as f:
        yaml.safe_dump(manifest, f, sort_keys=False, allow_unicode=True)

    if from_pack and (src_dir / "data").is_dir():
        for item in (src_dir / "data").iterdir():
            if item.is_file() and item.name != sample_doc.name:
                shutil.copy2(item, pack_dir / "data" / item.name)

    return pack_dir
