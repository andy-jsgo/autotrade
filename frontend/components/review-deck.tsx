"use client";

import { useMemo, useState, TouchEvent } from "react";
import { Fill, submitReview } from "../lib/api";

type Props = { fills: Fill[] };

const defaultTags = ["无量上涨", "强支撑突破", "追高", "逆势抄底"];

export default function ReviewDeck({ fills }: Props) {
  const [index, setIndex] = useState(0);
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [notes, setNotes] = useState("");
  const [pending, setPending] = useState(false);
  const [startX, setStartX] = useState<number | null>(null);

  const current = useMemo(() => fills[index], [fills, index]);

  if (!current) {
    return <p className="empty">今日已复盘完成。</p>;
  }

  const reset = () => {
    setSelectedTags([]);
    setNotes("");
  };

  const toggleTag = (tag: string) => {
    setSelectedTags((prev) =>
      prev.includes(tag) ? prev.filter((t) => t !== tag) : [...prev, tag]
    );
  };

  const mark = async (verdict: "good" | "bad") => {
    setPending(true);
    try {
      await submitReview({ fillId: current.id, verdict, tags: selectedTags, notes });
      setIndex((i) => i + 1);
      reset();
    } finally {
      setPending(false);
    }
  };

  const onTouchStart = (e: TouchEvent<HTMLDivElement>) => {
    setStartX(e.touches[0].clientX);
  };

  const onTouchEnd = async (e: TouchEvent<HTMLDivElement>) => {
    if (startX === null) return;
    const delta = e.changedTouches[0].clientX - startX;
    if (Math.abs(delta) < 50 || pending) return;
    await mark(delta > 0 ? "good" : "bad");
    setStartX(null);
  };

  return (
    <section className="deck">
      <div className="card" onTouchStart={onTouchStart} onTouchEnd={onTouchEnd}>
        <p className="symbol">{current.symbol} · {current.side}</p>
        <h2>{current.realizedPnl >= 0 ? "+" : ""}{current.realizedPnl.toFixed(2)} USDT</h2>
        <p>Price: {current.price.toFixed(2)} | Size: {current.size}</p>
        <p>{new Date(current.createdAt).toLocaleString()}</p>
      </div>

      <div className="tags">
        {defaultTags.map((tag) => (
          <button
            key={tag}
            type="button"
            onClick={() => toggleTag(tag)}
            className={selectedTags.includes(tag) ? "tag active" : "tag"}
          >
            {tag}
          </button>
        ))}
      </div>

      <textarea
        placeholder="补充主观复盘"
        value={notes}
        onChange={(e) => setNotes(e.target.value)}
      />

      <div className="actions">
        <button disabled={pending} onClick={() => mark("bad")} className="bad">左滑/劣</button>
        <button disabled={pending} onClick={() => mark("good")} className="good">右滑/优</button>
      </div>
    </section>
  );
}
