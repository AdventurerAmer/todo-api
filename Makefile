.PHONY: build
build:
	@go build -o ./bin/todo ./api

.PHONY: run
run: build
	@./bin/todo

.PHONY: psql
psql:
	@psql ${DB_DSN}

.PHONY: migration
migration:
	@migrate create -seq -ext=.sql -dir=./migrations ${name}

.PHONY: up
up:
	@migrate -path=./migrations -database=${DB_DSN} up