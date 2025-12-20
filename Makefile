GOBIN?=bin
BINARY=$(GOBIN)/nulyun
CMD=.
FRONTEND_DIR=www
WWW_DIR=www/dist
GHCR_OWNER?=$(shell echo $${GITHUB_REPOSITORY_OWNER:-})
TAG?=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")

.PHONY: help all build test fmt vet install-deps frontend-install frontend-build dist release docker-build clean

help:
	@echo "Makefile targets:"
	@echo "  docs            Project documentation in docs/"
	@echo "  build           Build backend binary (to $(BINARY))"
	@echo "  test            Run Go tests"
	@echo "  fmt             Run gofmt -w ."
	@echo "  vet             Run go vet ./..."
	@echo "  install-deps    Download Go module dependencies"
	@echo "  frontend-install  Install frontend deps (pnpm)"
	@echo "  frontend-build  Build frontend and copy to $(WWW_DIR)"
	@echo "  dist            Build backend and frontend"
	@echo "  release         Run goreleaser (requires goreleaser installed)"
	@echo "  docker-build    Build and push multi-arch image to GHCR (requires docker/buildx)"
	@echo "  clean           Remove build artifacts"

all: build

build:
	@mkdir -p $(GOBIN)
	go build -v -o $(BINARY) $(CMD)

test:
	go test ./... -v

fmt:
	gofmt -w .

vet:
	go vet ./...

install-deps:
	go mod download

frontend-install:
	cd $(FRONTEND_DIR) && pnpm install --frozen-lockfile
#	cd $(FRONTEND_DIR) && pnpm install && npm i --package-lock-only

frontend-build:
	cd $(FRONTEND_DIR) && pnpm run build

dist: frontend-build build

release:
	goreleaser release

docker-build:
	if [ -z "$(GHCR_OWNER)" ]; then echo "GHCR_OWNER not set. Export GHCR_OWNER=<owner> or set GITHUB_REPOSITORY_OWNER env var."; exit 1; fi
	docker buildx build --platform linux/amd64,linux/arm64 -f Dockerfile -t ghcr.io/$(GHCR_OWNER)/nulyun:$(TAG) --push .

clean:
	rm -rf $(BINARY) $(WWW_DIR)
