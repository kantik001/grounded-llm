# Разбор: `rag/vector_store.py` / Vector store

**Исходный файл / Source:** `rag/vector_store.py`  
**Данные / Data:** `data/{domain_id}/*.{txt,pdf,docx}`  
**Хранилище / Storage:** `chroma_db/` (Docker volume `chroma_data`)  
**Кто вызывает / Called by:** `rag/retrieval.py`, admin reindex

---

## Назначение / Purpose

Ядро **векторного RAG**: документы → embeddings → **Chroma**. LLM здесь нет.

---

## Пайплайн индексации / Indexing pipeline

```mermaid
flowchart LR
    A[data/domain/*] --> B[document_loaders]
    B --> C[metadata domain_id filename file_type]
    C --> D[RecursiveCharacterTextSplitter]
    D --> E[chunk 500 overlap 50]
    E --> F[HuggingFaceEmbeddings e5-small]
    F --> G[Chroma persist chroma_db]
```

### `rag/document_loaders.py`

| Расширение | Loader |
|------------|--------|
| `.txt` | `TextLoader` (UTF-8) |
| `.pdf` | `PyPDFLoader` |
| `.docx` | `Docx2txtLoader` |

Metadata на каждом документе: `filename`, `domain_id`, `source_file`, `file_type`.

---

## `load_all_documents()`

- Обходит домены из `domains.json`
- Для каждого — glob по поддерживаемым расширениям

---

## `load_vector_store(force_reindex=False)`

| Ситуация | Поведение |
|----------|-----------|
| RAM-кэш `_vector_store` | вернуть |
| `FORCE_RAG_REINDEX=true` | удалить `chroma_db`, пересоздать |
| `chroma_db` существует | открыть Chroma |
| иначе | `create_vector_store()` |

---

## `search(query, domain_id, k=8)`

```python
store.similarity_search(query, k=k, filter={"domain_id": domain_id})
```

---

## Docker

- `./data:/app/data:ro` (python)
- `chroma_data:/app/chroma_db`
- `./data:/app/data` rw (server) — admin upload

После upload новых файлов — **reindex** обязателен.

---

## Зависимости / Dependencies

`api/requirements.txt`: `langchain-chroma`, `sentence-transformers`, `pypdf`, `docx2txt`.

---

## Что читать дальше

| Тема | Файл |
|------|------|
| Домены | [rag-domains_config.md](./rag-domains_config.md) |
| Retrieval | [rag-retrieval.md](./rag-retrieval.md) |
| HTTP reindex | [python-api.md](./python-api.md) |
