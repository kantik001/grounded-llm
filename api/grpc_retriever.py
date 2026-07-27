"""gRPC Retriever service — wraps rag.retrieval.retrieve_rag_context."""

from __future__ import annotations

import logging
import os
from concurrent import futures

import grpc
from grpc_health.v1 import health, health_pb2, health_pb2_grpc
from rag.retrieval import retrieve_rag_context

logger = logging.getLogger(__name__)

# Generated at image build / local: python -m grpc_tools.protoc ...
try:
    from api import retriever_pb2, retriever_pb2_grpc
except ImportError:  # pragma: no cover - generated next to package
    import retriever_pb2  # type: ignore
    import retriever_pb2_grpc  # type: ignore


class RetrieverServicer(retriever_pb2_grpc.RetrieverServicer):
    def Retrieve(self, request, context):  # noqa: N802
        token = ""
        for key, value in context.invocation_metadata():
            if key.lower() in ("x-rag-service-token", "authorization"):
                token = value
                if key.lower() == "authorization" and value.lower().startswith("bearer "):
                    token = value[7:].strip()
        expected = (os.environ.get("RAG_SERVICE_TOKEN") or "").strip()
        if expected and token != expected:
            context.abort(grpc.StatusCode.UNAUTHENTICATED, "invalid RAG service token")

        domain = (request.domain_id or "default").strip() or "default"
        tenant = (request.tenant_id or "default").strip() or "default"
        locale = (request.locale or "en").strip() or "en"
        top_k = int(request.top_k) if request.top_k and request.top_k > 0 else None

        # retrieve_rag_context uses domain rag_k; top_k override via env for grpc callers
        if top_k:
            os.environ["RAG_GRPC_TOP_K"] = str(top_k)
        try:
            out = retrieve_rag_context(request.query, domain, tenant, locale)
        finally:
            os.environ.pop("RAG_GRPC_TOP_K", None)

        resp = retriever_pb2.RetrieveResponse(
            success=bool(out.get("success")),
            error=str(out.get("error") or ""),
            context=str(out.get("context") or ""),
        )
        for frag in out.get("fragments") or []:
            resp.chunks.append(
                retriever_pb2.Chunk(
                    id=str(frag.get("filename") or ""),
                    text=str(frag.get("content") or frag.get("excerpt") or ""),
                    score=0.0,
                    source=str(frag.get("filename") or ""),
                )
            )
        return resp


def serve(port: int = 50051) -> grpc.Server:
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=4))
    retriever_pb2_grpc.add_RetrieverServicer_to_server(RetrieverServicer(), server)
    health_servicer = health.HealthServicer()
    health_pb2_grpc.add_HealthServicer_to_server(health_servicer, server)
    health_servicer.set("", health_pb2.HealthCheckResponse.SERVING)
    health_servicer.set("grounded.rag.v1.Retriever", health_pb2.HealthCheckResponse.SERVING)
    bind = f"0.0.0.0:{port}"
    server.add_insecure_port(bind)
    server.start()
    logger.info("gRPC Retriever listening on %s", bind)
    return server
