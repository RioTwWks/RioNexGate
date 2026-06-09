#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BACKUP_DIR="${BACKUP_DIR:-$ROOT/backups}"
STAMP="$(date +%Y%m%d_%H%M%S)"
ARCHIVE="$BACKUP_DIR/proxy-mgr-data-$STAMP.tar.gz"

mkdir -p "$BACKUP_DIR"
tar -czf "$ARCHIVE" -C "$ROOT" data
echo "Backup saved to $ARCHIVE"
