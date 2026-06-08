# Tasks for proxy-mgr MVP

## Backend (Go)
- [ ] init go modules, config reader (viper)
- [ ] sqlite models (gorm) – User, Traffic, Node
- [ ] xray-core / sing-box config generator (templates)
- [ ] API: CRUD users, get link/qr, traffic stats
- [ ] Core controller: restart core on config change, collect stats
- [ ] Telegram bot (go-telegram-bot-api)
- [ ] Basic auth middleware

## Frontend (React)
- [ ] setup vite + tailwind + axios
- [ ] login page (API key)
- [ ] dashboard with traffic chart
- [ ] users table + add/edit/delete
- [ ] view links and QR codes
- [ ] switch between xray/sing-box

## Integration
- [ ] docker-compose for all services
- [ ] nginx reverse proxy config
- [ ] health checks
- [ ] backup scripts

## Deployment
- [ ] .env.example
- [ ] README with install steps