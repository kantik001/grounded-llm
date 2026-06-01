# Разбор: папка `tests/` (Python) и тесты проекта

**Папка:** `tests/` — **только Python** (pytest)  
**Связанные тесты Go:** `server/*_test.go` (отдельная папка, тот же смысл «проверки»)  
**CI:** job `python-test` и `go-test` в [github-ci.yml.md](./github-ci.yml.md)

---

## Зачем тесты в этом проекте

Проверяют **логику без Docker, без LLM, без Chroma, без Telegram**:

- верно ли парсятся числа в ответе RAG;
- корректно ли читается `config/crops.json`.

Это **unit-тесты** — быстрые, дешёвые, гоняются на каждый push в CI.

Не заменяют [smoke](scripts-overview.md) и ручной чат в webapp.

---

## Файлы в `tests/`

| Файл | Назначение |
|------|------------|
| `conftest.py` | Общая настройка pytest: корень проекта в `PYTHONPATH` |
| `test_verifier.py` | Тесты `rag/verifier.py` |
| `test_crops_config.py` | Тесты `rag/crops_config.py` |
| `requirements-test.txt` | Минимальные зависимости для pytest (без PyTorch/Chroma) |

Папки `tests/__pycache__/` и `.pytest_cache/` — автогенерация, в git не нужны.

---

## `conftest.py`

```python
_ROOT = .../doctor_gardens_ai
sys.path.insert(0, _ROOT)
```

Чтобы `from rag.verifier import ...` работало при запуске из любой директории.

**Фикстур с БД нет** — тесты чисто функциональные.

---

## `requirements-test.txt`

```
pytest>=8.0.0
langchain-core>=0.3.0
```

- **pytest** — раннер тестов;
- **langchain-core** — только класс `Document` в тестах verifier (как в проде).

Намеренно **нет** `torch`, `langchain-chroma`, `flask` — CI и локальный `pytest` остаются лёгкими.

---

## `test_verifier.py` — что проверяет

Модуль: [rag-verifier.md](./rag-verifier.md) (в проде дубль на Go в `rag_chat.go`).

### `test_extract_numbers_decimal_comma`

- Вход: `"304,7 кг"`
- Ожидание: `[304.7]` — запятая как десятичный разделитель.

### `test_verify_passes_with_matching_number`

- Фрагмент: «Среднее 77.»
- Ответ: «Среднее 77.» + дисклеймер
- `verify_answer` → **успех** (число есть в контексте).

### `test_verify_fails_on_hallucinated_number`

- Фрагмент: без цифр
- Ответ: «Рентабельность 72%»
- `verify_answer` → **провал**, в reason есть «72» или «не найдены».

Это тот же кейс, из-за которого в чате ответ мог не пройти verify.

### `test_strip_source_attribution`

- Убирает строку `Источник: "Журнал"`, оставляет «Факт».

---

## `test_crops_config.py` — что проверяет

Модуль: [rag-crops_config.md](./rag-crops_config.md).

### Фикстура `crops_config_path` (autouse)

Перед **каждым** тестом:

1. `monkeypatch.setenv("CROPS_CONFIG_PATH", .../config/crops.json)`
2. Сброс `rag.crops_config._CONFIG = None` — перечитать JSON заново.

Без этого тесты могли бы взять неверный путь или старый кэш.

### `test_normalize_crop_id_apple`

- `"apple"` и `" Apple "` → `"apple"`.

### `test_normalize_crop_id_unknown`

- `"banana_xyz"` → `ValueError` с текстом «Неизвестная».

### `test_list_crops_has_apple`

- `default_crop == "apple"`;
- в списке есть `apple`, `rag_enabled is True`.

---

## Как запустить локально

Из **корня** проекта:

```powershell
pip install -r tests/requirements-test.txt
$env:CROPS_CONFIG_PATH = "config/crops.json"
pytest tests/ -v --tb=short
```

Или:

```bash
make test-py
```

Ожидание: **8 passed** (4 verifier + 4 crops, включая `demo_hr`).

---

## Что **не** покрыто папкой `tests/`

| Не тестируется | Почему |
|----------------|--------|
| Chroma / embeddings | тяжело, медленно |
| `vector_store.py`, `retrieval.py` end-to-end | нужен индекс и HF-модель |
| LLM API | платно, недетерминировано |
| Flask `api/app.py` | нет HTTP-тестов |
| PostgreSQL | нет testcontainers |

Это нормально для текущего этапа; расширение — eval-набор (в плане).

---

## Go-тесты (не в `tests/`, но часть той же стратегии)

Лежат в **`server/`**, запуск: `go test ./...` или `make test-go`.

| Файл | Что проверяет |
|------|----------------|
| `rag_chat_test.go` | числа, verify, дисклеймер, очистка ответа (зеркало Python verifier) |
| `crops_test.go` | `normalizeCropID`, загрузка crops.json |
| `auth_telegram_test.go` | подпись `initData` (валидный / неверный hash) |
| `admin_test.go` | regex имён файлов для upload (только латиница, `.txt`) |

Синхронизация: при смене логики в `rag/verifier.py` проверьте и `rag_chat_test.go`.

---

## Связь с CI и Makefile

| Команда | Действие |
|---------|----------|
| `make test-py` | только `tests/` |
| `make test-go` | только `server/*_test.go` |
| `make test` | оба |
| GitHub `python-test` | как `make test-py` + env `CROPS_CONFIG_PATH` |

Smoke (`make smoke`) — **не** pytest, см. [scripts-overview.md](./scripts-overview.md).

---

## Как добавить новый Python-тест

1. Создать `tests/test_<модуль>.py`.
2. Импорт из `rag.*` или другого пакета — `conftest` уже добавляет корень в path.
3. Если нужен env — фикстура с `monkeypatch` как в `test_crops_config.py`.
4. Тяжёлые зависимости — только если без них нельзя; иначе CI раздуется.

Именование: `test_что_проверяем` — pytest подхватит автоматически.

---

## Краткий итог

`tests/` — **7 быстрых unit-тестов** на критичную бизнес-логику RAG (числа) и конфиг культур. Дешёвые, в CI на каждый push. Полный путь «статья → поиск → LLM» здесь не тестируется — для этого smoke, ручные проверки и будущий eval-набор.
