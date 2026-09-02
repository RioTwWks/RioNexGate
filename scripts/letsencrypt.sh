#!/usr/bin/env bash
# Example: obtain Let's Encrypt certificates for rionexgate panel (nginx).
#
# Prerequisites:
#   - Domain DNS points to this host
#   - Port 80 is reachable from the internet (for HTTP-01 challenge)
#   - certbot installed: sudo apt install certbot
#
# Usage:
#   ./scripts/letsencrypt.sh panel.example.com
#
# After success, certs are copied to data/nginx/ssl/ for the nginx container.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DOMAIN="${1:-}"

if [[ -z "$DOMAIN" ]]; then
  echo "Usage: $0 <domain>" >&2
  exit 1
fi

SSL_DIR="${ROOT}/data/nginx/ssl"
mkdir -p "$SSL_DIR"

echo "==> Requesting certificate for ${DOMAIN} (standalone mode)"
echo "    Stop anything listening on :80 before continuing (e.g. make down)"
read -r -p "Press Enter to continue or Ctrl+C to abort..."

sudo certbot certonly --standalone \
  -d "$DOMAIN" \
  --agree-tos \
  --email "admin@${DOMAIN}" \
  --non-interactive \
  --preferred-challenges http

LE_DIR="/etc/letsencrypt/live/${DOMAIN}"
sudo cp "${LE_DIR}/fullchain.pem" "${SSL_DIR}/fullchain.pem"
sudo cp "${LE_DIR}/privkey.pem" "${SSL_DIR}/privkey.pem"
sudo chown "$(whoami):" "${SSL_DIR}"/*.pem

echo ""
echo "Certificates copied to ${SSL_DIR}"
echo ""
echo "Next steps:"
echo "  1. cp nginx/nginx-https.conf.example nginx/nginx.conf"
echo "  2. Uncomment HTTPS_PORT in docker-compose.yml and .env"
echo "  3. make up"
echo ""
echo "Renewal (add to crontab):"
echo "  0 3 * * * certbot renew --quiet && cp /etc/letsencrypt/live/${DOMAIN}/*.pem ${SSL_DIR}/ && docker compose restart nginx"
