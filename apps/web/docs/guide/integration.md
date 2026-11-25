---
title: 博客接入
description: Site Key、CORS、canonical path、渲染安全与页面接入流程。
outline: [2, 3]
---

# 博客接入

Marku 是 API-first 的。任何能发起 HTTP 请求的静态博客都可以接入，不要求
Next.js、React、Vue 或其他前端框架。

## 1. 两种凭证

| 凭证 | 使用位置 | 是否公开 |
| --- | --- | --- |
| Site Key | 博客前端、公共 API | 会出现在浏览器源码中 |
| Admin JWT | 管理后台、管理 API | 仅管理员会话 |

公共请求使用请求头：

```http
X-Site-Key: mk_your_site_key
```

管理请求使用：

```http
Authorization: Bearer <admin-jwt>
```

:::warning Site Key 不是 Secret
Site Key 用于识别站点与隔离数据，但它会发给浏览器。真正防止滥用的手段是
评论反垃圾、频率限制、审核流程和可轮换 Key。
:::

## 2. CORS 与来源限制

浏览器接入时需要注意两层来源策略：

### 公共路径 CORS

`/api/v1/public/*` 按设计允许跨域读取，页面不需要把博客域名加入
`MARKU_CORS_ALLOWED_ORIGINS`。这使静态博客可以直接从浏览器调用公共 API。

`MARKU_CORS_ALLOWED_ORIGINS` 主要限制管理 API 与管理端页面来源。

### 站点允许来源

公共写请求会校验当前 `Origin` 是否属于站点自己的 `allowedOrigins`：

```text
站点 allowedOrigins 为空     → 不限制来源
站点 allowedOrigins 非空     → Origin 必须完全匹配
```

生产环境建议：

1. 每个站点显式配置 `allowedOrigins`，不要留空。
2. 管理端使用单独的域名，并加入全局 CORS。
3. 不要使用通配符 `*`。

## 3. canonical path

评论、浏览量和点赞都以 `path` 作为文章标识。API 会统一处理：

- 自动补充开头的 `/`
- 去掉查询参数与锚点
- 去掉末尾 `/`
- 路径最长 512 字符

例如下列路径会被视为同一页面：

```text
/posts/hello
posts/hello
/posts/hello/
/posts/hello?utm_source=feed#comments
```

推荐博客显式传入 canonical path：

```ts title="选择路径"
const canonical =
  document.querySelector('link[rel="canonical"]')?.getAttribute('href') ??
  window.location.pathname

const path = new URL(canonical, window.location.origin).pathname
```

不要使用带预览域名、查询参数或锚点的 URL 作为 path。

## 4. visitorId

点赞必须提供稳定的匿名 `visitorId`。浏览量可选提供，用于独立访客统计。

```js
function getVisitorId() {
  const key = 'marku-visitor-id'
  const existing = localStorage.getItem(key)
  if (existing) return existing

  const id = crypto.randomUUID()
  localStorage.setItem(key, id)
  return id
}
```

不要使用邮箱、完整 IP、用户真实身份或可跨站点追踪的标识。

## 5. 最小请求封装

```js
const API_BASE = 'https://api.example.com/api/v1'
const SITE_KEY = 'mk_your_site_key'

async function marku(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
      'X-Site-Key': SITE_KEY,
      ...options.headers,
    },
  })

  const payload = await response.json()
  if (!response.ok || payload.code !== 0) {
    throw new Error(payload.message || 'Marku request failed')
  }
  return payload.data
}
```

## 6. 评论内容必须按文本渲染

Marku 保留评论中的原始字符，包含可能的 `<`、`>` 和 URL。接入方必须使用
安全的文本渲染方式：

```js
element.textContent = comment.content
```

不要使用：

```js
// 错误：可能造成 XSS
element.innerHTML = comment.content
```

如果未来启用 Markdown，也必须使用可信的 Markdown 渲染器并做 HTML 清洗。

## 7. 失败不影响阅读

互动模块应局部降级：

- 评论区加载失败：文章继续显示，提供重试。
- 点赞失败：恢复原状态并提示。
- 评论提交 `429`：读取 `Retry-After`，暂时禁用提交。
- Site Key 无效：显示配置错误，而不是让整页崩溃。
- CORS 失败：提示检查 API 域名与允许来源。

## 8. 多站点隔离

每个公共请求都必须携带对应站点的 Site Key。不要：

- 在同一个站点 Key 下混合多个博客的数据。
- 使用管理 JWT 代替 Site Key。
- 在前端根据 `siteId` 自行切换数据，因为公共 API 不接受任意 `siteId`。

## 下一步

- [评论能力与审核流程](/guide/comments)
- [浏览量与点赞](/guide/analytics)
- [公共 API 参考](/api/public)
