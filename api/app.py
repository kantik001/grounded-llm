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
        if not question:
            return jsonify({"success": False, "error": "Пустой вопрос"}), 400

        payload = retrieve_rag_context(question, domain_id=domain_id)
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


@app.route("/admin/reindex", methods=["POST"])
def admin_reindex():
    expected = os.environ.get("ADMIN_SECRET", "")
    secret = request.headers.get("X-Admin-Secret", "")
    if not expected or secret != expected:
        return jsonify({"success": False, "error": "forbidden"}), 403
    try:
        vs.reset_vector_store()
        store = vs.load_vector_store(force_reindex=True)
        if store is None:
            return jsonify({"success": False, "error": "Нет документов для индексации"}), 400
        return jsonify({"success": True, "message": "RAG переиндексирован"}), 200
    except Exception as e:
        return jsonify({"success": False, "error": str(e)}), 500


if __name__ == "__main__":
    port = int(os.environ.get("PYTHON_SERVICE_PORT", 5000))
    app.run(host="0.0.0.0", port=port, debug=False)
