FROM golang:1.27-alpine AS build

WORKDIR /src
RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/marku-api ./cmd/server

FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
    && addgroup -S marku \
    && adduser -S -G marku marku \
    && mkdir -p /data \
    && chown -R marku:marku /data

WORKDIR /app
COPY --from=build /out/marku-api /app/marku-api

USER marku
EXPOSE 36580

ENV MARKU_HTTP_ADDR=:36580 \
    MARKU_DB_DRIVER=sqlite \
    MARKU_DB_DSN=/data/marku.db

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:36580/healthz || exit 1

ENTRYPOINT ["/app/marku-api"]
