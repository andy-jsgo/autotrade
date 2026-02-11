import asyncio
import logging
import random
from datetime import datetime, timezone

from sqlalchemy import text
from sqlalchemy.ext.asyncio import create_async_engine


class TraderEngine:
    def __init__(self, db_url: str) -> None:
        self.db = create_async_engine(db_url, pool_pre_ping=True)
        self.interval_seconds = 15

    async def loop(self) -> None:
        while True:
            try:
                await self.trade_once()
            except Exception as exc:  # pragma: no cover
                logging.exception("auto-trader cycle failed: %s", exc)
                await self._set_runtime_error(str(exc))
            await asyncio.sleep(self.interval_seconds)

    async def trade_once(self) -> None:
        async with self.db.begin() as conn:
            row = (
                await conn.execute(
                    text(
                        """
                        SELECT c.auto_trading, COALESCE(ws.agent_approved, false)
                        FROM control_state c
                        LEFT JOIN LATERAL (
                            SELECT agent_approved
                            FROM wallet_sessions
                            ORDER BY updated_at DESC
                            LIMIT 1
                        ) ws ON true
                        WHERE c.id = 1
                        """
                    )
                )
            ).first()

            if not row:
                return

            auto_trading, agent_approved = bool(row[0]), bool(row[1])
            if not auto_trading:
                await conn.execute(
                    text("UPDATE strategy_runtime SET runtime_status = 'paused', updated_at = now() WHERE id = 1")
                )
                return

            if not agent_approved:
                await conn.execute(
                    text(
                        """
                        UPDATE strategy_runtime
                        SET runtime_status = 'blocked', last_error = 'agent not approved', updated_at = now()
                        WHERE id = 1
                        """
                    )
                )
                return

            snap = (
                await conn.execute(
                    text(
                        """
                        SELECT symbol, ohlcv, exec_context
                        FROM market_snapshots
                        ORDER BY captured_at DESC
                        LIMIT 1
                        """
                    )
                )
            ).first()
            if not snap:
                return

            symbol, ohlcv, ctx = snap[0], snap[1], snap[2]
            close_px = float((ohlcv or {}).get("close", 0) or 0)
            volatility = float((ctx or {}).get("volatility", 0) or 0)
            if close_px <= 0:
                return

            side = "Buy" if random.random() > 0.5 else "Sell"
            size = 0.002 if symbol == "BTC" else 0.02
            sl = close_px * (0.995 if side == "Buy" else 1.005)
            tp = close_px * (1.008 if side == "Buy" else 0.992)

            await conn.execute(
                text(
                    """
                    INSERT INTO orders
                    (symbol, side, order_type, size, entry_price, stop_loss, take_profit, status, execution, client_tag)
                    VALUES
                    (:symbol, :side, 'market', :size, :entry, :sl, :tp, 'open', 'paper-auto', :tag)
                    """
                ),
                {
                    "symbol": symbol,
                    "side": side,
                    "size": size,
                    "entry": close_px,
                    "sl": sl,
                    "tp": tp,
                    "tag": f"auto-{datetime.now(timezone.utc).strftime('%H%M%S')}",
                },
            )

            await conn.execute(
                text(
                    """
                    UPDATE strategy_runtime
                    SET runtime_status = 'running',
                        last_signal = :signal,
                        last_error = '',
                        updated_at = now()
                    WHERE id = 1
                    """
                ),
                {
                    "signal": f"{symbol} {side} vol={volatility:.4f}",
                },
            )

            logging.info("auto-trader placed paper order: %s %s", symbol, side)

    async def _set_runtime_error(self, message: str) -> None:
        async with self.db.begin() as conn:
            await conn.execute(
                text(
                    """
                    UPDATE strategy_runtime
                    SET runtime_status = 'error',
                        last_error = :msg,
                        updated_at = now()
                    WHERE id = 1
                    """
                ),
                {"msg": message[:160]},
            )
