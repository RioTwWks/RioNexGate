# Tasks for RioNexGate

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

**Цель:** RioNexGate остаётся универсальной панелью для любых VLESS/VMess/Trojan-клиентов, но дополнительно умеет обслуживать [RioNexTunnel](https://github.com/RioTwWks/RioNexTunnel) через опциональный расширенный API. Управляющие фичи (подписки, телеметрия, удалённые команды) не ломают стандартную совместимость.

**Принцип — двухуровневая модель:**
1. **Базовый уровень** — стандартные протоколы и ссылки; работает с любыми клиентами.
2. **Расширенный уровень** — опциональный API для пары RioNexTunnel + RioNexGate: регистрация устройств, JSON-конфиг, телеметрия, команды. Аутентификация через отдельный `device_token`, не влияющий на транспортные протоколы.

### Контекст: «болячки», которые решает сервер

- Подписка обновляется, но клиент не применяет изменения или падает на невалидном JSON → нужен стабильный формат подписки и JSON-конфига.
- Потеря статистики при обрыве → сервер должен принимать телеметрию с дедупликацией и не блокировать пользователя при отсутствии heartbeat.
- Конфликт аутентификации (динамические SOCKS5-креды vs проверка подписки) → отдельный `device_token`, не связанный с транспортом.
- Неоднозначные форматы ссылок → строгая генерация по спецификациям протоколов.

---

### 1. Модель данных

- [x] Добавить поле `device_tokens` — список уникальных токенов устройств (отдельная таблица `Device` с FK на `User`: `token`, `label`, `last_seen_at`, `created_at`).
- [x] Добавить поле `subscription_token` у пользователя — персональный токен для URL подписки (`https://panel/sub/{token}`).
- [x] Миграция БД + обновление OpenAPI-схемы.

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
- [x] Реализовать handlers и роуты в `backend/internal/api/`.
- [x] Middleware `DeviceTokenAuth` (отдельно от `APIKeyAuth`).
- [x] `POST /api/client/register` — выдача токена, привязка к пользователю, возврат `subscription_url`.
- [x] `GET /api/client/config` — JSON-экспорт конфига пользователя (протоколы, хост, порт, параметры транспорта).
- [x] `POST /api/client/stats` — приём телеметрии от клиента.
- [x] `GET /api/client/commands` — long polling / SSE для удалённых команд.
- [x] `GET /api/subscription/{token}` — base64-подписка (VLESS/VMess/Trojan ссылки из `core/links.go`).
- [x] Обновить CORS: разрешить заголовки `X-Device-Token`, `X-API-Version`.
- [x] Документировать эндпоинты в OpenAPI (`/api/openapi.yaml`).

---

### 3. Генерация конфигов и ссылок

- [x] Проверить соответствие генерируемых ссылок спецификациям (обязательные поля VLESS: `id`, `encryption`, `host`, `port`; для WebSocket — `path`, `host` и т.д.).
- [x] Добавить экспорт полного JSON-конфига через `GET /api/client/config` (структура, удобная для RioNexTunnel).
- [x] Включить в JSON-конфиг поле `config_hash` (SHA-256 от канонического JSON) для дедупликации обновлений на клиенте.
- [x] Опционально: параметры входящего SOCKS5 (порт, метод аутентификации) в JSON-конфиге для согласованности с панелью.

---

### 4. Хранение и учёт статистики

- [x] Таблица `ClientStatsReport` (`device_token`, `session_id`, `bytes_in`, `bytes_out`, `reported_at`) для приёма телеметрии от клиентов.
- [x] Дедупликация по `session_id` + временным меткам — повторные отправки не удваивают трафик.
- [x] При отсутствии heartbeat от устройства — не блокировать пользователя; продолжать учёт с серверной стороны (xray Stats API).
- [x] Агрегация client-reported stats с server-side polling в единый ответ `/api/stats/user/{id}`.

---

### 5. Отказоустойчивость и кеширование

- [x] Кеш последнего известного конфига для устройства (в БД или in-memory с TTL) — клиент может получить конфиг при временных проблемах с генерацией.
- [x] Валидировать JSON-ответы перед отдачей; при ошибке генерации — отдавать последний валидный кеш + логировать ошибку.
- [x] Подписка `/api/subscription/{token}` должна отдавать корректный base64 даже при частичных сбоях (graceful degradation: хотя бы рабочие ссылки).

---

### 6. Версионирование API

- [x] Поддержать заголовок `X-API-Version: v1` в запросах клиента; отвечать тем же заголовком.
- [x] Зафиксировать контракт v1 в OpenAPI; breaking changes — только в v2.

---

### 7. Логирование и мониторинг

- [x] Логировать все запросы к `/api/client/*` с `device_token` (маскированным) и `user_id` для отладки синхронизации.
- [x] Метрики: количество активных устройств, частота sync, ошибки регистрации.

---

### 8. Тестирование (сервер)

- [x] Интеграционные тесты в `backend/internal/api/integration_test.go`:
  - регистрация устройства → получение конфига → приём статистики;
  - обновление пользователя → новый `config_hash` в ответе `/api/client/config`;
  - невалидный `device_token` → `401`;
  - подписка `/api/subscription/{token}` → валидный base64 со ссылками;
  - дедупликация stats по `session_id`.
- [x] E2E (Playwright): сценарий выдачи subscription URL из UI и проверка ответа эндпоинта. *(PR #5, `e2e/tests/rionex-stealth.spec.ts`)*

---

### 9. UI (минимальные доработки)

- [x] На странице пользователя: отображение `subscription_url`, кнопка «Скопировать подписку».
- [x] Список зарегистрированных устройств (`device_tokens`) с `last_seen_at`, возможность отозвать токен.
- [x] Индикатор: клиент синхронизирован / давно не отчитывался.

---

### Порядок реализации (рекомендуемый)

1. Модель данных + миграция (`Device`, `subscription_token`, `ClientStatsReport`).
2. `GET /api/subscription/{token}` — быстрая совместимость с любыми клиентами.
3. `POST /api/client/register`, `GET /api/client/config`, middleware `DeviceTokenAuth`.
4. `POST /api/client/stats` с дедупликацией.
5. `GET /api/client/commands` (long polling → при необходимости SSE/WebSocket).
6. OpenAPI, интеграционные тесты, UI для subscription URL и устройств.

---

## Устойчивость к DPI / ТСПУ (серверная часть)

**Цель:** RioNexGate генерирует и разворачивает конфигурации, устойчивые к современному DPI (2026), оставаясь на **официальном Xray-core** (и sing-box где применимо). Клиентские доработки (uTLS, split tunneling, auto-fallback) — в репозитории RioNexTunnel.

**Стратегическое решение по ядру:**
- **Не использовать** заброшенный форк [fwflunky/REALITY-rkn-fix](https://github.com/fwflunky/REALITY-rkn-fix) в production — нет тестов, нет поддержки, риск отставания от upstream.
- **Не вести собственный форк** Xray-core без крайней необходимости — постоянный merge с upstream, ответственность за безопасность.
- **Использовать официальный Xray-core** с актуальными настройками транспортов; идеи форка (динамические сертификаты, `ImpersonateCert`) изучать как reference и отслеживать появление аналогов в [XTLS/Xray-core](https://github.com/XTLS/Xray-core) (issues/PR).

### Контекст: угрозы, с которыми борется серверная конфигурация

- **Сигнатурный анализ** — статичная структура TLS-сертификата Reality, типичные отпечатки VLESS+Reality на TCP.
- **Атака на TLS-рукопожатие** — аномально долгое handshake, постоянное TCP без HTTP-паттерна после него.
- **Анализ размеров пакетов** — ServerHello ~200 байт в одном сегменте; тотальная фрагментация всех пакетов даёт аномальный PPS.

---

### 1. Целевые транспортные связки (генерация в шаблонах)

Сейчас шаблон `xray.json.tmpl` — наивный `VLESS + TCP`. Нужны пресеты inbounds:

| Приоритет | Связка | Порт (пример) | Назначение |
|-----------|--------|---------------|------------|
| **Основная** | `VLESS + Reality + XHTTP` (`mode: stream-one`) | 443 | Лучшая маскировка поведения трафика; стандарт 2026 |
| **Резервная** | `VLESS + Reality + Vision` (TCP) | 8443 | iOS и клиенты без XHTTP; разнообразие для DPI |
| **Мобильная** | `VLESS + TLS` (нестандартный порт) | настраиваемый | Запасной вариант; в ссылках — hint на `mux` (concurrency 8) |
| **Альтернатива** | `AmneziaWG` | UDP | Резерв при блокировке Xray-стека; отдельный inbound/профиль |

Задачи:
- [x] Расширить `backend/internal/config` — секция `core.stealth` (или `core.xray.inbounds[]`) с пресетами транспортов.
- [x] Переписать/дополнить `templates/xray.json.tmpl` — **два основных inbound** одновременно: XHTTP+Reality (443) и TCP+Reality+Vision (8443).
- [x] Для XHTTP явно задавать `"mode": "stream-one"` (не `auto` — известные баги).
- [x] Параметры Reality в шаблоне: `dest`, `serverNames`, `privateKey`, `shortIds`, `fingerprint` (по умолчанию `firefox` или `edge`, не `chrome`/`safari`).
- [x] Аналогичные пресеты для sing-box (`templates/singbox.json.tmpl`) где транспорт поддерживается.
- [x] Генерация **нескольких ссылок на пользователя** (по одной на каждый inbound) для подписки и API.

---

### 2. Маскировка Reality (Impersonation)

- [x] Поля конфига: `reality.dest`, `reality.server_names[]`, опционально `show` / `xver`.
- [x] **Валидация и рекомендации в UI/docs:** не использовать популярные цели (`yahoo.com`, `vk.com`) — повышенный контроль. *(StealthWarnings + docs/stealth.md)*
- [x] Рекомендовать малоизвестные легитимные CDN-домены (например, зона `.okcdn.ru`) или собственный сайт-донор.
- [x] Документировать выбор `dest` и проверку доступности сайта-донора с сервера (`curl -vI`).
- [ ] При появлении в upstream Xray расширенных опций impersonation — подключать через конфиг без форка.

---

### 3. Фрагментация ServerHello

- [ ] Поддержать в конфиге фрагментацию **только первого пакета ServerHello**, не всех пакетов (снижение PPS-аномалии).
- [ ] Параметры в `config.yaml` / UI: включение, размер/стратегия фрагментации (по документации актуальной версии Xray).
- [ ] Значения по умолчанию — консервативные; предупреждение в UI при агрессивной фрагментации.

---

### 4. Multi-hop (цепочка узлов)

Модель `Node` уже есть — использовать для схемы «клиент → узел в РФ → узел за рубежом → интернет».

- [x] Расширить модель `Node`: `role` (`entry` / `exit`), `protocol`, `credentials`, `region`, `priority`.
- [x] Генерация outbound `freedom` / `vless` / `chain` в xray-конфиге entry-узла на exit-узел.
- [x] API CRUD для узлов + привязка пользователей к цепочке (опционально).
- [x] В подписке и `/api/client/config` — отдавать entry-точку как основную, exit — прозрачно на сервере.
- [ ] UI: схема топологии, health-check узлов, переключение active/inactive.

---

### 5. Генерация ссылок и подписок под anti-DPI

Текущие ссылки в `core/links.go` — `VLESS/TCP/security=none`. Нужно:

- [x] `buildVLESSRealityXHTTPLink()` — параметры: `type=xhttp`, `security=reality`, `pbk`, `sid`, `fp`, `path`, `mode=stream-one`.
- [x] `buildVLESSRealityVisionLink()` — `flow=xtls-rprx-vision`, `security=reality`.
- [x] `buildVLESSTLSLink()` — нестандартный порт, `security=tls`, SNI, ALPN; комментарий/hint для mux в JSON-конфиге.
- [x] Подписка `/api/subscription/{token}` — **несколько строк** (все доступные профили пользователя) для fallback на стороне клиента.
- [x] `GET /api/client/config` — массив `profiles[]` с `priority`, `transport`, `tags` (`xhttp-primary`, `vision-ios-fallback`).
- [x] Дефолтный `fingerprint` в ссылках: `firefox` или `edge` (настраиваемо в `config.yaml`).

---

### 6. AmneziaWG (опциональный резервный протокол)

- [ ] Исследовать интеграцию: отдельный Docker-сервис / sidecar или внешний узел, управляемый из панели.
- [ ] Модель `WireGuardPeer` или расширение `User` — ключи AmneziaWG, параметры обфускации (Jc, Jmin, Jmax, S1, S2, H).
- [ ] API: выдача конфига/QR для AmneziaWG-профиля в подписке (отдельная строка или URI-схема).
- [ ] Документация: когда использовать AWG vs Xray-пресеты.

---

### 7. Версионирование и обновление ядра

- [x] Пиновать версию Xray-core в Docker (`ARG XRAY_VERSION`) с регулярным обновлением.
- [x] CI: smoke-тест генерации конфига на новой версии Xray (валидный JSON, `xray run -test`).
- [ ] Changelog/checklist при обновлении ядра: проверить breaking changes в XHTTP/Reality. *(частично в docs/stealth.md)*
- [ ] Watch upstream: issues по Reality fingerprint, XHTTP, DPI — без форка до появления официального решения.

---

### 8. Развёртывание и документация

- [x] Пример `config.yaml` с полным stealth-пресетом (два inbound, Reality, fingerprint).
- [x] Скрипт или раздел в README: генерация Reality keypair (`xray x25519`), выбор `dest`.
- [x] Docker Compose: порты 443 и 8443, `cap_net_bind_service` при необходимости.
- [x] `docs/stealth.md` — объяснение угроз, выбор связки, почему не форк, чеклист перед production.
- [ ] Ansible/terraform примеры (опционально) для entry-узла в РФ + exit за рубежом.

---

### 9. UI (панель управления)

- [x] Страница «Транспорты / Stealth»: включение пресетов, порты, Reality dest, fingerprint по умолчанию.
- [x] Просмотр сгенерированных ссылок по профилям (XHTTP, Vision, TLS, AWG).
- [x] Предупреждения при небезопасных настройках (популярный dest, `chrome` fingerprint, только TCP без XHTTP).
- [x] Тест «проверить доступность dest» с backend-запроса.

---

### 10. Тестирование (сервер)

- [x] Unit-тесты генерации конфига: два inbound, обязательные поля Reality/XHTTP.
- [x] Unit-тесты ссылок: корректные query-параметры для каждого пресета.
- [x] Integration: `xray run -test -c` на сгенерированном конфиге в CI.
- [ ] Документировать ручной чеклист: подключение с тестового клиента через каждый inbound (вне CI, на staging). *(частично в docs/stealth.md)*

---

### Порядок реализации (рекомендуемый)

1. Расширить `config.yaml` и структуры конфига (`core.stealth`, Reality, XHTTP).
2. Два inbound в `xray.json.tmpl` (XHTTP+Reality + Vision+Reality).
3. Генерация мульти-профильных ссылок и подписки.
4. UI для stealth-настроек и просмотра профилей.
5. Multi-hop через модель `Node`.
6. Фрагментация ServerHello (когда поддерживается целевой версией Xray).
7. AmneziaWG как опциональный модуль.
8. Документация и CI smoke-тесты на обновления ядра.
