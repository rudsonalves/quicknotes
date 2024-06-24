PASSWORD = $(shell echo $${DB_PASSWORD:-default_password})
POSTGRESQL_URL = postgres://postgres:${PASSWORD}@localhost:5432/postgres?sslmode=disable
MNAME ?= "no_migration"

run:
	go run cmd/http/*.go

exp:
	go run cmd/exp/exp.go

db_up:
	docker compose up

# db_down:
# 	docker compose down

db:
	docker exec -it quicknotes-db-1 bash -c "psql -h localhost -p 5432 -U postgres -d postgres"

migrate_up:
	@migrate -database ${POSTGRESQL_URL} -path db/migrations up

migrate_down:
	@migrate -database ${POSTGRESQL_URL} -path db/migrations down

migrate_create:
	@if [ "$(MNAME)" = "no_migration" ]; then \
		echo "Error: Migration name is required. Use MNAME=<name> make migration-create"; \
		exit 1; \
	fi
	@migrate create -ext sql -dir db/migrations -seq $(MNAME)

.PHONY: run exp db_up db migrate-up migrate-down