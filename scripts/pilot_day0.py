#!/usr/bin/env python3
"""Day-0 pilot bootstrap: install pack, print next steps, seed golden template.

Usage:
  python scripts/pilot_day0.py --pack hr --tenant pilot-acme
  python scripts/pilot_day0.py --pack it_support --tenant pilot-acme --skip-install
"""

from __future__ import annotations

import argparse
import os
import shutil
import sys

_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
_SCRIPTS = os.path.dirname(__file__)
for _p in (_ROOT, _SCRIPTS):
    if _p not in sys.path:
        sys.path.insert(0, _p)


def main() -> int:
    parser = argparse.ArgumentParser(description="Grounded LLM pilot day-0 bootstrap")
    parser.add_argument("--pack", default="hr", help="Pack id: hr | it_support | legal_faq | ...")
    parser.add_argument("--tenant", required=True, help="Tenant id for data/{tenant}/{domain}/")
    parser.add_argument("--force", action="store_true", help="Reinstall pack over existing domain")
    parser.add_argument("--skip-install", action="store_true", help="Only print paths / copy golden template")
    parser.add_argument(
        "--golden-out",
        default="",
        help="Path for pilot golden JSONL (default: eval/rag_pilot_<tenant>.jsonl)",
    )
    args = parser.parse_args()

    tenant = args.tenant.strip()
    if not tenant or "/" in tenant or "\\" in tenant:
        print("Invalid --tenant", file=sys.stderr)
        return 2

    golden_out = args.golden_out or os.path.join(_ROOT, "eval", f"rag_pilot_{tenant}.jsonl")
    template = os.path.join(_ROOT, "eval", "pilot_golden_template.jsonl")

    result: dict = {}
    if not args.skip_install:
        from pack_installer import install_pack

        result = install_pack(args.pack, tenant_id=tenant, force=args.force, dry_run=False)
        print(f"Installed pack '{args.pack}' for tenant '{tenant}'")
        print(f"  domain_id:  {result.get('domain_id')}")
        print(f"  data_dir:   {result.get('data_dir')}")
        print(f"  eval_suite: {result.get('eval_suite')}")
    else:
        print(f"Skip install · pack={args.pack} tenant={tenant}")

    if os.path.isfile(template):
        os.makedirs(os.path.dirname(golden_out) or ".", exist_ok=True)
        if not os.path.exists(golden_out) or args.force:
            shutil.copyfile(template, golden_out)
            print(f"Golden template → {golden_out}")
        else:
            print(f"Golden exists (use --force to overwrite): {golden_out}")
    else:
        print(f"WARNING: missing {template}", file=sys.stderr)

    print()
    print("Next:")
    print("  1. Copy customer documents into data_dir (txt/pdf/docx)")
    print("  2. Edit golden JSONL with real expect_contains facts")
    print("  3. python scripts/reindex_rag.py")
    if result.get("eval_suite"):
        print(f"  4. python scripts/run_rag_eval.py --suite {result['eval_suite']}")
    print("  5. Fill docs/en/PILOT_CHECKLIST.md before UAT")
    print("  6. Prod: VERIFY_FAITHFULNESS=enforce, METRICS_TOKEN, tenant-bound API keys")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
