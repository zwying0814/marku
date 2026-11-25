# @marku/embed

基于 Web Component 的评论、点赞与统计组件。默认使用 Shadow DOM，样式不会污染博客；
通过 CSS 变量与 `::part()` 自定义外观，也可以加 `no-shadow` 退回 Light DOM。

```html
<script type="module" src="https://api.example.com/embed/v1/marku-embed.js"></script>

<marku-comments
  api="https://api.example.com/api/v1"
  site-key="mk_your_site_key"
  path="/posts/hello"
></marku-comments>

<marku-stats record-view></marku-stats>
<marku-like></marku-like>
```

`api`、`site-key`、`path` 都可以省略：

- `api` 依次读取属性、`<meta name="marku:api">`、`/api/v1`。
- `path` 依次读取属性、`<meta name="marku:path">`、`<link rel="canonical">`、当前路径。
- `site-key` 可以从 `<meta name="marku:site-key">` 之外的任意方式注入；
  当前版本只读取元素属性，避免把 Key 写进全局状态。

## 主题

```css
marku-comments {
  --marku-accent: #b45309;
  --marku-radius: 6px;
  --marku-max-width: 640px;
}

marku-comments::part(comment) {
  border-style: dashed;
}
```

可用 part：`container`、`header`、`heading`、`count`、`list`、`comment`、`author`、
`badge`、`time`、`content`、`replies`、`reply`、`form`、`input`、`textarea`、`submit`、
`status`、`error`、`load-more`、`empty`、`button`（`marku-like`）、`stat` / `value` /
`label`（`marku-stats`）。

## 事件

| 事件 | 来源 | detail |
| --- | --- | --- |
| `marku:comment-submitted` | `marku-comments` | 提交结果 |
| `marku:like-change` | `marku-like` | `{ liked, likes }` |
| `marku:error` | 全部组件 | `{ error }` |

## 构建

`pnpm --filter @marku/embed build` 输出：

- `dist/marku-embed.js`：ES module。
- `dist/marku-embed.iife.js`：可直接用 `<script>` 引入，自动注册组件。

单二进制发行会把这些文件内嵌到 `/embed/v1/`。
