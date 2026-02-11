import asyncio
import logging
import os

from collector.engine import CollectorEngine
from derivation.engine import DerivationEngine


logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")


async def run_forever() -> None:
    db_url = os.getenv(
        "DATABASE_URL",
        "postgresql+psycopg://autotrade:autotrade@localhost:5432/autotrade",
    )
    api_base = os.getenv("HYPERLIQUID_API_BASE", "https://api.hyperliquid-testnet.xyz")
    poll_seconds = int(os.getenv("COLLECTOR_POLL_SECONDS", "60"))

    collector = CollectorEngine(db_url=db_url, api_base=api_base, poll_seconds=poll_seconds)
    derivation = DerivationEngine(db_url=db_url)

    await asyncio.gather(
        collector.loop(),
        derivation.loop(),
    )


if __name__ == "__main__":
    asyncio.run(run_forever())
