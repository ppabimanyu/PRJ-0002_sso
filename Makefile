.PHONY: help dev down logs observability migrate-up migrate-down backend-dev frontend-dev generate test test-integration lint build smoke

COMPOSE := docker compose -f deploy/docker-compose.yml
ifneq (,$(wildcard .env))
COMPOSE += --env-file .env
endif

help:
	@echo "dev             Start the full stack with Docker Compose"
	@echo "down            Stop the stack"
	@echo "logs            Follow stack logs"
	@echo "observability   Start the stack with Jaeger and Prometheus"
	@echo "migrate-up      Apply goose migrations"
	@echo "migrate-down    Roll back one goose migration"
	@echo "backend-dev     Run the API with air hot reload"
	@echo "frontend-dev    Run the Vite dev server"
	@echo "generate        Regenerate sqlc, OpenAPI, and frontend types"
	@echo "test            Run backend and frontend unit tests"
	@echo "test-integration Run backend integration tests (needs Docker)"
	@echo "lint            Run backend and frontend linters"
	@echo "build           Build backend and frontend artifacts"
	@echo "smoke           Check a running stack end to end"

dev:
	$(COMPOSE) --profile dev up -d --build

down:
	$(COMPOSE) --profile dev --profile observability down

logs:
	$(COMPOSE) --profile dev --profile observability logs -f

observability:
	OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318 $(COMPOSE) --profile dev --profile observability up -d --build

migrate-up:
	$(COMPOSE) run --rm migrate

migrate-down:
	$(COMPOSE) run --rm migrate -dir /migrations down

backend-dev:
	set -a; . ./.env; set +a; cd backend && air

frontend-dev:
	cd frontend && npm run dev

generate:
	cd backend && $(MAKE) generate
	cd frontend && npm run generate:api

test:
	cd backend && go test ./...
	cd frontend && npm run test -- --run

test-integration:
	cd backend && go test -tags=integration ./...

lint:
	cd backend && golangci-lint run
	cd frontend && npm run lint

build:
	cd backend && $(MAKE) build
	cd frontend && npm run build

smoke:
	./scripts/smoke.sh
