"""Python HTTP API: RAG retrieval (/rag/context) for the Go server."""

from __future__ import annotations

import os
import sys
import threading
import time

_root = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
if _root not in sys.path:
    sys.path.insert(0, _root)

from dotenv import load_dotenv
from flask import Flask, Response, jsonify, request
from flask_cors import CORS

load_dotenv(os.path.join(_root, ".env"))

from api.auth import (
    admin_secret_ok,
    rag_service_token_ok,
    resolve_admin_secret,
    resolve_rag_token,
    validate_production_secrets,
)
from api.retrieve_metrics import record_retrieve, retrieve_metrics_lines
from rag import vector_store as vs
from rag.domains_config import list_domains, normalize_domain_id
from rag.retrieval import retrieve_rag_context

app = Flask(__name__)
_reindex_lock = threading.Lock()


def _is_production() -> bool:
    for key in ("GROUNDED_ENV", "APP_ENV", "ENV"):
        val = (os.environ.get(key) or "").strip().lower()
        if val in ("production", "prod"):
            return True
    return False


def _configure_cors() -> None:
    """CORS is optional; Python RAG is an internal service (prefer network isolation)."""
    raw = (os.environ.get("PYTHON_CORS_ORIGINS") or "").strip()
    if not raw:
        return
    if raw == "*":
        if _is_production():
            raise RuntimeError("PYTHON_CORS_ORIGINS=* is not allowed in production")
        CORS(app)
        return
    origins = [o.strip() for o in raw.split(",") if o.strip()]
    if origins:
        CORS(app, resources={r"/*": {"origins": origins}})


_configure_cors()


def _require_production_secrets() -> None:
    if not _is_production():
        return
    validate_production_secrets()


_require_production_secrets()


def _admin_authorized() -> bool:
    return admin_secret_ok(
        resolve_admin_secret(
            header_secret=request.headers.get("X-Admin-Secret"),
            authorization=request.headers.get("Authorization"),
        )
    )


def _rag_service_authorized() -> bool:
    """Internal auth for Go server → Python RAG calls."""
    return rag_service_token_ok(
        resolve_rag_token(
            header_token=request.headers.get("X-RAG-Service-Token"),
            authorization=request.headers.get("Authorization"),
        )
    )


def _client_error_message(exc: BaseException) -> str:
    """Never leak internal exception text to API clients in production."""
    if _is_production():
        return "internal error"
    return str(exc) or type(exc).__name__


@app.route("/rag/context", methods=["POST"])
def rag_context():
    if not _rag_service_authorized():
        return jsonify({"success": False, "error": "forbidden"}), 403
    req_id = (request.headers.get("X-Request-ID") or "").strip()
    t0 = time.perf_counter()
    try:
        data = request.get_json(silent=True) or {}
        question = (data.get("question") or "").strip()
        domain_id = (data.get("domain_id") or "default").strip()
        tenant_id = (data.get("tenant_id") or os.environ.get("DEFAULT_TENANT_ID", "default")).strip()
        locale = (data.get("locale") or os.environ.get("DEFAULT_LOCALE", "en")).strip()
        if not question:
            return jsonify({"success": False, "error": "Empty question"}), 400

        payload = retrieve_rag_context(
            question, domain_id=domain_id, tenant_id=tenant_id, locale=locale
        )
        elapsed = time.perf_counter() - t0
        frags = payload.get("fragments") if isinstance(payload, dict) else None
        n = len(frags) if isinstance(frags, list) else 0
        ok = bool(payload.get("success")) if isinstance(payload, dict) else False
        record_retrieve(
            elapsed,
            protocol="http",
            ok=ok,
            business_failure=not ok,
        )
        app.logger.info(
            "req=%s step=retrieve ms=%s fragments=%s ok=%s domain_id=%s tenant_id=%s",
            req_id or "-",
            int(elapsed * 1000),
            n,
            ok,
            domain_id,
            tenant_id,
        )
        resp = jsonify(payload)
        resp.headers.set("Content-Type", "application/json; charset=utf-8")
        if req_id:
            resp.headers.set("X-Request-ID", req_id)
        return resp, 200
    except Exception as e:
        elapsed = time.perf_counter() - t0
        record_retrieve(elapsed, protocol="http", ok=False, business_failure=False)
        app.logger.exception(
            "req=%s step=retrieve ms=%s ok=False error=%s",
            req_id or "-",
            int(elapsed * 1000),
            type(e).__name__,
        )
        return jsonify({"success": False, "error": _client_error_message(e)}), 500


