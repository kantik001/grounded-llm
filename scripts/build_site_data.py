#!/usr/bin/env python3
"""Build static assets for site/ (pack registry JSON)."""

from __future__ import annotations

import os
import sys

_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
_SCRIPTS = os.path.dirname(__file__)
if _SCRIPTS not in sys.path:
    sys.path.insert(0, _SCRIPTS)

from pack_registry import export_registry_json  # noqa: E402


def main() -> int:
    site_dir = os.path.join(_ROOT, "site")
    os.makedirs(site_dir, exist_ok=True)
    out_path = os.path.join(site_dir, "packs.json")
    with open(out_path, "w", encoding="utf-8") as f:
        f.write(export_registry_json())
        f.write("\n")
    print(f"Wrote {out_path}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
