---
title: 快速开始
description: 在本地启动 Marku API、Admin 与文档站，并完成第一次接入。
outline: [2, 3]
---

# 快速开始

Marku 由三个应用组成：

| 应用 | 默认地址 | 用途 |
| --- | --- | --- |
| API | `http://localhost:36580` | 公共接口与管理接口 |
| Admin | `http://localhost:36581` | 创建站点、审核评论、查看数据 |
| Docs | `http://localhost:36582` | 当前文档站 |

## 环境要求

- Node.js 24
- pnpm 12+
- Go 1.27+

## 1. 安装依赖

```bash
pnpm install
```

## 2. 配置 API

```bash
cp apps/api/.env.example apps/api/.env
```

至少填写以下变量：

```dotenv title="apps/api/.env"
MARKU_JWT_SECRET=至少32位的随机字符串
MARKU_ADMIN_EMAIL=admin@example.com
MARKU_ADMIN_PASSWORD=你的管理员密码
MARKU_IP_HASH_SALT=用于评论IP哈希的随机字符串
```

生成随机值：

```bash
openssl rand -base64 48
openssl rand -hex 32
```

开发环境允许留空 JWT Secret，但管理员登录和 Site Key 哈希能力需要完整配置。
生产环境要求 `MARKU_JWT_SECRET` 至少 32 个字符。

## 3. 启动全部服务

```bash
pnpm dev
```

该命令并行启动 API、Admin 和 Rspress 文档站。

健康检查：

```bash
curl http://localhost:36580/healthz
curl http://localhost:36580/readyz
```

## 4. 创建管理员与会话

首次启动时，API 会使用 `MARKU_ADMIN_EMAIL` 和 `MARKU_ADMIN_PASSWORD`
创建管理员。打开 `http://localhost:36581` 登录。

JWT 只保存在当前标签页的 `sessionStorage`，关闭标签页或退出登录后清除。

## 5. 创建第一个站点

在 Admin 中进入「站点 → 新建站点」：

1. 填写站点名称与 slug。
2. 配置主域名。
3. 填写允许来源，例如 `http://localhost:36583`。
4. 选择「先审后发」或「自动公开」。
5. 创建后立即复制完整 Site Key。

完整 Site Key 只在创建和轮换时展示一次。服务端只保存哈希与展示前缀。

## 6. 发出第一个公共请求

```bash
export MARKU_SITE_KEY="mk_your_site_key"

curl http://localhost:36580/api/v1/public/site \
  -H "X-Site-Key: $MARKU_SITE_KEY"
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "name": "My Blog",
    "slug": "my-blog",
    "domain": "blog.example.com",
    "commentModeration": "pending"
  }
}
```

## 7. 生成演示数据（可选）

```bash
pnpm seed
```

该命令会创建两个演示站点，并写入评论、浏览量与点赞。站点 Site Key 会输出在日志中。

## 下一步

- [把 Marku 接入现有博客](/guide/integration)
- [运行纯 HTML 完整示例](/guide/html-example)
- [阅读 API 参考](/api/overview)
