# Marku 开发任务清单

## 1. 项目定位

Marku 是一个面向静态博客的「无头动态内容扩展平台」：

- 静态博客保持原有的构建与部署方式，通过公共 API 或可选官方组件接入动态能力。
- 一个 API 实例支持多个独立站点，每个站点的评论、浏览量、点赞数据完全隔离。
- 提供后台管理界面，用于创建站点、管理评论、查看数据。
- 默认零依赖启动（SQLite），同时保留切换 MySQL / PostgreSQL 的能力。

### 三个应用

| 应用 | 技术栈 | 职责 |
| --- | --- | --- |
| `apps/api` | Go + Gin + GORM | 多站点动态内容 API，内置评论、浏览量、点赞 |
| `apps/admin` | React + Vite + TypeScript + Tailwind CSS | 站点管理、评论审核、数据看板 |
| `apps/web` | Rspress | 官网、文档站与接入指南（独立部署） |

### 可选集成包

官方 UI 不是核心依赖。v0.1 以公共 API 为唯一必需的接入方式，以下包均为可选：

| 包 | 技术栈 | 职责 |
| --- | --- | --- |
| `packages/client` | TypeScript | Headless 数据客户端，无样式，适合 Next.js / Astro / Vue / Svelte |
| `packages/embed` | Web Components + Shadow DOM | 可选评论组件，内置默认 UI，通过 CSS 变量与 `::part` 定制 |

### 核心原则

- 多站点隔离：所有业务数据必须带 `site_id`，查询层强制按站点过滤。
- API-first：公共 REST API 是核心契约，官方 UI 与 SDK 都是可选层。
- 术语：对外公开的站点凭证统一称 **Site Key**，它不是 Secret；管理端使用 JWT。
- 前后端契约优先：统一响应结构、错误码、分页规则，并通过 OpenAPI 同步类型。
- 先可用再扩展：v0.1 不引入消息队列、微服务、复杂权限系统。
- 可自部署：本地一条命令启动，生产环境支持 Docker Compose。

---

## 2. 目标目录结构

```text
marku/
├── apps/
│   ├── api/
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── database/
│   │   │   ├── handler/
│   │   │   ├── middleware/
│   │   │   ├── model/
│   │   │   ├── repository/
│   │   │   ├── router/
│   │   │   ├── service/
│   │   │   └── response/
│   │   ├── pkg/
│   │   ├── data/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── package.json
│   ├── admin/
│   │   ├── src/
│   │   │   ├── components/
│   │   │   ├── features/
│   │   │   ├── layouts/
│   │   │   ├── lib/
│   │   │   ├── pages/
│   │   │   ├── routes/
│   │   │   └── styles/
│   │   ├── public/
│   │   ├── index.html
│   │   ├── vite.config.ts
│   │   └── package.json
│   └── docs/
│       ├── README.md
│       └── package.json
├── examples/
│   └── nextjs-blog/         # 可选：Next.js 接入参考实现
├── packages/
│   ├── shared/              # 后续放 OpenAPI 生成的类型与常量
│   ├── client/              # 可选：Headless 数据客户端
│   └── embed/               # 可选：Web Component 评论组件
├── scripts/                 # 根级开发/初始化脚本
├── .env.example
├── .gitignore
├── .npmrc
├── .nvmrc
├── package.json
├── pnpm-workspace.yaml
├── README.md
└── DEVELOPMENT_TASKS.md
```

说明：

- Go 应用不受 pnpm 管理，但保留 `apps/api/package.json` 作为脚本代理，使 `pnpm --filter @marku/api dev` 等命令可用。
- `apps/web` 使用 Rspress 构建，内容覆盖快速开始、API、部署与发布。
- `packages/shared` 在 API 契约稳定后再创建，避免过早抽象。
- `packages/client` 与 `packages/embed` 为可选交付，v0.1 可以只保留目录与设计文档，不阻塞核心验收。
- `examples/nextjs-blog` 是 API-first 的参考实现，优先级高于自研 Widget。

---

## 3. 技术选型与约束

