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

## API endpoints

- `GET /v1/me/state`
- `GET /v1/me/fills?limit=20`
- `POST /v1/me/review`
- `GET /v1/strategy/derives`
- `PATCH /v1/control/bias`

## Notes

- Default network target is Hyperliquid Testnet.
- `worker` will generate mock-safe market snapshots when external fetch fails.
- `ws://localhost:8081` currently sends periodic heartbeat and can be extended for live pushes.
