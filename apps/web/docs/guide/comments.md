---
title: 评论
description: 评论提交、审核、楼中楼、管理员回复与反垃圾策略。
outline: [2, 3]
---

# 评论

评论以站点和 canonical path 为作用域。不同站点、不同文章的评论完全隔离。

## 评论状态

| 状态 | 是否公开 | 说明 |
| --- | --- | --- |
| `pending` | 否 | 等待管理员审核 |
| `approved` | 是 | 公开展示 |
| `rejected` | 否 | 管理员拒绝 |
| `spam` | 否 | 标记为垃圾评论 |

公共评论列表只返回 `approved` 评论及其已通过的子回复。

## 审核模式

每个站点可选择：

- `pending`：先审后发，新评论默认进入待审核队列。
- `approved`：自动公开，适合信任度高的站点。

创建站点后可在 Admin 的「站点详情 → 编辑站点」中修改。

## 提交评论

```http
POST /api/v1/public/comments
X-Site-Key: mk_your_site_key
Content-Type: application/json
```

```json
{
  "path": "/posts/hello",
  "parentId": 12,
  "authorName": "Alice",
  "authorEmail": "alice@example.com",
  "authorWebsite": "https://alice.example.com",
  "content": "写得很清楚，谢谢分享。",
  "honeypot": "",
  "formStartedAt": 1763952000000
}
```

成功响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 42,
    "status": "pending",
    "createdAt": "2026-11-24T08:00:00Z"
  }
}
```

接入方应根据 `status` 显示不同提示：

- `approved`：评论已发布。
- `pending`：评论已提交，审核通过后展示。

## 字段限制

| 字段 | 限制 |
| --- | --- |
| `path` | 必填，最长 512 字符 |
| `authorName` | 1–100 字符 |
| `authorEmail` | 可选，合法邮箱，最长 255 字符 |
| `authorWebsite` | 可选，`http` / `https` URL，最长 255 字符 |
| `content` | 必填，默认最多 2000 字符 |
| `formStartedAt` | 表单打开时的 Unix 毫秒时间戳 |

邮箱、IP 哈希与 User Agent 不会出现在公共评论响应中。

## 楼中楼

回复时设置 `parentId`。默认最大深度为 3，即：

```text
根评论 → 第 1 层回复 → 第 2 层回复
```

超过深度返回 `400`。接入 UI 可以在默认深度内隐藏更深的回复按钮，也可以直接
让 API 返回错误并提示。

## 反垃圾

默认策略：

- 蜜罐字段 `honeypot` 必须为空。
- `formStartedAt` 与提交时间差至少 2 秒。
- 同一 IP 10 分钟内最多 5 条评论。
- 同一 IP、路径和内容 10 分钟内不允许重复。
- 评论中最多 3 个链接。
- 拒绝控制字符。

可通过环境变量调整：

```dotenv
MARKU_COMMENT_MIN_SUBMIT_DURATION=2s
MARKU_COMMENT_RATE_LIMIT_MAX=5
MARKU_COMMENT_RATE_LIMIT_WINDOW=10m
MARKU_COMMENT_DUPLICATE_WINDOW=10m
MARKU_COMMENT_MAX_DEPTH=3
MARKU_COMMENT_MAX_LENGTH=2000
MARKU_COMMENT_MAX_LINKS=3
```

## 管理员回复

管理员在 Admin 评论详情抽屉中回复。管理员回复：

- `isAdmin=true`
- 状态为 `approved`
- 与普通回复一样以楼中楼结构返回

## 渲染安全

评论内容按原文字符保存。接入方必须使用文本节点渲染，禁止直接 `innerHTML`：

```js
textElement.textContent = comment.content
```

## 相关接口

- [公共评论接口](/api/public#评论)
- [管理评论接口](/api/admin#评论管理)
