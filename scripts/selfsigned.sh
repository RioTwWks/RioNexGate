#!/usr/bin/env bash
# Generate a self-signed TLS certificate for the rionexgate nginx reverse proxy.
#
# Use scripts/letsencrypt.sh for public hosts — browsers and clients trust it.
# A self-signed certificate is acceptable for LAN/VPN-only panels and testing;
# clients will show a trust warning and some subscription apps must allow it.
#
# Usage:
#   ./scripts/selfsigned.sh [domain-or-ip] [days]
#   ./scripts/selfsigned.sh panel.lan 825
#
# Output (consumed by nginx-https.conf.example via the compose volume mount
# ./data/nginx/ssl -> /etc/nginx/ssl:ro):
#   data/nginx/ssl/fullchain.pem
#   data/nginx/ssl/privkey.pem   (0600)

set -euo pipefail

DOMAIN="${1:-localhost}"
DAYS="${2:-825}"
OUT_DIR="$(cd "$(dirname "$0")/.." && pwd)/data/nginx/ssl"

mkdir -p "$OUT_DIR"

SAN="DNS:${DOMAIN}"
if [[ "$DOMAIN" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
	SAN="IP:${DOMAIN}"
fi

openssl req -x509 -newkey rsa:2048 -nodes \
	-keyout "$OUT_DIR/privkey.pem" \
	-out "$OUT_DIR/fullchain.pem" \
	-days "$DAYS" \
	-subj "/CN=${DOMAIN}" \
	-addext "subjectAltName=${SAN}"

chmod 600 "$OUT_DIR/privkey.pem"
chmod 644 "$OUT_DIR/fullchain.pem"

cat <<EOF
Self-signed certificate for "${DOMAIN}" written to:
  ${OUT_DIR}/fullchain.pem
  ${OUT_DIR}/privkey.pem

Next steps:
  1. cp nginx/nginx-https.conf.example nginx/nginx.conf
  2. In docker-compose.yml uncomment the HTTPS_PORT mapping (host -> container 8443)
  3. make up   # panel at https://${DOMAIN}:8443 (expect a trust warning)
EOF