| 项目 | 选型 | 说明 |
| --- | --- | --- |
| 工作区 | pnpm workspace | 根目录统一脚本与依赖版本 |
| 运行时 | Node.js 24 LTS / pnpm 12 | 与当前开发环境保持一致，写入 `engines` 与 `.nvmrc` |
| API | Go 1.27+ / Gin / GORM | 分层结构，禁止 handler 直接访问数据库 |
| 数据库 | SQLite（默认）/ MySQL / PostgreSQL | SQLite 使用纯 Go 驱动，便于交叉编译与容器部署 |
| 鉴权 | 管理端 JWT + 站点 Site Key | 公共写接口使用站点级 Site Key，管理接口使用 JWT |
| 密码 | bcrypt | 管理后台账号密码 |
| 日志 | Go `log/slog` | 结构化日志，输出请求 ID |
| 前端构建 | Vite + React + TypeScript | 开启严格模式与路径别名 |
| 前端样式 | Tailwind CSS | 设计变量集中管理，支持暗色主题 |
| 路由 | React Router | 管理端 SPA 路由 |
| 服务端状态 | TanStack Query | 请求缓存、失效、加载态统一处理 |
| 客户端状态 | Zustand | 仅存登录态与 UI 偏好 |
| 表单 | react-hook-form + Zod | 表单校验与 API 类型复用 |
| 图表 | Recharts | 浏览量/点赞趋势与排行 |
| UI 基础 | shadcn/ui + Lucide（建议） | 可复制组件，不绑定重型组件库 |
| API 文档 | OpenAPI 3 + Swagger UI | 以代码注解或手写 spec 为准，需可导出 |
| 集成策略 | API-first | 公共 REST API 是核心契约，`client` / `embed` 为可选层 |
| Headless SDK | TypeScript + OpenAPI 生成类型 | 无 UI，完全由接入方渲染 |
| Embed | Web Components + Shadow DOM + CSS 变量 + `::part` | 默认样式隔离，开放有限但稳定的主题契约 |
| 测试 | Go `httptest` + SQLite 内存库 / Vitest / Playwright（可选） | 优先覆盖隔离、并发与鉴权 |

统一约定：

- 所有 API 前缀为 `/api/v1`。
- 成功响应：`{ "code": 0, "message": "ok", "data": ... }`。
- 失败响应：`{ "code": 40001, "message": "invalid request", "data": null }`，同时返回正确 HTTP 状态码。
- 分页参数：`page` 从 1 开始，`pageSize` 默认 20、最大 100。
- 分页响应：`{ "items": [], "total": 0, "page": 1, "pageSize": 20 }`。
- 时间统一使用 UTC 存储，ISO 8601 输出。
- 页面路径统一规范化：去掉 query 与 hash，去掉末尾 `/`（根路径 `/` 除外），限制最大长度。

---

## 4. 阶段任务

### M0：Monorepo 基建

> 状态：已完成

- [x] 初始化 Git 仓库，创建 `.gitignore`（Node、Go、SQLite、构建产物、环境变量）。
- [x] 创建根 `package.json`，声明 `private`、`packageManager`、Node 版本约束。
- [x] 创建 `pnpm-workspace.yaml`，包含 `apps/*`、`packages/*`。
- [x] 创建 `.npmrc`，统一依赖提升与安装行为。
- [x] 创建 `.nvmrc` 与 `.editorconfig`。
- [x] 建立 `apps/api`、`apps/admin`、`apps/web` 三个目录。
- [x] 建立可选包占位目录 `packages/client`、`packages/embed`，以及 `examples/nextjs-blog`。
- [x] 根脚本：`dev`、`dev:api`、`dev:admin`、`build`、`lint`、`test`、`typecheck`。
- [x] 配置根 README：项目简介、环境要求、快速开始、目录说明。
- [x] 创建各应用 `.env.example`，禁止提交真实密钥。
- [x] 配置基础 CI（可选但建议）：安装 pnpm、Go、执行 lint/test/build。

完成标准：

- `pnpm install` 成功。
- `pnpm -r --if-present build` 不报错。
- `go build ./...` 在 `apps/api` 下通过。
- 三个应用目录职责清晰，`docs` 不包含业务实现。

### M1：API 基础设施

> 状态：已完成（Go module path 暂定为 `marku/api`，正式组织域名待定）

- [x] 初始化 Go module，确定 module path（如 `github.com/<org>/marku/apps/api`）。
- [x] 建立分层目录：`cmd`、`config`、`database`、`handler`、`middleware`、`model`、`repository`、`service`、`router`、`response`。
- [x] 实现配置加载：端口、运行环境、数据库驱动/DSN、JWT 密钥、初始管理员、限流参数、CORS。
- [x] 实现数据库连接与连接池配置，支持 SQLite / MySQL / PostgreSQL。
- [x] 实现 GORM 自动迁移入口与初始化命令。
- [x] 实现统一响应封装、业务错误类型与错误码。
- [x] 实现中间件：Request ID、访问日志、Panic Recovery、CORS、请求体大小限制、基础限流。
- [x] 实现 `/healthz` 与 `/readyz`。
- [x] 实现优雅关闭（HTTP + 数据库）。
- [x] 编写配置、响应封装、健康检查的单元测试。

