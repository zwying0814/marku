---
title: TypeScript 客户端
description: 使用 @marku/client 在任意框架中以类型化方式调用 Marku 公共 API。
outline: [2, 3]
---

# TypeScript 客户端

`@marku/client` 是无 UI 的公共 API 封装。Astro、Next.js、Vue、Svelte 或原生脚本
都可以直接使用，只有两点要求：

1. 运行在浏览器侧（或你自己的 Edge / Node 代理里）。
2. 能访问 Marku API，并携带站点的 Site Key。

## 安装

仓库内使用 workspace 依赖：

```json title="package.json"
{
  "dependencies": {
    "@marku/client": "workspace:*"
  }
}
```

独立项目可以直接复制 `packages/client/dist/`，或从 Marku 源码构建：

```bash
pnpm --filter @marku/client build
```

## 初始化

```ts
import { MarkuClient } from '@marku/client'

export const marku = new MarkuClient({
  baseUrl: 'https://api.example.com/api/v1',
  siteKey: import.meta.env.PUBLIC_MARKU_SITE_KEY,
})
```

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| `siteKey` | 必填 | 以 `mk_` 开头，会出现在浏览器源码中 |
| `baseUrl` | `/api/v1` | API 根路径，可跨域 |
| `fetch` | 全局 `fetch` | 自定义请求实现，便于测试或代理 |
| `visitorId` | 自动生成 | 传 `null` 关闭；点赞会要求显式提供 |
| `timeoutMs` | `15000` | 单次请求超时 |
| `headers` | `{}` | 追加请求头 |

## 常用方法

```ts
const site = await marku.site()

const stats = await marku.pageStats('/posts/hello')
// { views, uniqueViews, comments, likes, liked }

const page = await marku.comments.list({
  path: '/posts/hello',
  page: 1,
  pageSize: 20,
  sort: 'asc',
})

await marku.comments.create({
  path: '/posts/hello',
  authorName: 'Ada',
  authorEmail: 'ada@example.com',
  content: '写得很好。',
})

await marku.views.record('/posts/hello')
await marku.likes.toggle('/posts/hello')
await marku.likes.set('/posts/hello', true)
```

### 绑定路径

文章页反复传 `path` 很啰嗦，用 `forPath()`：

```ts
const post = marku.forPath('/posts/hello')

const stats = await post.stats()
const comments = await post.comments.list({ page: 1 })
await post.comments.create({
  authorName: 'Ada',
  content: '收到。',
})
await post.views.record()
await post.likes.toggle()
```

### 批量统计

列表页可以一次取多篇文章的计数，避免 N 个请求：

```ts
const items = await marku.stats(['/posts/a', '/posts/b', '/posts/c'])
```

## 文章路径

客户端与服务端都会归一化 `path`：补 `/`、去查询参数、去锚点、去末尾 `/`。
推荐显式传入 canonical path：

```ts
const canonical =
  document.querySelector('link[rel="canonical"]')?.getAttribute('href') ??
  window.location.pathname
const path = new URL(canonical, window.location.origin).pathname
```

## 访客标识

未显式配置时，客户端会在 `localStorage` 写入 `marku:visitor-id`。这个 ID 只用于
点赞去重和独立访客统计，不包含邮箱、IP 或真实身份。

需要自己的账号体系时：

```ts
const marku = new MarkuClient({
  siteKey,
  visitorId: () => currentUser()?.id ?? null,
})
```

`visitorId: null` 会让点赞请求直接报错，浏览量和评论仍可用。

## 错误处理

所有失败都抛出 `MarkuError`：

```ts
import { MarkuError, isMarkuError } from '@marku/client'

try {
  await marku.comments.create({ path, authorName, content })
} catch (error) {
  if (!isMarkuError(error)) throw error

  if (error.status === 429) {
    console.warn(`触发限流，${error.retryAfter || 30}s 后重试`)
  } else if (error.status === 401) {
    console.error('Site Key 无效或已吊销')
  } else if (error.status === 403) {
    console.error('来源不在 allowedOrigins，或站点已停用')
  }
}
```

字段说明：

| 字段 | 说明 |
| --- | --- |
| `status` | HTTP 状态码，网络错误为 `0` |
| `code` | Marku 业务码，例如 `42900` |
| `retryAfter` | 429 的等待秒数 |
| `data` | 服务端附加数据 |
| `url` | 出错的请求地址 |

## Astro 示例

```astro
---
// 只做构建期变量注入，不要在这里发请求
const apiBase = import.meta.env.PUBLIC_MARKU_API
---
<script>
  import { MarkuClient } from '@marku/client'

  const marku = new MarkuClient({ siteKey: '<site-key>' })
  const stats = await marku.pageStats(window.location.pathname)
  document.querySelector('#views')!.textContent = String(stats.views ?? 0)
</script>
```

:::warning 不要在构建期调用
浏览量写入必须发生在浏览器。放到 `getStaticPaths`、服务端组件或构建脚本里，
每次 `build` 都会增加一次浏览。
:::

## 与 Web Component 的关系

`@marku/embed` 的组件内部就是用这个客户端发请求的。需要完全自定义 UI 时用
`@marku/client`；想零代码拿到评论区和点赞按钮时用 Web Component。
