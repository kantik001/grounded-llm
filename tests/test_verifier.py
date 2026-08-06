from langchain_core.documents import Document
from rag.verifier import (
    RAG_ANSWER_DISCLAIMER,
    extract_numbers,
    strip_source_attribution,
    unsupported_sentences,
    verify_answer,
)


def test_extract_numbers_decimal_comma():
    assert extract_numbers("304.7 kg") == [304.7]


def test_verify_pass_with_number_in_context():
    fragments = [Document(page_content="Average 77.", metadata={"filename": "Table"})]
    answer = f"Average 77.\n\n{RAG_ANSWER_DISCLAIMER}"
    ok, _ = verify_answer("", answer, fragments)
    assert ok


def test_verify_fail_hallucinated_number():
    fragments = [Document(page_content="No digits.", metadata={"filename": "Article"})]
    answer = f"Margin 72%.\n\n{RAG_ANSWER_DISCLAIMER}"
    ok, reason = verify_answer("", answer, fragments)
    assert not ok
    assert "72" in reason or "not found" in reason


def test_strip_source_line():
    raw = 'Fact.\n\nSource: "Journal"'
    body = strip_source_attribution(raw)
    assert "Source" not in body
    assert "Journal" not in body
    assert "Fact" in body


_FAITH_CTX = (
    "Возврат товара возможен в течение 14 дней с момента покупки. "
    "Для возврата необходим чек и заявление."
)


def test_unsupported_sentences_passes_grounded():
    body = "Возврат товара возможен в течение 14 дней. Для возврата необходим чек и заявление."
    assert unsupported_sentences(body, _FAITH_CTX) == []


def test_unsupported_sentences_flags_fabrication():
    body = (
        "Возврат товара возможен в течение 14 дней. "
        "Компенсация выплачивается биткоинами через мобильное приложение банка."
    )
    flagged = unsupported_sentences(body, _FAITH_CTX)
    assert len(flagged) == 1
    assert "биткоинами" in flagged[0]


def test_unsupported_sentences_tolerates_inflection():
    body = "Для возврата товара потребуется заявление вместе с чеком."
    assert unsupported_sentences(body, _FAITH_CTX) == []


def test_unsupported_sentences_skips_short():
    assert unsupported_sentences("Да, конечно.", _FAITH_CTX) == []
