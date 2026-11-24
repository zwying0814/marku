---
title: API 总览
description: 基础地址、认证方式、响应结构、分页、限流与路径规则。
outline: [2, 3]
---

# API 总览

## 基础地址

本地开发：

```text
http://localhost:36580/api/v1
```

生产环境应由反向代理提供 HTTPS 域名，例如：

```text
https://api.example.com/api/v1
```

## 两套认证

### 公共 API

```http
X-Site-Key: mk_your_site_key
```

适用于博客前端、构建脚本和服务端渲染。

### 管理 API

```http
Authorization: Bearer <admin-jwt>
```

JWT 由管理员登录接口签发，默认有效期 8 小时。

## 统一响应

成功与业务错误都使用统一 envelope：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

错误示例：

```json
{
  "code": 40100,
  "message": "invalid site key",
  "data": null
}
```

调用方应同时检查：

1. HTTP 状态码。
2. `code` 是否为 `0`。

## 主要接口

### 公共

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `GET` | `/public/site` | 当前站点公开配置 |
| `GET` | `/public/stats` | 批量页面互动统计 |
| `GET` | `/public/comments` | 已通过评论树 |
| `POST` | `/public/comments` | 提交评论 |
| `GET` | `/public/comments/count` | 已通过评论数 |
| `POST` | `/public/views` | 记录浏览量 |
| `GET` | `/public/views` | 批量读取浏览量 |
| `GET` | `/public/likes` | 读取点赞状态 |
| `PUT` | `/public/likes` | 点赞 |
| `DELETE` | `/public/likes` | 取消点赞 |
| `POST` | `/public/likes` | 切换点赞（兼容接口） |

### 管理

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `POST` | `/admin/auth/login` | 管理员登录 |
| `GET` | `/admin/auth/me` | 当前管理员 |
| `POST` | `/admin/auth/change-password` | 修改密码 |
| `GET` | `/admin/overview` | 全局概览 |
| `GET/POST` | `/admin/sites` | 站点列表 / 创建 |
| `GET/PATCH/DELETE` | `/admin/sites/:id` | 站点详情 / 更新 / 删除 |
| `POST` | `/admin/sites/:id/rotate-key` | 轮换 Site Key |
| `DELETE` | `/admin/sites/:id/key` | 吊销 Site Key |
| `GET` | `/admin/sites/:id/comments` | 评论列表 |
| `GET` | `/admin/sites/:id/comments/stats` | 评论趋势 |
| `POST` | `/admin/sites/:id/comments/bulk` | 批量评论操作 |
| `GET` | `/admin/sites/:id/stats` | 浏览量统计 |
| `GET` | `/admin/sites/:id/likes` | 点赞统计 |
| `PATCH/DELETE` | `/admin/comments/:id` | 审核 / 删除评论 |
| `POST` | `/admin/comments/:id/reply` | 管理员回复 |

## 分页

列表接口通用参数：

| 参数 | 默认 | 限制 |
| --- | --- | --- |
| `page` | `1` | 最小 1 |
| `pageSize` | `20` | 1–100 |

响应：

```json
{
  "items": [],
  "total": 0,
  "page": 1,
  "pageSize": 20
}
```

## 路径规范化

所有 `path` 会按以下规则规范化：

- 去掉 `?query` 与 `#hash`
- 补充开头的 `/`
- 去掉末尾 `/`（根路径除外）
- 最长 512 字符
- 拒绝换行与制表符

## 限流

API 默认按客户端 IP 进行令牌桶限流：

```dotenv
MARKU_RATE_LIMIT_ENABLED=true
MARKU_RATE_LIMIT_RPS=20
MARKU_RATE_LIMIT_BURST=40
```

响应头：

```http
X-RateLimit-Limit: 40
X-RateLimit-Remaining: 0
Retry-After: 1
```

限流状态保存在单个 API 进程内。多实例部署时，每个实例独立计数。

## Swagger

开发环境可直接查看交互式 OpenAPI：

```text
http://localhost:36580/swagger
http://localhost:36580/openapi.yaml
```

## 下一步

- [公共接口](/api/public)
- [管理接口](/api/admin)
- [错误与限流](/api/errors)
