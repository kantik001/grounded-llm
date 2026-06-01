"""Python HTTP API: RAG retrieval (/rag/context) for the Go server."""

import os
import sys

_root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
sys.path.insert(0, _root)

from dotenv import load_dotenv
from flask import Flask, jsonify, request
from flask_cors import CORS

load_dotenv(os.path.join(_root, ".env"))

from rag.domains_config import list_domains, normalize_domain_id
from rag.retrieval import retrieve_rag_context
from rag import vector_store as vs

app = Flask(__name__)
CORS(app)


@app.route("/rag/context", methods=["POST"])
def rag_context():
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


def _admin_authorized() -> bool:
    expected = os.environ.get("ADMIN_SECRET", "")
    secret = request.headers.get("X-Admin-Secret", "")
    return bool(expected) and secret == expected


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
