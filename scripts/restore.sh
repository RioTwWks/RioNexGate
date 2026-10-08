#!/usr/bin/env bash
set -euo pipefail

if [ $# -lt 1 ]; then
  echo "Usage: $0 <backup.tar.gz|backup.tar.gz.age>"
  echo "  For .age archives: set AGE_IDENTITY to a private key file (age identity)."
  exit 1
fi

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ARCHIVE="$1"
TMP=""

cleanup() {
  if [ -n "$TMP" ] && [ -f "$TMP" ]; then
    rm -f "$TMP"
  fi
}
trap cleanup EXIT

if [ ! -f "$ARCHIVE" ]; then
  echo "Archive not found: $ARCHIVE"
  exit 1
fi

case "$ARCHIVE" in
  *.age)
    if ! command -v age >/dev/null 2>&1; then
      echo "age not found; install https://github.com/FiloSottile/age" >&2
      exit 1
    fi
    if [ -z "${AGE_IDENTITY:-}" ] || [ ! -f "$AGE_IDENTITY" ]; then
      echo "Set AGE_IDENTITY to an age identity file for encrypted restore" >&2
      exit 1
    fi
    TMP="$(mktemp "${TMPDIR:-/tmp}/rionexgate-restore.XXXXXX.tar.gz")"
    age -d -i "$AGE_IDENTITY" -o "$TMP" "$ARCHIVE"
    ARCHIVE="$TMP"
    ;;
esac

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
# Ensure backend (uid 1000) can write after restore from root.
if [ "$(id -u)" -eq 0 ]; then
  chown -R 1000:1000 "$ROOT/data" 2>/dev/null || true
fi
echo "Restored data/ from $1"
