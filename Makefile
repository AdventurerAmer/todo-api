.PHONY: build
build:
	@go build -o ./bin/todo ./api

.PHONY: run
run: build
	@./bin/todo

TODO_MAIN_DB_DSN := postgres://${TODO_MAIN_DB_USERNAME}:${TODO_MAIN_DB_PASSWORD}@${TODO_MAIN_DB_HOST}:${TODO_MAIN_DB_PORT}/${TODO_MAIN_DB_NAME}?sslmode=${TODO_MAIN_DB_SSL_MODE}

.PHONY: psql
psql:
	@psql ${TODO_MAIN_DB_DSN}

.PHONY: create_migration
create_migration:
	@migrate create -seq -ext=.sql -dir=./migrations ${name}

.PHONY: migrate_up
migrate_up:
	@migrate -path=./migrations -database=${TODO_MAIN_DB_DSN} up

.PHONY: migrate_force
migrate_force:
	@migrate -path=./migrations -database=${TODO_MAIN_DB_DSN} force ${version}

.PHONY: migrate_one
migrate_one:
	@migrate -path=./migrations -database=${TODO_MAIN_DB_DSN} down 1

.PHONY: migrate_down
migrate_down:
	@migrate -path=./migrations -database=${TODO_MAIN_DB_DSN} down
