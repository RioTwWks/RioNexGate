.PHONY: dev run migrate test docker-dev

# Локальный backend (без Docker)
dev run:
	go run ./cmd/main.go

migrate:
	go run ./cmd/main.go migrate

test:
	go test ./...

# Полный стек через Docker Compose (из корня репозитория)
docker-dev:
	$(MAKE) -C .. dev
