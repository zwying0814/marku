---
title: 环境变量
description: API 生产配置、数据库、CORS、评论、统计和日志。
outline: [2, 3]
---

# 环境变量

## 生产必填

```dotenv
MARKU_ENV=production
MARKU_JWT_SECRET=至少32位随机字符串
MARKU_IP_HASH_SALT=用于评论IP哈希的随机字符串
```

生产环境规则：

- `MARKU_JWT_SECRET` 至少 32 字符。
- `MARKU_IP_HASH_SALT` 生产必填。
- 首次启动会创建默认管理员 `admin` / `123456`，登录后请立即修改。
- 管理员账号与密码不通过环境变量配置。

## HTTP

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| `MARKU_HTTP_ADDR` | `:36580` | API 监听地址 |
| `MARKU_HTTP_MAX_BODY_BYTES` | `1048576` | 最大请求体 |
| `MARKU_HTTP_READ_TIMEOUT` | `15s` | 读取超时 |
| `MARKU_HTTP_WRITE_TIMEOUT` | `30s` | 写入超时 |
| `MARKU_HTTP_IDLE_TIMEOUT` | `60s` | 空闲超时 |
| `MARKU_HTTP_SHUTDOWN_TIMEOUT` | `10s` | 优雅退出时间 |
| `MARKU_HTTP_TRUSTED_PROXIES` | 空 | 可信反向代理 IP/CIDR，禁止 `*` |

## 数据库

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| `MARKU_DB_DRIVER` | `sqlite` | `sqlite` / `postgres` / `mysql` |
| `MARKU_DB_DSN` | `./data/marku.db` | 数据库连接串 |
| `MARKU_DB_AUTO_MIGRATE` | 开发 `true`，生产 `false` | 启动时自动迁移 |
| `MARKU_DB_MAX_OPEN_CONNS` | `25` | SQLite 固定为 1 |
| `MARKU_DB_MAX_IDLE_CONNS` | `10` | SQLite 固定为 1 |
| `MARKU_DB_CONN_MAX_LIFETIME` | `30m` | 连接寿命 |

### SQLite

```dotenv
MARKU_DB_DRIVER=sqlite
MARKU_DB_DSN=/data/marku.db
```

### PostgreSQL

```dotenv
MARKU_DB_DRIVER=postgres
MARKU_DB_DSN=host=postgres user=marku password=change-me dbname=marku port=5432 sslmode=disable
```

### MySQL

```dotenv
MARKU_DB_DRIVER=mysql
MARKU_DB_DSN=marku:change-me@tcp(mysql:3306)/marku?charset=utf8mb4&parseTime=True&loc=UTC
```

Compose 默认设置 `MARKU_DB_AUTO_MIGRATE=true`。如果生产环境关闭自动迁移，
需要在发布前从源码仓库使用同一数据库连接串执行：

```bash
pnpm migrate
```

## 管理端认证

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| `MARKU_JWT_SECRET` | 空 | JWT 签名密钥 |
| `MARKU_JWT_TTL` | `8h` | 登录有效期 |

修改 JWT Secret 会让已有管理员会话立即失效。

## CORS 与站点来源

```dotenv
MARKU_CORS_ALLOWED_ORIGINS=
```

注意：

- 官方二进制中 Admin 与 API 同源，不需要额外 CORS 配置。
- 只有当管理端从其他域名访问管理 API 时，才需要加入全局 CORS。
- 公共路径按设计允许跨域读取，仍必须提供有效 Site Key。
- 公共写请求会进一步校验站点自己的 `allowedOrigins`。
- 生产环境应为每个站点配置正式博客 origin，而不是留空。

## 限流

```dotenv
MARKU_RATE_LIMIT_ENABLED=true
MARKU_RATE_LIMIT_RPS=20
MARKU_RATE_LIMIT_BURST=40
```

限流按客户端 IP 计数，默认存储在进程内存中。多实例部署时每个实例独立计数；
如需全局限流，应在反向代理或网关层实现。

## 评论

```dotenv
MARKU_COMMENT_RATE_LIMIT_WINDOW=10m
MARKU_COMMENT_RATE_LIMIT_MAX=5
MARKU_COMMENT_MIN_SUBMIT_DURATION=2s
MARKU_COMMENT_MAX_LENGTH=2000
MARKU_COMMENT_MAX_LINKS=3
MARKU_COMMENT_MAX_DEPTH=3
MARKU_COMMENT_DUPLICATE_WINDOW=10m
```

## 浏览量

```dotenv
MARKU_VIEW_DEDUPE_BY_VISITOR_DAY=false
MARKU_VIEW_FILTER_BOTS=true
MARKU_VIEW_MAX_BATCH_PATHS=50
```

`MARKU_VIEW_DEDUPE_BY_VISITOR_DAY=true` 且请求包含 `visitorId` 时，独立访客才会
按访客与日期去重累计。`views` 始终按上报次数累加。

## 日志

```dotenv
MARKU_LOG_LEVEL=info
MARKU_LOG_FORMAT=json
```

生产环境推荐：

- `MARKU_LOG_LEVEL=info`
- `MARKU_LOG_FORMAT=json`

每个请求都有请求 ID，可通过 `X-Request-ID` 关联日志与客户端问题。

## 配置检查

API 启动时会校验关键配置。配置错误会直接退出并输出原因。检查顺序建议：

1. JWT Secret 长度。
2. 默认管理员密码是否已在首次登录后修改。
3. 数据库驱动与 DSN。
4. CORS origin 是否包含协议和端口。
5. 可信代理是否填写了具体 IP，而不是 `*`。
