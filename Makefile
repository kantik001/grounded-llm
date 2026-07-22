.PHONY: build up up-prod down restart logs clean ps help test test-go test-py smoke load-smoke backup-smoke eval-retrieval eval-retrieval-ci reindex conformance-spec conformance-http

# Docker Compose project name
PROJECT_NAME := grounded_llm

# Main commands

## Build all images
build:
	docker compose -p $(PROJECT_NAME) build --no-cache

## Full rebuild without cache
build-no-cache:
	docker compose -p $(PROJECT_NAME) build --no-cache --pull

## Start all services in background
up:
	docker compose -p $(PROJECT_NAME) up -d

## Production overlay (required secrets in .env — see docker-compose.prod.yml)
up-prod:
	docker compose -p $(PROJECT_NAME) -f docker-compose.yml -f docker-compose.prod.yml up -d --build

## Start with rebuild of changed services
up-build:
	docker compose -p $(PROJECT_NAME) up -d --build

## Start in foreground (debug)
up-dev:
	docker compose -p $(PROJECT_NAME) up

## Stop all services
down:
	docker compose -p $(PROJECT_NAME) down

## Stop and remove volumes
down-volumes:
	docker compose -p $(PROJECT_NAME) down -v

## Restart all services
restart:
	docker compose -p $(PROJECT_NAME) restart

## Tail logs for all services
logs:
	docker compose -p $(PROJECT_NAME) logs -f

## Tail logs for one service (e.g. make logs-service SERVICE=webapp)
logs-service:
	docker compose -p $(PROJECT_NAME) logs -f $(SERVICE)

## Show service status
ps:
	docker compose -p $(PROJECT_NAME) ps

## Full cleanup: containers, images, volumes
clean:
	docker compose -p $(PROJECT_NAME) down -v --rmi all --remove-orphans

## Rebuild and restart one service (e.g. make rebuild SERVICE=webapp)
rebuild:
	docker compose -p $(PROJECT_NAME) up -d --build --force-recreate $(SERVICE)

## Health check (service status)
health:
	docker compose -p $(PROJECT_NAME) ps

## Go unit tests
test-go:
	cd server && go test -v -count=1 ./...

## Python unit tests
test-py:
	pip install -r tests/requirements-test.txt
	pytest tests/ -v

## Python SDK tests
test-sdk:
	pip install -e "sdk/python[dev]"
	pytest sdk/python/tests/ -v

test: test-go test-py

## OpenAPI conformance (offline, no server)
conformance-spec:
	pip install -r conformance/requirements.txt
	python -m conformance spec

## spec + HTTP (requires running server)
conformance-check:
	pip install -r conformance/requirements.txt
	python -m conformance check --url $(or $(URL),http://127.0.0.1:8080)

## Full HTTP conformance (requires running server: CONFORMANCE_BASE_URL)
conformance-http:
	pip install -r conformance/requirements.txt
	pytest conformance/test_openapi_http.py -v --tb=short

## Adversarial E2E (mock server)
adversarial-e2e:
	pip install requests
	python scripts/run_adversarial_e2e.py --base-url http://127.0.0.1:8080

## Smoke API (localhost:8080, TELEGRAM_AUTH_DISABLED=true)
smoke:
	powershell -ExecutionPolicy Bypass -File scripts/smoke.ps1

## Reindex Chroma (requires Python service or local env)
reindex:
	python scripts/reindex_rag.py

## RAG eval retrieval-only (PYTHON_RAG_URL, python на :5000)
eval-retrieval:
	pip install requests
	python scripts/run_rag_eval.py --suite default_en

## Full retrieval gate locally (reindex + start Python + all suites)
eval-retrieval-ci:
	bash scripts/ci_eval_retrieval.sh

## Template pack CLI
init-pack-list:
	python scripts/init_pack.py list

init-pack-install:
	python scripts/init_pack.py install $(PACK)

## Show available commands
help:
	@echo "Available commands:"
	@echo "  make build          - Build all Docker images"
	@echo "  make build-no-cache - Full rebuild without cache"
	@echo "  make up             - Start services in background"
	@echo "  make up-prod        - Start with production overlay (required secrets)"
	@echo "  make up-build       - Start with rebuild"
	@echo "  make up-dev         - Start in foreground (debug)"
	@echo "  make down           - Stop services"
	@echo "  make down-volumes   - Stop and remove volumes"
	@echo "  make restart        - Restart services"
	@echo "  make logs           - Tail logs for all services"
	@echo "  make logs-service SERVICE=<name> - Tail logs for one service"
	@echo "  make ps             - Show service status"
	@echo "  make clean          - Full cleanup (containers, images, volumes)"
	@echo "  make rebuild SERVICE=<name> - Rebuild and restart one service"
	@echo "  make health         - Health check (service status)"
	@echo "  make test-go        - Go unit tests (server/)"
	@echo "  make test-py        - Python unit tests (tests/)"
	@echo "  make test           - test-go + test-py"
	@echo "  make eval-retrieval   - RAG eval (needs Python on :5000)"
	@echo "  make eval-retrieval-ci - Reindex + RAG + all eval suites (local CI gate)"
	@echo "  make init-pack-list    - List official template packs"
	@echo "  make init-pack-install PACK=it_support - Install a template pack"
	@echo "  make smoke          - Smoke API (localhost:8080)"
	@echo "  make load-smoke     - Concurrent load smoke (mock server on :8080)"
	@echo "  make backup-smoke   - Postgres pg_dump/pg_restore round-trip"
	@echo "  make help           - This help"

## Concurrent load smoke (requires server with TELEGRAM_AUTH_DISABLED + mocks or real stack)
load-smoke:
	bash scripts/load_smoke.sh $(or $(URL),http://127.0.0.1:8080) $(or $(N),20) $(or $(ROUNDS),2)

## Postgres backup/restore smoke (needs psql/pg_dump + reachable Postgres)
backup-smoke:
	bash scripts/backup_postgres_smoke.sh
