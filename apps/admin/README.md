# @marku/admin

Marku 的管理后台，基于 React 19、Vite、TypeScript 和 Tailwind CSS。

## 功能

- 管理员登录、路由守卫、401 自动退出
- 全局概览、浏览量/点赞/评论趋势
- 多站点创建、编辑、停用、删除
- Site Key 一次性展示、复制、轮换与吊销
- 评论筛选、详情抽屉、审核、回复和批量操作
- 数据分析与热门页面排行
- 响应式布局与明暗主题

## 开发

在仓库根目录执行：

```bash
pnpm install
pnpm dev:admin
```

默认地址：`http://localhost:36581`

开发服务器默认把 `/api/v1` 代理到 `http://localhost:36580`。如需指向其他
API，可设置 `VITE_API_BASE_URL` 或 `MARKU_API_PROXY_TARGET`。

## 脚本

```bash
pnpm --filter @marku/admin dev
pnpm --filter @marku/admin build
pnpm --filter @marku/admin test
pnpm --filter @marku/admin typecheck
```

登录账号由 API 的 `MARKU_ADMIN_EMAIL` 和 `MARKU_ADMIN_PASSWORD` 首次启动时创建。
JWT 存放在当前标签页的 `sessionStorage`，关闭标签页或退出登录后清除。

## 环境变量

复制 `.env.example` 为 `.env.local`，按需修改：

```dotenv
VITE_API_BASE_URL=http://localhost:36580/api/v1
```
