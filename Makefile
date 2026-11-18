# Makefile for email-platform project
.PHONY: ci test lint test-module gen dev-up dev-down crash-suite

ci: test lint
	@echo "All CI checks passed"

test:
	go test ./...

test-module:
	go test -v ./$(M)

lint:
	golangci-lint run

gen:
	@echo "Sqlc generation skipped (empty queries directory)"

dev-up:
	@echo "Starting development environment (not implemented)"
dev-down:
	@echo "Stopping development environment (not implemented)"

crash-suite:
	@echo "Not implemented"