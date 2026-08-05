"""Unit tests for retrieve_rag_context top_k and fragment scores."""

from __future__ import annotations

import os
import sys

_root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
sys.path.insert(0, _root)

from rag.retrieval import retrieve_rag_context  # noqa: E402


class _Doc:
    def __init__(self, text: str, filename: str = "a.txt"):
        self.page_content = text
        self.metadata = {"filename": filename}


def test_top_k_passed_to_search(monkeypatch):
    seen: dict = {}

    def fake_search(q, domain_id, tenant_id="default", k=8):
        seen["k"] = k
        return [_Doc("vacation days are 28", "policy.txt")]

    monkeypatch.setattr("rag.retrieval.search", fake_search)
    monkeypatch.setattr(
        "rag.retrieval.get_domain",
        lambda _id: {"rag_enabled": True, "rag_k": 8, "name": "default"},
    )
    monkeypatch.setattr("rag.retrieval.normalize_domain_id", lambda x: x)
    monkeypatch.setattr("rag.retrieval.few_shot_for", lambda *a, **k: "")

    out = retrieve_rag_context("vacation days", domain_id="default", top_k=3)
    assert out["success"] is True
    assert seen["k"] == 3
    assert out["fragments"][0]["score"] > 0.0


def test_grpc_servicer_maps_score_and_top_k(monkeypatch):
    from api.grpc.retriever import RetrieverServicer

    captured: dict = {}

    def fake_retrieve(query, domain, tenant, locale, top_k=None):
        captured["top_k"] = top_k
        return {
            "success": True,
            "error": "",
            "context": "ctx",
            "fragments": [{"filename": "f.txt", "content": "hello", "score": 0.75}],
        }

    monkeypatch.setattr("api.grpc.retriever.retrieve_rag_context", fake_retrieve)
    monkeypatch.setattr("api.grpc.retriever.rag_service_token_ok", lambda _t: True)
    monkeypatch.setattr("api.grpc.retriever.record_retrieve", lambda *a, **k: None)

    class _Req:
        query = "q"
        domain_id = "default"
        tenant_id = "default"
        locale = "en"
        top_k = 5

    class _Ctx:
        def invocation_metadata(self):
            return (("x-request-id", "rid-1"),)

        def abort(self, *a, **k):
            raise AssertionError("abort should not be called")

    resp = RetrieverServicer().Retrieve(_Req(), _Ctx())
    assert captured["top_k"] == 5
    assert resp.success is True
    assert resp.chunks[0].score == 0.75
