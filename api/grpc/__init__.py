"""gRPC surface for the Python RAG service.

Run::

    python -m api.grpc
"""

from api.grpc.retriever import RetrieverServicer, main, serve

__all__ = ["RetrieverServicer", "main", "serve"]
