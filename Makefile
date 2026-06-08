.PHONY: build run dev clean migrate

build:
	docker-compose build

up:
	docker-compose up -d

down:
	docker-compose down

dev:
	docker-compose up --build

migrate:
	docker-compose exec backend ./proxy-mgr migrate

logs:
	docker-compose logs -f

clean:
	docker-compose down -v
	rm -rf data/