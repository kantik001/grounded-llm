"""Compatibility shim — prefer ``python -m api.grpc`` / ``api.grpc.retriever``."""

from __future__ import annotations

import os
import sys

_root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
if _root not in sys.path:
    sys.path.insert(0, _root)

from api.grpc.retriever import RetrieverServicer, main, serve

__all__ = ["RetrieverServicer", "main", "serve"]

if __name__ == "__main__":
    main()
