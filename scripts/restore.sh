#!/usr/bin/env bash
set -euo pipefail

if [ $# -lt 1 ]; then
  echo "Usage: $0 <backup.tar.gz>"
  exit 1
fi

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ARCHIVE="$1"

if [ ! -f "$ARCHIVE" ]; then
  echo "Archive not found: $ARCHIVE"
  exit 1
fi

tar -xzf "$ARCHIVE" -C "$ROOT"
echo "Restored data/ from $ARCHIVE"
