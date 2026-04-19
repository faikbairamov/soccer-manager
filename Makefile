include .env
export

BINARY=soccer-manager
MAIN=./cmd/api/main.go
MIGRATE=migrate -path ./internal/db/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)"

.PHONY: run build tidy docker-up docker-down migrate-up migrate-down migrate-create migrate-reset sqlc-gen mock-gen swagger-gen lint

## Run the application
run:
	set -a && . ./.env && go run $(MAIN)

## Build binary
build:
	go build -o bin/$(BINARY) $(MAIN)

## Tidy dependencies
tidy:
	go mod tidy

## Start docker services
docker-up:
	docker compose up -d

## Stop docker services
docker-down:
	docker compose down

## Run all migrations up
migrate-up:
	$(MIGRATE) up

## Rollback last migration
migrate-down:
	$(MIGRATE) down 1

## Drop and recreate the database, then run all migrations
migrate-reset:
	docker exec -it soccer-manager-postgres-1 psql -U $(DB_USER) -d postgres -c "DROP DATABASE IF EXISTS $(DB_NAME);"
	docker exec -it soccer-manager-postgres-1 psql -U $(DB_USER) -d postgres -c "CREATE DATABASE $(DB_NAME);"
	$(MIGRATE) up

## Create a new migration: make migrate-create name=create_users
migrate-create:
	migrate create -ext sql -dir ./internal/db/migrations -seq $(name)

## Generate sqlc code
sqlc-gen:
	cd internal/db && sqlc generate

## Generate mocks for repository interface
mock-gen:
	mockgen -source=internal/repository/querier.go -destination=internal/repository/mock/mock_querier.go -package=mock

## Generate swagger docs
swagger-gen:
	swag init -g $(MAIN) -o ./docs

## Run linter
lint:
	golangci-lint run ./...
