# Project Memory (HyperClaw V2.5)

## Current Architecture
- `frontend` (Next.js 14): mobile-first control panel with tabs `overview/trade/strategy/review/me`
- `backend-go` (Go): REST API gateway + websocket placeholder
- `python-worker` (Python): collector, derivation engine, and paper auto-trader
- `postgres` (PostgreSQL 16): storage for state/fills/reviews/strategies/orders/snapshots

## Key Runtime Contracts
- Manual order API enforces atomic risk fields: `stopLoss` + `takeProfit` required.
- Auto trading only runs when:
  - wallet session exists
  - agent is approved
  - `control_state.auto_trading = true`
- Frontend browser calls API via `/api` in production through reverse proxy.

## Production/Test Deployment
- Target server: `72.62.247.57`
- Domains:
  - `trade-test.hyperclaw.dev` (frontend + `/api` proxied to backend)
  - `api.trade-test.hyperclaw.dev` (direct backend/ws)
- Reverse proxy on server: `nginx` (ports 80/443 already occupied by host nginx)
- Caddy exists in compose but is not active on this host due port conflict
- Compose profile: base `docker-compose.yml` + `docker-compose.prod.yml`
- Server working tree: `/opt/autotrade`, branch `main`

## Operator Notes
- Wallet connect uses RainbowKit/wagmi.
- Before enabling auto-trade, user must run in order:
  1. connect wallet
  2. sign bind session
  3. approve agent
- Current execution mode is `paper`/`paper-auto` (safe simulation).
- `worker` DB URL is configurable via `WORKER_DATABASE_URL`; it must match DB password in `DATABASE_URL`.
- Current prod `.env` uses:
  - `POSTGRES_PASSWORD=HC_test_2026_secure`
  - `DATABASE_URL=postgres://autotrade:HC_test_2026_secure@postgres:5432/autotrade?sslmode=disable`
  - `WORKER_DATABASE_URL=postgresql+psycopg://autotrade:HC_test_2026_secure@postgres:5432/autotrade`
- Cloudflare DNS records for the two domains are set to DNS only (`proxied=false`) pointing to `72.62.247.57`.
