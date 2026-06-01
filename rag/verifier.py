"""Answer verification helpers (mirror of server/rag_verify.go for tests)."""

import re

RAG_ANSWER_DISCLAIMER = (
    "Reference information from the knowledge base. "
    "Not a substitute for official expert advice."
)

_SOURCE_LINE_RE = re.compile(r"(?im)^\s*(Источник|Source):.*\n?")


def extract_numbers(text: str) -> list[float]:
    text = text.replace(",", ".")
    return [float(m) for m in re.findall(r"\b\d+(?:\.\d+)?\b", text)]


def strip_source_attribution(answer: str) -> str:
    body = _SOURCE_LINE_RE.sub("", answer or "")
    body = body.replace(RAG_ANSWER_DISCLAIMER, "")
    return " ".join(body.split())


def verify_answer(question: str, answer: str, fragments) -> tuple[bool, str]:
    del question
    if answer is None:
        return False, "Empty answer (None)"
    if not isinstance(answer, str):
        return False, "Answer is not a string"
    context_text = "\n".join(getattr(f, "page_content", "") for f in fragments)
    body = strip_source_attribution(answer)
    answer_numbers = extract_numbers(body)
    if not answer_numbers:
        return True, "Verification passed"
    context_numbers = extract_numbers(context_text)
    missing_numbers = [
        n for n in answer_numbers if not any(abs(n - c) < 0.01 for c in context_numbers)
    ]
    if missing_numbers:
        return False, f"Number(s) {missing_numbers} not found in sources."
    return True, "Verification passed"
