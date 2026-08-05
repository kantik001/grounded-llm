"""Compatibility shim — prefer ``python -m api.grpc`` / ``api.grpc.retriever``."""

from api.grpc.retriever import RetrieverServicer, main, serve

__all__ = ["RetrieverServicer", "main", "serve"]

if __name__ == "__main__":
    main()
