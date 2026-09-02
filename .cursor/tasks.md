# Tasks for proxy-mgr

## MVP — выполнено

### Backend (Go)
- [x] init go modules, config reader (viper)
- [x] sqlite models (gorm) – User, Traffic, Node
- [x] xray-core / sing-box config generator (templates)
- [x] API: CRUD users, get link/qr, traffic stats
- [x] Core controller: restart core on config change, collect stats
- [x] Telegram bot (go-telegram-bot-api)
- [x] Basic auth middleware

### Frontend (React)
- [x] setup vite + tailwind + axios
- [x] login page (API key)
- [x] dashboard with traffic chart
- [x] users table + add/edit/delete
- [x] view links and QR codes
- [x] switch between xray/sing-box

### Integration
- [x] docker-compose for all services
- [x] nginx reverse proxy config (+ X-API-Key forwarding)
- [x] health checks
- [x] backup scripts
- [x] docker compose v2 wrapper (`scripts/docker-compose.sh`)

### Deployment
- [x] .env.example
- [x] README with install steps
- [x] `make init`, `make docker-doctor`

## Post-MVP

- [x] OpenAPI `/api/docs`
- [x] HTTPS / Let's Encrypt example
- [x] Доп. протоколы (VMess, Trojan) в UI
- [x] CI pipeline
- [x] E2E tests

---

## Интеграция с RioNexTunnel (серверная часть)

