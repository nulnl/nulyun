# syntax=docker/dockerfile:1.4

ARG TARGETARCH
FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
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
