#!/usr/bin/env bash
set -euo pipefail

MARKU_DIR="${MARKU_DIR:-/opt/marku}"
CONF_FILE="${MARKU_DIR}/install.conf"
ENV_FILE="${MARKU_DIR}/.env"
DEFAULT_IMAGE="ghcr.io/zwying0814/marku:0.3.2"
COMMAND="${1:-install}"
PURGE="${2:-}"

log()  { printf '\033[1;32m[marku]\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33m[marku]\033[0m %s\n' "$*" >&2; }
die()  { printf '\033[1;31m[marku]\033[0m %s\n' "$*" >&2; exit 1; }

read_conf() {
  local value=""
  if [ -f "$CONF_FILE" ]; then
    value=$(grep -E "^$1=" "$CONF_FILE" | tail -n 1 | cut -d= -f2- || true)
  fi
  printf '%s' "$value"
}

env_has() {
  [ -f "$ENV_FILE" ] && grep -qE "^$1=" "$ENV_FILE"
}

ensure_env() {
  local key="$1" value="$2"
  if ! env_has "$key"; then
    printf '%s=%s\n' "$key" "$value" >>"$ENV_FILE"
  fi
}

random_secret() {
  openssl rand -hex "$1"
}

need_docker() {
  command -v docker >/dev/null 2>&1 || die "未找到 docker，请先安装 Docker"
  command -v openssl >/dev/null 2>&1 || die "未找到 openssl，请先安装 openssl"
  docker info >/dev/null 2>&1 || die "无法访问 Docker，请用 root 或加入 docker 组后重试"
}

load_config() {
  IMAGE="${MARKU_IMAGE:-$(read_conf MARKU_IMAGE)}"
  IMAGE="${IMAGE:-$DEFAULT_IMAGE}"
  CONTAINER="${MARKU_CONTAINER:-$(read_conf MARKU_CONTAINER)}"
  CONTAINER="${CONTAINER:-marku}"
  VOLUME="${MARKU_VOLUME:-$(read_conf MARKU_VOLUME)}"
  VOLUME="${VOLUME:-marku-data}"
  PORT="${MARKU_PORT:-$(read_conf MARKU_PORT)}"
  PORT="${PORT:-36580}"
  BIND="${MARKU_BIND:-$(read_conf MARKU_BIND)}"
  BIND="${BIND:-0.0.0.0}"

  case "$PORT" in
    ''|*[!0-9]*) die "MARKU_PORT 必须是数字：$PORT" ;;
  esac
  if [ "$PORT" -lt 1 ] || [ "$PORT" -gt 65535 ]; then
    die "MARKU_PORT 超出范围：$PORT"
  fi
}

write_conf() {
  mkdir -p "$MARKU_DIR"
  chmod 700 "$MARKU_DIR" 2>/dev/null || true
  cat >"$CONF_FILE" <<EOF
MARKU_IMAGE=$IMAGE
MARKU_CONTAINER=$CONTAINER
MARKU_VOLUME=$VOLUME
MARKU_PORT=$PORT
MARKU_BIND=$BIND
EOF
  chmod 600 "$CONF_FILE"
}

prepare_env() {
  mkdir -p "$MARKU_DIR"
  chmod 700 "$MARKU_DIR" 2>/dev/null || true

  if [ ! -f "$ENV_FILE" ]; then
    NEW_ENV=1
    JWT_SECRET="${MARKU_JWT_SECRET:-$(random_secret 48)}"
    IP_SALT="${MARKU_IP_HASH_SALT:-$(random_secret 32)}"

    cat >"$ENV_FILE" <<EOF
MARKU_ENV=production
MARKU_HTTP_ADDR=:36580
MARKU_DB_DRIVER=sqlite
MARKU_DB_DSN=/data/marku.db
MARKU_DB_AUTO_MIGRATE=true
MARKU_JWT_SECRET=${JWT_SECRET}
MARKU_IP_HASH_SALT=${IP_SALT}
MARKU_CORS_ALLOWED_ORIGINS=${MARKU_CORS_ALLOWED_ORIGINS:-}
MARKU_HTTP_TRUSTED_PROXIES=${MARKU_HTTP_TRUSTED_PROXIES:-127.0.0.1}
EOF
  else
    NEW_ENV=0
    ensure_env MARKU_ENV production
    ensure_env MARKU_HTTP_ADDR :36580
    ensure_env MARKU_DB_DRIVER sqlite
    ensure_env MARKU_DB_DSN /data/marku.db
    ensure_env MARKU_DB_AUTO_MIGRATE true
    ensure_env MARKU_JWT_SECRET "$(random_secret 48)"
    ensure_env MARKU_IP_HASH_SALT "$(random_secret 32)"
    ensure_env MARKU_CORS_ALLOWED_ORIGINS "${MARKU_CORS_ALLOWED_ORIGINS:-}"
    ensure_env MARKU_HTTP_TRUSTED_PROXIES "${MARKU_HTTP_TRUSTED_PROXIES:-127.0.0.1}"
  fi

  chmod 600 "$ENV_FILE"
}

