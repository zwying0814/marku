---
title: 管理接口
description: 管理员认证、站点、Site Key、统计与评论管理。
outline: [2, 3]
---

# 管理接口

除登录外，所有管理接口都需要：

```http
Authorization: Bearer <admin-jwt>
```

## 认证

### 登录

```http
POST /api/v1/admin/auth/login
Content-Type: application/json
```

```json
{
  "email": "admin@example.com",
  "password": "your-password"
}
```

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "token": "eyJ...",
    "expiresAt": "2026-11-24T16:00:00Z",
    "admin": {
      "id": 1,
      "email": "admin@example.com",
      "nickname": "Admin",
      "status": "active",
      "lastLoginAt": "2026-11-24T08:00:00Z",
      "createdAt": "2026-11-01T00:00:00Z",
      "updatedAt": "2026-11-24T08:00:00Z"
    }
  }
}
```

浏览器端建议把 JWT 存入 `sessionStorage`，不要写入长期存储。

### 当前管理员

```http
GET /api/v1/admin/auth/me
```

### 修改密码

```http
POST /api/v1/admin/auth/change-password
```

```json
{
  "currentPassword": "current-password",
  "newPassword": "new-password"
}
```

成功返回 `204`。新密码至少 8 个字符。

## 全局概览

```http
GET /api/v1/admin/overview?range=7d
```

`range` 支持 `7d`、`30d`，也接受 `1d` 到 `365d`。

```json
{
  "sites": 2,
  "comments": 16,
  "pendingComments": 3,
  "views": 920,
  "likes": 74,
  "trend": [
    {
      "date": "2026-11-24",
      "views": 120,
      "likes": 8,
      "unlikes": 1,
      "netLikes": 7
    }
  ]
}
```

## 站点

### 列表

```http
GET /api/v1/admin/sites?page=1&pageSize=20&q=blog
```

```json
{
  "items": [
    {
      "id": 1,
      "name": "My Blog",
      "slug": "my-blog",
      "domain": "blog.example.com",
      "status": "active",
      "allowedOrigins": ["https://blog.example.com"],
      "commentModeration": "pending",
      "createdAt": "2026-11-24T08:00:00Z",
      "updatedAt": "2026-11-24T08:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "pageSize": 20
}
```

完整 Site Key 不会出现在列表、详情或更新响应中。

### 创建

```http
POST /api/v1/admin/sites
Content-Type: application/json
```

```json
{
  "name": "My Blog",
  "slug": "my-blog",
  "domain": "blog.example.com",
  "allowedOrigins": ["https://blog.example.com"],
  "commentModeration": "pending"
}
```

字段：

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `name` | 是 | 1–100 字符 |
| `slug` | 是 | 2–64 位小写字母、数字和连字符，创建后不可修改 |
| `domain` | 否 | 展示用主域名，最长 255 字符 |
| `allowedOrigins` | 否 | 写请求允许来源；为空表示不限制；禁止 `*` |
| `commentModeration` | 否 | `pending` 或 `approved`，默认 `pending` |

创建成功返回 `201`，并在响应顶层的 `siteKey` 字段中返回完整 Key。下面的示例
省略了该敏感值：

```json
{
  "id": 3,
  "name": "My Blog",
  "slug": "my-blog",
  "commentModeration": "pending"
}
```

### 详情与更新

```http
GET   /api/v1/admin/sites/:id
PATCH /api/v1/admin/sites/:id
```

更新字段均可选：

```json
{
  "name": "New Name",
  "domain": "new.example.com",
  "status": "disabled",
  "allowedOrigins": ["https://new.example.com"],
  "commentModeration": "approved"
}
```

`status`：

- `active`：公共 API 可用。
- `disabled`：该站点的公共 API 请求返回 `403`。

### 删除

```http
DELETE /api/v1/admin/sites/:id
```

成功返回 `204`。站点采用软删除，历史数据不会立即物理清除。

### 轮换 Site Key

```http
POST /api/v1/admin/sites/:id/rotate-key
```

成功返回完整 `siteKey`。旧 Key 立即失效。完整 Key 只在此响应中展示一次。

### 吊销 Site Key

```http
DELETE /api/v1/admin/sites/:id/key
```

成功返回 `204`。之后该站点的公共接口返回 `401`，直到再次轮换 Key。

## 统计

### 站点浏览量

```http
GET /api/v1/admin/sites/:id/stats?range=7d&limit=10
```

```json
{
  "totalViews": 920,
  "totalUniqueViews": 530,
  "topPages": [
    {
      "path": "/posts/hello",
      "views": 210,
      "uniqueViews": 130
    }
  ],
  "trend": [
    {
      "date": "2026-11-24",
      "views": 120,
      "uniqueViews": 81
    }
  ]
}
```

### 站点点赞

```http
GET /api/v1/admin/sites/:id/likes?range=7d&limit=10
GET /api/v1/admin/sites/:id/likes?range=7d&limit=10&path=/posts/hello
```

不传 `path` 时返回站点总点赞、热门点赞页面和趋势。传 `path` 时返回该页面总点赞，
不返回 `topPages`。

```json
{
  "totalLikes": 74,
  "topPages": [
    {
      "path": "/posts/hello",
      "likes": 30
    }
  ],
  "trend": [
    {
      "date": "2026-11-24",
      "likes": 8,
      "unlikes": 1,
      "net": 7
    }
  ]
}
```

### 评论趋势

```http
GET /api/v1/admin/sites/:id/comments/stats?range=7d
```

```json
{
  "total": 16,
  "pending": 3,
  "approved": 12,
  "trend": [
    {
      "date": "2026-11-24",
      "comments": 4,
      "pending": 1
    }
  ]
}
```

## 评论管理

### 列表

```http
GET /api/v1/admin/sites/:id/comments
```

参数：

| 参数 | 说明 |
| --- | --- |
| `status` | `pending` / `approved` / `rejected` / `spam` |
| `path` | 精确页面路径 |
| `q` | 作者名、邮箱或内容模糊搜索 |
| `from` | 起始日期，`YYYY-MM-DD`，按 UTC 日边界 |
| `to` | 结束日期，`YYYY-MM-DD`，包含当天 |
| `page` | 默认 1 |
| `pageSize` | 默认 20，最大 100 |
| `sort` | `asc` / `desc`，默认 `desc` |

管理评论响应包含公共接口不返回的字段：

```json
{
  "id": 42,
  "siteId": 1,
  "path": "/posts/hello",
  "parentId": null,
  "authorName": "Alice",
  "authorEmail": "alice@example.com",
  "authorWebsite": "https://alice.example.com",
  "content": "谢谢分享。",
  "status": "pending",
  "isAdmin": false,
  "ipHash": "9f2f...",
  "userAgent": "Mozilla/5.0 ...",
  "createdAt": "2026-11-24T08:00:00Z",
  "updatedAt": "2026-11-24T08:00:00Z"
}
```

### 审核

```http
PATCH /api/v1/admin/comments/:id
```

```json
{
  "status": "approved"
}
```

返回更新后的完整管理评论。

### 删除

```http
DELETE /api/v1/admin/comments/:id
```

成功返回 `204`。

### 管理员回复

```http
POST /api/v1/admin/comments/:id/reply
```

```json
{
  "content": "感谢阅读。"
}
```

返回 `201`，新回复为 `isAdmin=true` 且已通过。

### 批量操作

```http
POST /api/v1/admin/sites/:id/comments/bulk
```

```json
{
  "ids": [42, 43, 44],
  "action": "approve"
}
```

`action`：

- `approve`
- `reject`
- `spam`
- `delete`

```json
{
  "updated": 3
}
```

批量操作会校验评论属于 URL 中的站点，不能跨站点操作。
