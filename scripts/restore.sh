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

# Only allow members under data/; reject absolute paths and .. traversal.
while IFS= read -r member; do
  [ -z "$member" ] && continue
  case "$member" in
    data|data/*) ;;
    *)
      echo "Refusing to restore: archive member outside data/: $member" >&2
      exit 1
      ;;
  esac
  case "$member" in
    *..*|/*)
      echo "Refusing to restore: unsafe archive member: $member" >&2
      exit 1
      ;;
  esac
done < <(tar -tzf "$ARCHIVE")

tar -xzf "$ARCHIVE" -C "$ROOT" -- data
echo "Restored data/ from $ARCHIVE"
