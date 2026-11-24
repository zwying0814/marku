---
title: 错误与限流
description: 错误码、HTTP 状态、Retry-After 与调用方处理建议。
outline: [2, 3]
---

# 错误与限流

## 错误结构

```json
{
  "code": 40000,
  "message": "invalid request body",
  "data": null
}
```

## 错误码

| HTTP | `code` | 含义 |
| --- | --- | --- |
| `200/201` | `0` | 成功 |
| `400` | `40000` | 参数或请求体无效 |
| `401` | `40100` | JWT 无效或 Site Key 无效 |
| `403` | `40300` | 站点停用、管理员停用或来源不允许 |
| `404` | `40400` | 资源不存在 |
| `405` | `40500` | 方法不允许 |
| `409` | `40900` | slug 冲突或重复评论 |
| `413` | `41300` | 请求体超过限制 |
| `415` | `41500` | Content-Type 不支持 |
| `429` | `42900` | 频率限制 |
| `500` | `50000` | 内部错误 |
| `503` | `50300` | 服务未就绪或认证未配置 |

## 常见场景

### 401

- 缺少 `X-Site-Key`
- Site Key 拼写错误
- Site Key 已轮换或吊销
- 管理 JWT 过期

### 403

- 站点已停用
- 管理员账号已停用
- 写请求的 `Origin` 不在站点 `allowedOrigins`
- 管理端来源不在全局 CORS 列表

### 409

- 创建站点时 slug 已存在
- 相同 IP 在去重窗口内提交了相同内容

### 429

两类限制：

- 全局 API 速率限制：默认 20 req/s，burst 40。
- 评论提交限制：提交过快、同一 IP 超过窗口内上限。

响应会包含：

```http
Retry-After: 60
```

调用方应暂停重试并在倒计时结束后恢复。

### 503

- 未配置管理员认证，但访问了管理接口。
- 数据库未就绪时由 `/readyz` 返回 `503`。

## 客户端处理建议

```js
if (!response.ok) {
  const retryAfter = Number(response.headers.get('Retry-After') || 0)

  if (response.status === 401) {
    // Site Key 或 JWT 失效
  } else if (response.status === 403) {
    // 检查站点状态、来源和权限
  } else if (response.status === 429) {
    // 按 retryAfter 禁用操作
  }
}
```

## 自动重试规则

可以安全重试：

- `GET`
- 点赞 `PUT`
- 取消点赞 `DELETE`
- 幂等更新 `PATCH`

不要自动重试：

- `POST /public/comments`，除非先实现幂等键或客户端去重。
- `POST /public/likes` 切换接口。
- 创建站点等可能产生重复资源的请求。

:::tip 重试预算
建议最多重试 2 次，并使用指数退避。收到 `429` 时优先遵循 `Retry-After`，
而不是固定短间隔重试。
:::
