# Changelog

所有重要变更记录在此文件中。版本号遵循语义化版本。

## [Unreleased]

### Planned

- Next.js 接入示例（v0.1 不包含）
- `@marku/client` 与 `@marku/embed`（v0.2 候选）
- 评论提交幂等键

## [0.2.0] - 2026-11-24

### Added

- 单二进制发行：Admin 与 Rspress Docs 通过 `go:embed` 内置于 API。
- `pnpm build:binary`：自动构建前端、复制静态资源并编译 Go 二进制。
- 单容器 Docker 构建：一个 `marku-api` 进程同时提供 API、Admin 和 Docs。

### Changed

- Admin 移至根路径 `/`，Rspress Docs 使用 `/docs/`。
- Swagger UI 从 `/docs` 迁移到 `/swagger`。
- Docker Compose 移除独立 Admin 服务，仅保留 API 服务。
- 内置 Admin 与 API 同源，不再需要跨域配置。
- 版本升级至 `0.2.0`。

### Notes

- 静态资源仍可独立构建和部署，但官方 Release 以单二进制为主。

## [0.1.0] - 2026-11-24

### Added

- Go + Gin + GORM 多站点动态内容 API。
- 管理员 JWT、bcrypt 密码与初始管理员 seed。
- 站点 CRUD、Site Key 生成、轮换、吊销与来源白名单。
- 评论提交、审核、楼中楼、管理员回复、批量操作和反垃圾策略。
- 浏览量、独立访客、Bot 过滤、点赞与趋势统计。
- OpenAPI 3 文档、Swagger UI 与共享 TypeScript 类型。
- React + Tailwind CSS 管理后台，包含看板、站点、数据、评论与设置页面。
- Rspress 文档站，覆盖快速开始、接入、API、部署、备份和发布清单。
- 纯 HTML/CSS/JavaScript 接入示例，包含评论、浏览量和点赞。
- Docker Compose、Admin Nginx 静态托管与 SQLite/PostgreSQL 部署说明。
- GitHub Actions CI、Vitest 与 Go race 测试。

### Security

- 公共与管理接口使用不同凭证。
- 评论不公开邮箱、IP 哈希和 User Agent。
- Site Key 仅保存哈希和展示前缀。
- 公共写请求支持站点级 origin 校验。
- 点赞提供幂等 `PUT` / `DELETE`。

