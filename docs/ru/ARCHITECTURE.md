# Архитектура Grounded LLM

Репозиторий — **ядро платформы** для ассистентов с ответами по документам (RAG) в любой отрасли.  
Отраслевые пакеты — **domain pack**: `config/` + исходники в `packs/*/data/` (регистрируются в registry при install).

Локальные LLM и кэши: [../en/LLM_PROVIDERS.md](../en/LLM_PROVIDERS.md) (EN).

---

## Слои

```
┌─────────────────────────────────────────────────────────┐
│  Ядро платформы (этот репозиторий)                      │
│  Go · Python RAG · Redis · verify · админка · CI        │
└───────────────────────────┬─────────────────────────────┘
                            │
              ┌─────────────┴─────────────┐
              ▼                           ▼
        Domain pack A              Domain pack B
        config + data/               config + data/
```

| Слой | Папки | Меняется часто? |
|------|-------|-----------------|
| **Ядро** | `server/`, `api/`, `rag/`, `migrations/`, `webapp/`, `scripts/` | Нет |
| **Domain pack** | `config/domains.json`, `config/locales/{ru,en}/`, `data/*` | **Да** |
| **Опционально** | Redis (кэши), Ollama / vLLM (локальный LLM) | По задаче |

- **`domain_id`** — идентификатор домена / базы знаний.
- **`tenant_id`** — изоляция арендаторов (мультитенантность).

---

## Сервисы Compose

| Сервис | Роль |
|--------|------|
| **server** (Go) | Auth, сессии, LLM, verify (local или remote), citations, `/metrics` |
| **python** | HTTP RAG (`:5000`, Gunicorn) + **gRPC Retriever** (`:50051`) |
| **postgres** | Сессии, сообщения; опционально pgvector |
| **redis** | Кэш эмбеддингов + семантический кэш ответов LLM |
| **webapp** | Reference UI (nginx → Go) |
| **ollama** / **vllm** | Опциональный локальный LLM (`--profile ollama` / `vllm`) |
| **guardrails** (опц.) | [grounded-guardrails](https://github.com/kantik001/grounded-guardrails) gRPC `:50052` — см. [../en/GUARDRAILS.md](../en/GUARDRAILS.md) |

---

## Поток текстового чата

1. Клиент → Go `POST /message` (опционально `?stream=1`)
2. При пустой истории: опциональный **semantic response cache** (Redis) → заголовок `X-Cache: HIT`
3. Иначе Go → Python `POST /rag/context` (`domain_id`, `tenant_id`, `locale`)
4. Python: эмбеддинги (с Redis при `REDIS_URL`) → vector store (Chroma / Qdrant / pgvector) → hybrid/rerank → фрагменты
5. Go → OpenAI-совместимый LLM (`LLM_PROVIDER`)
6. **Verify** после LLM:
   - `GUARDRAILS_MODE=local` (по умолчанию): in-process проверка чисел в `internal/rag/verify.go`
   - `remote` / `hybrid`: gRPC `VerifyText` → grounded-guardrails `:50052` (hybrid при сбое сети откатывается на local)
7. Дисклеймер → Postgres (`citations[]`); прошедшие verify ответы могут попасть в response cache

**Агенты:** gRPC `grounded.rag.v1.Retriever/Retrieve` на Python `:50051` (metadata `x-rag-service-token`).

**Порты:** Retriever `:50051` · Guardrails (опц.) `:50052`.

Язык ответа и UI задаётся локалью (`ru` / `en`): см. `config/locales/`.

---

## Документы базы знаний

Форматы: **`.txt`**, **`.pdf`**, **`.docx`** → `rag/document_loaders.py` → чанки → vector backend.

**Prod SoT:** Postgres `kb_documents` + blobs — [KB_SOURCE_OF_TRUTH.md](./KB_SOURCE_OF_TRUTH.md).  
Каталог `data/` — только demo-файлы в git; для runtime используйте upload, pack install или backfill.

---

## Новый ассистент из template pack

Предпочтительно [packs/](../../packs/):

```bash
python scripts/init_pack.py list
python scripts/init_pack.py install it_support
python scripts/reindex_rag.py
```

---

## Документация

- Провайдеры / Redis / gRPC: [../en/LLM_PROVIDERS.md](../en/LLM_PROVIDERS.md)
- Деплой: [DEPLOY.md](./DEPLOY.md)
- Knowledge base: [knowledge-base/README.md](./knowledge-base/README.md)