完成标准：

- `go run ./cmd/server` 可启动并创建 SQLite 数据库。
- 日志包含请求 ID、方法、路径、状态码、耗时。
- 非法配置启动时给出明确错误。
- 健康检查可用于容器探针。

### M2：多站点与鉴权

> 状态：已完成（内容仓储从 M3 起统一通过 `repository.ForSite(siteID)` 访问）

- [x] 设计 `admins` 表：邮箱、密码哈希、昵称、状态、最后登录时间。
- [x] 设计 `sites` 表：名称、slug、域名、状态、Site Key 哈希与前缀、允许来源、评论审核模式、创建/更新时间。
- [x] 启动时按环境变量创建初始管理员，密码只在首次初始化时使用。
- [x] 实现管理员登录、当前用户信息、修改密码接口。
- [x] 实现 JWT 签发与校验中间件，区分 401（未登录）与 403（无权限）。
- [x] 实现管理端站点 CRUD：列表、创建、详情、更新、删除（软删除）。
- [x] 实现 Site Key 生成、一次性展示、轮换与吊销。
- [x] 实现公共接口站点解析中间件：读取 `X-Site-Key`，校验站点状态，注入 `site_id`。
- [x] 所有业务仓储接口强制接收 `siteID`，禁止跨站点查询。
- [x] 支持按站点配置允许来源（`allowed_origins`），公共写接口执行 Origin 校验。
- [x] 编写鉴权、站点隔离、Key 轮换测试。

完成标准：

- 创建两个站点并各自生成 Key。
- 使用站点 A 的 Key 无法读取或写入站点 B 的数据。
- 轮换 Key 后旧 Key 立即失效。
- 未登录访问管理接口返回 401。

### M3：评论功能

> 状态：已完成（评论按纯文本存储和返回，客户端必须使用文本渲染，禁止直接使用 `innerHTML`）

- [x] 设计 `comments` 表：站点、页面路径、父评论、作者名、邮箱（不公开）、网站、内容、状态、IP 哈希、UA、创建时间。
- [x] 评论状态：`pending`、`approved`、`rejected`、`spam`，支持站点级默认状态。
- [x] 公共接口：提交评论，校验必填项、长度、邮箱格式、页面路径。
- [x] 反垃圾：蜜罐字段、最小提交耗时、IP + 站点限流、内容长度限制、重复内容检测。
- [x] 公共接口：按页面路径分页获取已审核评论，支持楼中楼组装与最大层级限制。
- [x] 公共接口：获取页面评论数。
- [x] 管理接口：跨状态筛选、按页面/关键词搜索、分页、排序。
- [x] 管理接口：审核通过、拒绝、标记垃圾、删除、批量操作。
- [x] 管理接口：管理员回复（生成一条 `approved` 的管理员评论）。
- [x] 评论内容输出转义策略明确，禁止存储型 XSS。
- [x] 编写评论提交、审核、层级组装、限流与站点隔离测试。

完成标准：

- 提交评论后按站点审核模式进入对应状态。
- 未审核评论不出现在公共列表。
- 管理员可在后台完成审核、回复、删除全流程。
- 恶意 HTML 在输出侧被安全处理。

### M4：浏览量

> 状态：已完成（访客去重默认关闭；爬虫过滤默认开启）

- [x] 设计 `page_stats` 表：站点、页面路径、浏览量、最后访问时间；唯一索引 `(site_id, path)`。
- [x] 实现浏览量原子自增（UPSERT），避免并发丢失。
- [x] 公共接口：单页浏览上报。
- [x] 公共接口：按多个路径批量查询浏览量，供文章列表使用。
- [x] 可选去重：基于 `visitor_id + 页面 + 日期` 的独立访客统计，默认关闭。
- [x] 过滤明显爬虫 UA，并提供开关。
- [x] 管理接口：站点总浏览量、热门页面排行、按日期趋势。
- [x] 编写并发自增、批量查询、路径规范化测试。

完成标准：

- 并发上报 100 次后计数准确。
- 一次请求可批量获取文章列表浏览量。
- 两个站点的同名路径数据互不影响。

### M5：点赞

