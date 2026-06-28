.PHONY: build up down restart logs clean ps help test test-go test-py smoke eval-retrieval eval-retrieval-ci reindex

# Имя проекта Docker Compose
PROJECT_NAME := grounded_llm

# Основные команды

## Сборка всех образов
build:
	docker compose -p $(PROJECT_NAME) build --no-cache

## Сборка без кэша (полная пересборка)
build-no-cache:
	docker compose -p $(PROJECT_NAME) build --no-cache --pull

## Запуск всех сервисов в фоновом режиме
up:
	docker compose -p $(PROJECT_NAME) up -d

## Запуск с пересборкой изменённых сервисов
up-build:
	docker compose -p $(PROJECT_NAME) up -d --build

## Запуск в режиме foreground (для отладки)
up-dev:
	docker compose -p $(PROJECT_NAME) up

## Остановка всех сервисов
down:
	docker compose -p $(PROJECT_NAME) down

## Остановка с удалением томов
down-volumes:
	docker compose -p $(PROJECT_NAME) down -v

## Перезапуск всех сервисов
restart:
	docker compose -p $(PROJECT_NAME) restart

## Просмотр логов всех сервисов
logs:
	docker compose -p $(PROJECT_NAME) logs -f

## Просмотр логов конкретного сервиса (пример: make logs-service SERVICE=webapp)
logs-service:
	docker compose -p $(PROJECT_NAME) logs -f $(SERVICE)

## Показать статус сервисов
ps:
	docker compose -p $(PROJECT_NAME) ps

## Очистка: остановка, удаление контейнеров, образов и томов
clean:
	docker compose -p $(PROJECT_NAME) down -v --rmi all --remove-orphans

## Пересборка и запуск одного сервиса (пример: make rebuild SERVICE=webapp)
rebuild:
	docker compose -p $(PROJECT_NAME) up -d --build --force-recreate $(SERVICE)

## Проверка здоровья сервисов
health:
	docker compose -p $(PROJECT_NAME) ps

## Unit-тесты Go
test-go:
	cd server && go test -v -count=1 ./...

## Unit-тесты Python
test-py:
	pip install -r tests/requirements-test.txt
	pytest tests/ -v

test: test-go test-py

## Smoke API (localhost:8080, TELEGRAM_AUTH_DISABLED=true)
smoke:
	powershell -ExecutionPolicy Bypass -File scripts/smoke.ps1

## Переиндексация Chroma (нужен Python-сервис / локальное окружение)
reindex:
	python scripts/reindex_rag.py

## RAG eval retrieval-only (PYTHON_RAG_URL, python на :5000)
eval-retrieval:
	pip install requests
	python scripts/run_rag_eval.py --suite default_en

## Full retrieval gate (reindex + start Python + all suites) — same as CI job eval-retrieval-gate
eval-retrieval-ci:
	bash scripts/ci_eval_retrieval.sh

## Помощь по доступным командам
help:
	@echo "Доступные команды:"
	@echo "  make build          - Сборка всех образов"
	@echo "  make build-no-cache - Полная пересборка без кэша"
	@echo "  make up             - Запуск сервисов в фоне"
	@echo "  make up-build       - Запуск с пересборкой"
	@echo "  make up-dev         - Запуск в режиме отладки (foreground)"
	@echo "  make down           - Остановка сервисов"
	@echo "  make down-volumes   - Остановка с удалением томов"
	@echo "  make restart        - Перезапуск сервисов"
	@echo "  make logs           - Просмотр логов всех сервисов"
	@echo "  make logs-service SERVICE=<name> - Логи конкретного сервиса"
	@echo "  make ps             - Статус сервисов"
	@echo "  make clean          - Полная очистка (контейнеры, образы, тома)"
	@echo "  make rebuild SERVICE=<name> - Пересборка и запуск одного сервиса"
	@echo "  make health         - Проверка статуса сервисов"
	@echo "  make test-go        - Unit-тесты Go (server/)"
	@echo "  make test-py        - Unit-тесты Python (tests/)"
	@echo "  make test           - test-go + test-py"
	@echo "  make eval-retrieval   - RAG eval (needs Python on :5000)"
	@echo "  make eval-retrieval-ci - Reindex + RAG + all eval suites (local CI gate)"
	@echo "  make smoke          - Smoke API (localhost:8080)"
	@echo "  make help           - Эта справка"
