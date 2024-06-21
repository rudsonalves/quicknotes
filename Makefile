PASSWORD = $(shell echo $${DB_PASSWORD:-default_password})
POSTGRESQL_URL = postgres://postgres:${PASSWORD}@localhost:5432/postgres?sslmode=disable

run:
	go run cmd/http/*.go

exp:
	go run cmd/exp/exp.go

db:
	docker compose up

migrate-up:
	@migrate -database ${POSTGRESQL_URL} -path db/migrations up

migrate-down:
	@migrate -database ${POSTGRESQL_URL} -path db/migrations down

.PHONY: run exp db migrate-up