> 状态：已完成（提供幂等 `PUT` / `DELETE`，并保留 `POST` 切换接口）

- [x] 设计 `like_stats` 表：站点、页面路径、点赞数；唯一索引 `(site_id, path)`。
- [x] 设计 `like_records` 表：站点、页面路径、访客标识；唯一索引 `(site_id, path, visitor_id)`。
- [x] 公共接口：点赞/取消点赞切换，同一访客幂等。
- [x] 公共接口：查询点赞数与当前访客是否已点赞。
- [x] 使用事务保证记录与计数一致。
- [x] 管理接口：点赞总数、热门页面排行、单页趋势。
- [x] 编写并发切换、重复提交、站点隔离测试。

完成标准：

- 同一 `visitor_id` 重复点赞不会重复计数。
- 取消点赞后计数正确。
- 并发切换后计数与明细记录一致。

### M6：API 完善与交付

> 状态：已完成（关键接口由集成测试覆盖；Swagger UI 使用 CDN 资源）

- [x] 补齐公共聚合接口：一次获取指定页面的评论数、浏览量、点赞数。
- [x] 统一参数校验与错误码文档。
- [x] 完善分页、排序、搜索的边界处理。
- [x] 输出 OpenAPI 3 文档，提供 Swagger UI（开发环境可访问）。
- [x] 基于 OpenAPI 生成前端类型，放入 `packages/shared`。
- [x] 为外部接入方提供 OpenAPI 类型生成命令，供 `@marku/client` 与 Next.js 示例复用。
- [x] 增加演示数据种子命令。
- [x] 提供 Dockerfile（多阶段构建、非 root 运行、健康检查）。
- [x] 提供 `docker-compose.yml`：API + Admin 静态站点 + 可选数据库。
- [x] 增加数据目录挂载与 SQLite 备份说明。
- [x] 核心接口具备 Router、Service、Repository 自动化测试，覆盖鉴权、站点隔离、评论和统计关键路径。
- [x] 执行 `go vet`、`gofmt`、`go test ./...`。

完成标准：

- 新环境可按 README 在 5 分钟内跑通。
- OpenAPI 与实际接口一致，前端类型由文档生成。
- `docker compose up` 可启动完整服务。

### M7：Admin 前端

> 状态：已完成（按项目要求不引入 ESLint、Oxlint 或 Prettier）

#### 7.1 基础架构

- [x] 使用 Vite 初始化 React + TypeScript 项目，开启严格模式。
- [x] 配置 Tailwind CSS、设计变量、基础主题与暗色模式。
- [x] 配置路径别名 `@/`；按项目要求不引入 ESLint、Oxlint 或 Prettier。
- [x] 建立目录：`components`、`features`、`layouts`、`lib`、`pages`、`routes`，样式集中在 `src/index.css`。
- [x] 实现应用外壳：侧边栏、顶栏、页面层级标题、响应式折叠、主题切换。
- [x] 实现基础组件：按钮、输入框、表格、分页、弹窗、确认框、Toast、空状态、错误状态、骨架屏。
- [x] 配置环境变量 `VITE_API_BASE_URL`。

#### 7.2 请求与鉴权

- [x] 封装 API Client：基础 URL、JSON 解析、错误规范化、超时、取消请求。
- [x] 实现自动附带 JWT、401 跳转登录、403 错误提示。
- [x] 使用 Zustand 管理登录态，Token 存于 `sessionStorage`，401 时清理会话。
- [x] 实现登录页、登录后跳转、路由守卫、退出登录。
- [x] 使用 TanStack Query 管理服务端状态与缓存失效。

#### 7.3 站点管理

- [x] 站点列表：搜索、状态展示、创建时间、快捷进入。
- [x] 创建站点向导：名称、slug、域名、允许来源、审核模式。
- [x] 创建成功后一次性展示 Site Key，并提供复制。
- [x] 站点详情与编辑、启用/停用、软删除。
- [x] Site Key 轮换与二次确认。
- [x] 展示站点接入代码片段：Next.js / 纯 HTML API 示例，以及可选的 `@marku/embed` 示例。

#### 7.4 仪表盘与数据

- [x] 全局概览：站点数、总评论数、总浏览量、总点赞数。
- [x] 站点概览：核心指标卡片、近 7/30 天趋势图。
- [x] 热门页面排行：按浏览量、点赞数切换。
- [x] 评论趋势与待审核数量提醒。

#### 7.5 评论管理

