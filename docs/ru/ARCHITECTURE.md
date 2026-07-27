# Архитектура Grounded LLM

Репозиторий — **ядро платформы** для ассистентов с ответами по документам (RAG) в любой отрасли.  
Отраслевые пакеты (HR, юриспруденция, поддержка и т.д.) — это **domain pack**: `config/` + `data/{tenant_id}/{domain_id}/`.

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
| **server** (Go) | Auth, сессии, LLM, numeric verify, citations, `/metrics` |
| **python** | HTTP RAG (`:5000`, Gunicorn) + **gRPC Retriever** (`:50051`) |
| **postgres** | Сессии, сообщения; опционально pgvector |
| **redis** | Кэш эмбеддингов + семантический кэш ответов LLM |
| **webapp** | Reference UI (nginx → Go) |
| **ollama** / **vllm** | Опциональный локальный LLM (`--profile ollama` / `vllm`) |

---

## Поток текстового чата

1. Клиент → Go `POST /message` (опционально `?stream=1`)
2. При пустой истории: опциональный **semantic response cache** (Redis) → заголовок `X-Cache: HIT`
3. Иначе Go → Python `POST /rag/context` (`domain_id`, `tenant_id`, `locale`)
4. Python: эмбеддинги (с Redis при `REDIS_URL`) → vector store (Chroma / Qdrant / pgvector) → hybrid/rerank → фрагменты
5. Go → OpenAI-совместимый LLM (`LLM_PROVIDER`) → проверка чисел → дисклеймер → Postgres (`citations[]`)
6. Прошедшие verify ответы могут попасть в response cache (`X-Cache: MISS` на первый запрос)

**Агенты:** gRPC `grounded.rag.v1.Retriever/Retrieve` на Python `:50051` (metadata `x-rag-service-token`).

Язык ответа и UI задаётся локалью (`ru` / `en`): см. `config/locales/`.

---

## Документы базы знаний

Форматы: **`.txt`**, **`.pdf`**, **`.docx`** → `rag/document_loaders.py` → чанки → выбранный vector backend.

Рекомендуемый путь: `data/{tenant_id}/{domain_id}/`.  
Старый layout `data/{domain_id}/` по-прежнему поддерживается.

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
