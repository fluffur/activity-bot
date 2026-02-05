ifneq (,$(wildcard .env))
    include .env
    export
endif

.PHONY: up down sqlc logs

up:
	docker compose up -d

up-prod:
	docker compose -f docker-compose.prod.yml up -d --build


migrate-up:
	docker compose exec app ./cmd/migrate up

down:
	docker compose down


down-prod:
	docker compose -f docker-compose.prod.yml down


sqlc:
	sqlc generate

logs:
	docker compose logs app -f


logs-prod:
	docker compose logs -f docker-compose.prod.yml app -f