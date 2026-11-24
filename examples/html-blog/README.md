# 纯 HTML / CSS / JavaScript 接入示例

这是一个不依赖框架、不需要构建步骤的 Marku 接入示例。页面本身是静态
HTML，浏览器直接调用 Marku 公共 API，完成：

- 文章浏览量上报与展示
- 点赞 / 取消点赞
- 评论列表、楼中楼回复
- 评论提交与审核状态提示
- canonical path 处理
- `401` / `403` / `429`、CORS、网络失败等错误处理

## 目录

```text
examples/html-blog/
├── index.html   # 静态文章、评论表单与评论区
├── styles.css   # 响应式样式与暗色模式
├── app.js       # 公共 API 调用与 DOM 渲染
├── config.js    # API 地址和 Site Key
├── favicon.svg  # 示例站点图标
├── server.mjs   # 仅用于本地预览的静态服务器
└── README.md
```

## 1. 启动 API

在仓库根目录启动 API：

```bash
pnpm dev:api
```

公共 API 路径允许浏览器跨域读取。站点自身的 `allowedOrigins` 会在写请求时检查：

- 留空：不限制来源，适合本地演示。
- 配置后：把 `http://127.0.0.1:36583` 或正式博客域名加入列表。

管理端跨域仍受 `MARKU_CORS_ALLOWED_ORIGINS` 限制，但与这个公共示例无关。

## 2. 获取 Site Key

执行 `pnpm seed` 会在日志中输出演示站点的 Site Key，或在 Admin
中创建站点后复制一次性展示的 Key。

然后编辑 `config.js`：

```js
window.MARKU_EXAMPLE = {
  apiBaseUrl: 'http://localhost:36580/api/v1',
  siteKey: 'mk_your_site_key',
}
```

Site Key 会出现在浏览器源码中，它不是服务端密钥。不要把管理员 JWT
放进任何公开页面。

## 3. 启动示例

```bash
pnpm dev:html-example
```

打开：

```text
http://127.0.0.1:36583/posts/static-blogs-dynamic-content
```

`server.mjs` 会把无扩展名的路径回退到 `index.html`，用于模拟真实博客的文章 URL。
生产环境可把 HTML、CSS、JS、配置与图标文件部署到任意静态服务器，不需要这个
Node 服务。

## canonical path

页面会按以下优先级决定 Marku 使用的页面路径：

1. `<meta name="marku:path" content="/custom/path">`
2. `<link rel="canonical" href="...">` 的 pathname
3. 当前 `location.pathname`

示例在 `index.html` 中配置了：

```html
<link rel="canonical" href="/posts/static-blogs-dynamic-content" />
```

同一篇文章存在多个可访问 URL 时，应统一 canonical path，避免浏览量和评论被拆散。

## 浏览量去重

示例使用 `sessionStorage` 记录：

```text
marku:viewed:<siteId>:<path>
```

同一标签页会话内重复刷新不会重复上报。实际接入时可以改为：

- 每次页面加载都记录
- 由服务端判断 bot 与访客
- 在 SPA 路由切换时按页面记录

Marku API 本身支持可选的 `visitorId` 独立访客统计和 Bot UA 过滤。

## 错误处理

- `401`：Site Key 无效或已吊销，显示配置提示。
- `403`：站点停用，或写请求 origin 不在站点允许来源中。
- `409`：短时间内提交了重复评论。
- `429`：读取 `Retry-After`，暂时禁用提交按钮并显示倒计时。
- 网络 / CORS 失败：文章保持可读，互动区域显示局部错误与重试入口。

## 生产部署注意

- 页面与 API 都建议使用 HTTPS，避免混合内容拦截。
- API 域名需要加入全局 `MARKU_CORS_ALLOWED_ORIGINS`。
- 站点 `allowedOrigins` 建议填写正式博客 origin，而不是留空。
- Site Key 轮换后，需要同步更新 `config.js` 或运行时注入的配置。
- 如果通过服务端模板注入配置，可在 `config.js` 之前设置
  `window.MARKU_SITE_KEY` 和 `window.MARKU_API_BASE_URL`。
