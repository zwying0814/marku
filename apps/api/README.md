# @marku/api

Marku 的公共 API，基于 Go、Gin、GORM 和 SQLite。

当前版本 v0.3.1：

- 分层目录：`cmd`、`config`、`database`、`handler`、`middleware`、`model`、`repository`、`router`、`service`、`response`
- 环境变量配置与启动校验
- SQLite、MySQL、PostgreSQL 驱动切换
- 统一响应、业务错误码、Request ID、结构化日志、Panic Recovery
- CORS、请求体大小限制、基于 IP 的基础限流
- 管理员账号、bcrypt 密码、JWT 登录与鉴权
- 多站点 CRUD、软删除、Site Key 生成/轮换/吊销
- 公共接口 Site Key 解析、站点禁用校验、按站点 Origin 白名单
- `repository.ForSite(siteID)` 作为后续内容仓储的强制入口
- 评论提交、审核、公开列表、评论数、楼中楼组装
- 反垃圾：蜜罐、最短提交时间、IP + 站点限流、内容长度、链接数和重复内容检测
- 管理端评论筛选、审核、删除、管理员回复和批量操作
- 浏览量原子自增、批量查询、按访客/日期的可选去重
- 爬虫 UA 过滤开关、每日浏览趋势和热门页面统计
- 点赞幂等设置/取消、访客状态查询、点赞总数与每日净变化趋势
- 公共聚合统计接口，一次返回评论数、浏览量、点赞数和访客点赞状态
- OpenAPI 3 文档、Swagger UI、共享 TypeScript 类型生成
- Admin 与 `/embed/v1` 组件通过 `go:embed` 内置，单二进制同时提供 API 和 Web UI
- Dockerfile、Docker Compose、Demo seed 命令
- `/healthz`、`/readyz`、优雅关闭

## 开发

在仓库根目录执行：

```bash
pnpm dev:api
```

或直接进入应用目录：

```bash
cd apps/api
cp .env.example .env
go run ./cmd/server
```

构建内置 Admin 与组件的单二进制：

```bash
pnpm build:binary
```

产物：`dist/marku-api`。

默认监听 `:36580`，SQLite 数据文件写入 `apps/api/data/marku.db`。

## 首次启动与管理员

启用管理后台前，在 `apps/api/.env` 中设置：

```dotenv
MARKU_JWT_SECRET=<至少 32 位的随机字符串>
MARKU_ADMIN_EMAIL=admin@example.com
MARKU_ADMIN_PASSWORD=<至少 8 位的初始密码>
```

启动时会自动迁移数据库，并在管理员不存在时创建初始管理员。密码只在首次创建时使用，修改环境变量不会覆盖已有账号密码。

如果 JWT Secret 为空，服务仍可启动，但管理员登录返回 `503`。

## 管理员认证

```bash
curl -X POST http://localhost:36580/api/v1/admin/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.com","password":"<password>"}'
```

后续管理接口使用：

```text
Authorization: Bearer <token>
```

JWT 默认有效期 `8h`，可通过 `MARKU_JWT_TTL` 调整。

## 站点与 Site Key

创建站点时返回一次完整的 `siteKey`：

```bash
curl -X POST http://localhost:36580/api/v1/admin/sites \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"My Blog","slug":"my-blog","allowedOrigins":["https://blog.example.com"]}'
```

响应中的 `siteKey` 只在创建和轮换时返回。公共博客通过请求头使用：

```text
X-Site-Key: mk_xxxxxxxx
```

Site Key 会出现在博客前端，不是 Secret。它只能访问当前站点的公共能力，不能访问管理接口。

## 评论

提交评论：

```bash
curl -X POST http://localhost:36580/api/v1/public/comments \
  -H 'X-Site-Key: mk_xxxxxxxx' \
  -H 'Content-Type: application/json' \
  -d '{
    "path": "/posts/hello",
    "authorName": "Alice",
    "authorEmail": "alice@example.com",
    "authorWebsite": "https://alice.example.com",
    "content": "Nice post!",
    "honeypot": ""
  }'
```

- `path` 会被规范化为页面路径，忽略 query 和 hash。
- `formStartedAt` 可选，单位毫秒；提供时会校验最短提交时间。
- `honeypot` 必须为空。
- 评论是否直接公开取决于站点的 `commentModeration`。
- 公开列表只返回 `approved` 评论，并只包含公开字段；邮箱、IP 哈希和 User-Agent 不会出现在公共响应中。

公开列表与评论数：

```bash
curl -G http://localhost:36580/api/v1/public/comments \
  -H 'X-Site-Key: mk_xxxxxxxx' \
  --data-urlencode 'path=/posts/hello'

curl -G http://localhost:36580/api/v1/public/comments/count \
  -H 'X-Site-Key: mk_xxxxxxxx' \
  --data-urlencode 'path=/posts/hello'
```

评论内容按纯文本存储，并通过 JSON 字符串返回。接入方必须使用 `textContent`、React 默认转义等文本渲染方式，禁止把评论内容直接交给 `innerHTML`。官方 Embed 和管理后台也必须遵守这一策略。

