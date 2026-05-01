# Multi-stage Dockerfile for Haribon
FROM golang:1.23.4-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo dev) && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
      -ldflags="-s -w -X main.version=${VERSION}" \
      -o /out/haribon \
      ./cli


FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    addgroup -S haribon && \
    adduser -S -G haribon haribon

WORKDIR /app

RUN mkdir -p /etc/haribon

COPY --from=builder /out/haribon /usr/local/bin/haribon
COPY haribon-config.yml /etc/haribon/haribon-config.yml

RUN chown -R haribon:haribon /etc/haribon /app

USER haribon

EXPOSE 4444

# default env (can be overridden at runtime)
ENV HARIBON_CONFIG=/etc/haribon/haribon-config.yml
ENV HARIBON_PORT=4444

ENTRYPOINT ["sh", "-c", "haribon start --config ${HARIBON_CONFIG}"]