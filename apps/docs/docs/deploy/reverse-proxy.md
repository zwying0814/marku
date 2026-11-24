---
title: 反向代理与 HTTPS
description: Nginx、可信代理、CORS 与单域名单二进制部署。
outline: [2, 3]
---

# 反向代理与 HTTPS

## 单域名部署

官方二进制同时提供 Admin、Docs、Swagger 与 API，推荐使用一个管理域名：

```text
https://marku.example.com
```

Nginx 示例：

```nginx
server {
    listen 443 ssl http2;
    server_name marku.example.com;

    ssl_certificate     /etc/letsencrypt/live/marku.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/marku.example.com/privkey.pem;

    client_max_body_size 1m;

    location / {
        proxy_pass http://127.0.0.1:36580;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 30s;
    }
}
```

路由由 Marku 自身处理：

```text
/          Admin
/docs/     Rspress
/swagger    Swagger UI
/api/v1/*  API
```

## 可信代理

如果 API 运行在反向代理后面，必须设置：

```dotenv
MARKU_HTTP_TRUSTED_PROXIES=127.0.0.1
```

否则 `ClientIP()` 可能把所有请求识别为代理 IP，造成：

- 全局限流把所有人算作同一个客户端。
- 评论 IP 限流与去重失真。

Docker 网络中可以填写具体网关或子网，但不要使用 `*`。

## CORS

Admin 与 API 同源，不需要为 Admin 配置 CORS。静态博客从自己的域名调用公共
API 时，公共路径按设计允许跨域读取；站点写请求仍根据每个站点的
`allowedOrigins` 校验来源：

```text
https://blog.example.com
```

## Admin 访问控制

内置 Admin 位于站点根路径。建议在反向代理或网关层增加：

- IP allowlist
- VPN
- HTTP Basic Auth
- SSO / Access Proxy

如果无法保护整个域名，也可以继续使用 `apps/admin/Dockerfile` 或
`apps/admin/dist` 单独部署 Admin，并通过 `VITE_API_BASE_URL` 指向 API。

## HTTPS 要求

- 博客若使用 HTTPS，API 也必须使用 HTTPS，否则会被混合内容策略拦截。
- Cookie、sessionStorage 与 Site Key 都不应通过明文 HTTP 传输。
- 建议开启 HSTS、TLS 1.2+ 和自动证书续期。

## 健康检查

反向代理健康检查应使用：

```text
GET /healthz
GET /readyz
```

`/readyz` 会检查数据库连接，不应被公开缓存。
