const API_BASE =
  process.env.API_BASE_INTERNAL ||
  process.env.NEXT_PUBLIC_API_BASE ||
  "http://localhost:8080";

export type Fill = {
  id: number;
  symbol: string;
  side: string;
  price: number;
  size: number;
  realizedPnl: number;
  status: string;
  createdAt: string;
};

export async function fetchFills(limit = 20): Promise<Fill[]> {
  const res = await fetch(`${API_BASE}/v1/me/fills?limit=${limit}`, { cache: "no-store" });
  if (!res.ok) throw new Error("failed to fetch fills");
  const data = await res.json();
  return data.fills || [];
}

export async function submitReview(payload: {
  fillId: number;
  verdict: "good" | "bad";
  tags: string[];
  notes: string;
}) {
  const res = await fetch(`${API_BASE}/v1/me/review`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload)
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || "failed to submit review");
  }
  return res.json();
}

export async function fetchState() {
  const res = await fetch(`${API_BASE}/v1/me/state`, { cache: "no-store" });
  if (!res.ok) throw new Error("failed to fetch state");
  return res.json();
}
