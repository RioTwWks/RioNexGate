# proxy-mgr — текущий статус

**Версия:** MVP  
**Дата:** 2026-06

## Работает

- Docker Compose: backend + frontend + nginx
- Веб-панель: Login, Dashboard, Users, Settings
- REST API с `X-API-Key`
- SQLite в `./data/proxy-mgr.db`
- Генерация xray/sing-box конфигов при CRUD
- Telegram-бот (при настроенном `bot_token`)
- Backup/restore скрипты

## Быстрый старт

```bash
make init
# отредактировать backend/config.yaml (api_key) и .env
make dev
# в другом терминале:
make migrate
```

Открыть http://localhost:8888, войти с `server.api_key` из `config.yaml`.

Опционально ядра:

```bash
make dev-cores
```

## Локальная разработка без Docker

```bash
make -C backend migrate
make -C backend dev          # :8080
cd frontend && npm run dev   # :5173, proxy /api → :8080
```

## Файлы конфигурации

| Файл | Назначение |
|------|------------|
| `backend/config.yaml` | API key, БД, ядра, Telegram |
| `.env` | `HTTP_PORT`, `TELEGRAM_BOT_TOKEN`, `BACKUP_DIR` |
| `docker-compose.yml` | Сервисы, healthchecks, profile `cores` |

## Диагностика

```bash
make docker-doctor
curl -H "X-API-Key: YOUR_KEY" http://localhost:8888/api/health
```

401 в панели → неверный ключ в localStorage; Logout или `localStorage.removeItem('proxy_mgr_api_key')`.
