GOBIN?=bin
BINARY=$(GOBIN)/nulyun
CMD=.
FRONTEND_DIR=www
WWW_DIR=www/dist
GHCR_OWNER?=$(shell echo $${GITHUB_REPOSITORY_OWNER:-})
TAG?=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")

.PHONY: help all build build-backend build-frontend test fmt vet \
	install-deps frontend-install frontend-build frontend-fmt \
	dist release docker-build clean

# Default target: format, test, build everything
all: fmt vet test build

help:
	@echo "Makefile targets:"
	@echo ""
	@echo "Default targets:"
	@echo "  all                 Format, test and build everything (default)"
	@echo "  build               Build both frontend and backend"
	@echo ""
	@echo "Backend targets:"
	@echo "  build-backend       Build backend binary only (to $(BINARY))"
	@echo "  test                Run Go tests"
	@echo "  fmt                 Format Go code with gofmt"
	@echo "  vet                 Run go vet on Go code"
	@echo "  install-deps        Download Go module dependencies"
	@echo ""
	@echo "Frontend targets:"
	@echo "  build-frontend      Build frontend only"
	@echo "  frontend-install    Install frontend dependencies (pnpm)"
	@echo "  frontend-fmt        Format frontend code (prettier)"
	@echo ""
	@echo "Release targets:"
	@echo "  dist                Build frontend and backend for distribution"
	@echo "  release             Run goreleaser (requires goreleaser installed)"
	@echo "  docker-build        Build and push multi-arch image to GHCR"
	@echo ""
	@echo "Utility targets:"
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

# Frontend install
frontend-install:
	@echo "Installing frontend dependencies..."
	@cd $(FRONTEND_DIR) && pnpm install --frozen-lockfile

# Frontend build
build-frontend: frontend-install
	@echo "Building frontend..."
	@cd $(FRONTEND_DIR) && pnpm run build

# Frontend format (optional, if prettier is configured)
frontend-fmt:
	@echo "Formatting frontend code..."
	@cd $(FRONTEND_DIR) && pnpm run format || echo "No format script found, skipping"

# Distribution build
dist: build-frontend build-backend

# Release
release:
	@echo "Running goreleaser..."
	@goreleaser release

# Docker build
docker-build:
	@if [ -z "$(GHCR_OWNER)" ]; then \
		echo "Error: GHCR_OWNER not set. Export GHCR_OWNER=<owner> or set GITHUB_REPOSITORY_OWNER env var."; \
		exit 1; \
	fi
	@echo "Building multi-arch Docker image..."
	@docker buildx build --platform linux/amd64,linux/arm64 \
		-f Dockerfile \
		-t ghcr.io/$(GHCR_OWNER)/nulyun:$(TAG) \
		--push .

# Clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BINARY) $(WWW_DIR)
