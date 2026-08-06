"""Optional Redis cache for embedding vectors.

Key: embedding:{md5(text)}:{model_name}
Metric counter: rag_embedding_cache_hit_total (exported via module-level int for /metrics scrape later).
"""

from __future__ import annotations

import hashlib
import json
import logging
import os
import threading
from typing import Sequence

logger = logging.getLogger(__name__)

_cache_hits = 0
_cache_misses = 0
_lock = threading.Lock()
_client = None
_client_failed = False


def cache_stats() -> dict[str, int]:
    with _lock:
        return {"hits": _cache_hits, "misses": _cache_misses}


def _redis():
    global _client, _client_failed
    if _client_failed:
        return None
    if _client is not None:
        return _client
    url = (os.environ.get("REDIS_URL") or "").strip()
    if not url:
        _client_failed = True
        return None
    try:
        import redis  # type: ignore

        _client = redis.Redis.from_url(url, decode_responses=True, socket_connect_timeout=1.0)
        _client.ping()
        return _client
    except Exception as exc:  # noqa: BLE001
        logger.warning("Redis embedding cache unavailable: %s", exc)
        _client_failed = True
        _client = None
        return None


def _key(text: str, model: str) -> str:
    digest = hashlib.md5(text.encode("utf-8")).hexdigest()
    return f"embedding:{digest}:{model}"


def get_embedding(text: str, model: str) -> list[float] | None:
    global _cache_hits, _cache_misses
    r = _redis()
    if r is None:
        return None
    try:
        raw = r.get(_key(text, model))
        if not raw:
            with _lock:
                _cache_misses += 1
            return None
        vec = json.loads(raw)
        if not isinstance(vec, list):
            return None
        with _lock:
            _cache_hits += 1
            hits = _cache_hits
        try:
            r.incr("rag_embedding_cache_hit_total")
        except Exception:  # noqa: BLE001
            pass
        logger.debug("embedding cache HIT (%s)", hits)
        return [float(x) for x in vec]
    except Exception as exc:  # noqa: BLE001
        logger.debug("embedding cache get failed: %s", exc)
        return None


def set_embedding(text: str, model: str, vector: Sequence[float], ttl_sec: int = 3600) -> None:
    r = _redis()
    if r is None:
        return
    try:
        r.setex(_key(text, model), ttl_sec, json.dumps(list(vector)))
    except Exception as exc:  # noqa: BLE001
        logger.debug("embedding cache set failed: %s", exc)


def e5_prefixes_enabled(model_name: str) -> bool:
    """E5-family models are trained with "query:"/"passage:" prefixes; skipping
    them measurably hurts dense retrieval. Auto-on for e5 models, overridable
    via RAG_E5_PREFIXES=0/1."""
    raw = (os.environ.get("RAG_E5_PREFIXES") or "").strip().lower()
    if raw in ("1", "true", "yes", "on"):
        return True
    if raw in ("0", "false", "no", "off"):
        return False
    return "e5" in model_name.lower()


class CachedHuggingFaceEmbeddings:
    """LangChain-compatible wrapper around HuggingFaceEmbeddings with Redis cache on embed_query."""

    def __init__(self, model_name: str):
        from langchain_huggingface import HuggingFaceEmbeddings

        self._model_name = model_name
        self._inner = HuggingFaceEmbeddings(model_name=model_name)
        self._ttl = int(os.environ.get("EMBEDDING_CACHE_TTL_SEC", "3600") or "3600")
        self._e5 = e5_prefixes_enabled(model_name)

    @property
    def uses_e5_prefixes(self) -> bool:
        return self._e5

    def _query_text(self, text: str) -> str:
        return f"query: {text}" if self._e5 else text

    def _passage_text(self, text: str) -> str:
        return f"passage: {text}" if self._e5 else text

    def embed_query(self, text: str) -> list[float]:
        text = self._query_text(text)
        cached = get_embedding(text, self._model_name)
        if cached is not None:
            return cached
        vec = self._inner.embed_query(text)
        set_embedding(text, self._model_name, vec, self._ttl)
        return vec

    def embed_documents(self, texts: list[str]) -> list[list[float]]:
        texts = [self._passage_text(t) for t in texts]
        out: list[list[float]] = []
        missing_idx: list[int] = []
        missing_texts: list[str] = []
        placeholders: list[list[float] | None] = [None] * len(texts)
        for i, t in enumerate(texts):
            cached = get_embedding(t, self._model_name)
            if cached is not None:
                placeholders[i] = cached
            else:
                missing_idx.append(i)
                missing_texts.append(t)
        if missing_texts:
            computed = self._inner.embed_documents(missing_texts)
            for i, vec in zip(missing_idx, computed):
                placeholders[i] = vec
                set_embedding(texts[i], self._model_name, vec, self._ttl)
        for p in placeholders:
            assert p is not None
            out.append(p)
        return out
