---
title: Docker Compose
description: 使用 Docker Compose 启动内置 Admin 与组件脚本的单二进制 Marku。
outline: [2, 3]
---

# Docker Compose

## 1. 准备环境变量

```bash
cp .env.example .env
```

至少填写：

```dotenv
MARKU_JWT_SECRET=至少32位随机字符串
MARKU_ADMIN_EMAIL=admin@example.com
MARKU_ADMIN_PASSWORD=强密码
MARKU_IP_HASH_SALT=随机字符串
```

Admin 与 API 同源，默认使用相对路径 `/api/v1`，不需要额外配置
`VITE_API_BASE_URL`。

## 2. 启动

```bash
docker compose up -d --build
```

## GHCR 镜像

推送 `v*` tag 后，GitHub Actions 会构建 `linux/amd64` 与 `linux/arm64`
多架构镜像并发布到：

```text
ghcr.io/zwying0814/marku:<version>
ghcr.io/zwying0814/marku:latest
```

服务器可以直接拉取，不需要在本地构建：

```bash
docker pull ghcr.io/zwying0814/marku:0.3.1
```

GitHub Packages 默认是私有包。改成 Public 后可直接拉取；保持私有则需要在
服务器执行 `docker login ghcr.io`，使用具备 `read:packages` 的 PAT。

默认端口：

| 服务 | 地址 |
| --- | --- |
| Admin | `http://localhost:36580/` |
| Swagger UI | `http://localhost:36580/swagger` |
| 健康检查 | `http://localhost:36580/healthz` |

官网与文档站（`apps/web`）不再打进镜像，需要单独执行
`pnpm --filter @marku/web build` 后部署 `apps/web/doc_build`。

## 3. 容器结构

```text
web-build
  └── 构建 Admin（Vite）

runtime
  ├── Go 编译时嵌入 Admin 与组件脚本
  ├── 非 root 用户运行
  ├── SQLite 数据挂载到 /data
  └── 单端口 36580
```

## 4. PostgreSQL

在 `.env` 中启用：

```dotenv
MARKU_DB_DRIVER=postgres
MARKU_DB_DSN=host=postgres user=marku password=change-me dbname=marku port=5432 sslmode=disable
MARKU_POSTGRES_DB=marku
MARKU_POSTGRES_USER=marku
MARKU_POSTGRES_PASSWORD=change-me
```

启动 profile：

```bash
docker compose --profile postgres up -d --build
```

## 5. 查看状态

```bash
docker compose ps
docker compose logs -f api
curl http://localhost:36580/readyz
```

## 6. 升级

```bash
git pull
docker compose build
docker compose up -d
```

升级前先备份数据库。若生产环境关闭了自动迁移，需要先执行迁移任务。

## 7. 数据卷

默认 SQLite volume：

```text
marku_marku-data
```

PostgreSQL volume：

```text
marku_marku-postgres
```

备份与恢复见 [备份与恢复](/deploy/backup)。
