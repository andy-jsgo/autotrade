# HyperClaw V2.5 (Starter)

Go API gateway + Python data/derivation workers + Next.js mobile-first UI + PostgreSQL.

## Services

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- Backend WS placeholder: ws://localhost:8081
- PostgreSQL: localhost:5432

## Quick start

```bash
cp .env.example .env
docker compose up --build
```

## Production deploy (test server)

```bash
cp deploy/.env.prod.example .env
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
```

- Ensure both DB URLs are aligned to the same password:
  - `DATABASE_URL` for Go backend
  - `WORKER_DATABASE_URL` for Python worker (`postgresql+psycopg://...`)

- Caddy routes:
  - `trade-test.hyperclaw.dev` -> frontend (`/api/*` proxied to backend)
  - `api.trade-test.hyperclaw.dev` -> backend
- DNS helper:
  - `scripts/upsert_cloudflare_dns.sh <token> hyperclaw.dev trade-test.hyperclaw.dev api.trade-test.hyperclaw.dev`

## API endpoints

- `GET /v1/me/state`
- `GET /v1/me/fills?limit=20`
- `POST /v1/me/review`
- `GET /v1/strategy/derives`
- `PATCH /v1/control/bias`
- `GET /v1/auth/wallet/session`
- `POST /v1/auth/wallet/connect`
- `POST /v1/auth/approve-agent`
- `GET /v1/strategy/status`
- `PATCH /v1/strategy/auto-trade`
- `POST /v1/trade/order`
- `GET /v1/trade/orders?limit=20`

## Frontend tabs

- `/overview` 总览
- `/trade` 交易
- `/strategy` 策略
- `/review` 复盘
- `/me` 我的

## Notes

- Default network target is Hyperliquid Testnet.
- `worker` will generate mock-safe market snapshots when external fetch fails.
- `ws://localhost:8081` currently sends periodic heartbeat and can be extended for live pushes.
- Auto trading is currently `paper-auto` execution (safe local simulation), requiring wallet connect + agent approval state.
- Frontend wallet connection uses RainbowKit/wagmi. Set `NEXT_PUBLIC_WALLETCONNECT_PROJECT_ID` for WalletConnect support.
