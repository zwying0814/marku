# @marku/docs

Marku 官方文档站，使用 Rspress 2 构建。

## 开发

```bash
pnpm dev:docs
```

默认地址：`http://localhost:36582/docs/`

## 构建

```bash
pnpm --filter @marku/docs build
```

静态产物输出到 `apps/docs/doc_build`，可部署到任意静态服务器。
官方单二进制构建会把该目录嵌入到 `/docs/`。

## 目录

```text
apps/docs/
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