- [x] 评论表格：作者、内容摘要、页面路径、状态、时间。
- [x] 筛选：站点、状态、页面路径、关键词、时间范围。
- [x] 详情抽屉：完整内容、作者信息、IP 哈希、UA。
- [x] 操作：通过、拒绝、标记垃圾、删除、管理员回复。
- [x] 批量审核与批量删除。
- [x] 操作成功后刷新列表与统计数据。

#### 7.6 体验与质量

- [x] 所有列表具备加载态、空状态、错误态与重试。
- [x] 表单统一 Zod 校验，错误信息可定位到字段。
- [x] 移动端与窄屏可用。
- [x] 关键页面支持键盘操作与基础可访问性。
- [x] 增加工具函数与 API Client 的 Vitest 测试。
- [ ] 可选：Playwright 覆盖「登录 → 创建站点 → 审核评论」主流程。
- [x] 执行 `pnpm typecheck`、`pnpm test`、`pnpm build`。

完成标准：

- 管理员可完成站点创建、Key 获取、评论审核、数据查看的完整闭环。
- 刷新页面后登录态与路由行为正确。
- 所有写操作有成功/失败反馈。
- 生产构建产物可由任意静态服务器托管。

### M8：集成、文档与发布

> 状态：v0.1 已完成（Next.js 示例不纳入 v0.1）

- [x] Next.js 示例不纳入 v0.1，由纯 HTML 示例覆盖框架无关接入。
- [x] 增加纯 HTML/JS 接入示例 `examples/html-blog`：不依赖 `embed.js`，展示评论、浏览量、点赞与回复。
- [x] 纯 HTML 示例覆盖 canonical path、CORS、重复浏览上报、评论审核状态与 `429` 处理。
- [x] 接入文档补充 SPA 重复请求、canonical path 与审核状态建议。
- [x] 使用 Rspress 搭建文档站，覆盖快速开始、接入、API、部署与发布。
- [x] 编写 API 使用文档：鉴权、错误码、限流、分页、接入步骤。
- [x] 编写部署文档：SQLite 单机、MySQL/PostgreSQL、反向代理、HTTPS、备份恢复。
- [x] 根 README 补充架构图与本地开发说明。
- [x] 根脚本串联：安装、开发、构建、测试、格式化。
- [x] CI 执行：Go 测试与静态检查、前端 typecheck/test/build、文档构建。
- [x] 增加版本号与 CHANGELOG 模板。
- [x] 完成 v0.1 发布检查清单。

完成标准：

- 新用户仅阅读 README 即可完成本地启动与博客接入。
- CI 全绿。
- 通过两个模拟站点的端到端验收。

### v0.2：单二进制发行

> 状态：已完成

- [x] Admin 与 Rspress Docs 构建产物复制到 Go 包内并通过 `go:embed` 嵌入。
- [x] 根路径 `/` 提供 Admin，`/docs/` 提供 Rspress，`/swagger` 提供 Swagger UI。
- [x] 增加 `pnpm build:binary`，自动构建前端并生成 `dist/marku-api`。
- [x] Docker Compose 收敛为单个 API 容器。
- [x] Release 提供包含 Admin 的跨平台单二进制；v0.3 起官网与文档站独立部署。

### M9（可选，后续版本）：集成包与主题契约

- [x] 实现 `@marku/client`：提供站点、评论、浏览量、点赞、统计方法，支持浏览器与服务端。
- [x] 实现 `@marku/embed`：`<marku-comments>`、`<marku-like>`、`<marku-stats>`，默认 Shadow DOM。
- [x] 开放 CSS 变量：`--marku-accent`、`--marku-text`、`--marku-muted`、`--marku-surface`、`--marku-border`、`--marku-radius`、`--marku-font`、`--marku-gap`、`--marku-max-width`。
- [x] 开放 `::part()` 契约：`container`、`header`、`comment`、`author`、`time`、`content`、`actions`、`form`、`submit`、`empty`、`load-more` 等。
- [x] 支持 `theme="light|dark"` 与可选的 `no-shadow` Light DOM 模式。
- [x] 编写样式契约文档，明确内部 DOM 不属于公共 API，只有变量与 `part` 名称遵循 semver。
- [x] 单二进制通过 `/embed/v1/` 分发组件脚本，并补充单元测试。
- [ ] 为 `embed` 增加移动端、键盘导航与屏幕阅读器的专项测试。
- [ ] 配置 npm 发布、类型声明、exports 与包体积预算。

---

## 5. 数据模型草案

