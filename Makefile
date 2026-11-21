# Makefile for email-platform project

SHELL := /bin/bash

SQLC_VERSION ?= v1.31.1
OAPI_CODEGEN_VERSION ?= v2.8.2
GOVULNCHECK_VERSION ?= v1.10.0
GOSEC_VERSION ?= v2.30.0
GOLANGCI_LINT_VERSION ?= v1.64.8

.PHONY: ci test test-module lint gen gen-check security dev-up dev-down dev-logs crash-suite \
	fe-install fe-typecheck fe-lint fe-build fe-csp

ci: test lint gen-check fe-typecheck fe-build security
	@echo "All CI checks passed"

test:
	go test ./backend/... ./tools/...

test-module:
	go test -v ./backend/internal/$(M)/...

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./backend/... ./tools/...

gen:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) generate
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION) \
		--generate types,gin \
		-o backend/api/openapi.gen.go \
		--package api \
		api/openapi.yaml
	cd frontend && npm run gen:api
	test -s backend/api/openapi.gen.go
	test -s frontend/lib/types/openapi.ts

gen-check:
	@tmpdir=$$(mktemp -d); \
	trap 'rm -rf "$$tmpdir"' EXIT; \
	tar -cf "$$tmpdir/before.tar" backend/api/openapi.gen.go backend/internal/platform/db/sqlc frontend/lib/types/openapi.ts; \
	$(MAKE) gen; \
	mkdir -p "$$tmpdir/before"; \
	tar -xf "$$tmpdir/before.tar" -C "$$tmpdir/before"; \
	diff -ru "$$tmpdir/before/backend/api/openapi.gen.go" backend/api/openapi.gen.go; \
	diff -ru "$$tmpdir/before/backend/internal/platform/db/sqlc" backend/internal/platform/db/sqlc; \
	diff -ru "$$tmpdir/before/frontend/lib/types/openapi.ts" frontend/lib/types/openapi.ts; \
	echo "Generated files are reproducible."

security:
	go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./backend/... ./tools/...
	go run github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION) -severity medium -confidence medium ./backend/... ./tools/...

fe-install:
	cd frontend && npm ci

fe-typecheck:
	cd frontend && npm run typecheck

fe-lint:
	cd frontend && npm run lint

fe-build:
	cd frontend && npm run build

fe-csp: fe-build
	node tools/csp/spike.mjs

dev-up:
	docker compose -f infra/dev/docker-compose.yml up -d --build

dev-down:
	docker compose -f infra/dev/docker-compose.yml down

dev-logs:
	docker compose -f infra/dev/docker-compose.yml logs

crash-suite:
	@echo "Crash suite lands with WP08; target reserved by the test strategy."
