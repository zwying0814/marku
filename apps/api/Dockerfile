FROM node:24-alpine AS web-build

RUN corepack enable && corepack prepare pnpm@12.23.0 --activate

WORKDIR /repo
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY apps/admin/package.json apps/admin/package.json
COPY packages/shared/package.json packages/shared/package.json
RUN pnpm install --frozen-lockfile --filter @marku/admin

COPY packages/shared packages/shared
COPY apps/admin apps/admin
RUN pnpm --filter @marku/admin build

FROM golang:1.27-alpine AS build

WORKDIR /src
RUN apk add --no-cache ca-certificates

COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

COPY apps/api ./
COPY --from=web-build /repo/apps/admin/dist ./internal/webui/admin-dist

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