**Цель:** proxy-mgr остаётся универсальной панелью для любых VLESS/VMess/Trojan-клиентов, но дополнительно умеет обслуживать [RioNexTunnel](https://github.com/RioTwWks/RioNexTunnel) через опциональный расширенный API. Управляющие фичи (подписки, телеметрия, удалённые команды) не ломают стандартную совместимость.

**Принцип — двухуровневая модель:**
1. **Базовый уровень** — стандартные протоколы и ссылки; работает с любыми клиентами.
2. **Расширенный уровень** — опциональный API для пары RioNexTunnel + proxy-mgr: регистрация устройств, JSON-конфиг, телеметрия, команды. Аутентификация через отдельный `device_token`, не влияющий на транспортные протоколы.

### Контекст: «болячки», которые решает сервер

- Подписка обновляется, но клиент не применяет изменения или падает на невалидном JSON → нужен стабильный формат подписки и JSON-конфига.
- Потеря статистики при обрыве → сервер должен принимать телеметрию с дедупликацией и не блокировать пользователя при отсутствии heartbeat.
- Конфликт аутентификации (динамические SOCKS5-креды vs проверка подписки) → отдельный `device_token`, не связанный с транспортом.
- Неоднозначные форматы ссылок → строгая генерация по спецификациям протоколов.

---

### 1. Модель данных

- [ ] Добавить поле `device_tokens` — список уникальных токенов устройств (отдельная таблица `Device` с FK на `User`: `token`, `label`, `last_seen_at`, `created_at`).
- [ ] Добавить поле `subscription_token` у пользователя — персональный токен для URL подписки (`https://panel/sub/{token}`).
- [ ] Миграция БД + обновление OpenAPI-схемы.

---

### 2. Новые API-эндпоинты (опциональны для сторонних клиентов)

Все эндпоинты, кроме `/api/subscription/{token}`, аутентифицируются заголовком `X-Device-Token`. При отсутствии или невалидном токене — `401`; базовые ссылки через веб/API по `X-API-Key` продолжают работать.

| Метод | Путь | Назначение |
|-------|------|------------|
| `POST` | `/api/client/register` | Регистрация устройства: `user_id` или логин/пароль → `device_token`, `subscription_url`. |
| `GET` | `/api/client/config` | Актуальный конфиг в JSON (серверы, inbounds, DNS) + `config_hash` для условного обновления. |
| `POST` | `/api/client/stats` | Тело: `{session_id, bytes_in, bytes_out, sessions, status}`; обновление статистики пользователя. |
| `GET` | `/api/client/commands` | Long polling или SSE: команды `refresh_config`, `disconnect` и т.п. |
| `GET` | `/api/subscription/{token}` | Стандартная подписка: base64-encoded список ссылок (совместимость с любыми клиентами). |

Задачи:
- [ ] Реализовать handlers и роуты в `backend/internal/api/`.
- [ ] Middleware `DeviceTokenAuth` (отдельно от `APIKeyAuth`).
- [ ] `POST /api/client/register` — выдача токена, привязка к пользователю, возврат `subscription_url`.
- [ ] `GET /api/client/config` — JSON-экспорт конфига пользователя (протоколы, хост, порт, параметры транспорта).
- [ ] `POST /api/client/stats` — приём телеметрии от клиента.
- [ ] `GET /api/client/commands` — long polling / SSE для удалённых команд.
- [ ] `GET /api/subscription/{token}` — base64-подписка (VLESS/VMess/Trojan ссылки из `core/links.go`).
- [ ] Обновить CORS: разрешить заголовки `X-Device-Token`, `X-API-Version`.
- [ ] Документировать эндпоинты в OpenAPI (`/api/openapi.yaml`).

---

### 3. Генерация конфигов и ссылок

- [ ] Проверить соответствие генерируемых ссылок спецификациям (обязательные поля VLESS: `id`, `encryption`, `host`, `port`; для WebSocket — `path`, `host` и т.д.).
- [ ] Добавить экспорт полного JSON-конфига через `GET /api/client/config` (структура, удобная для RioNexTunnel).
- [ ] Включить в JSON-конфиг поле `config_hash` (SHA-256 от канонического JSON) для дедупликации обновлений на клиенте.
- [ ] Опционально: параметры входящего SOCKS5 (порт, метод аутентификации) в JSON-конфиге для согласованности с панелью.

---

### 4. Хранение и учёт статистики

- [ ] Таблица `ClientStatsReport` (`device_token`, `session_id`, `bytes_in`, `bytes_out`, `reported_at`) для приёма телеметрии от клиентов.
- [ ] Дедупликация по `session_id` + временным меткам — повторные отправки не удваивают трафик.
- [ ] При отсутствии heartbeat от устройства — не блокировать пользователя; продолжать учёт с серверной стороны (xray Stats API).
- [ ] Агрегация client-reported stats с server-side polling в единый ответ `/api/stats/user/{id}`.

---

### 5. Отказоустойчивость и кеширование

- [ ] Кеш последнего известного конфига для устройства (в БД или in-memory с TTL) — клиент может получить конфиг при временных проблемах с генерацией.
- [ ] Валидировать JSON-ответы перед отдачей; при ошибке генерации — отдавать последний валидный кеш + логировать ошибку.
- [ ] Подписка `/api/subscription/{token}` должна отдавать корректный base64 даже при частичных сбоях (graceful degradation: хотя бы рабочие ссылки).

---

### 6. Версионирование API

- [ ] Поддержать заголовок `X-API-Version: v1` в запросах клиента; отвечать тем же заголовком.
- [ ] Зафиксировать контракт v1 в OpenAPI; breaking changes — только в v2.

---

### 7. Логирование и мониторинг

- [ ] Логировать все запросы к `/api/client/*` с `device_token` (маскированным) и `user_id` для отладки синхронизации.
- [ ] Метрики: количество активных устройств, частота sync, ошибки регистрации.

---

### 8. Тестирование (сервер)

- [ ] Интеграционные тесты в `backend/internal/api/integration_test.go`:
  - регистрация устройства → получение конфига → приём статистики;
  - обновление пользователя → новый `config_hash` в ответе `/api/client/config`;
  - невалидный `device_token` → `401`;
  - подписка `/api/subscription/{token}` → валидный base64 со ссылками;
  - дедупликация stats по `session_id`.
- [ ] E2E (Playwright): сценарий выдачи subscription URL из UI и проверка ответа эндпоинта.

---

### 9. UI (минимальные доработки)

- [ ] На странице пользователя: отображение `subscription_url`, кнопка «Скопировать подписку».
- [ ] Список зарегистрированных устройств (`device_tokens`) с `last_seen_at`, возможность отозвать токен.
- [ ] Индикатор: клиент синхронизирован / давно не отчитывался.

---

### Порядок реализации (рекомендуемый)

1. Модель данных + миграция (`Device`, `subscription_token`, `ClientStatsReport`).
2. `GET /api/subscription/{token}` — быстрая совместимость с любыми клиентами.
3. `POST /api/client/register`, `GET /api/client/config`, middleware `DeviceTokenAuth`.
4. `POST /api/client/stats` с дедупликацией.
5. `GET /api/client/commands` (long polling → при необходимости SSE/WebSocket).
6. OpenAPI, интеграционные тесты, UI для subscription URL и устройств.
