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

- [ ] OpenAPI `/api/docs`
- [ ] HTTPS / Let's Encrypt example
- [ ] Доп. протоколы (VMess, Trojan) в UI
- [ ] CI pipeline
- [ ] E2E tests
