.PHONY: \
	build build-no-cache up up-prod up-build up-dev down down-volumes restart \
	logs logs-service ps clean rebuild health \
	test test-go test-py test-sdk \
	conformance-spec conformance-check conformance-http adversarial-e2e \
	smoke reindex eval-retrieval eval-retrieval-ci \
	init-pack-list init-pack-install \
	load-smoke backup-smoke help

# Docker Compose project name
PROJECT_NAME := grounded_llm

# ---------------------------------------------------------------------------
# Docker lifecycle
# ---------------------------------------------------------------------------

## Build all images (no cache)
build:
	docker compose -p $(PROJECT_NAME) build --no-cache

## Full rebuild without cache + pull base images
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

## Stop and remove volumes (destructive: DB / chroma / uploads data)
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

## Alias of ps (compose status)
health: ps

## Full cleanup: containers, images, volumes
clean:
	docker compose -p $(PROJECT_NAME) down -v --rmi all --remove-orphans

## Rebuild and restart one service (e.g. make rebuild SERVICE=webapp)
rebuild:
	docker compose -p $(PROJECT_NAME) up -d --build --force-recreate $(SERVICE)

# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------

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

## Go + Python unit tests (local PR minimum)
test: test-go test-py

# ---------------------------------------------------------------------------
# Conformance / quality
# ---------------------------------------------------------------------------

## OpenAPI conformance (offline, no server)
conformance-spec:
	pip install -r conformance/requirements.txt
	python -m conformance spec

## Spec + HTTP check (requires running server)
conformance-check:
	pip install -r conformance/requirements.txt
	python -m conformance check --url $(or $(URL),http://127.0.0.1:8080)

## Full HTTP conformance (requires running server)
conformance-http:
	pip install -r conformance/requirements.txt
	python -m conformance http --url $(or $(URL),http://127.0.0.1:8080)

## Adversarial E2E against /message (server on :8080)
adversarial-e2e:
	pip install requests
	python scripts/run_adversarial_e2e.py --base-url $(or $(URL),http://127.0.0.1:8080)

## Smoke API (localhost:8080; TELEGRAM_AUTH_DISABLED=true recommended)
smoke:
ifeq ($(OS),Windows_NT)
	powershell -ExecutionPolicy Bypass -File scripts/smoke.ps1
else
	bash scripts/smoke.sh $(or $(URL),http://127.0.0.1:8080)
endif

## Reindex vector store (requires Python RAG service or local env)
reindex:
	python scripts/reindex_rag.py

## RAG retrieval-only eval (needs Python on :5000; PYTHON_RAG_URL optional)
eval-retrieval:
	pip install requests
	python scripts/run_rag_eval.py --suite $(or $(SUITE),default_en)

## Full retrieval gate locally (reindex + Python + all suites; same family as CI)
eval-retrieval-ci:
	bash scripts/ci_eval_retrieval.sh

## Concurrent load smoke (TELEGRAM_AUTH_DISABLED + mocks or real stack)
load-smoke:
	bash scripts/load_smoke.sh $(or $(URL),http://127.0.0.1:8080) $(or $(N),20) $(or $(ROUNDS),2)

## Postgres backup/restore smoke (needs psql/pg_dump + reachable Postgres)
backup-smoke:
	bash scripts/backup_postgres_smoke.sh

# ---------------------------------------------------------------------------
# Template packs
# ---------------------------------------------------------------------------

## List official template packs
init-pack-list:
	python scripts/init_pack.py list

## Install a pack (e.g. make init-pack-install PACK=it_support)
init-pack-install:
	python scripts/init_pack.py install $(PACK)

# ---------------------------------------------------------------------------
# Help
# ---------------------------------------------------------------------------

help:
	@echo "Available commands:"
	@echo "  make build / build-no-cache - Build Docker images"
	@echo "  make up / up-build / up-dev  - Start services"
	@echo "  make up-prod                - Start with production overlay"
	@echo "  make down / down-volumes    - Stop (optionally wipe volumes)"
	@echo "  make restart / ps / health  - Restart or show status"
	@echo "  make logs / logs-service SERVICE=<name>"
	@echo "  make rebuild SERVICE=<name> - Rebuild one service"
	@echo "  make clean                  - Remove containers, images, volumes"
	@echo "  make test / test-go / test-py / test-sdk"
	@echo "  make reindex                - Reindex knowledge base"
	@echo "  make eval-retrieval [SUITE=default_en]"
	@echo "  make eval-retrieval-ci      - Local CI retrieval gate"
	@echo "  make smoke / load-smoke / backup-smoke / adversarial-e2e"
	@echo "  make conformance-spec|check|http"
	@echo "  make init-pack-list"
	@echo "  make init-pack-install PACK=it_support"
	@echo "  make help"
