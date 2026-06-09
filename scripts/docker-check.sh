#!/usr/bin/env bash
set -euo pipefail
unset DOCKER_HOST

if docker info >/dev/null 2>&1; then
  exit 0
fi

if sg docker -c "docker info" >/dev/null 2>&1; then
  echo "Note: 'docker' group not active in this shell; compose will use sg docker."
  echo "      For a permanent fix: log out and log back in."
  exit 0
fi

echo "Error: cannot connect to Docker daemon." >&2
echo "  sudo systemctl start docker" >&2
echo "  sudo usermod -aG docker \$USER  # then log out and back in" >&2
echo "  groups  # should list 'docker'" >&2
echo "  Or: make dev-local" >&2
exit 1
