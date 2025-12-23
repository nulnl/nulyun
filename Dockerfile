# syntax=docker/dockerfile:1.4

ARG TARGETARCH

# Build frontend using Node (with layer caching)
FROM node:20-alpine AS frontend-builder
WORKDIR /src/www
RUN apk add --no-cache git && \
    corepack enable && \
    corepack prepare pnpm@latest --activate

# Copy package files first for better layer caching
COPY www/package.json www/pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile

# Copy source and build
COPY www .
RUN pnpm run build

# Build backend with frontend artifacts
FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source and frontend dist
COPY . .
COPY --from=frontend-builder /src/www/dist ./www/dist

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -ldflags='-s -w -extldflags "-static"' -trimpath -o /out/nulyun .

# Final minimal image
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /out/nulyun /nulyun

EXPOSE 8080
ENTRYPOINT ["/nulyun"]
