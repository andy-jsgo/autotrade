import asyncio
import json
import logging
from pathlib import Path

from sqlalchemy import text
from sqlalchemy.ext.asyncio import create_async_engine


class DerivationEngine:
    def __init__(self, db_url: str) -> None:
        self.db = create_async_engine(db_url, pool_pre_ping=True)
        self.interval_seconds = 300
        self.out_dir = Path("/app/derived_strategies")
        self.out_dir.mkdir(parents=True, exist_ok=True)

    async def loop(self) -> None:
        while True:
            try:
                await self.derive_once()
            except Exception as exc:  # pragma: no cover
                logging.exception("derivation cycle failed: %s", exc)
            await asyncio.sleep(self.interval_seconds)

    async def derive_once(self) -> None:
        async with self.db.begin() as conn:
            rows = await conn.execute(
                text(
                    """
                    SELECT f.id, f.symbol, f.realized_pnl, ms.open_interest, ms.exec_context
                    FROM fills f
                    LEFT JOIN LATERAL (
                        SELECT open_interest, exec_context
                        FROM market_snapshots ms
                        WHERE ms.symbol = f.symbol
                        ORDER BY captured_at DESC
                        LIMIT 1
                    ) ms ON true
                    WHERE f.status = 'closed' AND f.realized_pnl < 0
                    ORDER BY f.created_at DESC
                    LIMIT 200
                    """
                )
            )
            losses = rows.fetchall()

            if not losses:
                return

            weak_env_count = 0
            for loss in losses:
                exec_ctx = loss.exec_context or {}
                volatility = float(exec_ctx.get("volatility", 0.0))
                oi = float(loss.open_interest or 0.0)
                if volatility >= 0.015 and oi > 0:
                    weak_env_count += 1

            score = weak_env_count / max(len(losses), 1)
            if score < 0.35:
                return

            name = "Strategy_V1_HighVolatilityFilter"
            rule = {
                "name": name,
                "base": "Strategy_V1",
                "filters": {
                    "max_volatility": 0.015,
                    "min_open_interest": 10000000,
                },
                "reason": "Stop-loss attribution: high volatility and unstable OI context",
            }

            await conn.execute(
                text(
                    """
                    INSERT INTO strategy_derives (name, base_strategy, win_rate, pnl_ratio, condition, recommendation)
                    VALUES (:name, :base, :win_rate, :pnl_ratio, :condition, :recommendation)
                    """
                ),
                {
                    "name": name,
                    "base": "Strategy_V1",
                    "win_rate": round(0.52 + score * 0.2, 4),
                    "pnl_ratio": round(1.1 + score * 0.7, 4),
                    "condition": "volatility >= 0.015 with unstable OI",
                    "recommendation": "switch_candidate" if score > 0.5 else "watch",
                },
            )

            path = self.out_dir / f"{name}.json"
            path.write_text(json.dumps(rule, ensure_ascii=True, indent=2), encoding="utf-8")
            logging.info("derived strategy updated at %s", path)
