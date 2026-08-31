# `api/` — Python RAG service

**Актуальное описание:** [../../en/knowledge-base/python-api.md](../../en/knowledge-base/python-api.md)  
**Карта пакета:** [`api/README.md`](../../../api/README.md)

Кратко: внутренний сервис поиска (не публичный Go `/api/v1`). Admin: reindex + **ingest** (`/admin/ingest/*`) — см. [INGESTION.md](../INGESTION.md).

```bash
python -m flask --app api.http.app run -p 5000
python -m api.grpc
sh api/entrypoint.sh
```

Docker: `ENTRYPOINT tini` + `CMD ["/app/api/entrypoint.sh"]` в `Dockerfile.python`.
