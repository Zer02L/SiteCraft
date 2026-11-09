.DEFAULT_GOAL := help

SHELL := /usr/bin/env bash
export PATH := $(PWD)/.devenv/state/go/bin:$(PWD)/.devenv/profile/bin:$(PATH)

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

## feature: Start a fresh feature branch from latest main (e.g. make feature name=cart)
feature:
	@if [ -z "$(name)" ]; then echo "Error: specify branch name, e.g. make feature name=my-task"; exit 1; fi
	git checkout main
	git pull origin main
	git checkout -b $(if $(findstring /,$(name)),$(name),feat/$(name))

## done: Push current branch, create PR, merge to main, and clean up branches
done:
	@BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
	if [ "$$BRANCH" = "main" ]; then echo "Error: currently on main branch"; exit 1; fi; \
	echo "==> Pushing branch $$BRANCH..."; \
	git push -u origin $$BRANCH; \
	echo "==> Creating and merging Pull Request..."; \
	gh pr create --fill; \
	gh pr merge --merge --delete-branch; \
	echo "==> Switching back to main..."; \
	git checkout main; \
	git pull origin main; \
	echo "==> Deleting local branch $$BRANCH..."; \
	git branch -d $$BRANCH; \
	echo "==> Done! Back on up-to-date main."

## sync: Update local main and prune dead local branches
sync:
	git checkout main
	git pull origin main
	git fetch --prune
	@DEAD=$$(git branch -vv | grep ': gone]' | awk '{print $$1}'); \
	if [ -n "$$DEAD" ]; then \
		echo "Cleaning up deleted local branches: $$DEAD"; \
		echo "$$DEAD" | xargs git branch -D; \
	else \
		echo "All local branches are up to date."; \
	fi

