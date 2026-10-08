# RioNexGate — план доработок безопасности и hardening

**Статус:** P0–P2 done; P3 бэклог  

**Контекст:** аудит MVP (2026-10-08). Панель — single-admin self-hosted; дефолтный деплой небезопасен для internet-facing.

Связано с: [.cursor/plans/mvp.md](mvp.md) · PR: https://github.com/RioTwWks/RioNexGate/pull/30

---

## Фазы

```mermaid
flowchart TD
  P0[P0 Critical / High] --> P1[P1 Medium]
  P1 --> P2[P2 Hardening]
  P2 --> P3[P3 Ops / UX]
```

| Фаза | Цель | Состояние |
|------|------|-----------|
| **P0** | Закрыть угон доступа и утечку секретов в дефолте | сделано |
| **P1** | Enforce лимитов, SSRF/injection, nginx | сделано |
| **P2** | Контейнеры, CORS, deps, бэкапы | сделано |
| **P3** | Invite-токены, ротация ключей, OpenAPI auth | частично (invites) |

---

## P0 — Critical / High (этот PR)

### P0.1 Закрыть `POST /api/client/register`
- [x] По умолчанию регистрация **не открыта**
- [x] Разрешено при: `X-API-Key` **или** `X-Registration-Secret` (= `server.registration_secret`)
- [x] Opt-in `server.allow_open_register: true` только для доверенной LAN (логировать warning)
- [x] Обновить OpenAPI, README, e2e, integration tests

### P0.2 Отвергать слабый / пустой `api_key` при старте
- [x] Пустой ключ и плейсхолдеры (`change-me-to-secure-key`, `change-me`, …) → fatal
- [x] Constant-time compare в middleware
- [x] `make init` генерирует случайный `api_key`

### P0.3 Xray Stats API не на `0.0.0.0`
- [x] Ввести `core.xray.api_listen` (bind); `api_address` — куда ходит backend
- [x] Default listen: `127.0.0.1:<port>` (не `0.0.0.0`)
- [x] Пример конфига + комментарий про Docker bridge / firewall

### P0.4 Не публиковать backend/frontend наружу
- [x] Compose: `127.0.0.1:8080:8080`, `127.0.0.1:3000:80`
- [x] Единая точка входа — nginx `:8888`

### P0.5 Права на конфиги с ключами
- [x] `writeConfigAtomic` / AWG → `0600`

### P0.6 Device token не в query + логи
- [x] Убрать `?token=` из device auth
- [x] Access log без query string

### P0.7 `restore.sh` path traversal
- [x] Проверка членов архива + extract только `data/`

---

## P1 — Medium (следом / частично в этом PR)

### P1.1 Enforce `ExpiresAt` (и опционально квоты)
- [x] `ListActiveUsers`: `active AND (expires_at IS zero OR expires_at > now)`
- [x] Device auth / subscription: reject expired
- [x] Quota: `TrafficGB` enforced for access + exclude from active core users

### P1.2 JSON-escape в шаблонах ядер
- [x] `jsonString` / `jsonStringList` через `encoding/json`

### P1.3 Nginx: headers + rate limit
- [x] `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`
- [x] `limit_req` на `/api/`

### P1.4 SSRF-ограничения
- [x] Блок private/link-local/metadata для `test-dest` и node health
- [x] Opt-in `server.allow_private_probes` для LAN multihop
- [x] Dial по конкретному IP (анти DNS-rebinding)

---

## P2 — Hardening

- [x] CORS: явные origins (localhost panel + Vite); `cors_origins` в конфиге
- [x] Не класть default `config.yaml` в образ (только `config.example.yaml`)
- [x] SQLite `0600`; backup archive `0600` + warning
- [x] MaxBytesReader 1 MiB на body
- [x] axios bump; npm audit fix (остались breaking: vite/react-router major)
- [x] `server.enable_docs` (default true; выключать в prod)
- [x] Non-root backend/frontend/nginx; `read_only` + `tmpfs`; `no-new-privileges`; `cap_drop: ALL`
- [x] Pin `sing-box` tag (v1.14.2; override via `SINGBOX_IMAGE`)
- [x] Pin amneziawg tag `3.1.20260828`; drop `SYS_MODULE`; `cap_drop: ALL` + `NET_ADMIN`
- [x] Шифрование бэкапов через `age` (`AGE_RECIPIENT` / `AGE_IDENTITY`)
- [x] Security checklist в README (prod deploy)

---

## P3 — Ops / продукт (бэклог)

- [x] Per-user invite / one-time registration tokens (`/users/{id}/invites`, `invite_token`)
- [ ] Ротация API key из UI
- [ ] Подписка: короткий TTL / signed URL (уже есть token URL — усилить)
- [ ] Авто-отключение по трафику (quota уже режет доступ; UI/auto `active=false` — отдельно)

---

## Критерии приёмки P0

1. Без API key / registration secret → `POST /api/client/register` = 401  
2. `api_key: change-me-to-secure-key` → процесс не стартует  
3. Сгенерированный xray config: API listen на `127.0.0.1`, не `0.0.0.0` (если не задан override)  
4. `docker compose` публикует 8080/3000 только на loopback  
5. Конфиги ядер на диске mode `0600`  
6. `?token=` больше не принимает device auth  
7. `restore.sh` не извлекает `../` из архива  
8. `go test ./...` зелёный  

---

## Вне скоупа этого плана

- Полный multi-user RBAC / OAuth  
- Переписывание стека (не Docker → systemd и т.п.)  
- Идеальная защита от DPI на стороне клиента  