### `admins`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | uint | 主键 |
| `email` | string | 唯一 |
| `password_hash` | string | bcrypt |
| `nickname` | string | 显示名 |
| `status` | string | active / disabled |
| `last_login_at` | datetime | 可空 |
| `created_at` / `updated_at` | datetime | 时间戳 |

### `sites`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | uint | 主键 |
| `name` | string | 站点名称 |
| `slug` | string | 唯一标识 |
| `domain` | string | 主域名，可空 |
| `status` | string | active / disabled |
| `site_key_prefix` | string | 用于展示与排查 |
| `site_key_hash` | string | 仅存哈希 |
| `allowed_origins` | text | 逗号分隔或 JSON 数组 |
| `comment_moderation` | string | pending / approved |
| `created_at` / `updated_at` / `deleted_at` | datetime | 软删除 |

### `comments`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | uint | 主键 |
| `site_id` | uint | 索引，站点隔离 |
| `path` | string | 页面路径，联合索引 |
| `parent_id` | uint | 可空，父评论 |
| `author_name` | string | 昵称 |
| `author_email` | string | 不公开 |
| `author_website` | string | 可空 |
| `content` | text | 评论内容 |
| `status` | string | pending / approved / rejected / spam |
| `is_admin` | bool | 管理员回复标记 |
| `ip_hash` | string | 不存明文 IP |
| `user_agent` | string | 截断存储 |
| `created_at` / `updated_at` / `deleted_at` | datetime | 软删除 |

### `page_stats`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | uint | 主键 |
| `site_id` | uint | 站点 |
| `path` | string | 页面路径 |
| `views` | uint64 | 浏览量 |
| `last_viewed_at` | datetime | 最后访问时间 |
| `unique_key` | string | 唯一索引 `(site_id, path)` |

### `like_stats` / `like_records`

| 表 | 关键字段 | 说明 |
| --- | --- | --- |
| `like_stats` | `site_id`、`path`、`likes`、唯一索引 `(site_id, path)` | 聚合计数 |
| `like_records` | `site_id`、`path`、`visitor_id`、唯一索引 `(site_id, path, visitor_id)` | 访客点赞明细，防重复 |

### 站点隔离规则

- 所有内容表必须包含 `site_id`。
- Repository 方法签名必须显式传入 `siteID`。
- 管理端查询同样按站点过滤，禁止「先按 ID 查再校验」的写法。
- 测试中至少包含一组「同 ID 不同站点」的越权用例。

---

## 6. API 路由草案

### 公共接口（需 `X-Site-Key`）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/v1/public/site` | 获取站点公开配置 |
| GET | `/api/v1/public/comments?path=&page=&pageSize=` | 获取已审核评论 |
| POST | `/api/v1/public/comments` | 提交评论 |
| GET | `/api/v1/public/comments/count?path=` | 获取评论数 |
| POST | `/api/v1/public/views` | 上报浏览量 |
| GET | `/api/v1/public/views?paths=a,b` | 批量查询浏览量 |
| POST | `/api/v1/public/likes` | 点赞/取消点赞 |
| GET | `/api/v1/public/likes?path=&visitorId=` | 查询点赞状态 |
| GET | `/api/v1/public/stats?paths=a,b` | 聚合评论/浏览/点赞 |

公共请求头：

```text
X-Site-Key: mk_live_xxx
X-Visitor-Id: <客户端生成的匿名 ID>
Origin: https://blog.example.com
```

### 管理接口（需 `Authorization: Bearer <JWT>`）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/v1/admin/auth/login` | 登录 |
| GET | `/api/v1/admin/auth/me` | 当前管理员 |
| POST | `/api/v1/admin/auth/change-password` | 修改密码 |
| GET / POST | `/api/v1/admin/sites` | 站点列表 / 创建 |
| GET / PATCH / DELETE | `/api/v1/admin/sites/:id` | 站点详情 / 更新 / 删除 |
| POST | `/api/v1/admin/sites/:id/rotate-key` | 轮换 Site Key |
| GET | `/api/v1/admin/sites/:id/stats?range=7d` | 站点统计 |
| GET | `/api/v1/admin/sites/:id/comments` | 评论列表 |
| PATCH | `/api/v1/admin/comments/:id` | 修改评论状态 |
| DELETE | `/api/v1/admin/comments/:id` | 删除评论 |
| POST | `/api/v1/admin/comments/:id/reply` | 管理员回复 |
| GET | `/api/v1/admin/sites/:id/pages?sort=views,likes` | 页面数据排行 |

