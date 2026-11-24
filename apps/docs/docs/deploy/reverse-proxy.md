---
title: 反向代理与 HTTPS
description: Nginx、可信代理、CORS、安全头和静态站点部署。
outline: [2, 3]
---

# 反向代理与 HTTPS

## API 域名

推荐独立域名：

```text
https://api.example.com
```

Nginx 示例：

```nginx
limit_req_zone $binary_remote_addr zone=marku_api:10m rate=20r/s;

server {
    listen 443 ssl http2;
    server_name api.example.com;

    ssl_certificate     /etc/letsencrypt/live/api.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.example.com/privkey.pem;

    client_max_body_size 1m;

    location / {
        limit_req zone=marku_api burst=40 nodelay;

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

Admin 域名：

```dotenv
MARKU_CORS_ALLOWED_ORIGINS=https://admin.example.com
```

公共 API 路径按设计允许跨域读取。站点写请求还会根据站点 `allowedOrigins`
校验来源，因此仍需在每个站点里配置：

```text
https://blog.example.com
```

## Admin 静态站点

```nginx
server {
    listen 443 ssl http2;
    server_name admin.example.com;

    root /var/www/marku-admin;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /assets/ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

Admin 建议额外增加：

- IP allowlist
- VPN
- HTTP Basic Auth
- SSO / Access Proxy

公共博客不需要访问 Admin 域名。

## 文档站

```nginx
server {
    listen 443 ssl http2;
    server_name docs.example.com;

    root /var/www/marku-docs;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

将 `apps/docs/doc_build` 的内容部署到 `root`。

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
