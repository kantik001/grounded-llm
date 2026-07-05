#!/usr/bin/env python3
"""CLI: list, install, and scaffold Grounded LLM template packs."""

from __future__ import annotations

import argparse
import os
import sys

_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
_SCRIPTS = os.path.dirname(__file__)
for _p in (_ROOT, _SCRIPTS):
    if _p not in sys.path:
        sys.path.insert(0, _p)

from pack_installer import (  # noqa: E402
    install_pack,
    list_packs,
    load_pack_manifest,
    scaffold_new_pack,
)


def cmd_list(_: argparse.Namespace) -> int:
    packs = list_packs()
    if not packs:
        print("No packs found under packs/")
        return 0
    for name in packs:
        _, manifest = load_pack_manifest(name)
        domain = (manifest.get("domain") or {}).get("id", "?")
        desc = manifest.get("description") or ""
        print(f"  {name:16} domain={domain:12} {desc}")
    return 0


def cmd_install(args: argparse.Namespace) -> int:
    result = install_pack(
        args.pack,
        tenant_id=args.tenant,
        force=args.force,
        dry_run=args.dry_run,
    )
    if args.dry_run:
        print("Dry run — would install:")
        for k, v in result.items():
            print(f"  {k}: {v}")
        return 0
    print(f"Installed pack '{args.pack}'")
    print(f"  domain_id:  {result['domain_id']}")
    print(f"  data_dir:   {result['data_dir']}")
    print(f"  eval_suite: {result['eval_suite']}")
    print(f"  eval_path:  {result.get('eval_path')}")
    print("Next:")
    print("  python scripts/reindex_rag.py")
    print(f"  python scripts/run_rag_eval.py --suite {result['eval_suite']}")
    return 0


def cmd_registry(args: argparse.Namespace) -> int:
    from pack_registry import export_registry_json, load_registry, validate_registry

    if args.validate:
        errors = validate_registry()
        if errors:
            for err in errors:
                print(f"  ERROR: {err}", file=sys.stderr)
            return 1
        print("Registry OK")
        return 0

    if args.json:
        print(export_registry_json())
        return 0

    registry = load_registry()
    for entry in registry.get("packs", []):
        print(f"  {entry.get('id'):16} domain={entry.get('domain_id'):12} {entry.get('guide', '')}")
    return 0


def cmd_new(args: argparse.Namespace) -> int:
    pack_dir = scaffold_new_pack(
        args.name,
        domain_id=args.domain_id,
        locale=args.locale,
        from_pack=args.from_pack,
    )
    print(f"Created pack scaffold: {pack_dir}")
    if args.install:
        return cmd_install(
            argparse.Namespace(
                pack=args.name,
                tenant=args.tenant,
                force=False,
                dry_run=False,
            )
        )
    print("Next:")
    print(f"  Edit {pack_dir}/pack.yaml and data/")
    print(f"  python scripts/init_pack.py install {args.name}")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Grounded LLM template pack installer (pack.yaml v1)",
    )
    sub = parser.add_subparsers(dest="command", required=True)

    p_list = sub.add_parser("list", help="List official packs")
    p_list.set_defaults(func=cmd_list)

    p_install = sub.add_parser("install", help="Install pack into config/data/eval")
    p_install.add_argument("pack", help="Pack folder name under packs/")
    p_install.add_argument("--tenant", default="default", help="Tenant id (default: default)")
    p_install.add_argument("--force", action="store_true", help="Overwrite locale/domain entries")
    p_install.add_argument("--dry-run", action="store_true", help="Show plan only")
    p_install.set_defaults(func=cmd_install)

    p_registry = sub.add_parser("registry", help="Show or validate packs/registry.yaml")
    p_registry.add_argument("--validate", action="store_true", help="Validate registry and pack files")
    p_registry.add_argument("--json", action="store_true", help="Print registry as JSON")
    p_registry.set_defaults(func=cmd_registry)

    p_new = sub.add_parser("new", help="Scaffold a new pack under packs/")
    p_new.add_argument("name", help="New pack folder name (slug)")
    p_new.add_argument("--domain-id", help="Domain id (default: same as pack name)")
    p_new.add_argument("--locale", default="en", choices=["en", "ru"])
    p_new.add_argument("--from", dest="from_pack", help="Clone manifest from existing pack")
    p_new.add_argument("--install", action="store_true", help="Run install after scaffold")
    p_new.add_argument("--tenant", default="default")
    p_new.set_defaults(func=cmd_new)

    args = parser.parse_args()
    try:
        return args.func(args)
    except (FileNotFoundError, FileExistsError, ValueError) as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
