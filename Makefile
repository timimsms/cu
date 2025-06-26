.PHONY: build test lint fmt install clean help

# Variables
BINARY_NAME=cu
BINARY_PATH=./$(BINARY_NAME)
MAIN_PATH=./cmd/cu
VERSION=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse --short HEAD)
DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X github.com/tim/cu/internal/version.Version=$(VERSION) \
	-X github.com/tim/cu/internal/version.Commit=$(COMMIT) \
	-X github.com/tim/cu/internal/version.Date=$(DATE) \
	-X github.com/tim/cu/internal/version.BuiltBy=make"

# Default target
all: build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_PATH)"

## test: Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

## lint: Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: brew install golangci-lint"; \
		exit 1; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@go mod tidy

## install: Install the binary
install: build
	@echo "Installing $(BINARY_NAME)..."
	@go install $(LDFLAGS) $(MAIN_PATH)
	@echo "Installed to $$(go env GOPATH)/bin/$(BINARY_NAME)"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_PATH)
	@rm -rf dist/
	@echo "Clean complete"

## run: Build and run
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BINARY_PATH)

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

## coverage: Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## ci: Run CI checks locally (mirrors GitHub Actions)
ci:
	@echo "Running CI checks..."
	@./scripts/ci.sh

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'