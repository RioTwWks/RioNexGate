#!/usr/bin/env bash
set -euo pipefail

echo "=== Docker doctor ==="
echo

if grep -q 'credsStore.*desktop\|docker-credential-desktop' "$HOME/.docker/config.json" 2>/dev/null; then
  echo "FAIL: ~/.docker/config.json still uses Docker Desktop credential helper"
  echo "  Remove \"credsStore\": \"desktop\" from ~/.docker/config.json"
  echo "  Or run: echo '{\"auths\":{}}' > ~/.docker/config.json"
  exit 1
fi

unset DOCKER_HOST

if docker info >/dev/null 2>&1; then
  echo "OK: docker info"
elif sg docker -c "docker info" >/dev/null 2>&1; then
  echo "OK: docker info (via sg docker — log out/in to fix shell group)"
else
  echo "FAIL: cannot connect to Docker daemon"
  echo "  sudo systemctl start docker"
  echo "  sudo usermod -aG docker \$USER  # then log out and back in"
  echo "  Or: make dev-local"
  exit 1
fi

if systemctl is-active docker >/dev/null 2>&1; then
  echo "OK: docker.service is active"
else
  echo "WARN: docker.service not active (or not using systemd)"
fi

echo
echo "Testing image store..."
docker pull hello-world:latest >/dev/null
if docker image inspect hello-world:latest >/dev/null 2>&1; then
  echo "OK: docker pull + image inspect"
  docker rmi hello-world:latest >/dev/null 2>&1 || true
else
  echo "FAIL: image inspect after pull"
  exit 1
fi

echo
echo "Testing container run..."
docker run --rm hello-world >/dev/null
echo "OK: docker run"

echo
echo "All checks passed. Run: make dev"
