# syntax=docker/dockerfile:1.4

ARG TARGETARCH

# Build frontend using Node
FROM node:20-alpine AS frontend-builder
WORKDIR /src/www
RUN apk add --no-cache git
COPY www/package.json www/pnpm-lock.yaml* ./
COPY www/pnpm-workspace.yaml* ./
RUN corepack enable && corepack prepare pnpm@latest --activate
RUN pnpm install --frozen-lockfile || pnpm install
COPY www .
RUN pnpm run build

# Build backend and include frontend build artifacts
FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

# copy source and frontend dist
COPY . .
COPY --from=frontend-builder /src/www/dist ./www/dist

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -ldflags='-s -w' -o /out/nulyun .

FROM alpine:3.18
ARG TARGETARCH

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/nulyun /usr/local/bin/nulyun

RUN addgroup -S nulyun && adduser -S -G nulyun nulyun || true
RUN chown nulyun:nulyun /usr/local/bin/nulyun || true

USER nulyun

WORKDIR /home/nulyun

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/nulyun"]
CMD ["--help"]
