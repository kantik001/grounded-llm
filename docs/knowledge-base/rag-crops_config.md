# Разбор: `rag/crops_config.py`

**Исходный файл:** `rag/crops_config.py`  
**Конфиг:** `config/crops.json`  
**Кто использует:** `rag/vector_store.py`, `rag/retrieval.py`, `cv/registry.py`, Go (`server/crops.go`), тесты

---

## Зачем этот файл

Единая точка правды для **списка культур** (яблоня, груша, слива) и флагов:

- `rag_enabled` — можно ли искать статьи;
- `cv_enabled` — можно ли классифицировать фото.

Один JSON — и Python RAG/CV, и Go API читают одно и то же (через свой код / env `CROPS_CONFIG_PATH`).

---

## Кэш конфига

```python
_CONFIG: Optional[Dict[str, Any]] = None
```

`load_crops_config()` читает JSON **один раз** за жизнь процесса Python. После правки `crops.json` нужен **перезапуск** classifier/server.

---

## Где ищется `crops.json`

1. `CROPS_CONFIG_PATH` из env (в Docker: `/config/crops.json`)
2. `{корень}/config/crops.json`
3. `/config/crops.json`

---

## Функции

### `load_crops_config()`

Возвращает весь JSON, например:

```json
{
  "default_crop": "apple",
  "crops": {
    "apple": { "name_ru": "Яблоня", "cv_enabled": true, "rag_enabled": true },
    "pear": { "cv_enabled": false, "rag_enabled": false }
  }
}
```

### `default_crop_id()`

Обычно `"apple"` — подставляется, если `crop_id` не передали.

### `normalize_crop_id(crop_id)`

- приводит к нижнему регистру;
- пустое → `default_crop`;
- неизвестная культура → **`ValueError`** с текстом «Неизвестная культура».

Используется **везде** перед RAG и CV, чтобы не искать по несуществующему `crop_id`.

### `get_crop(crop_id)`

Словарь одной культуры после нормализации — для проверки `rag_enabled` / `cv_enabled`.

### `list_crops()`

Упрощённый ответ для API `GET /crops` в `api/app.py`.

---

## Связь с продуктом

| Флаг | Эффект |
|------|--------|
| `rag_enabled: false` | `retrieve_rag_context` вернёт ошибку «база статей не подключена» |
| `cv_enabled: false` | `get_classifier_for_crop` выбросит ValueError |

Сейчас только **apple** с обоими `true`.

---

## Краткий итог

`crops_config.py` — тонкая обёртка над `config/crops.json`: загрузка, кэш, нормализация `crop_id`. Без ML; только конфигурация домена.
