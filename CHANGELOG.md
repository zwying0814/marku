# Changelog

所有重要变更记录在此文件中。版本号遵循语义化版本。

## [Unreleased]

### Planned

- Next.js 接入示例
- 评论提交幂等键
- npm 自动化发布工作流

## [0.3.2] - 2026-11-26

### Added

- 首次启动默认创建管理员 `admin` / `123456`，无需任何管理员环境变量。
- `PATCH /api/v1/admin/auth/profile`：验证当前密码后修改登录账号与密码。
- Admin「账号与安全」支持同时修改登录账号和密码，修改后当前会话继续有效。
- 登录页与文档提示默认账号，并要求首次登录后立即修改。
- `deploy/install.sh`：主机侧安装脚本，支持 install / upgrade / status / logs /
  backup / uninstall。

### Changed

- 移除 `MARKU_ADMIN_EMAIL` 与 `MARKU_ADMIN_PASSWORD`；管理员不再通过环境变量配置。
- Docker Compose 与安装脚本不再要求或生成管理员凭据。
- 管理员登录标识改为通用“账号”，不再强制邮箱格式。

### Security

- 默认密码仅用于首次初始化，部署后应立即在后台修改账号与密码。

## [0.3.1] - 2026-11-25

### Fixed

- Docker 镜像在全新克隆构建时缺少 `/embed/v1` 组件脚本。镜像内现在会构建
  `@marku/embed` 并复制产物，不再依赖本地被忽略的构建文件。

### Added

- GitHub Actions 镜像工作流：推送 `v*` tag 时构建 `linux/amd64` 与
  `linux/arm64` 多架构镜像并发布到 GHCR。

### Changed

- 版本升级至 `0.3.1`。

## [0.3.0] - 2026-11-25

### Added

- `@marku/client`：类型化公共 API 客户端，覆盖站点、评论、浏览量、点赞与聚合统计。
- `@marku/embed`：`<marku-comments>`、`<marku-like>`、`<marku-stats>` Web Components。
- 两个包补充 npm 发布配置：`publishConfig`（public）、`exports`、
  `prepublishOnly`、`unpkg`/`jsdelivr` 入口与 MIT 许可证。
- Shadow DOM 默认隔离、CSS 变量与 `::part()` 主题契约，以及 `no-shadow` Light DOM 模式。
- 组件事件：`marku:comment-submitted`、`marku:like-change`、`marku:error`。
- 单二进制新增 `/embed/v1/marku-embed.js` 与 `/embed/v1/marku-embed.iife.js` 分发。
- 纯 HTML 示例新增 `embed.html` 组件版本。
- 文档新增《TypeScript 客户端》与《Web Component 嵌入》。

### Changed

- `apps/docs` 重命名为 `apps/web`，定位为独立官网与文档站，Rspress `base` 改为 `/`。
- 文档站不再构建进 Docker 镜像或单二进制；二进制只内置 Admin 与 `/embed/v1`。
- 根脚本改为 `dev:web`，新增 `publish:client` / `publish:embed`。
- 版本升级至 `0.3.0`。

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

