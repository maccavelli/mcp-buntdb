GOPATH_BIN     := $(shell go env GOPATH)/bin
GOLANGCI_LINT  ?= $(GOPATH_BIN)/golangci-lint

.PHONY: all help test fmt vet lint tidy

all: help

test: ## Runs all tests
	go test ./...

fmt: ## Formats all Go source files
	go fmt ./...

vet: ## Runs go vet
	go vet ./...

lint: ## Runs golangci-lint
	@if [ ! -x "$(GOLANGCI_LINT)" ]; then \
		echo "golangci-lint not found at $(GOLANGCI_LINT)"; \
		echo "Install: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
	$(GOLANGCI_LINT) run -c .golangci.yml ./...

tidy: ## Runs go mod tidy
	go mod tidy

help: ## Displays this help message
	@echo "Usage: make [target]"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'
