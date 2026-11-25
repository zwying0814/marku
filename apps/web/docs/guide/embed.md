---
title: Web Component 嵌入
description: 用一行 script 给静态博客加上评论区、浏览量和点赞，并通过 CSS 变量与 ::part 自定义样式。
outline: [2, 3]
---

# Web Component 嵌入

`@marku/embed` 提供三个自定义元素：

| 元素 | 功能 |
| --- | --- |
| `<marku-comments>` | 评论列表、楼中楼回复、提交表单、审核提示、加载更多 |
| `<marku-like>` | 点赞按钮与计数 |
| `<marku-stats>` | 浏览量、评论数、点赞数，可选自动记一次浏览 |

## 一行接入

官方二进制会把组件内嵌在 `/embed/v1/`：

```html
<script type="module" src="https://api.example.com/embed/v1/marku-embed.js"></script>

<marku-comments
  api="https://api.example.com/api/v1"
  site-key="mk_your_site_key"
  path="/posts/hello"
></marku-comments>
```

同一页面上的组件可以共享配置：

```html
<meta name="marku:api" content="https://api.example.com/api/v1" />
<meta name="marku:path" content="/posts/hello" />

<marku-stats record-view></marku-stats>
<marku-like></marku-like>
<marku-comments></marku-comments>
```

`path` 的解析顺序：元素属性 → `<meta name="marku:path">` → `<link rel="canonical">`
→ 当前 `location.pathname`。`api` 的解析顺序：元素属性 → `<meta name="marku:api">`
→ `/api/v1`。

## 加载方式

三种都可用：

```html
<!-- 1. 官方二进制内置 -->
<script type="module" src="https://api.example.com/embed/v1/marku-embed.js"></script>

<!-- 2. 传统 script，自动注册 -->
<script src="https://api.example.com/embed/v1/marku-embed.iife.js"></script>
```

```ts
// 3. 打包工具
import '@marku/embed'
```

## 属性

### 通用

| 属性 | 说明 |
| --- | --- |
| `api` | API 根地址，默认 `/api/v1` |
| `site-key` | 站点 Key，必填 |
| `path` | 文章路径，默认自动推断 |
| `lang` | 文案与日期语言，默认跟随 `<html lang>` |
| `theme` | `light` / `dark`，缺省跟随系统 |
| `no-shadow` | 使用 Light DOM，样式注入到 `document.head` |

### `<marku-comments>`

| 属性 | 默认 | 说明 |
| --- | --- | --- |
| `heading` | `评论` | 标题 |
| `page-size` | `20` | 每页条数，1–100 |
| `sort` | `asc` | `asc` / `desc` |
| `show-form` | `true` | 设为 `false` 只展示列表 |
| `placeholder` | `说点什么…` | 输入框提示 |
| `empty-text` | `还没有评论，来写第一条吧。` | 空状态文案 |
| `submit-label` | `发表评论` | 提交按钮 |

提交成功后，若站点为审核制，组件会提示“审核通过后显示”，不会假装评论已出现。
表单会记住昵称、邮箱和网站，但不保存评论正文。

### `<marku-like>`

| 属性 | 默认 | 说明 |
| --- | --- | --- |
| `label` | `点赞` | 未点赞文案 |
| `liked-label` | `已点赞` | 已点赞文案 |
| `show-count` | `true` | 设为 `false` 隐藏计数 |

### `<marku-stats>`

| 属性 | 默认 | 说明 |
| --- | --- | --- |
| `show` | `views,comments,likes` | 显示哪些指标 |
| `separator` | `·` | 指标分隔符，设为空字符串可去掉 |
| `record-view` | 关闭 | 存在即记录一次浏览，同一会话同一路径只记一次 |

## 主题

组件默认使用 Shadow DOM，页面样式不会穿透进去；需要调整时用 CSS 变量和 `::part()`：

```css
marku-comments,
marku-like,
marku-stats {
  --marku-font: "Noto Serif SC", serif;
  --marku-accent: #b45309;
  --marku-accent-text: #fffaf0;
  --marku-radius: 6px;
  --marku-max-width: 640px;
  --marku-gap: 20px;
}

marku-comments::part(comment) {
  border-style: dashed;
}

marku-comments::part(submit) {
  letter-spacing: 0.04em;
}
```

可用变量：`--marku-font`、`--marku-font-size`、`--marku-text`、`--marku-muted`、
`--marku-bg`、`--marku-surface`、`--marku-border`、`--marku-accent`、
`--marku-accent-text`、`--marku-danger`、`--marku-radius`、`--marku-gap`、
`--marku-max-width`。

可用 part：

- `marku-comments`：`container`、`header`、`heading`、`count`、`list`、`comment`、
  `author`、`author-link`、`badge`、`time`、`content`、`replies`、`reply`、`actions`、
  `form`、`reply-banner`、`input`、`textarea`、`submit`、`status`、`error`、`retry`、
  `load-more`、`empty`、`loading`。
- `marku-like`：`container`、`button`、`icon`、`label`、`count`、`error`。
- `marku-stats`：`container`、`stat`、`value`、`label`、`separator`、`error`。

## Light DOM

需要让博客现有 CSS 直接命中组件内部时：

```html
<marku-comments no-shadow></marku-comments>
```

此时样式会以 `marku-comments { ... }` 的形式注入 `document.head`，内部类名统一以
`marku-` 开头。`no-shadow` 需要在元素挂载前设置。

## 事件

```js
document.querySelector('marku-comments').addEventListener('marku:comment-submitted', (event) => {
  console.log(event.detail) // { id, status, createdAt }
})

document.querySelector('marku-like').addEventListener('marku:like-change', (event) => {
  console.log(event.detail) // { liked, likes }
})

document.addEventListener('marku:error', (event) => {
  console.error(event.detail.error)
})
```

所有事件都会冒泡且 `composed: true`，可以统一挂在 `document` 上做埋点。

## 框架接入

### Astro

```astro
<marku-comments
  api={import.meta.env.PUBLIC_MARKU_API}
  site-key={import.meta.env.PUBLIC_MARKU_SITE_KEY}
  path={Astro.url.pathname}
>
</marku-comments>

<script>
  import '@marku/embed'
</script>
```

### Next.js

```tsx
'use client'

import { useEffect } from 'react'

export function Comments({ path }: { path: string }) {
  useEffect(() => {
    void import('@marku/embed')
  }, [])

  return (
    <marku-comments
      api={process.env.NEXT_PUBLIC_MARKU_API}
      site-key={process.env.NEXT_PUBLIC_MARKU_SITE_KEY}
      path={path}
    />
  )
}
```

TypeScript 不认识自定义元素时，加一份声明：

```ts
declare module 'react' {
  namespace JSX {
    interface IntrinsicElements {
      'marku-comments': React.DetailedHTMLProps<
        React.HTMLAttributes<HTMLElement>,
        HTMLElement
      > & { api?: string; 'site-key'?: string; path?: string }
    }
  }
}
```

## 安全与限制

- Site Key 会出现在页面源码中，这是设计如此。真正的防护是审核、限流、来源白名单和
  可轮换 Key。
- 评论内容一律以纯文本渲染，不解析 HTML 或 Markdown，避免 XSS。
- 组件由 JS 注入 `<style>`。启用严格 CSP 时需要允许内联样式（`style-src`），
  或把组件样式表自行提取到允许的来源。
