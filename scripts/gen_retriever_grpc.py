#!/usr/bin/env python3
"""Generate api/gen/retriever_pb2*.py from api/proto/retriever.proto.

Usage (repo root)::

    pip install grpcio-tools==1.69.0
    python scripts/gen_retriever_grpc.py
    python scripts/gen_retriever_grpc.py --check
"""

from __future__ import annotations

import argparse
import re
import subprocess
import sys
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
PROTO = ROOT / "api" / "proto" / "retriever.proto"
OUT_DIR = ROOT / "api" / "gen"
PB2 = OUT_DIR / "retriever_pb2.py"
PB2_GRPC = OUT_DIR / "retriever_pb2_grpc.py"

GRPC_TOOLS_SPEC = "grpcio-tools==1.69.0"

_IMPORT_RE = re.compile(
    r"^import retriever_pb2 as (\w+)$",
    re.MULTILINE,
)
_IMPORT_FIXED = r"from api.gen import retriever_pb2 as \1"


def _fix_grpc_imports(text: str) -> str:
    """Import generated messages as ``api.gen.retriever_pb2``."""
    if "from api.gen import retriever_pb2" in text:
        return text
    # Drop legacy rewrite if regenerating over old files.
    text = text.replace("from api import retriever_pb2 as", "from api.gen import retriever_pb2 as")
    if "from api.gen import retriever_pb2" in text:
        return text
    fixed, n = _IMPORT_RE.subn(_IMPORT_FIXED, text, count=1)
    if n != 1:
        fixed2, n2 = re.subn(
            r"^import retriever_pb2\s*$",
            "from api.gen import retriever_pb2 as retriever_pb2",
            text,
            count=1,
            flags=re.MULTILINE,
        )
        if n2 != 1:
            raise RuntimeError(
                "could not rewrite retriever_pb2 import in generated *_grpc.py"
            )
        return fixed2
    return fixed


def _run_protoc(out_dir: Path) -> None:
    out_dir.mkdir(parents=True, exist_ok=True)
    proc = subprocess.run(
        [
            sys.executable,
            "-m",
            "grpc_tools.protoc",
            f"-I{PROTO.parent}",
            f"--python_out={out_dir}",
            f"--grpc_python_out={out_dir}",
            str(PROTO),
        ],
        cwd=ROOT,
        capture_output=True,
        text=True,
    )
    if proc.returncode != 0:
        sys.stderr.write(proc.stdout)
        sys.stderr.write(proc.stderr)
        raise SystemExit(f"protoc failed with code {proc.returncode}")


def generate(out_dir: Path = OUT_DIR) -> tuple[str, str]:
    try:
        import grpc_tools  # noqa: F401
    except ImportError as exc:
        raise SystemExit(
            f"grpcio-tools is required ({GRPC_TOOLS_SPEC}). "
            f"Install: pip install {GRPC_TOOLS_SPEC}"
        ) from exc

    _run_protoc(out_dir)
    pb2_path = out_dir / "retriever_pb2.py"
    grpc_path = out_dir / "retriever_pb2_grpc.py"
    pb2_text = pb2_path.read_text(encoding="utf-8")
    grpc_text = _fix_grpc_imports(grpc_path.read_text(encoding="utf-8"))
    grpc_path.write_text(grpc_text, encoding="utf-8", newline="\n")
    pb2_path.write_text(pb2_text, encoding="utf-8", newline="\n")
    return pb2_text, grpc_text


def check_fresh() -> int:
    with tempfile.TemporaryDirectory(prefix="retriever-grpc-") as tmp:
        tmp_path = Path(tmp)
        generate(tmp_path)
        want_pb2 = (tmp_path / "retriever_pb2.py").read_text(encoding="utf-8")
        want_grpc = (tmp_path / "retriever_pb2_grpc.py").read_text(encoding="utf-8")

    have_pb2 = PB2.read_text(encoding="utf-8") if PB2.is_file() else ""
    have_grpc = PB2_GRPC.read_text(encoding="utf-8") if PB2_GRPC.is_file() else ""

    def norm(s: str) -> str:
        return "\n".join(line.rstrip() for line in s.replace("\r\n", "\n").split("\n")).strip() + "\n"

    ok = norm(have_pb2) == norm(want_pb2) and norm(have_grpc) == norm(want_grpc)
    if ok:
        print("retriever gRPC stubs are up to date")
        return 0

    print(
        "retriever gRPC stubs are stale or import rewrite missing.\n"
        "Run: python scripts/gen_retriever_grpc.py",
        file=sys.stderr,
    )
    return 1


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--check",
        action="store_true",
        help="exit 1 if committed stubs do not match proto (+ import rewrite)",
    )
    args = parser.parse_args()
    if args.check:
        return check_fresh()
    generate(OUT_DIR)
    print(f"wrote {PB2.relative_to(ROOT)}")
    print(f"wrote {PB2_GRPC.relative_to(ROOT)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
