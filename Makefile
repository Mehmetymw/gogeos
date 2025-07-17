# GoGEOS Makefile

.PHONY: test test-verbose test-coverage test-race test-bench clean build examples lint

# Default target
all: test

# Run tests
test:
	go test ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Run tests with coverage report
test-coverage-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run tests with race detection
test-race:
	go test -race ./...

# Run benchmarks
test-bench:
	go test -bench=. ./...

# Run all tests (unit, integration, benchmarks)
test-all:
	go test -v -race -cover -bench=. ./...

# Lint the code
lint:
	golangci-lint run

# Format the code
fmt:
	go fmt ./...

# Build the library
build:
	go build ./...

# Build examples
examples:
	go build -o bin/basic_usage ./cmd/basic_usage
	go build -o bin/advanced_operations ./cmd/advanced_operations

# Run basic usage example
run-basic:
	go run ./cmd/basic_usage

# Run advanced operations example
run-advanced:
	go run ./cmd/advanced_operations

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	go mod tidy
	go mod download

# Verify dependencies
deps-verify:
	go mod verify

# Run go vet
vet:
	go vet ./...

# Run all quality checks
quality: fmt vet lint test-race test-coverage

# Check if GEOS is properly installed
check-geos:
	@echo "Checking GEOS installation..."
	@pkg-config --exists geos || (echo "GEOS not found. Please install GEOS development library." && exit 1)
	@pkg-config --cflags --libs geos
	@echo "GEOS installation OK"

# Help target
help:
	@echo "Available targets:"
	@echo "  test                Run tests"
	@echo "  test-verbose        Run tests with verbose output"
	@echo "  test-coverage       Run tests with coverage"
	@echo "  test-coverage-html  Generate HTML coverage report"
	@echo "  test-race          Run tests with race detection"
	@echo "  test-bench         Run benchmarks"
	@echo "  test-all           Run all tests with all options"
	@echo "  lint               Run linter"
	@echo "  fmt                Format code"
	@echo "  build              Build the library"
	@echo "  examples           Build example programs"
	@echo "  run-basic          Run basic usage example"
	@echo "  run-advanced       Run advanced operations example"
	@echo "  clean              Clean build artifacts"
	@echo "  deps               Install dependencies"
	@echo "  deps-verify        Verify dependencies"
	@echo "  vet                Run go vet"
	@echo "  quality            Run all quality checks"
	@echo "  check-geos         Check GEOS installation"
	@echo "  help               Show this help"