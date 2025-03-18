.PHONY:  run/server run/migrations-up run/db ensure-db-ready run/client

run/server: ensure-db-ready run/migrations-up
	go run cmd/server/main.go

ensure-db-ready:
	@echo "Ensuring database is ready..."
	@docker compose up -d
	@./scripts/wait-for-db.sh

run/db:
	@echo "Starting database..."
	@docker compose up -d

run/migrations-up:
	@GOOSE_DRIVER=postgres \
	GOOSE_DBSTRING="host=localhost port=5432 user=gophkeeper password=gophkeeper dbname=gophkeeper sslmode=disable" \
	GOOSE_MIGRATION_DIR=./migrations \
	goose -dir ./migrations up

run/client:
	@echo "Starting client..."
	@go run cmd/client/main.go