---
title: 浏览量与点赞
description: 浏览上报、独立访客、点赞幂等与统计接口。
outline: [2, 3]
---

# 浏览量与点赞

## 浏览量

记录一次页面访问：

```http
POST /api/v1/public/views
X-Site-Key: mk_your_site_key
```

```json
{
  "path": "/posts/hello",
  "visitorId": "1f46c3f4-1b7c-4cea-9a52-5c66ee7f4bb5"
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "views": 101,
    "uniqueViews": 63,
    "recorded": true
  }
}
```

### 独立访客

`visitorId` 可选。独立访客统计同时依赖 `visitorId` 和以下开关：

```dotenv
MARKU_VIEW_DEDUPE_BY_VISITOR_DAY=false
```

- `false`：只累计浏览量，不累计独立访客。
- `true`：浏览量仍每次增加；独立访客按 `visitorId + 日期 + 页面` 去重累计。

不提供 `visitorId` 或关闭该开关时，`uniqueViews` 通常保持为 0。

### Bot 过滤

默认过滤明显 Bot UA：

```dotenv
MARKU_VIEW_FILTER_BOTS=true
```

被过滤时响应 `recorded=false`，计数不变。

### 批量读取

```http
GET /api/v1/public/views?paths=/posts/hello,/posts/second
X-Site-Key: mk_your_site_key
```

单次最多 50 个路径，可用 `MARKU_VIEW_MAX_BATCH_PATHS` 调整。

## 点赞

点赞以 `(siteId, path, visitorId)` 唯一确定。推荐使用幂等接口：

### 点赞

```http
PUT /api/v1/public/likes
X-Site-Key: mk_your_site_key
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
X-Site-Key: mk_your_site_key
```

### 查询状态

```http
GET /api/v1/public/likes?path=%2Fposts%2Fhello&visitorId=1f46c3f4-1b7c-4cea-9a52-5c66ee7f4bb5
X-Site-Key: mk_your_site_key
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "likes": 18,
    "liked": true
  }
}
```

:::tip 优先使用 PUT / DELETE
`POST /public/likes` 是切换接口，重试可能反向操作。网络重试场景应使用
幂等的 `PUT` 和 `DELETE`。
:::

## 聚合统计

一次请求读取指定页面的评论数、浏览量、独立访客、点赞数和当前访客点赞状态：

```http
GET /api/v1/public/stats?paths=/posts/hello,/about&visitorId=...
X-Site-Key: mk_your_site_key
```

适合文章列表页一次性获取多篇文章的互动数据。

## 管理端数据

Admin 提供：

- 全局站点、评论、待审核、浏览量和点赞概览
- 站点浏览量趋势与热门页面
- 站点点赞趋势与热门点赞页面
- 评论新增趋势与审核概况
- 支持近 7 天与近 30 天

对应接口见 [管理 API](/api/admin#统计)。
