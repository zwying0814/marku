---
title: 纯 HTML 示例
description: 不依赖框架和构建步骤的完整接入示例。
outline: [2, 3]
---

# 纯 HTML 示例

仓库内置了一个可直接运行的接入示例：

```text
examples/html-blog/
├── index.html
├── embed.html
├── styles.css
├── app.js
├── config.js
├── favicon.svg
└── server.mjs
```

它演示了：

- canonical path
- 浏览量上报
- 点赞 / 取消点赞
- 评论列表与楼中楼
- 评论提交与审核状态
- `401`、`403`、`409`、`429` 与 CORS 错误处理
- 移动端与暗色模式

`embed.html` 是同一功能的 Web Component 版本，只保留三个自定义元素和一行
脚本引入；对比两个页面即可看出手写请求与组件封装的差别。

## 1. 启动 API

公共 API 允许跨域读取。如果站点配置了 `allowedOrigins`，需要把
`http://127.0.0.1:36583` 加入该站点的允许来源；留空则本地示例可直接运行。

## 2. 配置 Site Key

编辑 `examples/html-blog/config.js`：

```js
window.MARKU_EXAMPLE = {
  apiBaseUrl: 'http://localhost:36580/api/v1',
  siteKey: 'mk_your_site_key',
}
```

也可以通过运行时变量覆盖：

```js
window.MARKU_API_BASE_URL = 'https://api.example.com/api/v1'
window.MARKU_SITE_KEY = 'mk_your_site_key'
```

## 3. 启动示例

```bash
pnpm dev:html-example
```

打开：

```text
http://127.0.0.1:36583/posts/static-blogs-dynamic-content
```

`server.mjs` 会把无扩展名路径回退到 `index.html`。生产环境不需要这个
Node 服务，可以直接部署静态文件。

## 4. 浏览去重

示例在 `sessionStorage` 中记录：

```text
marku:viewed:<siteId>:<path>
```

同一标签页会话内刷新不会重复上报。接入方可以根据产品需求选择：

- 每次页面加载都记录。
- 每篇文章每个会话记录一次。
- SPA 路由切换时按新 path 记录。

## 5. 生产部署

- 将示例文件部署到博客域名。
- 修改 canonical 与文章内容。
- 使用 HTTPS。
- 将博客 origin 加入全局 CORS 与站点允许来源。
- 在构建或服务端模板中注入 Site Key。

完整说明见仓库中的 `examples/html-blog/README.md`。
