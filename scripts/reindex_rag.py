#!/usr/bin/env python3
"""Reindex Chroma after adding documents under data/{domain_id}/."""

import os
import sys

_root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
sys.path.insert(0, _root)

os.environ["FORCE_RAG_REINDEX"] = "true"

from rag.vector_store import create_vector_store  # noqa: E402

if __name__ == "__main__":
    create_vector_store()
    print("RAG reindex completed.")
