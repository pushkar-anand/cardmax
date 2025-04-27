.DEFAULT_GOAL := build
.PHONY: dev new_migration build gen

dev:
	go tool air

new_migration: ## Create a new migration file. Usage: make new_migration name=<migration_name>
	 migrate create -dir=internal/db/migrations/ -seq -ext sql $(name)

build: gen
	go build

gen:
	go generate ./...