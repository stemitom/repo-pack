.PHONY: help build test clean install lint fmt vet coverage bench profile

# Variables
BINARY_NAME=repo-pack
GO=go
GOFLAGS=-v
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Default target
help:
	@echo "repo-pack - Makefile targets"
	@echo ""
	@echo "Build & Installation:"
	@echo "  make build          - Build the binary"
	@echo "  make install        - Install to $$GOPATH/bin"
	@echo "  make clean          - Remove built binary and artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  make test           - Run all tests"
	@echo "  make test-verbose   - Run tests with verbose output"
	@echo "  make coverage       - Generate coverage report"
	@echo "  make coverage-html  - Generate HTML coverage report"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint           - Run linter (golangci-lint)"
	@echo "  make fmt            - Format code with gofmt"
	@echo "  make vet            - Run go vet"
	@echo "  make fmt-check      - Check code formatting"
	@echo ""
	@echo "Benchmarks:"
	@echo "  make bench          - Run benchmarks"
	@echo "  make bench-verbose  - Run benchmarks with verbose output"
	@echo ""
	@echo "Development:"
	@echo "  make deps           - Download dependencies"
	@echo "  make tidy           - Clean up go.mod and go.sum"
	@echo "  make all            - Build, test, and lint"

# Build the binary
build:
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME)
	@echo "✓ Built $(BINARY_NAME)"

# Install the binary
install: build
	$(GO) install $(GOFLAGS)
	@echo "✓ Installed $(BINARY_NAME)"

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	$(GO) clean -cache -testcache
	@echo "✓ Cleaned artifacts"

# Run tests
test:
	$(GO) test ./... -race -timeout 30s

# Run tests with verbose output
test-verbose:
	$(GO) test ./... -race -timeout 30s -v

# Generate coverage report
coverage:
	$(GO) test ./... -race -cover -coverprofile=$(COVERAGE_FILE)
	$(GO) tool cover -func=$(COVERAGE_FILE)
	@echo "✓ Coverage report generated"

# Generate HTML coverage report
coverage-html: coverage
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "✓ HTML coverage report: $(COVERAGE_HTML)"

# Format code
fmt:
	$(GO) fmt ./...
	@echo "✓ Code formatted"

# Check code formatting
fmt-check:
	@if [ -n "$$($(GO) fmt ./...)" ]; then \
		echo "✗ Code is not formatted. Run 'make fmt'"; \
		exit 1; \
	fi
	@echo "✓ Code is properly formatted"

# Run go vet
vet:
	$(GO) vet ./...
	@echo "✓ Vet check passed"

# Run linter (requires golangci-lint)
lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not found. Install with:"; \
		echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin"; \
		exit 1; \
	}
	golangci-lint run ./...
	@echo "✓ Lint check passed"

# Download dependencies
deps:
	$(GO) mod download
	@echo "✓ Dependencies downloaded"

# Tidy up go.mod and go.sum
tidy:
	$(GO) mod tidy
	@echo "✓ Dependencies tidied"

# Run benchmarks
bench:
	$(GO) test ./... -bench=. -benchmem -run=^$

# Run benchmarks with verbose output
bench-verbose:
	$(GO) test ./... -bench=. -benchmem -benchtime=3s -v -run=^$

# Run all checks (build, test, lint, vet)
all: fmt vet test coverage lint
	@echo "✓ All checks passed"

# Development quick build (no testing)
quick:
	$(GO) build -o $(BINARY_NAME)
	@echo "✓ Quick build complete"

# Run the binary with example URL
run: build
	./$(BINARY_NAME) --help

# Run go mod verify
verify:
	$(GO) mod verify
	@echo "✓ Module verification passed"

# Update dependencies
update-deps:
	$(GO) get -u ./...
	$(GO) mod tidy
	@echo "✓ Dependencies updated"

# Docker image build (requires docker)
docker-build:
	docker build -t $(BINARY_NAME):latest .
	@echo "✓ Docker image built"

# Docker image run
docker-run: docker-build
	docker run --rm $(BINARY_NAME):latest --help
