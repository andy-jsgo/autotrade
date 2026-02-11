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
- Reverse proxy: Caddy (`deploy/Caddyfile`)
- Compose profile: base `docker-compose.yml` + `docker-compose.prod.yml`

## Operator Notes
- Wallet connect uses RainbowKit/wagmi.
- Before enabling auto-trade, user must run in order:
  1. connect wallet
  2. sign bind session
  3. approve agent
- Current execution mode is `paper`/`paper-auto` (safe simulation).
