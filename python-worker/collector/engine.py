import asyncio
import json
import logging
from datetime import datetime, timezone

import aiohttp
from sqlalchemy import text
from sqlalchemy.ext.asyncio import create_async_engine


class CollectorEngine:
    def __init__(self, db_url: str, api_base: str, poll_seconds: int) -> None:
        self.db = create_async_engine(db_url, pool_pre_ping=True)
        self.api_base = api_base.rstrip("/")
        self.poll_seconds = poll_seconds
        self.symbols = ["BTC", "ETH"]
        self.timeframes = ["1m", "5m", "1h"]

    async def loop(self) -> None:
        while True:
            try:
                await self.collect_once()
            except Exception as exc:  # pragma: no cover
                logging.exception("collector cycle failed: %s", exc)
            await asyncio.sleep(self.poll_seconds)

    async def collect_once(self) -> None:
        payloads = await self._fetch_market_data()
        now = datetime.now(timezone.utc)

        async with self.db.begin() as conn:
            for item in payloads:
                await conn.execute(
                    text(
                        """
                        INSERT INTO market_snapshots
                        (symbol, timeframe, ohlcv, open_interest, funding_rate, exec_context, captured_at)
                        VALUES
                        (:symbol, :timeframe, cast(:ohlcv as jsonb), :oi, :funding, cast(:ctx as jsonb), :captured_at)
                        """
                    ),
                    {
                        "symbol": item["symbol"],
                        "timeframe": item["timeframe"],
                        "ohlcv": json.dumps(item["ohlcv"]),
                        "oi": item["open_interest"],
                        "funding": item["funding_rate"],
                        "ctx": json.dumps(item["exec_context"]),
                        "captured_at": now,
                    },
                )
        logging.info("collector saved %s snapshots", len(payloads))

    async def _fetch_market_data(self) -> list[dict]:
        url = f"{self.api_base}/info"
        results: list[dict] = []

        try:
            async with aiohttp.ClientSession(timeout=aiohttp.ClientTimeout(total=8)) as session:
                for symbol in self.symbols:
                    funding = await self._funding_rate(session, url, symbol)
                    oi = await self._open_interest(session, url, symbol)
                    for tf in self.timeframes:
                        candle = await self._candle(session, url, symbol, tf)
                        results.append(
                            {
                                "symbol": symbol,
                                "timeframe": tf,
                                "ohlcv": candle,
                                "open_interest": oi,
                                "funding_rate": funding,
                                "exec_context": {
                                    "rsi": 52.0,
                                    "volatility": 0.017,
                                    "btc_correlation": 1.0 if symbol == "BTC" else 0.84,
                                    "source": "hyperliquid-testnet",
                                },
                            }
                        )
            return results
        except Exception as exc:
            logging.warning("collector fallback to mock data: %s", exc)
            return self._mock_data()

    async def _funding_rate(self, session: aiohttp.ClientSession, url: str, symbol: str) -> float:
        req = {"type": "metaAndAssetCtxs"}
        async with session.post(url, json=req) as resp:
            body = await resp.json()
            asset_ctxs = body[1] if isinstance(body, list) and len(body) > 1 else []
            for ctx in asset_ctxs:
                if ctx.get("coin") == symbol:
                    return float(ctx.get("funding", 0.0))
        return 0.0

    async def _open_interest(self, session: aiohttp.ClientSession, url: str, symbol: str) -> float:
        req = {"type": "metaAndAssetCtxs"}
        async with session.post(url, json=req) as resp:
            body = await resp.json()
            asset_ctxs = body[1] if isinstance(body, list) and len(body) > 1 else []
            for ctx in asset_ctxs:
                if ctx.get("coin") == symbol:
                    return float(ctx.get("openInterest", 0.0))
        return 0.0

    async def _candle(self, session: aiohttp.ClientSession, url: str, symbol: str, timeframe: str) -> dict:
        req = {
            "type": "candleSnapshot",
            "req": {
                "coin": symbol,
                "interval": timeframe,
                "startTime": int((datetime.now(timezone.utc).timestamp() - 3600) * 1000),
                "endTime": int(datetime.now(timezone.utc).timestamp() * 1000),
            },
        }
        async with session.post(url, json=req) as resp:
            candles = await resp.json()
            if isinstance(candles, list) and candles:
                c = candles[-1]
                return {
                    "open": float(c.get("o", 0)),
                    "high": float(c.get("h", 0)),
                    "low": float(c.get("l", 0)),
                    "close": float(c.get("c", 0)),
                    "volume": float(c.get("v", 0)),
                }
        return {"open": 0, "high": 0, "low": 0, "close": 0, "volume": 0}

    def _mock_data(self) -> list[dict]:
        base = []
        for symbol in self.symbols:
            for tf in self.timeframes:
                base.append(
                    {
                        "symbol": symbol,
                        "timeframe": tf,
                        "ohlcv": {
                            "open": 100000.0 if symbol == "BTC" else 3000.0,
                            "high": 100800.0 if symbol == "BTC" else 3025.0,
                            "low": 99800.0 if symbol == "BTC" else 2988.0,
                            "close": 100350.0 if symbol == "BTC" else 3010.0,
                            "volume": 120.0,
                        },
                        "open_interest": 25000000.0,
                        "funding_rate": 0.0001,
                        "exec_context": {
                            "rsi": 49.5,
                            "volatility": 0.016,
                            "btc_correlation": 1.0 if symbol == "BTC" else 0.82,
                            "source": "mock",
                        },
                    }
                )
        return base
