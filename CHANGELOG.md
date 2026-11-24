# Changelog

所有重要变更记录在此文件中。版本号遵循语义化版本。

## [Unreleased]

### Planned

- Next.js 接入示例（v0.1 不包含）
- `@marku/client` 与 `@marku/embed`（v0.2 候选）
- 评论提交幂等键

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

