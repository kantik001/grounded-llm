from langchain_core.documents import Document
from rag.verifier import RAG_ANSWER_DISCLAIMER, extract_numbers, strip_source_attribution, verify_answer


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
