"""Unit tests for RAG answer verifier (no LLM or Chroma)."""

from langchain_core.documents import Document

from rag.verifier import RAG_ANSWER_DISCLAIMER, extract_numbers, strip_source_attribution, verify_answer


def test_extract_numbers_decimal_comma():
    assert extract_numbers("304,7 кг") == [304.7]


def test_verify_passes_with_matching_number():
    fragments = [Document(page_content="Среднее 77.", metadata={"filename": "Таблица"})]
    answer = f"Среднее 77.\n\n{RAG_ANSWER_DISCLAIMER}"
    ok, reason = verify_answer("q", answer, fragments)
    assert ok, reason


def test_verify_fails_on_hallucinated_number():
    fragments = [Document(page_content="Без цифр.", metadata={"filename": "Статья"})]
    answer = f"Рентабельность 72%.\n\n{RAG_ANSWER_DISCLAIMER}"
    ok, reason = verify_answer("q", answer, fragments)
    assert not ok
    assert "72" in reason or "не найдены" in reason


def test_strip_source_attribution():
    raw = 'Факт.\n\nИсточник: "Журнал"'
    body = strip_source_attribution(raw)
    assert "Источник" not in body
    assert "Журнал" not in body
    assert "Факт" in body
