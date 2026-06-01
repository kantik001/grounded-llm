# `rag/verifier.py`

**Source:** `rag/verifier.py`  
**Tests:** `tests/test_verifier.py`  
**Production:** primary check in Go — `server/rag_verify.go` (`verifyRAGAnswer`, `appendRAGDisclaimer`)

---

## Purpose

Guard against **numeric hallucinations**: if the LLM writes “72%” or “748.5 cm”, that number must **appear in retrieved fragments**.

Article titles (`Source: "..."`) are **not shown** to users — replaced by a general disclaimer (on Go; constant duplicated here for tests).

---

## Constant `RAG_ANSWER_DISCLAIMER`

Text appended at the end of answers (Go `appendRAGDisclaimer`):

> Reference information from the knowledge base. Not a substitute for official expert advice.

In `verifier.py` used by `strip_source_attribution` so disclaimer numbers do not affect verification.

---

## `extract_numbers(text)`

- Replaces `,` with `.` (European decimals).
- Regex: `\b\d+(?:\.\d+)?\b`.
- Returns list of `float`.

Examples: `72`, `748.5`, `496,0` → `496.0`.

---

## `strip_source_attribution(answer)`

1. Remove lines `Source: ...` (regex `_SOURCE_LINE_RE`).
2. Strip disclaimer text.
3. Collapse whitespace.

Used to verify **answer body** without trailing boilerplate.

---

## `verify_answer(question, answer, fragments)`

### Input

- `question` — **not used** in logic today (reserved);
- `answer` — LLM string;
- `fragments` — LangChain `Document` or objects with `page_content`.

### Algorithm

1. Join all `page_content` → `context_text`.
2. Clean answer → `body`.
3. Extract numbers from `body` and `context_text`.
4. For each number in the answer: present in context within **±0.01**?
5. If extra numbers → `(False, "Number(s) [...] not found in sources.")`.
6. If no numbers in answer → `(True, "Verification passed")`.

### Examples

| Answer | Context | Result |
|--------|---------|--------|
| “Spots on leaves” | no digits | pass |
| “Margin 496%” | 496 in article | pass |
| “Margin 72%” | no 72 | fail |

---

## Python vs Go

| | `rag/verifier.py` | `server/rag_verify.go` |
|--|-------------------|------------------------|
| Production | tests / reference | **yes**, after each RAG reply |
| Number logic | same idea | `verifyRAGAnswer`, `extractNumbersFromText` |
| Disclaimer | constant for strip | `appendRAGDisclaimer` |

Keep logic **in sync** when changing rules.

---

## Why users sometimes see “not in materials” on verify fail

Go may **not return** raw LLM text on failed verify — model invented a number → verifier catches it.

---

## Tests

`tests/test_verifier.py`:

- decimal comma;
- pass with number in context;
- fail on 72 without 72 in context;
- strip `Source:` lines.

Run: `pytest tests/test_verifier.py` (no Chroma, no LLM).

---

## What to read next

| Topic | File |
|-------|------|
| Prompt & post-processing | [server-rag_chat.md](./server-rag_chat.md) |
| Fragment source | [rag-retrieval.md](./rag-retrieval.md) |

---

## Summary

`verifier.py` — **anti-hallucination for numbers** and answer cleanup utilities. Production duplicate on Go; Python is reference for pytest and documentation.
