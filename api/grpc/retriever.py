"""gRPC Retriever service — wraps rag.retrieval.retrieve_rag_context.

Run as a module (Docker entrypoint / local)::

    python -m api.grpc

Auth: metadata ``x-rag-service-token`` or ``authorization: Bearer <token>``.
Business failures stay in the response body (``success=false``).
Unexpected exceptions use gRPC ``INTERNAL``.
TLS is not enabled here — rely on private network / service mesh.
"""

from __future__ import annotations

import logging
import os
import signal
import threading
import time
from concurrent import futures

from grpc_health.v1 import health, health_pb2, health_pb2_grpc
from rag.retrieval import retrieve_rag_context

import grpc
from api.auth import rag_service_token_ok, rag_token_from_metadata
from api.gen import retriever_pb2, retriever_pb2_grpc
from api.retrieve_metrics import record_retrieve

logger = logging.getLogger(__name__)


def _metadata_value(metadata: tuple, key: str) -> str:
    want = key.lower()
    for k, v in metadata:
        if k.lower() == want:
            return (v or "").strip()
    return ""


def _worker_count() -> int:
    raw = (os.environ.get("GRPC_MAX_WORKERS") or "4").strip()
    try:
        return max(1, min(int(raw), 64))
    except ValueError:
        return 4


class RetrieverServicer(retriever_pb2_grpc.RetrieverServicer):
    def Retrieve(self, request, context):  # noqa: N802
        meta = tuple(context.invocation_metadata() or ())
        token = rag_token_from_metadata(meta)
        if not rag_service_token_ok(token):
            context.abort(grpc.StatusCode.UNAUTHENTICATED, "invalid RAG service token")

        req_id = _metadata_value(meta, "x-request-id")
        domain = (request.domain_id or "default").strip() or "default"
        tenant = (request.tenant_id or "default").strip() or "default"
        locale = (request.locale or "en").strip() or "en"
        top_k = int(request.top_k) if request.top_k and request.top_k > 0 else None

        t0 = time.perf_counter()
        try:
            out = retrieve_rag_context(
                request.query,
                domain,
                tenant,
                locale,
                top_k=top_k,
            )
        except Exception:
            elapsed = time.perf_counter() - t0
            record_retrieve(elapsed, protocol="grpc", ok=False, business_failure=False)
            logger.exception(
                "req=%s step=grpc_retrieve ms=%s ok=False",
                req_id or "-",
                int(elapsed * 1000),
            )
            context.abort(grpc.StatusCode.INTERNAL, "internal error")

        elapsed = time.perf_counter() - t0
        ok = bool(out.get("success")) if isinstance(out, dict) else False
        record_retrieve(
            elapsed,
            protocol="grpc",
            ok=ok,
            business_failure=not ok,
        )
        frags = out.get("fragments") if isinstance(out, dict) else None
        n = len(frags) if isinstance(frags, list) else 0
        logger.info(
            "req=%s step=grpc_retrieve ms=%s fragments=%s ok=%s domain_id=%s tenant_id=%s",
            req_id or "-",
            int(elapsed * 1000),
            n,
            ok,
            domain,
            tenant,
        )

        resp = retriever_pb2.RetrieveResponse(
            success=ok,
            error=str(out.get("error") or ""),
            context=str(out.get("context") or ""),
        )
        for frag in frags or []:
            try:
                score = float(frag.get("score") or 0.0)
            except (TypeError, ValueError):
                score = 0.0
            resp.chunks.append(
                retriever_pb2.Chunk(
                    id=str(frag.get("filename") or ""),
                    text=str(frag.get("content") or frag.get("excerpt") or ""),
                    score=score,
                    source=str(frag.get("filename") or ""),
                )
            )
        return resp


def serve(port: int = 50051) -> grpc.Server:
    workers = _worker_count()
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=workers))
    retriever_pb2_grpc.add_RetrieverServicer_to_server(RetrieverServicer(), server)
    health_servicer = health.HealthServicer()
    health_pb2_grpc.add_HealthServicer_to_server(health_servicer, server)
    health_servicer.set("", health_pb2.HealthCheckResponse.SERVING)
    health_servicer.set("grounded.rag.v1.Retriever", health_pb2.HealthCheckResponse.SERVING)
    bind = f"0.0.0.0:{port}"
    # Insecure by design for private Docker/K8s networks; terminate TLS at the mesh/ingress.
    server.add_insecure_port(bind)
    server.start()
    logger.info("gRPC Retriever listening on %s (workers=%s)", bind, workers)
    return server


def main() -> None:
    logging.basicConfig(
        level=logging.INFO,
        format="%(levelname)s %(name)s: %(message)s",
    )
    port = int(os.environ.get("PYTHON_GRPC_PORT", "50051"))
    server = serve(port)
    stopped = threading.Event()

    def _on_signal(signum: int, _frame: object | None) -> None:
        logger.info("received signal %s, shutting down gRPC", signum)
        stopped.set()

    signal.signal(signal.SIGTERM, _on_signal)
    signal.signal(signal.SIGINT, _on_signal)

    stopped.wait()
    server.stop(grace=5)
    logger.info("gRPC Retriever stopped")


if __name__ == "__main__":
    main()
