.DEFAULT_GOAL := help

SHELL := /usr/bin/env bash

.PHONY: help check test vet generate build dev clean check-clean

## help: Show available commands
help:
	@echo "SiteCraft development commands:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed -e 's/## //'

## check: Fast all-in-one verification for AI agent (generate -> vet -> test)
check: generate vet test
	@echo "==> All checks passed successfully."

## generate: Re-generate Templ components and compile Tailwind CSS v4
generate:
	templ generate
	tailwindcss -i ./ui/assets/input.css -o ./ui/assets/styles.css --minify

## test: Run all Go tests
test:
	go test ./...

## vet: Run go vet static analysis
vet:
	go vet ./...

## build: Full production binary build (generate + compile to bin/server)
build: generate
	go build -ldflags="-s -w" -o bin/server main.go

## dev: Run local development server with hot reload via air
dev:
	air

## check-clean: Verify that working tree has no uncommitted generated diffs
check-clean: check
	git diff --exit-code ui/assets/styles.css ui/*_templ.go

## clean: Remove build output and temporary files
clean:
	rm -rf bin/ tmp/
