ifneq (,$(wildcard .env))
    include .env
    export
endif

.PHONY: up down sqlc logs

up:
	docker compose up -d

migrate-up:
	docker compose exec app ./cmd/migrate up

down:
	docker compose down

sqlc:
	sqlc generate

logs:
	docker compose logs app -f