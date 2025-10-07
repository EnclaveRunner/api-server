.PHONY: test test-verbose test-coverage test-race bench clean fmt lint

# Default target
all: test

# Run tests
test:
	go test

# Run tests with verbose output
test-verbose:
	go test -v

# Run tests with coverage
test-coverage:
	go test -cover

# Run tests with race detection
test-race:
	go test -race

# Run benchmarks
bench:
	go test -bench=.

test-all: test-verbose test-coverage test-race bench

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint to be installed)
lint:
	golangci-lint run

# Clean test cache
clean:
	go clean -testcache

# Show help
help:
	@echo "Available targets:"
	@echo "  test          - Run basic tests"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  test-race     - Run tests with race detection"
	@echo "  bench         - Run benchmarks"
	@echo "  test-all      - Run all tests"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  clean         - Clean test cache"
	@echo "  help          - Show this help"
