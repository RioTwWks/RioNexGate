#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BACKUP_DIR="${BACKUP_DIR:-$ROOT/backups}"
STAMP="$(date +%Y%m%d_%H%M%S)"
ARCHIVE="$BACKUP_DIR/rionexgate-data-$STAMP.tar.gz"

mkdir -p "$BACKUP_DIR"
chmod 700 "$BACKUP_DIR" 2>/dev/null || true

# Archive contains SQLite DB, Reality/WG private keys, and optional TLS keys.
tar -czf "$ARCHIVE" -C "$ROOT" data
chmod 600 "$ARCHIVE" 2>/dev/null || true

encrypt_with_age() {
  local src="$1"
  local dst="$2"
  local recipient="${AGE_RECIPIENT:-${AGE_PUBLIC_KEY:-}}"
  if [ -z "$recipient" ]; then
    return 1
  fi
  if ! command -v age >/dev/null 2>&1; then
    echo "age not found; install https://github.com/FiloSottile/age (leaving unencrypted archive)" >&2
    return 1
  fi
  age -r "$recipient" -o "$dst" "$src"
  chmod 600 "$dst" 2>/dev/null || true
  rm -f "$src"
  echo "Encrypted backup saved to $dst"
  return 0
}

if encrypt_with_age "$ARCHIVE" "${ARCHIVE}.age"; then
  exit 0
fi

echo "Backup saved to $ARCHIVE (contains secrets — set AGE_RECIPIENT to encrypt with age)"
