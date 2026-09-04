**English** | [Русский](README.ru.md)

# RioNexGate — proxy management panel

**RioNexGate** is a management panel for **Xray-core** and **sing-box** proxies: Go API, React UI, Telegram bot, and Docker Compose.

It supports user management, VLESS/VMess/Trojan links, QR codes, traffic limits, stealth anti-DPI presets, multi-hop nodes, and an optional [RioNexTunnel](https://github.com/RioTwWks/RioNexTunnel) client API (devices, subscription, telemetry).

## Features

- **Dual core**: **xray** and **sing-box** (switch in UI / API)
- **Users**: CRUD, traffic limits, expiry, connection links and QR codes (VLESS, VMess, Trojan)
- **Dashboard**: aggregate traffic stats and charts
- **Stealth / anti-DPI**: VLESS + Reality + XHTTP (`stream-one`), Vision (TCP), optional TLS, ServerHello fragmentation, AmneziaWG reserve — see [docs/stealth.md](docs/stealth.md)
- **Multi-hop nodes**: entry/exit chains (e.g. RU → EU) with health checks — see [docs/multihop.md](docs/multihop.md)
- **RioNexTunnel client API**: device registration, JSON config sync, subscription URLs, telemetry, remote commands
- **Telegram bot** for administrators
- **REST API** with API-key authentication; OpenAPI / Swagger UI
- **Nginx** as a single entry point

## Documentation

| Guide | Description |
|-------|-------------|
| [Stealth / anti-DPI](docs/stealth.md) | Reality, XHTTP, Vision, TLS, fragmentation, AmneziaWG |
| [Multi-hop chains](docs/multihop.md) | RU entry → EU exit topology, nodes API, credentials |
| [docs/README.md](docs/README.md) | Documentation index (EN / RU) |
| [OpenAPI / Swagger](http://localhost:8888/api/docs) | Interactive API reference (after `make dev`) |

## Requirements

- Linux (tested on Ubuntu 24.04)
- **Docker Engine** + **Docker Compose v2** plugin (`docker compose`, not Python `docker-compose` v1)
- Go 1.21+ and Node 18+ — only for local development without Docker

### Install Docker (Linux)

```bash
# Official Docker repo — see https://docs.docker.com/engine/install/ubuntu/
sudo apt install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo usermod -aG docker $USER
# log out and back in so the docker group applies
```

Verify:

```bash
docker compose version   # v2.x or v5.x
make docker-doctor
```

> **Do not use** the legacy `docker-compose` v1 package from apt (`sudo apt remove docker-compose`).  
> After removing Docker Desktop, reset `~/.docker/config.json`: `echo '{"auths":{}}' > ~/.docker/config.json`

## Quick start

```bash
git clone https://github.com/RioTwWks/RioNexGate.git
cd RioNexGate

make init
```

Edit:

- `backend/config.yaml` — change `server.api_key` (required)
- `.env` — `HTTP_PORT`, `TELEGRAM_BOT_TOKEN` (optional)

```bash
make dev
```

In another terminal (first run):

```bash
make migrate
```

### URLs after startup

| URL | Description |
|-----|-------------|
| http://localhost:8888 | **Web panel** (nginx, port from `HTTP_PORT`) |
| http://localhost:8888/api/health | API health check |
| http://localhost:8888/api/docs | Swagger UI |
| http://localhost:8080/api/health | Backend directly |
| http://localhost:3000 | Frontend directly (API not proxied) |

Sign in to the panel with the API key from `backend/config.yaml` → `server.api_key`.

Optional — start xray/sing-box cores:

```bash
make dev-cores
```

For stealth inbounds (host ports 443 / 8443), use the `cores` profile above. AmneziaWG reserve: `docker compose --profile awg up -d amneziawg`.

## Web panel

1. Open http://localhost:8888
2. Enter the API key from `backend/config.yaml`
3. **Dashboard** — total traffic and chart
4. **Users** — create, edit, links, QR, devices, chain binding
5. **Nodes** — entry/exit nodes, health checks, multi-hop topology
6. **Stealth** — Reality/XHTTP/Vision/TLS/fragmentation/AWG settings
7. **Settings** — switch xray ↔ sing-box, reload core

On 401 errors: click **Logout** or clear `localStorage` (`rionexgate_api_key`).

In the link modal you can choose **VLESS**, **VMess**, or **Trojan**. With stealth enabled, use `?all=true` for multi-profile links.

## RioNexTunnel client API

[RioNexTunnel](https://github.com/RioTwWks/RioNexTunnel) can register devices and sync JSON configs. Standard VLESS/VMess/Trojan clients keep working without changes.

| Endpoint | Auth | Description |
|----------|------|-------------|
| `POST /api/client/register` | none | Register device → `device_token`, `subscription_url` |
| `GET /api/client/config` | `X-Device-Token` | JSON config with `config_hash` and transport profiles |
| `POST /api/client/stats` | `X-Device-Token` | Session telemetry |
| `GET /api/client/commands` | `X-Device-Token` | Long-poll or SSE (`?stream=sse`) for `refresh_config` etc. |
| `GET /api/subscription/{token}` | none | Base64 subscription (newline-separated links) |
| `GET /api/users/{id}/devices` | `X-API-Key` | List registered devices |
| `DELETE /api/users/{id}/devices/{deviceId}` | `X-API-Key` | Revoke a device |

Example registration:

```bash
curl -X POST http://localhost:8888/api/client/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","label":"laptop"}'
```

## Stealth / anti-DPI (Reality + XHTTP)

The `core.stealth` section in `backend/config.yaml` enables DPI-resistant presets:

- **VLESS + Reality + XHTTP** (`mode: stream-one`) on port **443**
- **VLESS + Reality + Vision** (TCP) on port **8443**
- Optional **VLESS + TLS** on a non-standard port
- Optional **AmneziaWG** UDP reserve

Details: [docs/stealth.md](docs/stealth.md) · [docs/stealth.ru.md](docs/stealth.ru.md)

## Multi-hop chains (RU entry → EU exit)

Route users through an entry server (e.g. Russia) and exit from another region (e.g. EU). Clients connect only to the entry; the panel generates Xray outbounds on the entry host.

Details: [docs/multihop.md](docs/multihop.md) · [docs/multihop.ru.md](docs/multihop.ru.md)

```bash
# Register exit node and bind user (see docs for full credentials JSON)
curl -X POST http://localhost:8888/api/nodes -H "X-API-Key: YOUR_KEY" -H "Content-Type: application/json" \
  -d '{"name":"exit-eu","address":"eu.example.com","port":8443,"role":"exit","credentials":"{...}"}'
curl -X PUT http://localhost:8888/api/users/1/chain -H "X-API-Key: YOUR_KEY" -H "Content-Type: application/json" \
  -d '{"entry_node_id":1,"exit_node_id":2}'
```

### Generate Reality keypair

```bash
# In the xray container or with xray installed:
xray x25519
# PrivateKey → core.stealth.reality.private_key
# Password   → core.stealth.reality.public_key (pbk in client links)
```

Multi-profile links via API:

```bash
curl -H "X-API-Key: YOUR_KEY" "http://localhost:8888/api/users/1/link?all=true"
```

## Makefile

| Command | Action |
|---------|--------|
| `make init` | Create `config.yaml`, `.env`, `data/` directories |
| `make dev` | Build and run (`docker compose up --build`) |
| `make up` | Run detached |
| `make down` | Stop containers |
| `make dev-cores` | Start xray-core and sing-box (profile `cores`) |
| `make dev-local` | Hint for running without Docker |
| `make migrate` | Run DB migration in backend container |
| `make docker-doctor` | Docker diagnostics |
| `make test` | `go test` + `npm test` |
| `make test-e2e` | Playwright E2E (backend + frontend dev) |
| `make logs` | Compose logs |
| `make clean` | Stop and remove volumes |

## OpenAPI / Swagger

After starting the panel: http://localhost:8888/api/docs  
Spec: http://localhost:8888/api/openapi.yaml

For **Try it out** in Swagger UI, use the key from `localStorage` (`rionexgate_api_key`) if you are already signed in.

## HTTPS (Let's Encrypt)

Example TLS setup for nginx:

1. Ensure the domain points to the server and port 80 is free.
2. Obtain a certificate:
   ```bash
   chmod +x scripts/letsencrypt.sh
   ./scripts/letsencrypt.sh panel.example.com
   ```
3. Enable HTTPS nginx config:
   ```bash
   cp nginx/nginx-https.conf.example nginx/nginx.conf
   ```
4. In `docker-compose.yml`, uncomment `HTTPS_PORT` (default `8443`).
5. `make up` — panel at `https://panel.example.com:8443`.

See comments in [`nginx/nginx-https.conf.example`](nginx/nginx-https.conf.example) and [`scripts/letsencrypt.sh`](scripts/letsencrypt.sh).

## CI and E2E tests

GitHub Actions (`.github/workflows/ci.yml`):

- `go test` (backend, including API integration tests)
- `npm run build` (frontend)
- Playwright smoke tests (`e2e/`)

Locally:

```bash
make test
make test-e2e   # requires Node 20+, Go 1.21+, gcc (CGO)
```

## Telegram bot

Configure in `backend/config.yaml` or via `.env` (`TELEGRAM_BOT_TOKEN`).

| Command | Description |
|---------|-------------|
| `/start` | List commands |
| `/users` | List users |
| `/add <email> [gb] [days]` | Create user |
| `/link <user_id>` | Connection link |
| `/traffic <user_id>` | Used traffic |
| `/reload` | Regenerate core config |

## Configuration

### `backend/config.yaml`

```yaml
server:
  port: 8080
  api_key: "change-me"          # panel and API key

database:
  path: "./data/rionexgate.db"

core:
  type: "xray"                  # or "sing-box"
  listen_port: 443
  public_host: "your.domain"    # host in VLESS links
  stats_poll_seconds: 60
  xray:
    config_path: "./data/xray/config.json"
    api_address: "host.docker.internal:10085"
  singbox:
    config_path: "./data/sing-box/config.json"
    api_address: "127.0.0.1:9090"

telegram:
  bot_token: "${TELEGRAM_BOT_TOKEN}"
  admin_ids: [123456789]

limits:
  default_traffic_gb: 50
  default_expire_days: 30
```

Full example (no secrets): [`backend/config.example.yaml`](backend/config.example.yaml).

### `.env`

```ini
HTTP_PORT=8888
TELEGRAM_BOT_TOKEN=
BACKUP_DIR=./backups
```

Port **8888** is the default — port 80 is often taken by system nginx/apache.

## API

Base URL: `http://localhost:8888/api`  
Authentication: header `X-API-Key: <api_key>`

```bash
curl -H "X-API-Key: YOUR_KEY" http://localhost:8888/api/health
curl -H "X-API-Key: YOUR_KEY" http://localhost:8888/api/users
```

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check (no auth) |
| GET | `/docs` | Swagger UI (OpenAPI) |
| GET | `/openapi.yaml` | OpenAPI spec |
| GET | `/protocols` | Supported link protocols |
| GET/POST | `/users` | List / create users |
| GET/PUT/DELETE | `/users/{id}` | User CRUD |
| GET | `/users/{id}/link?proto=vless\|vmess\|trojan` | Connection string |
| GET | `/users/{id}/profiles` | Stealth transport profiles |
| GET | `/users/{id}/qr` | QR code (PNG) |
| GET/DELETE | `/users/{id}/devices/...` | RioNexTunnel devices |
| PUT | `/users/{id}/chain` | Bind entry/exit nodes |
| GET/POST | `/nodes` | Multi-hop nodes |
| GET/PUT | `/stealth/settings` | Stealth panel settings |
| GET | `/stats/total` | Aggregate stats |
| GET | `/stats/user/{id}` | Per-user traffic |
| GET/PUT | `/core/type` | Active core |
| POST | `/core/reload` | Reload core config |

## Local development (without Docker)

```bash
make init
make -C backend migrate
make -C backend dev          # API :8080

cd frontend && npm install && npm run dev   # UI :5173, proxy /api → :8080
```

## Project structure

```
RioNexGate/
├── backend/
│   ├── cmd/main.go              # serve + migrate
│   ├── internal/
│   │   ├── api/                 # REST, chi, middleware
│   │   ├── core/                # templates, CoreManager, links
│   │   ├── db/                  # GORM
│   │   ├── telegram/
│   │   └── config/
│   ├── config.example.yaml
│   └── go.mod
├── frontend/
│   └── src/
│       ├── pages/               # Login, Dashboard, Users, Nodes, Stealth, Settings
│       ├── components/
│       └── services/api.ts
├── nginx/
│   ├── nginx.conf               # reverse proxy
│   └── default.conf             # SPA
├── scripts/
│   ├── docker-compose.sh        # compose v2 + sg docker
│   ├── docker-doctor.sh
│   ├── backup.sh
│   └── restore.sh
├── docs/
│   ├── README.md                # documentation index
│   ├── stealth.md               # stealth guide (EN)
│   └── stealth.ru.md              # stealth guide (RU)
├── data/                        # SQLite, core configs (gitignored)
├── docker-compose.yml
├── Makefile
└── .env.example
```

## Backup

```bash
./scripts/backup.sh
./scripts/restore.sh backups/rionexgate-data-YYYYMMDD_HHMMSS.tar.gz
```

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `permission denied` (docker.sock) | `sudo usermod -aG docker $USER`, log out/in |
| `docker-credential-desktop not found` | `echo '{"auths":{}}' > ~/.docker/config.json` |
| `KeyError: 'id'` (compose) | Remove `docker-compose` v1, use `docker compose` v2 |
| Port 80 busy | Use `HTTP_PORT=8888` in `.env` |
| 401 in panel | Wrong API key; Logout or clear `rionexgate_api_key` in localStorage |
| `make dev` cannot access docker | Re-login after `usermod`; `make` falls back to `sg docker` |
| xray stats not working | `make dev-cores`, check `core.xray.api_address` |

Diagnostics: `make docker-doctor`

## Developer documentation

- MVP plan: [`.cursor/plans/mvp.md`](.cursor/plans/mvp.md)
- Status: [`.cursor/STATUS.md`](.cursor/STATUS.md)

## License

MIT License.
