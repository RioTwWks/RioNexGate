**[English](README.md)** | Русский

# RioNexGate — панель управления прокси

**RioNexGate** — панель для управления прокси на **Xray-core** и **sing-box**: Go API, React UI, Telegram-бот, Docker Compose.

Поддерживаются пользователи, VLESS/VMess/Trojan-ссылки, QR-коды, лимиты трафика, stealth-пресеты anti-DPI, мульти-хоп узлы и опциональный клиентский API для [RioNexTunnel](https://github.com/RioTwWks/RioNexTunnel) (устройства, подписка, телеметрия).

## Возможности

- **Два ядра**: **xray** и **sing-box** (переключение в UI / API)
- **Пользователи**: CRUD, лимиты трафика и срока действия, ссылки и QR (VLESS, VMess, Trojan)
- **Дашборд**: общая статистика трафика и графики
- **Stealth / anti-DPI**: VLESS + Reality + XHTTP (`stream-one`), Vision (TCP), опциональный TLS, фрагментация ServerHello, резерв AmneziaWG — см. [docs/stealth.ru.md](docs/stealth.ru.md)
- **Мульти-хоп узлы**: цепочки entry/exit с проверкой здоровья (страница **Nodes**)
- **Клиентский API RioNexTunnel**: регистрация устройств, синхронизация JSON-конфига, URL подписки, телеметрия, удалённые команды
- **Telegram-бот** для администратора
- **REST API** с аутентификацией по API-ключу; OpenAPI / Swagger UI
- **Nginx** как единая точка входа

## Документация

| Руководство | Описание |
|-------------|----------|
| [Stealth / anti-DPI](docs/stealth.ru.md) | Reality, XHTTP, Vision, TLS, фрагментация, AmneziaWG |
| [docs/README.md](docs/README.md) | Индекс документации (EN / RU) |
| [OpenAPI / Swagger](http://localhost:8888/api/docs) | Интерактивная справка API (после `make dev`) |

## Требования

- Linux (протестировано на Ubuntu 24.04)
- **Docker Engine** + плагин **Docker Compose v2** (`docker compose`, не Python `docker-compose` v1)
- Go 1.21+ и Node 18+ — только для локальной разработки без Docker

### Установка Docker (Linux)

```bash
# Официальный репозиторий Docker — см. https://docs.docker.com/engine/install/ubuntu/
sudo apt install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo usermod -aG docker $USER
# перелогиниться, чтобы группа docker применилась
```

Проверка:

```bash
docker compose version   # v2.x или v5.x
make docker-doctor
```

> **Не используйте** устаревший пакет `docker-compose` v1 из apt (`sudo apt remove docker-compose`).  
> После удаления Docker Desktop сбросьте `~/.docker/config.json`: `echo '{"auths":{}}' > ~/.docker/config.json`

## Быстрый старт

```bash
git clone https://github.com/RioTwWks/RioNexGate.git
cd RioNexGate

make init
```

Отредактируйте файлы:

- `backend/config.yaml` — обязательно смените `server.api_key`
- `.env` — `HTTP_PORT`, `TELEGRAM_BOT_TOKEN` (опционально)

```bash
make dev
```

В другом терминале (первый запуск):

```bash
make migrate
```

### URL после запуска

| URL | Описание |
|-----|----------|
| http://localhost:8888 | **Веб-панель** (nginx, порт из `HTTP_PORT`) |
| http://localhost:8888/api/health | Health check API |
| http://localhost:8888/api/docs | Swagger UI |
| http://localhost:8080/api/health | Backend напрямую |
| http://localhost:3000 | Frontend напрямую (без прокси API) |

Войдите в панель с API-ключом из `backend/config.yaml` → `server.api_key`.

Опционально — поднять ядра xray/sing-box:

```bash
make dev-cores
```

Для stealth inbound (порты 443 / 8443 на хосте) используйте профиль `cores` выше. Резерв AmneziaWG: `docker compose --profile awg up -d amneziawg`.

## Веб-панель

1. Откройте http://localhost:8888
2. Введите API-ключ из `backend/config.yaml`
3. **Dashboard** — общий трафик и график
4. **Users** — создание, редактирование, ссылки, QR, устройства, привязка цепочки
5. **Nodes** — entry/exit узлы, health check, топология мульти-хоп
6. **Stealth** — настройки Reality/XHTTP/Vision/TLS/фрагментации/AWG
7. **Settings** — переключение xray ↔ sing-box, reload ядра

При ошибке 401: нажмите **Logout** или очистите `localStorage` (`rionexgate_api_key`).

В модальном окне ссылки можно выбрать протокол: **VLESS**, **VMess** или **Trojan**. При включённом stealth используйте `?all=true` для ссылок по всем профилям.

## Клиентский API RioNexTunnel

[RioNexTunnel](https://github.com/RioTwWks/RioNexTunnel) может регистрировать устройства и синхронизировать JSON-конфиг. Стандартные VLESS/VMess/Trojan-клиенты продолжают работать без изменений.

| Эндпоинт | Аутентификация | Описание |
|----------|----------------|----------|
| `POST /api/client/register` | нет | Регистрация устройства → `device_token`, `subscription_url` |
| `GET /api/client/config` | `X-Device-Token` | JSON-конфиг с `config_hash` и транспортными профилями |
| `POST /api/client/stats` | `X-Device-Token` | Телеметрия сессии |
| `GET /api/client/commands` | `X-Device-Token` | Long-poll или SSE (`?stream=sse`) для `refresh_config` и др. |
| `GET /api/subscription/{token}` | нет | Подписка в base64 (ссылки через перевод строки) |
| `GET /api/users/{id}/devices` | `X-API-Key` | Список зарегистрированных устройств |
| `DELETE /api/users/{id}/devices/{deviceId}` | `X-API-Key` | Отзыв устройства |

Пример регистрации:

```bash
curl -X POST http://localhost:8888/api/client/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","label":"laptop"}'
```

## Stealth / Anti-DPI (Reality + XHTTP)

Секция `core.stealth` в `backend/config.yaml` включает пресеты для устойчивости к DPI:

- **VLESS + Reality + XHTTP** (`mode: stream-one`) на порту **443**
- **VLESS + Reality + Vision** (TCP) на порту **8443**
- Опционально **VLESS + TLS** на нестандартном порту
- Опциональный резерв **AmneziaWG** (UDP)

Подробности: [docs/stealth.ru.md](docs/stealth.ru.md) · [docs/stealth.md](docs/stealth.md)

### Генерация Reality keypair

```bash
# В контейнере xray или с установленным xray:
xray x25519
# PrivateKey → core.stealth.reality.private_key
# Password   → core.stealth.reality.public_key (pbk в клиентских ссылках)
```

Мульти-профильные ссылки через API:

```bash
curl -H "X-API-Key: YOUR_KEY" "http://localhost:8888/api/users/1/link?all=true"
```

## Makefile

| Команда | Действие |
|---------|----------|
| `make init` | Создать `config.yaml`, `.env`, каталоги `data/` |
| `make dev` | Сборка и запуск (`docker compose up --build`) |
| `make up` | Запуск в фоне |
| `make down` | Остановить контейнеры |
| `make dev-cores` | Запустить xray-core и sing-box (profile `cores`) |
| `make dev-local` | Подсказка для запуска без Docker |
| `make migrate` | Миграция БД в контейнере backend |
| `make docker-doctor` | Проверка Docker |
| `make test` | `go test` + `npm test` |
| `make test-e2e` | Playwright E2E (backend + frontend dev) |
| `make logs` | Логи compose |
| `make clean` | Остановить и удалить volumes |

## OpenAPI / Swagger

После запуска панели: http://localhost:8888/api/docs  
Спецификация: http://localhost:8888/api/openapi.yaml

Для «Try it out» в Swagger UI используется ключ из `localStorage` (`rionexgate_api_key`), если вы уже вошли в панель.

## HTTPS (Let's Encrypt)

Пример настройки TLS для nginx:

1. Убедитесь, что домен указывает на сервер и порт 80 свободен.
2. Получите сертификат:
   ```bash
   chmod +x scripts/letsencrypt.sh
   ./scripts/letsencrypt.sh panel.example.com
   ```
3. Включите HTTPS-конфиг nginx:
   ```bash
   cp nginx/nginx-https.conf.example nginx/nginx.conf
   ```
4. В `docker-compose.yml` раскомментируйте порт `HTTPS_PORT` (по умолчанию `8443`).
5. `make up` — панель на `https://panel.example.com:8443`.

Подробности в комментариях [`nginx/nginx-https.conf.example`](nginx/nginx-https.conf.example) и [`scripts/letsencrypt.sh`](scripts/letsencrypt.sh).

## CI и E2E-тесты

GitHub Actions (`.github/workflows/ci.yml`):

- `go test` (backend, включая API integration tests)
- `npm run build` (frontend)
- Playwright smoke-тесты (`e2e/`)

Локально:

```bash
make test
make test-e2e   # требует Node 20+, Go 1.21+, gcc (CGO)
```

## Telegram-бот

Настройте в `backend/config.yaml` или через `.env` (`TELEGRAM_BOT_TOKEN`).

| Команда | Описание |
|---------|----------|
| `/start` | Список команд |
| `/users` | Список пользователей |
| `/add <email> [gb] [days]` | Создать пользователя |
| `/link <user_id>` | Ссылка подключения |
| `/traffic <user_id>` | Использованный трафик |
| `/reload` | Перегенерировать конфиг ядра |

## Конфигурация

### `backend/config.yaml`

```yaml
server:
  port: 8080
  api_key: "change-me"          # ключ для панели и API

database:
  path: "./data/rionexgate.db"

core:
  type: "xray"                  # или "sing-box"
  listen_port: 443
  public_host: "your.domain"    # для VLESS-ссылок
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

Пример без секретов: [`backend/config.example.yaml`](backend/config.example.yaml).

### `.env`

```ini
HTTP_PORT=8888
TELEGRAM_BOT_TOKEN=
BACKUP_DIR=./backups
```

Порт **8888** по умолчанию — порт 80 часто занят системным nginx/apache.

## API

Базовый URL: `http://localhost:8888/api`  
Аутентификация: заголовок `X-API-Key: <api_key>`

```bash
curl -H "X-API-Key: YOUR_KEY" http://localhost:8888/api/health
curl -H "X-API-Key: YOUR_KEY" http://localhost:8888/api/users
```

| Метод | Эндпоинт | Описание |
|-------|----------|----------|
| GET | `/health` | Health check (без auth) |
| GET | `/docs` | Swagger UI (OpenAPI) |
| GET | `/openapi.yaml` | OpenAPI спецификация |
| GET | `/protocols` | Поддерживаемые протоколы ссылок |
| GET/POST | `/users` | Список / создание пользователей |
| GET/PUT/DELETE | `/users/{id}` | CRUD |
| GET | `/users/{id}/link?proto=vless\|vmess\|trojan` | Строка подключения |
| GET | `/users/{id}/profiles` | Stealth-профили транспорта |
| GET | `/users/{id}/qr` | QR-код (PNG) |
| GET/DELETE | `/users/{id}/devices/...` | Устройства RioNexTunnel |
| PUT | `/users/{id}/chain` | Привязка entry/exit узлов |
| GET/POST | `/nodes` | Мульти-хоп узлы |
| GET/PUT | `/stealth/settings` | Настройки stealth в панели |
| GET | `/stats/total` | Общая статистика |
| GET | `/stats/user/{id}` | Трафик пользователя |
| GET/PUT | `/core/type` | Активное ядро |
| POST | `/core/reload` | Перезагрузить конфиг |

## Локальная разработка (без Docker)

```bash
make init
make -C backend migrate
make -C backend dev          # API :8080

cd frontend && npm install && npm run dev   # UI :5173, proxy /api → :8080
```

## Структура проекта

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
│   ├── README.md                # индекс документации
│   ├── stealth.md               # stealth (EN)
│   └── stealth.ru.md            # stealth (RU)
├── data/                        # SQLite, конфиги ядер (gitignored)
├── docker-compose.yml
├── Makefile
└── .env.example
```

## Резервное копирование

```bash
./scripts/backup.sh
./scripts/restore.sh backups/rionexgate-data-YYYYMMDD_HHMMSS.tar.gz
```

## Устранение неполадок

| Симптом | Решение |
|---------|---------|
| `permission denied` (docker.sock) | `sudo usermod -aG docker $USER`, перелогиниться |
| `docker-credential-desktop not found` | `echo '{"auths":{}}' > ~/.docker/config.json` |
| `KeyError: 'id'` (compose) | Удалить `docker-compose` v1, использовать `docker compose` v2 |
| Порт 80 занят | Используйте `HTTP_PORT=8888` в `.env` |
| 401 в панели | Неверный API key; Logout или очистить `rionexgate_api_key` в localStorage |
| `make dev` не видит docker | Перелогиниться после `usermod`; `make` использует `sg docker` как fallback |
| xray stats не работают | `make dev-cores`, проверить `core.xray.api_address` |

Диагностика: `make docker-doctor`

## Документация для разработчиков

- План MVP: [`.cursor/plans/mvp.md`](.cursor/plans/mvp.md)
- Статус: [`.cursor/STATUS.md`](.cursor/STATUS.md)

## Лицензия

MIT License.
