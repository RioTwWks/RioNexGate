# Запускать из корня репозитория: cd ~/RioNexGate && make dev
ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
DOCKER_COMPOSE := $(ROOT)scripts/docker-compose.sh

.PHONY: build up down dev dev-cores dev-cores-singbox dev-cores-skadi dev-local docker-doctor test clean migrate init logs docker-check

init:
	@if [ ! -f backend/config.yaml ]; then \
		cp backend/config.example.yaml backend/config.yaml; \
		KEY="$$(openssl rand -hex 16)"; \
		sed -i "s/change-me-to-secure-key/$$KEY/" backend/config.yaml; \
		echo "Generated server.api_key in backend/config.yaml"; \
	else \
		echo "backend/config.yaml already exists (not overwritten)"; \
	fi
	cp -n .env.example .env 2>/dev/null || true
	mkdir -p data/xray data/sing-box data/skadi data/awg data/nginx/ssl backups
	@chmod 640 backend/config.yaml 2>/dev/null || true
	@chmod 700 data backups 2>/dev/null || true
	@# Backend container runs as uid 1000 — ensure data/ is writable.
	@if [ "$$(id -u)" -eq 0 ]; then chown -R 1000:1000 data; \
	elif command -v docker >/dev/null 2>&1; then \
		docker run --rm -v "$$(pwd)/data:/data" alpine:3.20 chown -R 1000:1000 /data 2>/dev/null || \
		echo "Note: if backend cannot write ./data, run: sudo chown -R 1000:1000 data"; \
	fi
	@echo "Panel URL after make dev: http://localhost:$${HTTP_PORT:-8888}"
	@echo "Sign in with server.api_key from backend/config.yaml (placeholder keys are rejected at startup)."

docker-check:
	@$(ROOT)scripts/docker-check.sh

docker-doctor:
	@$(ROOT)scripts/docker-doctor.sh

build: docker-check
	$(DOCKER_COMPOSE) build

up: docker-check
	$(DOCKER_COMPOSE) up -d --build

down: docker-check
	$(DOCKER_COMPOSE) down

dev: docker-check
	$(DOCKER_COMPOSE) up --build

dev-cores: docker-check
	$(DOCKER_COMPOSE) --profile cores up -d xray-core

dev-cores-singbox: docker-check
	$(DOCKER_COMPOSE) --profile cores up -d sing-box

dev-cores-skadi: docker-check
	$(DOCKER_COMPOSE) --profile cores up -d --build skadi-core

dev-local:
	@echo "Local development without Docker:"
	@echo "  Terminal 1: make -C backend dev    # API http://localhost:8080"
	@echo "  Terminal 2: cd frontend && npm run dev   # UI http://localhost:5173"
	@echo ""
	@echo "Migrate DB first (once): make -C backend migrate"

test:
	cd backend && CGO_ENABLED=1 go test ./...
	cd frontend && npm test --if-present

test-e2e:
	cd e2e && npm ci && npx playwright install chromium && npm test

migrate: docker-check
	$(DOCKER_COMPOSE) exec backend ./rionexgate migrate

logs: docker-check
	$(DOCKER_COMPOSE) logs -f

clean: docker-check
	$(DOCKER_COMPOSE) down -v
	rm -rf data/
