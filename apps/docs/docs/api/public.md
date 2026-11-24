---
title: 公共接口
description: 博客前端使用的 Site Key 接口。
outline: [2, 3]
---

# 公共接口

所有请求必须携带：

```http
X-Site-Key: mk_your_site_key
```

## 站点信息

```http
GET /api/v1/public/site
```

响应 `data`：

```json
{
  "id": 1,
  "name": "My Blog",
  "slug": "my-blog",
  "domain": "blog.example.com",
  "commentModeration": "pending"
}
```

## 聚合统计

```http
GET /api/v1/public/stats?paths=/posts/hello,/about&visitorId=...
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `paths` | 是 | 逗号分隔或重复参数，单次最多 50 个路径 |
| `visitorId` | 否 | 用于返回当前访客的 `liked` |

响应 `data.items[]`：

```json
{
  "path": "/posts/hello",
  "comments": 12,
  "views": 101,
  "uniqueViews": 63,
  "likes": 18,
  "liked": false
}
```

## 评论

### 列表

```http
GET /api/v1/public/comments?path=/posts/hello&page=1&pageSize=20&sort=asc
```

| 参数 | 默认 | 说明 |
| --- | --- | --- |
| `path` | 必填 | canonical path |
| `page` | `1` | 根评论分页 |
| `pageSize` | `20` | 1–100 |
| `sort` | `asc` | `asc` / `desc` |

响应：

```json
{
  "items": [
    {
      "id": 1,
      "parentId": null,
      "authorName": "Alice",
      "authorWebsite": "https://alice.example.com",
      "content": "谢谢分享。",
      "isAdmin": false,
      "createdAt": "2026-11-24T08:00:00Z",
      "replies": [
        {
          "id": 2,
          "parentId": 1,
          "authorName": "Admin",
          "authorWebsite": "",
          "content": "感谢阅读。",
          "isAdmin": true,
          "createdAt": "2026-11-24T08:05:00Z"
        }
      ]
    }
  ],
  "total": 1,
  "page": 1,
  "pageSize": 20
}
```

`total` 是根评论数量，`replies` 内嵌在返回的根评论中。

### 提交

```http
POST /api/v1/public/comments
Content-Type: application/json
```

```json
{
  "path": "/posts/hello",
  "parentId": null,
  "authorName": "Alice",
  "authorEmail": "alice@example.com",
  "authorWebsite": "https://alice.example.com",
  "content": "谢谢分享。",
  "honeypot": "",
  "formStartedAt": 1763952000000
}
```

成功返回 `201`：

```json
{
  "id": 42,
  "status": "pending",
  "createdAt": "2026-11-24T08:00:00Z"
}
```

常见错误：

| 状态 | 原因 |
| --- | --- |
| `400` | 字段无效、父评论无效、嵌套过深、蜜罐或时间戳异常 |
| `409` | 重复评论 |
| `429` | 同一 IP 提交过快或超过频率限制 |

### 已通过数量

```http
GET /api/v1/public/comments/count?path=/posts/hello
```

```json
{
  "count": 12
}
```

## 浏览量

### 记录

```http
POST /api/v1/public/views
Content-Type: application/json
```

```json
{
  "path": "/posts/hello",
  "visitorId": "1f46c3f4-1b7c-4cea-9a52-5c66ee7f4bb5"
}
```

```json
{
  "views": 102,
  "uniqueViews": 64,
  "recorded": true
}
```

`recorded=false` 表示请求被 Bot 过滤，计数未变化。

### 批量读取

```http
GET /api/v1/public/views?paths=/posts/hello,/posts/second
```

```json
{
  "items": [
    {
      "path": "/posts/hello",
      "views": 102,
      "uniqueViews": 64
    }
  ]
}
```

## 点赞

### 读取

```http
GET /api/v1/public/likes?path=/posts/hello&visitorId=...
```

```json
{
  "likes": 18,
  "liked": true
}
```

### 点赞

```http
PUT /api/v1/public/likes
Content-Type: application/json
```

```json
{
  "path": "/posts/hello",
  "visitorId": "1f46c3f4-1b7c-4cea-9a52-5c66ee7f4bb5"
}
```

### 取消点赞

```http
DELETE /api/v1/public/likes?path=%2Fposts%2Fhello&visitorId=1f46c3f4-1b7c-4cea-9a52-5c66ee7f4bb5
```

`PUT` 与 `DELETE` 都是幂等的，推荐用于页面交互。

### 切换

```http
POST /api/v1/public/likes
```

请求体同 `PUT`。该接口会反转当前状态，不建议带自动重试。