pull_image() {
  log "拉取镜像 $IMAGE"
  docker pull "$IMAGE" >/dev/null
}

run_container() {
  docker rm -f "$CONTAINER" >/dev/null 2>&1 || true
  log "启动容器 ${CONTAINER}（${BIND}:${PORT} -> 36580）"
  docker run -d \
    --name "$CONTAINER" \
    --restart unless-stopped \
    --env-file "$ENV_FILE" \
    -p "${BIND}:${PORT}:36580" \
    -v "${VOLUME}:/data" \
    "$IMAGE" >/dev/null
}

wait_healthy() {
  local status
  for _ in $(seq 1 40); do
    status=$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$CONTAINER" 2>/dev/null || true)
    case "$status" in
      healthy|running) return 0 ;;
    esac
    sleep 1
  done

  warn "容器未在预期时间内就绪，最近日志："
  docker logs --tail 50 "$CONTAINER" || true
  return 1
}

print_result() {
  echo
  log "Marku 已启动"
  log "管理后台: http://<服务器IP>:${PORT}/"
  log "API:      http://<服务器IP>:${PORT}/api/v1"
  log "配置文件: ${ENV_FILE}"
  log "数据卷:   ${VOLUME}"
  echo
  warn "首次部署默认账号: admin / 123456"
  warn "登录后请立即在「账号与安全」中修改账号和密码。"
  if [ "${NEW_ENV:-0}" != "1" ]; then
    log "已沿用现有配置与数据库；默认密码仅在首次初始化时生效。"
  fi
  echo
}

cmd_install() {
  load_config
  write_conf
  prepare_env
  pull_image
  run_container
  wait_healthy
  print_result
}

cmd_upgrade() {
  load_config
  [ -f "$ENV_FILE" ] || die "未找到 ${ENV_FILE}，请先执行 install"
  write_conf
  NEW_ENV=0
  pull_image
  run_container
  wait_healthy
  log "已升级到 ${IMAGE}（数据卷 ${VOLUME} 保持不变）"
}

cmd_status() {
  load_config
  if ! docker inspect "$CONTAINER" >/dev/null 2>&1; then
    die "容器 $CONTAINER 不存在"
  fi
  docker ps -a --filter "name=^/${CONTAINER}$" \
    --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}'
  docker inspect -f 'health={{if .State.Health}}{{.State.Health.Status}}{{else}}n/a{{end}}' "$CONTAINER"
}

cmd_logs() {
  load_config
  docker logs --tail 200 "$CONTAINER"
}

cmd_backup() {
  load_config
  mkdir -p "$MARKU_DIR/backups"
  local name
  name="marku-$(date +%Y%m%d-%H%M%S).tar.gz"
  log "备份数据卷 $VOLUME 到 $MARKU_DIR/backups/$name"
  docker run --rm --entrypoint tar \
    -v "${VOLUME}:/data:ro" \
    -v "$MARKU_DIR/backups:/backup" \
    "$IMAGE" czf "/backup/$name" -C /data .
  ls -lh "$MARKU_DIR/backups/$name"
}

cmd_uninstall() {
  load_config
  docker rm -f "$CONTAINER" >/dev/null 2>&1 || true
  log "已删除容器 ${CONTAINER}，数据卷 ${VOLUME} 保留"
  if [ "$PURGE" = "--purge" ]; then
    docker volume rm "$VOLUME" >/dev/null 2>&1 || true
    log "已删除数据卷 $VOLUME"
  fi
}

usage() {
  cat <<EOF
用法: $0 <command>

命令:
  install     安装并启动（默认）
  upgrade     拉取当前配置的镜像并重建容器
  status      查看容器状态与健康检查
  logs        查看最近 200 行日志
  backup      备份数据卷到 ${MARKU_DIR}/backups
  uninstall   删除容器，保留数据卷；加 --purge 同时删除卷

可用环境变量:
  MARKU_IMAGE       镜像地址，默认 ${DEFAULT_IMAGE}
  MARKU_DIR         配置目录，默认 /opt/marku
  MARKU_CONTAINER   容器名，默认 marku
  MARKU_VOLUME      数据卷名，默认 marku-data
  MARKU_PORT        宿主机端口，默认 36580
  MARKU_BIND        监听地址，默认 0.0.0.0；反代时可设为 127.0.0.1
  MARKU_JWT_SECRET / MARKU_IP_HASH_SALT
  MARKU_CORS_ALLOWED_ORIGINS
  MARKU_HTTP_TRUSTED_PROXIES

示例:
  bash install.sh
  MARKU_PORT=36581 bash install.sh install
  MARKU_IMAGE=ghcr.io/zwying0814/marku:0.3.2 bash install.sh upgrade
  bash install.sh backup
EOF
}

need_docker

case "$COMMAND" in
  install)   cmd_install ;;
  upgrade)   cmd_upgrade ;;
  status)    cmd_status ;;
  logs)      cmd_logs ;;
  backup)    cmd_backup ;;
  uninstall) cmd_uninstall ;;
  help|-h|--help) usage ;;
  *) usage; exit 1 ;;
esac
