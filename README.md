# proxy-mgr — универсальная панель управления прокси

**proxy-mgr** — это лёгкая, самописная панель для управления прокси-серверами на базе **Xray-core** и **sing-box**. Проект включает веб-интерфейс (React), Telegram-бота и единый API на Go. Позволяет создавать пользователей, выдавать клиентские ссылки (VLESS, VMess, Trojan, Shadowsocks), отслеживать трафик и переключать ядро в один клик.

## Возможности

- Поддержка двух ядер: **Xray-core** и **sing-box** (переключение через конфиг или UI).
- Управление пользователями (создание, удаление, лимиты по трафику и времени).
- Генерация клиентских ссылок и QR-кодов.
- Статистика использования трафика в реальном времени.
- **Telegram-бот** для администратора: просмотр пользователей, выдача ключей, остатки трафика.
- **REST API** с аутентификацией по API-ключу (можно интегрировать с биллингом).
- **Nginx** как reverse-proxy для API и фронтенда.
- Простой запуск через **Docker Compose**.

## Требования

- Linux / macOS / WSL2
- Docker и Docker Compose (≥ 2.0)
- Домен (опционально, для HTTPS)
- Telegram Bot Token (можно получить у [@BotFather](https://t.me/botfather))

## Быстрый старт

### 1. Клонировать репозиторий

```bash
git clone https://github.com/your-username/proxy-mgr.git
cd proxy-mgr
```

### 2. Создать конфигурационные файлы

Скопируйте примеры и отредактируйте под себя:

```bash
cp backend/config.example.yaml backend/config.yaml
cp .env.example .env
```

В `.env` укажите токен Telegram-бота и список администраторов (Telegram ID):

```ini
TELEGRAM_BOT_TOKEN=123456:ABC-DEF...
ADMIN_IDS=123456789,987654321
```

В `backend/config.yaml` при необходимости измените порты, путь к БД, лимиты по умолчанию.

### 3. Запустить через Docker Compose

```bash
docker-compose up -d
```

В первый раз также нужно выполнить миграцию базы данных:

```bash
docker-compose exec backend ./proxy-mgr migrate
```

### 4. Проверить работу

- Веб-интерфейс: [http://localhost](http://localhost) (или ваш домен)
- API: [http://localhost/api/health](http://localhost/api/health)
- Telegram-бот: отправьте `/start` — он ответит списком команд.

## Использование

### Веб-панель

1. Авторизуйтесь с помощью API-ключа (по умолчанию `your-secure-api-key`, можно изменить в `config.yaml`).
2. На дашборде отображается общий трафик и график использования.
3. В разделе **Пользователи**:
   - Добавьте нового пользователя (e‑mail, лимит трафика в ГБ, срок действия).
   - После создания нажмите **Получить ссылку** — появится строка подключения и QR-код.
   - Редактируйте или удаляйте пользователей.
4. В разделе **Настройки** можно переключить активное ядро (`xray` ↔ `sing-box`). После смены ядра конфигурация всех пользователей будет пересоздана автоматически.

### Telegram-бот

Команды для администратора:

- `/users` — список пользователей (имя, остаток трафика, активен ли).
- `/add <email> [traffic_gb] [expire_days]` — создать нового пользователя.  
  Пример: `/add user@example.com 100 30`
- `/link <user_id>` — получить клиентскую ссылку для пользователя.
- `/traffic <user_id>` — показать использованный трафик.
- `/reload` — принудительно перезагрузить конфиг ядра (после ручного вмешательства).

Бот понимает инлайн-кнопки: при вызове `/users` можно сразу перейти к управлению конкретным пользователем.

## Настройка

### Структура конфигурационного файла `backend/config.yaml`

```yaml
server:
  port: 8080
  api_key: "your-secure-api-key"

database:
  path: "./data/proxy-mgr.db"

core:
  type: "xray"   # или "sing-box"
  xray:
    config_path: "./data/xray/config.json"
    binary_path: "/usr/local/bin/xray"
    api_address: "127.0.0.1:10085"
  singbox:
    config_path: "./data/sing-box/config.json"
    binary_path: "/usr/local/bin/sing-box"
    api_address: "127.0.0.1:10086"

telegram:
  bot_token: "${TELEGRAM_BOT_TOKEN}"
  admin_ids: [123456789]

limits:
  default_traffic_gb: 50
  default_expire_days: 30
```

### HTTPS и домен

Если у вас есть домен, настройте Nginx для автоматического получения SSL-сертификатов через Let's Encrypt:

1. Отредактируйте `nginx/nginx.conf`, добавив блок `server` с портом 443.
2. Используйте `certbot` или образ `nginx-certbot`.  
   Готовый пример можно найти в папке `examples/nginx-ssl.conf`.

### Ручная сборка (без Docker)

Если вы предпочитаете запускать без контейнеров:

- **Backend**: `cd backend && go run cmd/main.go`
- **Frontend**: `cd frontend && npm install && npm run dev`
- **Xray / sing-box** установите на хост и укажите пути в конфиге.

## API

Базовый адрес: `http://localhost/api`  
Аутентификация: заголовок `X-API-Key: <api_key>`

| Метод | Эндпоинт | Описание |
|-------|----------|-----------|
| GET | `/users` | Список всех пользователей |
| POST | `/users` | Создать пользователя (JSON: `email, traffic_gb, expire_days`) |
| GET | `/users/{id}` | Получить информацию о пользователе |
| PUT | `/users/{id}` | Обновить лимиты |
| DELETE | `/users/{id}` | Удалить пользователя |
| GET | `/users/{id}/link?proto=vless` | Получить строку подключения |
| GET | `/users/{id}/qr` | QR-код в формате PNG |
| GET | `/stats/total` | Общая статистика трафика |
| GET | `/stats/user/{id}` | Трафик конкретного пользователя |
| POST | `/core/reload` | Перезагрузить активное ядро |

Подробное описание (OpenAPI) будет доступно после запуска по адресу `/api/docs`.

## Структура проекта

```
proxy-mgr/
├── backend/                # Go-бэкенд
│   ├── cmd/main.go         # точка входа
│   ├── internal/           # внутренние пакеты (api, core, db, telegram)
│   ├── config.yaml         # конфигурация
│   └── go.mod
├── frontend/               # React-приложение (Vite + Tailwind)
│   ├── src/
│   ├── public/
│   └── package.json
├── nginx/                  # конфиги Nginx
│   └── nginx.conf
├── data/                   # монтируемые данные: БД, конфиги ядер, логи
├── scripts/                # вспомогательные скрипты
├── docker-compose.yml
├── Dockerfile.backend
├── Dockerfile.frontend
├── Makefile
└── README.md
```

## Разработка и доработка

### Локальная разработка (горячая перезагрузка)

```bash
# Запустить только базу данных и ядра
docker-compose up -d xray-core sing-box

# Бэкенд в режиме разработки
cd backend && go run cmd/main.go

# Фронтенд (порт 5173)
cd frontend && npm run dev
```

### Добавление нового протокола

1. В `backend/internal/core/manager.go` добавьте генерацию конфига для нового протокола.
2. Обновите API (эндпоинт `/users/{id}/link`).
3. В фронтенде добавьте кнопку выбора протокола.

### Создание резервных копий

```bash
./scripts/backup.sh
```

Скрипт архивирует папку `data/` и выгружает в указанную директорию (настраивается в `.env`).

## Устранение неполадок

- **Ошибка "cannot connect to xray-api"** — проверьте, что в `config.yaml` указан правильный `api_address` и что xray-core запущен с флагом `-api`.
- **Telegram бот не отвечает** — убедитесь, что `bot_token` корректен и бот был добавлен в чат администратора.
- **Пустая страница фронтенда** — откройте консоль браузера (F12) и проверьте, что API доступен и CORS не блокирует запросы (в Go-бэкенде CORS разрешён для всех источников в dev-режиме).

## Лицензия

MIT License. Автор не несёт ответственности за использование данного ПО в странах, где VPN-технологии ограничены законодательством.

## Благодарности

- [Xray-core](https://github.com/XTLS/Xray-core)
- [sing-box](https://github.com/SagerNet/sing-box)
- [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)

---