@app.route("/domains", methods=["GET"])
def domains_list():
    if not _rag_service_authorized():
        return jsonify({"success": False, "error": "forbidden"}), 403
    return jsonify({"success": True, **list_domains()}), 200


@app.route("/health", methods=["GET"])
def health_check():
    return jsonify({"status": "healthy", "service": "grounded-llm-python"}), 200


@app.route("/metrics", methods=["GET"])
def python_metrics():
    """Prometheus text: retrieve latency + embedding cache counters."""
    from rag.embedding_cache import cache_stats

    stats = cache_stats()
    lines = [
        *retrieve_metrics_lines(),
        "# HELP rag_embedding_cache_hit_total Embedding cache hits (process)",
        "# TYPE rag_embedding_cache_hit_total counter",
        f"rag_embedding_cache_hit_total {stats.get('hits', 0)}",
        "# HELP rag_embedding_cache_miss_total Embedding cache misses (process)",
        "# TYPE rag_embedding_cache_miss_total counter",
        f"rag_embedding_cache_miss_total {stats.get('misses', 0)}",
    ]
    return Response("\n".join(lines) + "\n", mimetype="text/plain; version=0.0.4; charset=utf-8")


@app.route("/ready", methods=["GET"])
def readiness_check():
    if not _rag_service_authorized():
        return jsonify({"status": "not_ready", "checks": {"auth": "forbidden"}}), 403
    checks: dict[str, str] = {"process": "ok"}
    data_root = os.path.join(_root, "data")
    if os.path.isdir(data_root):
        checks["data"] = "ok"
    else:
        checks["data"] = "missing"
        return jsonify({"status": "not_ready", "checks": checks}), 503

    chroma_dir = vs.PERSIST_DIR
    if os.path.isdir(chroma_dir):
        checks["chroma"] = "ok"
    else:
        checks["chroma"] = "pending"

    index_label, index_ok = vs.readiness_index_check()
    checks["index"] = index_label
    if not index_ok:
        return jsonify({"status": "not_ready", "checks": checks}), 503

    return jsonify({"status": "ready", "checks": checks}), 200


@app.route("/admin/index-stats", methods=["GET"])
def admin_index_stats():
    if not _admin_authorized():
        return jsonify({"success": False, "error": "forbidden"}), 403
    domain_id = (request.args.get("domain_id") or "default").strip()
    tenant_id = (request.args.get("tenant_id") or os.environ.get("DEFAULT_TENANT_ID", "default")).strip()
    try:
        normalize_domain_id(domain_id)
    except ValueError as e:
        return jsonify({"success": False, "error": str(e)}), 400
    files = vs.index_stats_for_domain(domain_id, tenant_id=tenant_id)
    return jsonify({"success": True, "domain_id": domain_id, "tenant_id": tenant_id, "files": files}), 200


@app.route("/admin/reindex", methods=["POST"])
def admin_reindex():
    if not _admin_authorized():
        return jsonify({"success": False, "error": "forbidden"}), 403
    if not _reindex_lock.acquire(blocking=False):
        return jsonify({"success": False, "error": "reindex already in progress"}), 409
    try:
        vs.reset_vector_store()
        store = vs.load_vector_store(force_reindex=True)
        if store is None:
            return jsonify({"success": False, "error": "No documents to index"}), 400
        return jsonify({"success": True, "message": "RAG reindexed"}), 200
    except Exception as e:
        app.logger.exception("admin reindex failed")
        return jsonify({"success": False, "error": _client_error_message(e)}), 500
    finally:
        _reindex_lock.release()


if __name__ == "__main__":
    port = int(os.environ.get("PYTHON_SERVICE_PORT", 5000))
    app.run(host="0.0.0.0", port=port, debug=False)
