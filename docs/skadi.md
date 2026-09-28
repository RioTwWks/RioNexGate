# SkadiCore integration

RioNexGate can drive [SkadiCore](https://github.com/RioTwWks/SkadiCore) as a third proxy core alongside Xray and sing-box.

## Enable

1. Set `core.type: skadi` in `backend/config.yaml` (or switch in **Settings**).
2. Configure Reality keys under `core.stealth` (same as Xray). Prefer generating with:

   ```bash
   # from a SkadiCore release binary or container
   skadicore genkey reality
   ```

3. Start the core container:

   ```bash
   make init
   make up                 # panel
   make dev-cores-skadi    # SkadiCore (profile cores)
   ```

Generated config: `data/skadi/skadi.toml` (mounted at `/config/skadi.toml`).

## Behaviour vs Xray

| Topic | SkadiCore |
|-------|-----------|
| Config format | TOML (`skadi.toml.tmpl`) |
| Listen | Single port (XHTTP port preferred when stealth XHTTP is on) |
| Reality | `[transport.reality]` |
| XHTTP | `[transport.xhttp]` on the same listen port |
| Vision (`xtls-rprx-vision`) | Not supported — panel emits TCP+Reality **without** Vision flow |
| Multi-hop chain outbounds | Not generated (same limitation as sing-box) |
| Per-user traffic stats | Not polled yet (Xray Stats API only) |
| Hot reload | Container restarts on config mtime change (`scripts/skadi-run.sh`) |

## Client links

When `core.type` is `skadi`, subscription / RioNexTunnel profiles are:

1. **skadi-xhttp** — VLESS + Reality + XHTTP (if XHTTP enabled)
2. **skadi-reality-tcp** — VLESS + Reality + TCP (no Vision flow)
3. Optional AWG reserve (unchanged)

## Version pin

`SKADI_VERSION` in `.env` / `.env.example` (default **`0.1.3`**, latest stable as of 2026-09-27) selects the GitHub Release used by `Dockerfile.skadi`.

Upstream release: [SkadiCore v0.1.3](https://github.com/RioTwWks/SkadiCore/releases/tag/v0.1.3) (rustls REALITY 0.23.45, dual-stack `listen`, stricter REALITY key validation, SSRF hardening on outbound).

## Config validation

Generated `skadi.toml` matches SkadiCore server schema (`[server]`, `[protocol.vless]`, `[transport.reality]`, optional `[transport.xhttp]`, `[api]`, `[metrics]`).

Backend test `TestSkadiConfigValidate` runs `skadicore check-config` when a binary is available:

```bash
export SKADICORE_BIN=/path/to/skadicore   # optional
export RUN_SKADI_TEST=1
cd backend && CGO_ENABLED=1 go test ./internal/core/ -run TestSkadiConfigValidate -v
```
