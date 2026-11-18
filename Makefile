# Makefile for email-platform project
.PHONY: ci test lint test-module gen dev-up dev-down dev-logs crash-suite

ci: test lint
	@echo "All CI checks passed"

test:
	go test ./...

test-module:
	go test -v ./$(M)

lint:
	golangci-lint run

gen:
	sqlc generate

dev-up:
	docker compose -f infra/dev/docker-compose.yml up -d --build

dev-down:
	docker compose -f infra/dev/docker-compose.yml down

dev-logs:
	docker compose -f infra/dev/docker-compose.yml logs

crash-suite:
	@echo "Not implemented"