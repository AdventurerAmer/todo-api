.PHONY: build
build:
	@go build -o ./bin/todo ./api

.PHONY: run
run: build
	@./bin/todo

.PHONY: psql
psql:
	@psql ${DB_DSN}

.PHONY: create_migration
create_migration:
	@migrate create -seq -ext=.sql -dir=./migrations ${name}

.PHONY: migrate_up
migrate_up:
	@migrate -path=./migrations -database=${DB_DSN} up

.PHONY: migrate_force
migrate_force:
	@migrate -path=./migrations -database=${DB_DSN} force ${version}

.PHONY: migrate_one
migrate_one:
	@migrate -path=./migrations -database=${DB_DSN} down 1

.PHONY: migrate_down
migrate_down:
	@migrate -path=./migrations -database=${DB_DSN} down
