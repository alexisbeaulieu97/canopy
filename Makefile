# Makefile for Canopy

# Binary name
BINARY := canopy

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOVET := $(GOCMD) vet
GOMOD := $(GOCMD) mod

# Version information (embedded via ldflags)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)

# Build directory
BUILD_DIR := ./build

# Source paths
CMD_PATH := ./cmd/canopy

.PHONY: all build build-release clean test test-race test-short lint vet check install deps version help

# Default target
all: build

## Build targets

# Build binary with version information
build:
	@mkdir -p $(BUILD_DIR)
	@echo "Building $(BINARY)..."
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) $(CMD_PATH)
	@echo "Built $(BUILD_DIR)/$(BINARY) ($(VERSION))"

# Build release binary with optimizations
build-release:
	@mkdir -p $(BUILD_DIR)
	@echo "Building release $(BINARY)..."
	CGO_ENABLED=0 $(GOBUILD) -ldflags "$(LDFLAGS) -s -w" -trimpath -o $(BUILD_DIR)/$(BINARY) $(CMD_PATH)
	@echo "Built $(BUILD_DIR)/$(BINARY) ($(VERSION))"

# Install to GOPATH/bin
install:
	@echo "Installing $(BINARY)..."
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(shell go env GOPATH)/bin/$(BINARY) $(CMD_PATH)
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY)"

## Test targets

# Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	$(GOTEST) -race -v ./...

# Run short tests only
test-short:
	@echo "Running short tests..."
	$(GOTEST) -short -v ./...

## Quality targets

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Run all quality checks
check: vet lint test

## Utility targets

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	@echo "Cleaned"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Show version info that would be embedded
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"

# Help
help:
	@echo "Canopy Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build         Build binary with version info"
	@echo "  make build-release Build optimized release binary"
	@echo "  make install       Install to GOPATH/bin"
	@echo "  make test          Run all tests"
	@echo "  make test-race     Run tests with race detector"
	@echo "  make lint          Run golangci-lint"
	@echo "  make vet           Run go vet"
	@echo "  make check         Run all quality checks"
	@echo "  make clean         Remove build artifacts"
	@echo "  make deps          Download and tidy dependencies"
	@echo "  make version       Show version info"
	@echo "  make help          Show this help"

