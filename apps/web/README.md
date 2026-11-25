# @marku/web

Marku 官网与文档站，使用 Rspress 2 构建。该站点独立部署，不打包进 API 镜像或单二进制。

## 开发

```bash
pnpm dev:web
```

默认地址：`http://localhost:36582/`

## 构建

```bash
pnpm --filter @marku/web build
```

静态产物输出到 `apps/web/doc_build`，可部署到任意静态服务器、对象存储或 CDN。

## 目录

```text
apps/web/
├── docs/
│   ├── index.md
│   ├── guide/
│   ├── api/
│   └── deploy/
├── styles/global.css
├── rspress.config.ts
└── doc_build/
```

内容维护原则：

- API 契约以 `apps/api/internal/docs/openapi.yaml` 为准。
- 环境变量以 `apps/api/.env.example` 为准。
- 示例代码优先链接或复用 `examples/html-blog`。
- 单二进制只内置 Admin 与 `/embed/v1`，站点部署方式见 `docs/deploy/`。