错误码建议：

| 范围 | 含义 |
| --- | --- |
| `400xx` | 参数或状态错误 |
| `401xx` | 未认证 |
| `403xx` | 无权限 / 站点被禁用 / 来源被拒绝 |
| `404xx` | 资源不存在 |
| `409xx` | 冲突或重复提交 |
| `429xx` | 触发限流 |
| `500xx` | 服务端错误 |

> 点赞当前草案为 `POST` 切换；若优先考虑 API 接入方的重试安全，建议在 M5 改为 `PUT`（点赞）/ `DELETE`（取消）。

---

## 7. 集成方式与样式契约

### 7.1 接入层次

Marku 采用 API-first：

1. **公共 REST API（必需）**：任何博客框架都可以直接调用，v0.1 的验收只依赖这一层。
2. **`@marku/client`（可选）**：Headless 数据客户端，仅封装请求、类型、错误与去重，不提供 UI。
3. **`@marku/embed`（可选）**：内置默认评论 UI 的 Web Component，适合不想自己写界面的博客。

`embed` 不是 v0.1 的核心依赖。官方优先保证 API、OpenAPI 文档和纯 HTML 示例可用。

### 7.2 UI 接入方式对比

| 方式 | 样式自由度 | 隔离性 | 适用场景 |
| --- | --- | --- | --- |
| Headless / 自研 UI | 完全自由 | 接入方决定 | Next.js、Astro、Vue、Svelte |
| Embed + Shadow DOM | 中高：CSS 变量 + `::part()` + slot | 强 | 官方默认组件 |
| Embed + Light DOM | 高：页面 CSS 可覆盖 | 弱 | 需要完全融入主题的站点 |
| iframe | 低 | 最强 | 不推荐用于评论 |

### 7.3 `@marku/embed` 基础用法

```html
<marku-comments
  site-key="mk_pub_xxxxxxxx"
  page-path="/posts/hello"
  theme="auto"
></marku-comments>
```

### 7.4 主题变量

| 变量 | 说明 |
| --- | --- |
| `--marku-accent` | 主色，用于按钮、链接、选中态 |
| `--marku-text` | 正文颜色 |
| `--marku-muted` | 次要文字颜色 |
| `--marku-surface` | 组件背景 |
| `--marku-border` | 边框与分隔线 |
| `--marku-radius` | 圆角 |
| `--marku-font` | 字体，默认 `inherit` |
| `--marku-spacing` | 基础间距 |
| `--marku-max-width` | 评论区域最大宽度 |

```css
marku-comments {
  --marku-accent: #2563eb;
  --marku-radius: 10px;
  --marku-font: inherit;
  --marku-surface: transparent;
}

marku-comments::part(comment) {
  padding-block: 0.75rem;
  border-bottom: 1px solid var(--marku-border);
}

marku-comments::part(submit) {
  background: var(--marku-accent);
  color: #fff;
  border-radius: var(--marku-radius);
}
```

### 7.5 `::part()` 契约

建议稳定开放以下 part 名称：

| Part | 说明 |
| --- | --- |
| `container` | 组件最外层容器 |
| `header` | 标题与评论数区域 |
| `comment` | 单条评论 |
| `avatar` | 头像 |
| `author` | 作者名 |
| `meta` | 时间等元信息 |
| `content` | 评论正文 |
| `actions` | 回复、点赞等操作区 |
| `form` | 评论表单 |
| `submit` | 提交按钮 |
| `empty` | 空状态 |
| `pagination` | 分页 |

规则：

- 只有上述变量和 `part` 名称属于公共样式契约；内部 DOM 结构与 class 名不是公共 API。
- `::part()` 只能直接命中标记了 `part` 的节点，不能任意选择其后代；嵌套结构如需暴露，使用独立 part 或 `exportparts`。
- 新增变量和 part 属于兼容变更；删除或重命名需要主版本升级。
- 默认不要求使用 `!important`，页面不应通过全局选择器强行覆盖 Shadow DOM。
- 字体、颜色等优先通过 CSS 变量继承，避免主题切换失效。

### 7.6 Light DOM 模式

```html
<marku-comments data-render="light"></marku-comments>
```

Light DOM 模式允许博客用普通 CSS 覆盖 `.marku-*` 类，适合需要深度融入现有主题的站点，但存在全局样式冲突风险。该模式应作为显式选项，不作为默认行为。

### 7.7 不建议 iframe

