# 部署脚本

`install.sh` 是主机侧的安装与维护脚本，负责拉取官方镜像、生成配置、创建容器、
备份数据卷。它不进入镜像，也不需要把 Docker socket 暴露给容器。

## 快速安装

```bash
curl -fsSL https://raw.githubusercontent.com/zwying0814/marku/main/deploy/install.sh -o install.sh
less install.sh
bash install.sh
```

默认行为：

- 配置目录：`/opt/marku`（`.env` 与 `install.conf`）
- 数据卷：`marku-data`
- 端口：`36580`
- 镜像：`ghcr.io/zwying0814/marku:0.4.0`
- 首次安装自动生成 JWT 密钥与 IP 盐
- 默认管理员：`admin` / `123456`，登录后请立即修改

## 常用命令

```bash
bash install.sh status
bash install.sh logs
bash install.sh backup
bash install.sh upgrade
bash install.sh uninstall
bash install.sh uninstall --purge
```

`uninstall` 默认保留数据卷，`--purge` 会连数据卷一起删除。

## 自定义

```bash
# 只监听本机，交给 1Panel / Nginx 反代
MARKU_BIND=127.0.0.1 bash install.sh

# 安装指定版本
MARKU_IMAGE=ghcr.io/zwying0814/marku:0.3.2 bash install.sh upgrade

# 更换端口与数据卷
MARKU_PORT=36581 MARKU_VOLUME=marku-prod-data bash install.sh
```

重复执行 `install` 会沿用已有的 `.env` 与数据库，不会重置账号或密码。

## 为什么不做进镜像

- 安装脚本需要操作宿主机的 Docker、端口、数据卷和 `/opt/marku`；
- 容器内执行就必须挂载 `/var/run/docker.sock`，等于把宿主机 root 权限交给容器；
- 官方镜像应保持“只跑应用”，升级与备份放在宿主机侧更安全，也更容易审计。

镜像本身已包含应用、内嵌 Admin 与 `/embed/v1` 组件，不需要额外脚本。
