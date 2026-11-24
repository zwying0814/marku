---
title: 备份与恢复
description: SQLite、PostgreSQL、MySQL 的备份策略与恢复步骤。
outline: [2, 3]
---

# 备份与恢复

## 备份范围

需要备份：

- 数据库
- 环境变量 / 密钥
- 反向代理与 TLS 配置
- 站点 Site Key 的使用记录（不是完整 Key 本身）

Marku 数据库只保存 Site Key 哈希，无法从数据库恢复完整 Key。遗失后只能轮换。

## SQLite

### 主机运行时

```bash
sqlite3 apps/api/data/marku.db ".backup '/backup/marku-$(date +%F).db'"
```

`.backup` 比直接复制正在写入的数据库更安全。

### Docker volume

最简单的一致性备份方式：

```bash
docker compose stop api

docker run --rm \
  -v marku_marku-data:/data:ro \
  -v "$PWD:/backup" \
  alpine sh -c 'tar czf /backup/marku-data.tar.gz -C /data .'

docker compose start api
```

### 恢复

```bash
docker compose stop api

docker run --rm \
  -v marku_marku-data:/data \
  -v "$PWD:/backup" \
  alpine sh -c 'rm -rf /data/* && tar xzf /backup/marku-data.tar.gz -C /data'

docker compose start api
curl http://localhost:36580/readyz
```

## PostgreSQL

```bash
docker compose exec -T postgres \
  pg_dump -U "${MARKU_POSTGRES_USER:-marku}" "${MARKU_POSTGRES_DB:-marku}" \
  > "marku-$(date +%F).sql"
```

恢复：

```bash
docker compose stop api

cat marku-2026-11-24.sql | docker compose exec -T postgres \
  psql -U "${MARKU_POSTGRES_USER:-marku}" "${MARKU_POSTGRES_DB:-marku}"

docker compose start api
```

定期使用 `pg_dump -Fc` 做压缩备份，并验证 `pg_restore --list`。

## MySQL

```bash
mysqldump --single-transaction --routines --triggers \
  -h 127.0.0.1 -u marku -p marku > "marku-$(date +%F).sql"
```

恢复：

```bash
mysql -h 127.0.0.1 -u marku -p marku < marku-2026-11-24.sql
```

## 备份策略

建议：

- 每日自动备份。
- 至少保留 7 个日备份、4 个周备份。
- 备份文件加密并异地保存。
- 每季度执行一次真实恢复演练。
- 恢复后验证 `/readyz`、管理员登录、评论列表和统计数据。

## 恢复后检查

1. API 健康检查通过。
2. Admin 可以登录。
3. 站点列表和评论数量正确。
4. 公共 API 使用现有 Site Key 可以访问。
5. 写入一条测试评论或浏览量并确认成功。
