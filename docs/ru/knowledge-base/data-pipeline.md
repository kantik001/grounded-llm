# Пайплайн данных базы знаний

**Цель:** как документы попадают в RAG и доходят до ответа в чате.

Подробно (async ingest): [INGESTION.md](../INGESTION.md) · legacy reindex: [config/REINDEX.md](../../config/REINDEX.md)

---

## Два пути индексации

| Путь | Когда |
|------|--------|
| **Ingest pipeline** | Прод: job в Postgres, parse → embed → index, status API |
| **Reindex** | Dev/CI: один синхронный прогон (`reindex_rag.py`, `POST /admin/reindex`) |

Оба используют одни и те же loaders, chunker, embeddings и manifest/sha1 в Chroma.

---

## От файла до ответа

```mermaid
flowchart TB
    subgraph sources
        U[Admin upload]
        C[Connectors]
        G[Git / packs]
    end
    U --> D[data/tenant/domain/]
    C --> D
    G --> D
    D --> I{Индекс}
    I -->|ingest| P[parse → staging → embed → Chroma + BM25]
    I -->|reindex| R[sync refresh]
    P --> S[search / hybrid]
    R --> S
    S --> A[Go + LLM + verify]
    A --> User[Пользователь]
```

| Этап | Где |
|------|-----|
| Файл на диске | upload / connectors / git → `data/{tenant}/{domain}/` |
| **Ingest** | `POST /admin/ingest` → `ingest_jobs` → Redis → `ingest-worker` |
| **Reindex** | `scripts/reindex_rag.py` или `POST /admin/reindex` |
| Парсинг | `rag/document_loaders.py` |
| Chunk | `rag/indexing.py` (500/50) |
| Embed | e5-small + Redis cache |
| Dense | Chroma / Qdrant / pgvector |
| Sparse | BM25 при `hybrid` |
| Retrieval | `POST /rag/context` |
| Ответ | `internal/rag/pipeline.go` |

---

## Ingest (рекомендуется для admin)

```bash
curl -u admin:pass -X POST "http://localhost:8080/api/admin/ingest?domain_id=default" \
  -H "Content-Type: application/json" \
  -d '{"sync": false}'

curl -u admin:pass "http://localhost:8080/api/admin/ingest/status?job_id=1"
```

Пустой `files` = все файлы домена. Статусы: `queued → parsing → embedding → indexing → succeeded | failed | partial`.

См. [INGESTION.md](../INGESTION.md).

---

## Форматы

`.txt`, `.pdf`, `.docx` — до 10 МБ, латиница в имени upload.

---

## Шаг 1 — документы в `data/`

```
data/default/default/vacation_policy_en.txt
```

---

## Шаг 2 — индекс

**Прод:** `docker compose up -d ingest-worker` + `POST /admin/ingest`

**Dev/CI:** `python scripts/reindex_rag.py`

Без индексации новые файлы **не** попадут в Chroma.

---

## Шаг 3 — проверка

```bash
python scripts/run_rag_eval.py --suite default
```

---

## Дальше

| Тема | Файл |
|------|------|
| Ingestion | [INGESTION.md](../INGESTION.md) |
| Admin API | [server-admin-and-ux-api.md](./server-admin-and-ux-api.md) |
| Connectors | [CONNECTORS.md](../CONNECTORS.md) |