评论列表的 `total` 表示顶层评论分页总数；`/public/comments/count` 返回包含回复在内的已审核评论总数。

## 浏览量

上报单页浏览量：

```bash
curl -X POST http://localhost:36580/api/v1/public/views \
  -H 'X-Site-Key: mk_xxxxxxxx' \
  -H 'Content-Type: application/json' \
  -H 'User-Agent: Mozilla/5.0' \
  -d '{"path":"/posts/hello","visitorId":"visitor-uuid"}'
```

批量查询：

```bash
curl -G http://localhost:36580/api/v1/public/views \
  -H 'X-Site-Key: mk_xxxxxxxx' \
  --data-urlencode 'paths=/posts/hello,/about'
```

- `path` 会忽略 query 和 hash，并统一末尾斜杠。
- `visitorId` 可选。开启 `MARKU_VIEW_DEDUPE_BY_VISITOR_DAY` 后，同一访客、同一路径、同一 UTC 日期只计入一次 `uniqueViews`，但 `views` 仍会按上报次数累加。
- `MARKU_VIEW_FILTER_BOTS=true` 时会过滤常见爬虫 UA，包括 `curl`、`wget`、`Googlebot` 等；本地用 curl 测试时请显式传浏览器 UA，或关闭该开关。
- 批量查询最多 `MARKU_VIEW_MAX_BATCH_PATHS` 个路径，缺失路径返回 0。

管理端统计：

```bash
curl 'http://localhost:36580/api/v1/admin/sites/1/stats?range=7d&limit=10' \
  -H "Authorization: Bearer $TOKEN"
```

返回站点总浏览量、总独立访客、热门页面和按日期补齐的趋势数据。

## 点赞

幂等点赞与取消点赞：

```bash
curl -X PUT http://localhost:36580/api/v1/public/likes \
  -H 'X-Site-Key: mk_xxxxxxxx' \
  -H 'Content-Type: application/json' \
  -d '{"path":"/posts/hello","visitorId":"visitor-uuid"}'

curl -X DELETE http://localhost:36580/api/v1/public/likes \
  -H 'X-Site-Key: mk_xxxxxxxx' \
  -H 'Content-Type: application/json' \
  -d '{"path":"/posts/hello","visitorId":"visitor-uuid"}'
```

查询点赞数与当前访客状态：

```bash
curl -G http://localhost:36580/api/v1/public/likes \
  -H 'X-Site-Key: mk_xxxxxxxx' \
  --data-urlencode 'path=/posts/hello' \
  --data-urlencode 'visitorId=visitor-uuid'
```

- `PUT` 表示点赞，`DELETE` 表示取消点赞，两者均幂等，适合网络重试。
- `POST /public/likes` 提供切换语义，适合简单的按钮交互。
- 同一访客对同一页面最多只有一条有效点赞记录。
- 管理端统计：`GET /api/v1/admin/sites/:id/likes?range=7d&limit=10`；追加 `path=/posts/hello` 可查看单页数据。

## 聚合统计

一次请求获取多个页面的评论数、浏览量、点赞数和当前访客点赞状态：

```bash
curl -G http://localhost:36580/api/v1/public/stats \
  -H 'X-Site-Key: mk_xxxxxxxx' \
  --data-urlencode 'paths=/posts/hello,/about' \
  --data-urlencode 'visitorId=visitor-uuid'
```

返回项包含 `comments`、`views`、`uniqueViews`、`likes` 和 `liked`。

## API 文档与类型

- Swagger UI：`http://localhost:36580/swagger`
- OpenAPI 3：`http://localhost:36580/openapi.yaml`

Swagger UI 静态资源默认从 unpkg CDN 加载；离线环境需要自行托管 `swagger-ui-dist`。

生成共享 TypeScript 类型：

```bash
pnpm generate:types
```

生成文件：`packages/shared/src/api.d.ts`。

## Demo 数据

```bash
pnpm seed
```

会创建两个演示站点、示例评论、浏览量和点赞数据。站点已存在时会复用，不会重复创建点赞记录。

