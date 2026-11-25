---
title: v0.3 发布检查
description: 发布前验证、部署步骤与回滚预案。
outline: [2, 3]
---

# v0.3 发布检查

## 发布前

- [ ] `pnpm install --frozen-lockfile` 成功。
- [ ] `gofmt -l .` 无输出。
- [ ] `go vet ./...` 通过。
- [ ] `go test -race ./...` 通过。
- [ ] `pnpm typecheck` 通过。
- [ ] `pnpm test` 通过。
- [ ] `pnpm build` 通过。
- [ ] `pnpm build:binary` 通过。
- [ ] `pnpm --filter @marku/web build` 通过，官网静态产物已部署到独立域名或 CDN。
- [ ] 纯 HTML 示例脚本通过 `node --check`。
- [ ] OpenAPI 与共享类型已重新生成且无差异。
- [ ] Docker 镜像构建成功。

## 安全配置

- [ ] `MARKU_ENV=production`。
- [ ] `MARKU_JWT_SECRET` 不少于 32 字符且未提交到仓库。
- [ ] `MARKU_IP_HASH_SALT` 已设置。
- [ ] 管理员初始密码为强密码。
- [ ] 管理端已加 IP allowlist、VPN 或 Basic Auth。
- [ ] Admin、官网与 API 使用 HTTPS。
- [ ] 每个站点都配置了正式 `allowedOrigins`。
- [ ] 反向代理 IP 已加入 `MARKU_HTTP_TRUSTED_PROXIES`。

## 功能验收

- [ ] 可以创建至少两个站点。
- [ ] 每个站点获得独立 Site Key。
- [ ] 两个站点的评论、浏览量和点赞互不影响。
- [ ] 评论可以提交、审核、回复和批量处理。
- [ ] 自动公开与先审后发都符合预期。
- [ ] 浏览量、独立访客与 Bot 过滤符合预期。
- [ ] 点赞 `PUT` / `DELETE` 幂等。
- [ ] Site Key 轮换后旧 Key 立即失效。
- [ ] 吊销 Key 后公共 API 返回 `401`。
- [ ] 禁用站点后公共 API 返回 `403`。
- [ ] 纯 HTML 示例完成评论、浏览量和点赞。
- [ ] Admin 的 7 天 / 30 天趋势和热门页面正确。

## 数据与运维

- [ ] 备份任务已启用。
- [ ] 已执行一次恢复演练。
- [ ] 日志采集包含 `X-Request-ID`。
- [ ] 磁盘、内存和数据库连接数有监控。
- [ ] 429、5xx 和 `readyz` 失败有告警。
- [ ] 升级与回滚步骤已记录。

## 发布步骤

```bash
pnpm install --frozen-lockfile
pnpm typecheck
pnpm test
pnpm build
pnpm build:binary

docker compose build
docker compose up -d
curl -f https://api.example.com/readyz
curl -f https://marku.example.com/
curl -f https://marku.example.com/swagger
curl -f https://marku.example.com/embed/v1/marku-embed.js
```

## 回滚

1. 停止新版本 API。
2. 恢复上一版本镜像或二进制。
3. 必要时恢复发布前数据库备份。
4. 检查 `/readyz`、管理员登录与公共 API。
5. 保留故障日志和请求 ID。

## 版本

当前版本：`0.3.0`

发布后更新：

- `CHANGELOG.md`
- Git tag
- 部署版本记录
