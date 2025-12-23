GOBIN?=bin
BINARY=$(GOBIN)/nulyun
CMD=.
FRONTEND_DIR=www
WWW_DIR=www/dist
GHCR_OWNER?=$(shell echo $${GITHUB_REPOSITORY_OWNER:-})
TAG?=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")

.PHONY: help all build build-backend build-frontend test fmt vet clean install-deps release

# Default target: build frontend and backend
all: build

help:
	@echo "Makefile targets:"
	@echo ""
	@echo "  all (default)       Build frontend and backend"
	@echo "  build               Build both frontend and backend"
	@echo "  build-backend       Build backend binary only"
	@echo "  build-frontend      Build frontend only"
	@echo "  test                Run Go tests"
	@echo "  fmt                 Format Go code"
	@echo "  vet                 Run go vet"
	@echo "  install-deps        Download Go dependencies"
	@echo "  release             Run goreleaser"
	@echo "  clean               Remove build artifacts"

# Build both frontend and backend
build: build-frontend build-backend

# Backend build
build-backend:
	@mkdir -p $(GOBIN)
	@echo "Building backend..."
	@go build -v -o $(BINARY) $(CMD)

# Backend test
test:
	@echo "Running tests..."
	@go test ./... -v

# Backend format
fmt:
	@echo "Formatting Go code..."
	@gofmt -w .

# Backend vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Backend dependencies
install-deps:
	@echo "Downloading Go dependencies..."
	@go mod download

# Frontend build (with dependency install)
build-frontend:
	@echo "Installing frontend dependencies..."
	@cd $(FRONTEND_DIR) && pnpm install --frozen-lockfile
	@echo "Building frontend..."
	@cd $(FRONTEND_DIR) && pnpm run build

# Release
release:
	@echo "Running goreleaser..."
	@goreleaser release

# Clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BINARY) $(WWW_DIR)
