"""Python RAG service package (HTTP + gRPC).

Folder name ``api`` is historical: this is **not** the public product HTTP API
(that lives in Go ``server/``). This package is the internal retrieval service
used by Go and agents.

Layout::

    api/
      http/       Flask app (Gunicorn: api.http.app:app)
      grpc/       Retriever service (python -m api.grpc)
      gen/        Generated protobuf/gRPC stubs (do not edit)
      proto/      Retriever .proto (source of truth for gen/)
      auth.py     Shared secrets (HTTP + gRPC)
      retrieve_metrics.py
      entrypoint.sh
"""

from __future__ import annotations

__all__: list[str] = []
