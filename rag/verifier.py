"""Answer verification helpers (mirror of server/internal/rag/verify.go for tests)."""

import re

_STEM_PREFIX_LEN = 5
_MIN_SENTENCE_TOKENS = 4

_SENTENCE_SPLIT_RE = re.compile(r"[.!?…]+\s+|[.!?…]+$|\n+")
_CONTENT_WORD_RE = re.compile(r"[^\W_]+", re.UNICODE)

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


def _stem(word: str) -> str:
    return word.lower()[:_STEM_PREFIX_LEN]


def _content_stems(text: str) -> list[str]:
    out = []
    for w in _CONTENT_WORD_RE.findall(text or ""):
        if not w.isdigit() and len(w) < 4:
            continue
        out.append(_stem(w))
    return out


def unsupported_sentences(body: str, context_text: str, min_support: float = 0.5) -> list[str]:
    """Answer sentences whose lexical (stem-based) support in the context is
    below min_support. Mirror of Go UnsupportedSentences (Spec v2)."""
    context_stems = set(_content_stems(context_text))
    flagged: list[str] = []
    for raw in _SENTENCE_SPLIT_RE.split(body or ""):
        sentence = raw.strip()
        if not sentence:
            continue
        stems = _content_stems(sentence)
        if len(stems) < _MIN_SENTENCE_TOKENS:
            continue
        found = sum(1 for s in stems if s in context_stems)
        if found / len(stems) < min_support:
            flagged.append(sentence)
    return flagged


def verify_answer(question: str, answer: str, fragments) -> tuple[bool, str]:
    del question
    if answer is None:
        return False, "Empty answer (None)"
    if not isinstance(answer, str):
        return False, "Answer is not a string"
    context_text = "\n".join(getattr(f, "page_content", "") for f in fragments)
    body = strip_source_attribution(answer)
    answer_numbers = extract_numbers(body)
    if answer_numbers:
        context_numbers = extract_numbers(context_text)
        missing_numbers = [
            n for n in answer_numbers if not any(abs(n - c) < 0.01 for c in context_numbers)
        ]
        if missing_numbers:
            return False, f"Number(s) {missing_numbers} not found in sources."
    # Spec v2 faithfulness — mirrors Go VERIFY_FAITHFULNESS (default warn).
    import os

    mode = (os.environ.get("VERIFY_FAITHFULNESS") or "warn").strip().lower()
    if mode != "off":
        flagged = unsupported_sentences(body, context_text)
        if flagged and mode == "enforce":
            return False, f"Unsupported claim(s): {flagged[0][:120]}"
    return True, "Verification passed"
