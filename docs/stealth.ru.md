**[English](stealth.md)** | Русский

# Stealth / Anti-DPI — конфигурация

RioNexGate генерирует конфигурации Xray-core и sing-box, устойчивые к современному DPI (2026), на базе **официального Xray-core** — без форков.

## Транспортные профили

| Приоритет | Стек | Порт (по умолчанию) | Назначение |
|-----------|------|---------------------|------------|
| Основной | VLESS + Reality + XHTTP (`mode: stream-one`) | 443 | Лучшая маскировка трафика; рекомендуемый default |
| Запасной | VLESS + Reality + Vision (TCP) | 8443 | iOS и клиенты без XHTTP |
| Мобильный | VLESS + TLS (опционально) | настраивается | Последний вариант; включите `mux` на клиенте (concurrency 8) |
| Резерв | AmneziaWG (опционально) | 51820 | UDP-запас, когда стек Xray заблокирован |

Каждый активный пользователь получает **по одной ссылке на каждый включённый inbound** для fallback в подписке.

## Почему только официальный Xray-core

- Форки вроде `fwflunky/REALITY-rkn-fix` без тестов и отстают от upstream.
- Исправления безопасности и улучшения XHTTP/Reality сначала попадают в [XTLS/Xray-core](https://github.com/XTLS/Xray-core).
- RioNexGate фиксирует проверенный `XRAY_VERSION` в Docker и валидирует конфиги через `xray run -test` в CI.

## Конфигурация (`core.stealth`)

Полный пример: [`backend/config.example.yaml`](../backend/config.example.yaml).

```yaml
core:
  public_host: "your.domain"
  stealth:
    enabled: true
    fingerprint: firefox    # или edge — избегайте chrome/safari по умолчанию
    reality:
      dest: "cdn.example.com:443"
      server_names: ["cdn.example.com"]
      private_key: "..."
      public_key: "..."
      short_ids: ["a1b2c3d4"]
    xhttp:
      enabled: true
      port: 443
      path: "/api/v1/data"
      mode: stream-one      # обязательно — не используйте auto
    vision:
      enabled: true
      port: 8443
```

### Reality keypair

```bash
# В контейнере xray или с установленным бинарником xray:
xray x25519
# PrivateKey: ...  → core.stealth.reality.private_key
# Password: ...    → core.stealth.reality.public_key (pbk в клиентских ссылках)
```

### Выбор `dest` (цель имперсонации)

1. Выберите **малоизвестный легитимный HTTPS-сайт** в CDN-зоне, которой вы доверяете (например `.okcdn.ru`), а не популярные цели (`yahoo.com`, `vk.com`) — к ним повышенное внимание.
2. Проверьте доступность с сервера:

```bash
curl -vI --connect-timeout 5 "https://cdn.example.com"
```

3. Укажите `dest` как `host:443`, а `server_names` — SNI hostname TLS.

Страница **Stealth** в панели и `POST /api/stealth/test-dest` могут проверить доступность.

### Fingerprint

По умолчанию `firefox` или `edge`. Настраивается через `core.stealth.fingerprint`. Передаётся клиентам как `fp=` в VLESS URL.

### Фрагментация ServerHello (`core.stealth.fragmentation`)

Xray **v26.3.27** поддерживает `finalmask.fragment` с `packets: tlshello` (первый TLS record) или `1-3` (агрессивно). По умолчанию: выключено, `length: 50-100`, `delay: 10-20`, `max_split: 2-4`.

**Частично:** REALITY inbound пока не поддерживает finalmask — падает Xray ([#6453](https://github.com/XTLS/Xray-core/issues/6453)). Фрагментация выдаётся только на VLESS+TLS inbound. Панель: `GET`/`PUT` `/api/stealth/settings`.

## Мульти-хоп узлы

Entry и exit узлы позволяют маршрутизировать пользователей через цепочку (например entry в одном регионе, exit в другом).

1. Создайте узлы на странице **Nodes** или через `POST /api/nodes` (`role`: `entry` или `exit`).
2. Привяжите пользователя: `PUT /api/users/{id}/chain` с `entry_node_id` / `exit_node_id`.
3. Health check: `GET /api/nodes/{id}/health`.

Сгенерированные ссылки и конфиги RioNexTunnel используют хост entry-узла, если он задан.

## Развёртывание

1. Включите `core.stealth` в `backend/config.yaml`.
2. Откройте на хосте порты **443** и **8443** (профиль xray-core использует `network_mode: host`).
3. Запустите `make dev-cores` после создания пользователей в панели.
4. Получите мульти-профильные ссылки: `GET /api/users/{id}/link?all=true`.

Docker фиксирует Xray через `XRAY_VERSION` (по умолчанию `26.3.27`) в `docker-compose.yml` / `Dockerfile.xray`.

## Клиентские ссылки

Генерируются в `core/links.go`:

- **XHTTP**: `type=xhttp`, `security=reality`, `pbk`, `sid`, `fp`, `path`, `mode=stream-one`
- **Vision**: `flow=xtls-rprx-vision`, `security=reality`, `type=tcp`
- **TLS**: `security=tls`, `sni`, `alpn`

Подписка (`GET /api/subscription/{token}`): base64 со ссылками через перевод строки (плюс `awg://` при включённом AWG).

### RioNexTunnel `GET /api/client/config`

Каждое зарегистрированное устройство получает JSON с транспортными профилями и `config_hash`:

```json
{
  "profiles": [
    {
      "priority": 1,
      "transport": "xhttp",
      "tags": "xhttp-primary",
      "port": 443,
      "link": "vless://..."
    }
  ],
  "config_hash": "sha256..."
}
```

Аутентификация: заголовок `X-Device-Token`. При изменении конфига сервера устройства получают `refresh_config` через `GET /api/client/commands`.

## Чеклист перед production

- [ ] Сгенерирован Reality keypair; `public_key` соответствует `private_key`
- [ ] `dest` доступен с сервера (`curl -vI` или тест в панели)
- [ ] XHTTP `mode` = `stream-one`
- [ ] Порты 443 и 8443 слушают (`ss -tlnp | grep -E '443|8443'`)
- [ ] `xray run -test -c data/xray/config.json` проходит
- [ ] Каждый inbound проверен с реального клиента на staging
- [ ] Зафиксирован `XRAY_VERSION`; перед обновлением читайте release notes Xray
- [ ] При использовании узлов: health-check entry/exit перед привязкой пользователей

## Известные ограничения

- Фрагментация ServerHello на REALITY inbound (падение upstream — issue #6453)
- Агрессивная фрагментация может ломать часть клиентов; тестируйте перед production

## Модель угроз (сторона сервера)

| Угроза | Смягчение |
|--------|-----------|
| Фингерпринтинг VLESS+Reality TCP | Основной путь — XHTTP (`stream-one`) |
| Длинный TCP без HTTP после TLS | XHTTP имитирует HTTP upload/download |
| Статическая структура Reality cert | Разнообразный `dest`; следите за опциями имперсонации upstream |
| Блокировка одного транспорта | Несколько inbound + мульти-ссылочная подписка |
| Блокировка стека Xray | Опциональный UDP-резерв AmneziaWG |

## Обновление Xray-core

1. Проверьте [релизы Xray](https://github.com/XTLS/Xray-core/releases) на breaking changes в XHTTP/Reality.
2. Поднимите `XRAY_VERSION` в `.env` или `docker-compose.yml`.
3. Пересоберите: `docker compose --profile cores build xray-core`.
4. Запустите `xray run -test -c ./data/xray/config.json`.
5. Smoke-test всех включённых профилей с клиента.

## AmneziaWG (UDP-резерв)

Опциональный fallback, когда стек Xray заблокирован. См. `core.stealth.awg` в [`backend/config.example.yaml`](../backend/config.example.yaml).

| Сценарий | Использование |
|----------|---------------|
| Anti-DPI по умолчанию | XHTTP + Reality |
| Xray заблокирован | UDP-резерв AmneziaWG |

Подписка добавляет строку `awg://`; API profiles возвращает сырой INI в поле `config`.

```bash
docker compose --profile awg up -d amneziawg
```

## См. также

- [README.ru.md](../README.ru.md) — установка и обзор панели
- [docs/README.md](README.md) — индекс документации
