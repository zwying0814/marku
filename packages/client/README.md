# @marku/client

无 UI、框架无关的 Marku 公共 API 客户端。覆盖站点信息、评论、浏览量、点赞与聚合统计。

```ts
import { MarkuClient } from '@marku/client'

const marku = new MarkuClient({
  baseUrl: 'https://api.example.com/api/v1',
  siteKey: 'mk_your_site_key',
})

const post = marku.forPath('/posts/hello')
const stats = await post.stats()
const comments = await post.comments.list({ page: 1, pageSize: 20 })
```

## 能力

- 统一 `X-Site-Key`、JSON 信封解析和 `{ code, message, data }` 错误映射。
- 自动生成并持久化匿名 `visitorId`（`marku:visitor-id`）。
- 评论列表、提交、计数；浏览量记录与批量读取；点赞切换、设置与状态查询。
- `forPath()` 绑定文章路径，避免重复传 `path`。
- 超时、外部 `AbortSignal`、自定义 `fetch`。

构建产物在 `dist/`，类型定义随包发布。完整说明见 `apps/web/docs/guide/client.md`。
