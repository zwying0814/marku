---
title: Docker Compose
description: 使用 Docker Compose 启动 API、Admin，以及可选的 PostgreSQL。
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
MARKU_CORS_ALLOWED_ORIGINS=http://localhost:36581
VITE_API_BASE_URL=http://localhost:36580/api/v1
```

`VITE_API_BASE_URL` 会在 Admin 构建时写入静态产物：

- 浏览器能访问 API 的完整地址。
- 生产环境必须是 HTTPS API 域名。

修改该地址后需要重新构建 Admin 镜像。

## 2. 启动

```bash
docker compose up -d --build
```

默认端口：

| 服务 | 地址 |
| --- | --- |
| API | `http://localhost:36580` |
| Admin | `http://localhost:36581` |
| 健康检查 | `http://localhost:36580/healthz` |

## 3. 容器结构

```text
api
  ├── Go 多阶段构建
  ├── 非 root 用户运行
  ├── SQLite 数据挂载到 /data
  └── 容器内端口 36580

admin
  ├── Vite 构建静态文件
  ├── nginx-unprivileged
  └── 容器内端口 36581
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
docker compose logs -f admin
curl http://localhost:36580/readyz
```

## 6. 升级

```bash
git pull
docker compose build
docker compose up -d
```

升级前先备份数据库。若生产环境关闭了自动迁移，需要先执行迁移任务。

## 7. 文档站

Rspress 文档站不包含在 Compose 中。构建静态产物：

```bash
pnpm --filter @marku/docs build
```

产物位于 `apps/docs/doc_build`，可部署到任意静态服务器。

## 8. 数据卷

默认 SQLite volume：

```text
marku_marku-data
```

PostgreSQL volume：

```text
marku_marku-postgres
```

备份与恢复见 [备份与恢复](/deploy/backup)。
