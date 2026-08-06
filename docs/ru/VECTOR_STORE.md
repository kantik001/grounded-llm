**Канон (EN):** [VECTOR_STORE.md](../en/VECTOR_STORE.md)

# Векторное хранилище (адаптер)

Grounded LLM поддерживает сменные vector backend для Python RAG. Референс — **Chroma** (локальный persist). **Qdrant** и **pgvector** — опционально для команд со своим managed-хранилищем.

Зачем: один API поиска (`rag/vector_store.py`), разный способ хранения эмбеддингов — без переписывания retrieval.

---

## Конфигурация

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `VECTOR_STORE` | `chroma` | Backend: `chroma`, `qdrant` или `pgvector` |
| `CHROMA_PERSIST_DIR` | `./chroma_db` | Путь Chroma на диске |
| `QDRANT_URL` | `http://127.0.0.1:6333` | HTTP endpoint Qdrant |
| `QDRANT_COLLECTION` | `grounded_llm` | Имя коллекции Qdrant |
| `PGVECTOR_URL` | `DATABASE_URL` | DSN Postgres для pgvector (`postgresql+psycopg://...`) |
| `PGVECTOR_COLLECTION` | `grounded_chunks` | Имя коллекции pgvector |
| `RAG_RETRIEVAL_MODE` | `vector` | `vector` или `hybrid` (BM25 + dense + RRF) |
| `RAG_RRF_K` | `60` | Константа RRF для гибридного слияния |
| `SPARSE_INDEX_DIR` | `./sparse_index` | Persist-путь BM25 индекса |
| `RAG_RERANKER` | `none` | `none`, `keyword` или `cross_encoder` |
| `RAG_CROSS_ENCODER_MODEL` | `cross-encoder/ms-marco-MiniLM-L-6-v2` | Модель cross-encoder |
| `FORCE_RAG_REINDEX` | `false` | Полная перестройка индекса при старте |
| `REDIS_URL` | (не задан) | Опциональный Redis для кэша эмбеддингов (`embedding:{md5}:{model}`) |
| `EMBEDDING_CACHE_TTL_SEC` | `3600` | TTL кэша эмбеддингов |

Подробнее про кэш: [LLM_PROVIDERS.md](./LLM_PROVIDERS.md) · [EN](../en/LLM_PROVIDERS.md).

---

## Chroma (по умолчанию)

Используется в Docker Compose, Helm и CI (`eval-retrieval-gate`).

```bash
VECTOR_STORE=chroma python scripts/reindex_rag.py
```

Фильтр метаданных на каждом чанке: `domain_id` + `tenant_id`.

---

## Qdrant (опционально)

```bash
pip install -r api/requirements-qdrant.txt
docker run -p 6333:6333 qdrant/qdrant

VECTOR_STORE=qdrant QDRANT_URL=http://127.0.0.1:6333 FORCE_RAG_REINDEX=true python scripts/reindex_rag.py
```

Смена backend или embedding-модели → **полный reindex** и повторный прогон eval gate.

---

## pgvector (опционально)

Эмбеддинги в PostgreSQL с расширением **pgvector** — при желании та же БД, что и для сессий.

**Требования:**

- образ Postgres с pgvector (референс: `pgvector/pgvector:pg16`)
- миграция `009_pgvector.sql` (`CREATE EXTENSION vector`)

```bash
pip install -r api/requirements-pgvector.txt

VECTOR_STORE=pgvector \
DATABASE_URL=postgres://grounded:grounded@localhost:5432/grounded?sslmode=disable \
FORCE_RAG_REINDEX=true \
python scripts/reindex_rag.py
```

Compose: `VECTOR_STORE=pgvector docker compose up -d --build`.

Чанки со стабильным `chunk_id`; фильтры: `domain_id` + `tenant_id`.

---

## Reranking и hybrid retrieval

| Режим | Env | Заметки |
|-------|-----|---------|
| Только vector | `RAG_RETRIEVAL_MODE=vector` (default) | Top-k по dense search |
| **Hybrid (BM25 + dense + RRF)** | `RAG_RETRIEVAL_MODE=hybrid` | Sparse BM25 + dense, слияние через RRF |
| Keyword rerank | `RAG_RERANKER=keyword` | Второй этап по пересечению токенов |
| Cross-encoder | `RAG_RERANKER=cross_encoder` | Медленнее, часто лучше на policy Q&A |

### Поток hybrid

1. Взять `3× rag_k` хитов из dense (Chroma / Qdrant / pgvector)  
2. Взять `3× rag_k` из BM25 (`rag/sparse_index.py`)  
3. Слить через RRF (`RAG_RRF_K`, default `60`)  
4. Опционально `RAG_RERANKER`  
5. Вернуть top `rag_k`  

Sparse-индекс пересобирается при `FORCE_RAG_REINDEX=true` или `python scripts/reindex_rag.py`, persist в `sparse_index/` (или `SPARSE_INDEX_DIR`).

```bash
RAG_RETRIEVAL_MODE=hybrid python scripts/run_rag_eval.py --suite hybrid
RAG_RETRIEVAL_MODE=hybrid python scripts/run_rag_eval.py --suite default_en
```

CI по умолчанию держит `vector`, чтобы gate оставался быстрым.

---

## Реализация

| Модуль | Роль |
|--------|------|
| `rag/indexing.py` | Chunking + metadata `chunk_id` |
| `rag/sparse_index.py` | BM25 sparse index |
| `rag/rrf.py` | Reciprocal rank fusion |
| `rag/vector_backend/` | Интерфейс + Chroma / Qdrant / pgvector |
| `rag/vector_store.py` | Публичный API (`search`, `index_stats_for_domain`) |
| `rag/rerank.py` | Keyword + cross-encoder |

---

## См. также

- [COMPATIBILITY.md](./COMPATIBILITY.md) · [EN](../en/COMPATIBILITY.md)
- [BACKUP_RESTORE.md](./BACKUP_RESTORE.md) — бэкап Chroma PVC / снимков Qdrant
