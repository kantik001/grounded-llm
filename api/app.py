"""Python HTTP API: RAG retrieval (/rag/context) for the Go server."""

import os
import sys

_root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
sys.path.insert(0, _root)

from dotenv import load_dotenv
from flask import Flask, jsonify, request
from flask_cors import CORS

load_dotenv(os.path.join(_root, ".env"))

from rag import vector_store as vs
from rag.domains_config import list_domains, normalize_domain_id
from rag.retrieval import retrieve_rag_context

app = Flask(__name__)


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
    if not (os.environ.get("RAG_SERVICE_TOKEN") or "").strip():
        raise RuntimeError("RAG_SERVICE_TOKEN must be set when GROUNDED_ENV=production")


_require_production_secrets()


def _admin_authorized() -> bool:
    expected = os.environ.get("ADMIN_SECRET", "")
    secret = request.headers.get("X-Admin-Secret", "")
    return bool(expected) and secret == expected


def _rag_service_authorized() -> bool:
    """Internal auth for Go server → Python RAG calls.

    Open when RAG_SERVICE_TOKEN is unset (local/dev only).
    In production, startup requires the token and every call must present it.
    """
    expected = os.environ.get("RAG_SERVICE_TOKEN", "")
    if not expected:
        return True
    token = request.headers.get("X-RAG-Service-Token", "")
    return token == expected


@app.route("/rag/context", methods=["POST"])
def rag_context():
    if not _rag_service_authorized():
        return jsonify({"success": False, "error": "forbidden"}), 403
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
        resp = jsonify(payload)
        resp.headers.set("Content-Type", "application/json; charset=utf-8")
        return resp, 200
    except Exception as e:
        return jsonify({"success": False, "error": str(e)}), 500


@app.route("/domains", methods=["GET"])
def domains_list():
    return jsonify({"success": True, **list_domains()}), 200


@app.route("/health", methods=["GET"])
def health_check():
    return jsonify({"status": "healthy", "service": "grounded-llm-python"}), 200


@app.route("/ready", methods=["GET"])
def readiness_check():
    if not _rag_service_authorized():
        return jsonify({"status": "not_ready", "checks": {"auth": "forbidden"}}), 403
    checks = {"process": "ok"}
    chroma_dir = vs.PERSIST_DIR
    if os.path.isdir(chroma_dir):
        checks["chroma"] = "ok"
    else:
        checks["chroma"] = "pending"
    data_root = os.path.join(_root, "data")
    if os.path.isdir(data_root):
        checks["data"] = "ok"
    else:
        checks["data"] = "missing"
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
    try:
        vs.reset_vector_store()
        store = vs.load_vector_store(force_reindex=True)
        if store is None:
            return jsonify({"success": False, "error": "No documents to index"}), 400
        return jsonify({"success": True, "message": "RAG reindexed"}), 200
    except Exception as e:
        return jsonify({"success": False, "error": str(e)}), 500


if __name__ == "__main__":
    port = int(os.environ.get("PYTHON_SERVICE_PORT", 5000))
    app.run(host="0.0.0.0", port=port, debug=False)
