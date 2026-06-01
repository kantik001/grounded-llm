"""RAG retrieval: build context for Go orchestration by domain_id."""

import json
import os
from typing import Any, Dict, List

from rag.domains_config import get_domain, normalize_domain_id
from rag.vector_store import search

_PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
_few_shot_cache = None


def _load_few_shot() -> dict:
    global _few_shot_cache
    if _few_shot_cache is not None:
        return _few_shot_cache
    path = os.path.join(_PROJECT_ROOT, "config", "few_shot.json")
    with open(path, encoding="utf-8") as f:
        _few_shot_cache = json.load(f)
    return _few_shot_cache


def _rag_k_for_domain(domain: dict) -> int:
    raw = domain.get("rag_k", 8)
    try:
        k = int(raw)
    except (TypeError, ValueError):
        k = 8
    return max(1, min(k, 20))


def few_shot_for(domain_id: str, category: str = "general") -> str:
    domain_shots = _load_few_shot().get(domain_id, {})
    return domain_shots.get(category) or domain_shots.get("general", "")


def _excerpt(text: str, max_len: int = 280) -> str:
    s = (text or "").strip()
    if len(s) <= max_len:
        return s
    return s[:max_len] + "…"


def retrieve_rag_context(
    user_question: str, domain_id: str = "default"
) -> Dict[str, Any]:
    q = (user_question or "").strip()
    empty: Dict[str, Any] = {
        "success": False,
        "error": "",
        "context": "",
        "few_shot": "",
        "category": "general",
        "fragments": [],
        "domain_id": domain_id,
    }
    if not q:
        empty["error"] = "Пустой вопрос"
        return empty

    try:
        domain_id = normalize_domain_id(domain_id)
    except ValueError as e:
        empty["error"] = str(e)
        return empty

    domain = get_domain(domain_id)
    if not domain.get("rag_enabled", True):
        name = domain.get("name") or domain.get("name_ru") or domain_id
        empty["error"] = (
            f"База документов для «{name}» пока не подключена. "
            "Выберите другой домен или вернитесь позже."
        )
        return empty

    k = _rag_k_for_domain(domain)
    fragments = search(q, domain_id=domain_id, k=k)
    if not fragments:
        name = domain.get("name") or domain.get("name_ru") or domain_id
        empty["error"] = f"Не нашёл информации в документах домена «{name}»."
        return empty

    for f in fragments:
        print(f"[RAG:{domain_id}] источник: {f.metadata.get('filename')}")

    context_parts: List[str] = []
    fr_json: List[Dict[str, str]] = []
    for frag in fragments:
        source_name = frag.metadata.get("filename", "Неизвестный источник")
        page = frag.metadata.get("page")
        page_label = f", стр. {int(page) + 1}" if page is not None else ""
        context_parts.append(f"Фрагмент '{source_name}'{page_label}:\n{frag.page_content}")
        entry: Dict[str, Any] = {
            "filename": source_name,
            "content": frag.page_content,
            "excerpt": _excerpt(frag.page_content),
        }
        if page is not None:
            try:
                entry["page"] = int(page) + 1
            except (TypeError, ValueError):
                pass
        fr_json.append(entry)

    context = "\n\n---\n\n".join(context_parts)
    few_shot = few_shot_for(domain_id, "general")

    return {
        "success": True,
        "error": "",
        "context": context,
        "few_shot": few_shot,
        "category": "general",
        "fragments": fr_json,
        "domain_id": domain_id,
    }
