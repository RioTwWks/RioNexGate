#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BACKUP_DIR="${BACKUP_DIR:-$ROOT/backups}"
STAMP="$(date +%Y%m%d_%H%M%S)"
ARCHIVE="$BACKUP_DIR/rionexgate-data-$STAMP.tar.gz"

mkdir -p "$BACKUP_DIR"
# WARNING: archive contains SQLite DB, Reality/WG private keys, and optional TLS keys — unencrypted.
# Restrict permissions on backups/ and prefer encrypting before off-host copy (e.g. age/gpg).
tar -czf "$ARCHIVE" -C "$ROOT" data
chmod 600 "$ARCHIVE" 2>/dev/null || true
echo "Backup saved to $ARCHIVE (contains secrets — store securely)"
