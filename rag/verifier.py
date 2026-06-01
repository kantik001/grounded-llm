# RAG answer verification: numeric claims must appear in retrieved fragments.
import re
from typing import List, Tuple

from langchain_core.documents import Document

RAG_ANSWER_DISCLAIMER = (
    "Справочная информация из базы знаний. "
    "Не заменяет официальную консультацию ответственного специалиста."
)

_SOURCE_LINE_RE = re.compile(r"(?im)^\s*Источник:.*\n?")


def extract_numbers(text: str) -> List[float]:
    if not text:
        return []
    text = text.replace(",", ".")
    numbers = re.findall(r"\b\d+(?:\.\d+)?\b", text)
    return [float(n) for n in numbers]


def strip_source_attribution(answer: str) -> str:
    s = _SOURCE_LINE_RE.sub("", answer or "")
    s = s.replace(RAG_ANSWER_DISCLAIMER, "")
    return " ".join(s.split())


def verify_answer(question: str, answer: str, fragments: List[Document]) -> Tuple[bool, str]:
    if answer is None:
        return False, "Ответ отсутствует (None)"
    if not isinstance(answer, str):
        return False, "Ответ не является строкой"

    context_text = "\n".join([f.page_content for f in fragments])
    body = strip_source_attribution(answer)
    numbers_in_answer = extract_numbers(body)
    if numbers_in_answer:
        numbers_in_context = extract_numbers(context_text)
        missing_numbers = []
        for num in numbers_in_answer:
            if not any(abs(num - ctx_num) < 0.01 for ctx_num in numbers_in_context):
                missing_numbers.append(num)
        if missing_numbers:
            return False, f"Число(а) {missing_numbers} не найдены в источниках."

    return True, "Верификация пройдена"