## 主要接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/v1/admin/auth/login` | 管理员登录 |
| GET | `/api/v1/admin/auth/me` | 当前管理员 |
| POST | `/api/v1/admin/auth/change-password` | 修改密码 |
| GET / POST | `/api/v1/admin/sites` | 站点列表 / 创建 |
| GET / PATCH / DELETE | `/api/v1/admin/sites/:id` | 站点详情 / 更新 / 软删除 |
| POST | `/api/v1/admin/sites/:id/rotate-key` | 轮换 Site Key |
| DELETE | `/api/v1/admin/sites/:id/key` | 吊销 Site Key |
| GET | `/api/v1/admin/sites/:id/comments` | 站点评论列表与筛选 |
| POST | `/api/v1/admin/sites/:id/comments/bulk` | 批量审核 / 删除评论 |
| PATCH | `/api/v1/admin/comments/:id` | 修改评论状态 |
| DELETE | `/api/v1/admin/comments/:id` | 删除评论 |
| POST | `/api/v1/admin/comments/:id/reply` | 管理员回复 |
| GET | `/api/v1/admin/sites/:id/stats` | 站点浏览量统计、热门页面和趋势 |
| GET | `/api/v1/admin/sites/:id/likes` | 站点点赞统计、热门页面和单页趋势 |
| GET | `/api/v1/public/site` | 当前 Site Key 对应的站点公开配置 |
| GET | `/api/v1/public/comments` | 已审核评论列表，支持楼中楼 |
| POST | `/api/v1/public/comments` | 提交评论 |
| GET | `/api/v1/public/comments/count` | 页面已审核评论数 |
| GET | `/api/v1/public/views` | 批量查询浏览量 |
| POST | `/api/v1/public/views` | 上报浏览量 |
| GET | `/api/v1/public/likes` | 查询点赞数与当前访客状态 |
| POST | `/api/v1/public/likes` | 点赞 / 取消点赞切换 |
| PUT | `/api/v1/public/likes` | 幂等点赞 |
| DELETE | `/api/v1/public/likes` | 幂等取消点赞 |
| GET | `/api/v1/public/stats` | 聚合评论数、浏览量、点赞数 |
| GET | `/healthz` | 进程健康检查 |
| GET | `/readyz` | 数据库就绪检查 |
| GET | `/` | 内置 Admin SPA |
| GET | `/embed/v1/*` | 官方 Web Component 脚本 |
| GET | `/swagger` | Swagger UI |
| GET | `/openapi.yaml` | OpenAPI 3 文档 |

所有响应使用统一结构：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

## 数据隔离

公共 Site Key 中间件会解析站点并注入 `site_id`。评论仓储以及后续内容仓储必须通过站点作用域访问：

```go
scope := store.ForSite(siteID)
scope.Query().Where("path = ?", path).Find(&comments)
```

禁止只按页面路径查询评论、浏览量或点赞。

## 数据库迁移

```bash
pnpm --filter @marku/api migrate
```

等价于：

```bash
cd apps/api
go run ./cmd/migrate
```

当前迁移模型为 `Admin`、`Site`、`Comment`、`PageStat`、`PageViewDaily`、`PageViewVisitor`、`LikeStat`、`LikeRecord` 和 `LikeDaily`。启动时如果 `MARKU_DB_AUTO_MIGRATE=true` 会自动执行。

## 配置说明

完整变量见 `.env.example`。生产环境必须设置：

- `MARKU_JWT_SECRET`：至少 32 个字符
- `MARKU_ADMIN_EMAIL`
- `MARKU_ADMIN_PASSWORD`：至少 8 个字符
- `MARKU_IP_HASH_SALT`：至少 16 个字符

开发环境如果设置了管理员账号，同样必须设置 `MARKU_JWT_SECRET`。

评论相关变量：

```dotenv
MARKU_COMMENT_RATE_LIMIT_WINDOW=10m
MARKU_COMMENT_RATE_LIMIT_MAX=5
MARKU_COMMENT_MIN_SUBMIT_DURATION=2s
MARKU_COMMENT_MAX_LENGTH=2000
MARKU_COMMENT_MAX_LINKS=3
MARKU_COMMENT_MAX_DEPTH=3
MARKU_COMMENT_DUPLICATE_WINDOW=10m
MARKU_IP_HASH_SALT=<random-string>
```

浏览量相关变量：

```dotenv
MARKU_VIEW_DEDUPE_BY_VISITOR_DAY=false
MARKU_VIEW_FILTER_BOTS=true
MARKU_VIEW_MAX_BATCH_PATHS=50
```

数据库驱动：

```dotenv
# SQLite
MARKU_DB_DRIVER=sqlite
MARKU_DB_DSN=./data/marku.db

# MySQL
MARKU_DB_DRIVER=mysql
MARKU_DB_DSN=user:password@tcp(127.0.0.1:3306)/marku?charset=utf8mb4&parseTime=True&loc=Local

# PostgreSQL
MARKU_DB_DRIVER=postgres
MARKU_DB_DSN=host=127.0.0.1 user=marku password=secret dbname=marku port=5432 sslmode=disable
```

全局 IP 限流是单进程内存实现，仅作用于 `/api/v1`，多实例部署时需要改为网关限流或 Redis 共享限流。评论频率限制基于数据库查询，会随实例共享。

公共路径允许任意浏览器 Origin 发起请求，以满足多站点静态博客；实际写接口仍会通过站点 `allowedOrigins` 做 Origin 校验。CORS 不是身份认证。

## 脚本

```bash
pnpm --filter @marku/api dev
pnpm --filter @marku/api migrate
pnpm --filter @marku/api build
pnpm --filter @marku/api test
pnpm --filter @marku/api lint
```

## 注意

Go module path 暂定为 `marku/api`。确定仓库组织域名后，应统一替换为正式 module path。
