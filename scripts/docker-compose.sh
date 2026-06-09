#!/usr/bin/env bash
# Wrapper: clears stale DOCKER_HOST, uses sg docker if group not active in shell
set -euo pipefail

unset DOCKER_HOST

# Prefer Docker Compose v2 plugin (docker compose), NOT legacy Python docker-compose v1
if docker compose version >/dev/null 2>&1; then
  COMPOSE=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
  echo "Warning: using legacy docker-compose v1; install plugin: docker-compose-plugin" >&2
  COMPOSE=(docker-compose)
else
  echo "Error: docker compose not found. Install docker-compose-plugin." >&2
  exit 1
fi

run() {
  "${COMPOSE[@]}" "$@"
}

if docker info >/dev/null 2>&1; then
  run "$@"
elif sg docker -c "docker info" >/dev/null 2>&1; then
  quoted=$(printf ' %q' "${COMPOSE[@]}" "$@")
  exec sg docker -c "${quoted# }"
else
  echo "Error: cannot connect to Docker daemon." >&2
  echo "  sudo systemctl start docker" >&2
  echo "  sudo usermod -aG docker \$USER  # then log out and back in" >&2
  echo "  Or: make dev-local" >&2
  exit 1
fi