iframe 样式隔离最强，但评论场景需要处理高度自适应、滚动、移动端键盘、无障碍和 SEO，且页面 CSS 无法直接定制内部样式。Marku 不采用 iframe 作为默认实现。

### 7.8 版本与兼容策略

- `@marku/client` 与 `@marku/embed` 独立 semver 版本。
- 样式变量与 `part` 名称写入文档，作为稳定契约。
- 内部 DOM 可随版本重构，不承诺可被外部选择器命中。
- 废弃变量或 part 至少保留一个次版本周期，并在 CHANGELOG 中标注。

---

## 8. 环境变量草案

### `apps/api/.env.example`

```dotenv
MARKU_ENV=development
MARKU_HTTP_ADDR=:36580
MARKU_DB_DRIVER=sqlite
MARKU_DB_DSN=./data/marku.db
MARKU_JWT_SECRET=replace-with-a-long-random-secret
MARKU_ADMIN_EMAIL=admin@example.com
MARKU_ADMIN_PASSWORD=change-me-on-first-run
MARKU_IP_HASH_SALT=replace-with-random-salt
MARKU_COMMENT_MODERATION=pending
MARKU_RATE_LIMIT_ENABLED=true
```

### `apps/admin/.env.example`

```dotenv
VITE_API_BASE_URL=http://localhost:36580/api/v1
```

管理端不得包含任何服务端密钥；站点 Site Key 由管理员在后台查看并提供给公开博客使用。

---

## 9. v0.1 验收标准

- [x] `pnpm install` 后，`pnpm dev` 可同时启动 API、Admin 与 Web。
- [x] 可创建至少两个站点，并获取各自 Site Key。
- [x] 两个站点的评论、浏览量、点赞数据完全隔离。
- [x] `examples/html-blog` 可完成评论提交、评论展示、浏览量上报、点赞切换。
- [x] v0.1 接入不依赖 `@marku/embed`，可选集成包不阻塞核心验收。
- [x] 管理后台可审核评论、回复评论、查看趋势与热门页面。
- [x] Site Key 轮换后旧 Key 失效。
- [x] 核心接口具备自动化测试，站点隔离与并发计数用例通过。
- [x] `docker compose up` 可启动完整服务。
- [x] README、API 文档、部署文档齐全。
- [x] `apps/web` 提供可构建的 Rspress 官网与文档站。

---

## 10. 开发顺序建议

1. M0 Monorepo 基建
2. M1 API 基础设施
3. M2 多站点与鉴权
4. M3 评论
5. M4 浏览量
6. M5 点赞
7. M7.1 - M7.2 Admin 基础与登录
8. M7.3 - M7.5 站点、看板、评论管理
9. M6 API 文档、类型同步、容器化
10. M8 集成示例、部署文档、发布验收
11. M9（可选）集成包与主题契约

建议每完成一个阶段执行一次完整构建与测试，避免最后集中修复。

---

## 11. 暂不纳入 v0.1 的能力

以下内容记录为后续方向，本期不做：

- 站点级管理员与细粒度 RBAC。
- 评论邮件通知、Webhook、第三方反垃圾服务。
- OAuth 登录与游客账号体系。
- 评论附件、图片上传、富文本编辑器。
- 多语言内容与多地区部署。
- 数据导出、审计日志、操作回滚。
- 可视化主题编辑器、主题市场、第三方框架官方组件库。
- Next.js 官方示例。

---

## 12. 待确认决策

- [ ] Go module path 使用哪个组织域名。
- [x] 管理端 JWT 存储方案：使用 `sessionStorage`，关闭标签页或退出后清除。
- [x] SQLite 之外同时保留 MySQL / PostgreSQL 驱动支持。
- [x] 评论内容按纯文本保存与展示，不在 v0.1 启用 Markdown。
- [x] 公共写接口使用站点 `allowedOrigins` 校验来源；公共读取仍要求 Site Key。
- [x] Admin 使用自定义 React + Tailwind 组件，不引入 shadcn/ui。
- [x] `@marku/client` / `@marku/embed` 不纳入 v0.1，只提供 API 与 HTML 示例。
- [ ] `@marku/embed` 是否默认 Shadow DOM，并保留 Light DOM 模式。
- [ ] `::part` 名称与 CSS 变量是否从首个版本起作为稳定契约。
- [x] 点赞接口提供幂等 `PUT` / `DELETE`，保留 `POST` 切换兼容接口。
- [ ] 评论提交是否要求 `Idempotency-Key`，避免客户端重试产生重复评论。
