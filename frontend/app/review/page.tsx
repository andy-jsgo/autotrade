"use client";

import { useEffect, useState } from "react";
import ReviewDeck from "../../components/review-deck";
import { fetchFills, Fill } from "../../lib/api";

export default function ReviewPage() {
  const [fills, setFills] = useState<Fill[]>([]);

  useEffect(() => {
    const load = async () => {
      const list = await fetchFills(20);
      setFills(list);
    };
    load().catch(() => undefined);
    const timer = setInterval(() => load().catch(() => undefined), 5000);
    return () => clearInterval(timer);
  }, []);

  return (
    <main className="container page-content">
      <header>
        <p className="kicker">HyperClaw V2.5</p>
        <h1>复盘</h1>
      </header>
      <ReviewDeck fills={fills} />
    </main>
  );
}
