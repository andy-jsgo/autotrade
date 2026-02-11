const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE ||
  (typeof window !== "undefined"
    ? "/api"
    : process.env.API_BASE_INTERNAL || "http://localhost:8080");

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

export type WalletSession = {
  address: string;
  connected: boolean;
  agentApproved: boolean;
  agentPubKey: string;
  updatedAt?: string;
};

export type StrategyStatus = {
  bias: "Long" | "Short" | "Hybrid";
  autoTrading: boolean;
  runtimeStatus: string;
  lastSignal: string;
  lastError: string;
  updatedAt: string;
};

export type StrategyDerive = {
  name: string;
  baseStrategy: string;
  winRate: number;
  pnlRatio: number;
  condition: string;
  recommendation: string;
};

export type Order = {
  id: number;
  symbol: string;
  side: string;
  orderType: string;
  size: number;
  entryPrice: number;
  stopLoss: number;
  takeProfit: number;
  status: string;
  execution: string;
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

export async function fetchWalletSession(): Promise<WalletSession> {
  const res = await fetch(`${API_BASE}/v1/auth/wallet/session`, { cache: "no-store" });
  if (!res.ok) throw new Error("failed to fetch wallet session");
  const data = await res.json();
  return data.session || { address: "", connected: false, agentApproved: false, agentPubKey: "" };
}

export async function connectWallet(payload: { address: string; signature: string; message: string }) {
  const res = await fetch(`${API_BASE}/v1/auth/wallet/connect`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload)
  });
  if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || "connect wallet failed");
  return res.json();
}

export async function approveAgent() {
  const res = await fetch(`${API_BASE}/v1/auth/approve-agent`, { method: "POST" });
  if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || "approve agent failed");
  return res.json();
}

export async function fetchStrategyStatus(): Promise<StrategyStatus> {
  const res = await fetch(`${API_BASE}/v1/strategy/status`, { cache: "no-store" });
  if (!res.ok) throw new Error("failed to fetch strategy status");
  const data = await res.json();
  return data.status;
}

export async function fetchStrategyDerives(): Promise<StrategyDerive[]> {
  const res = await fetch(`${API_BASE}/v1/strategy/derives`, { cache: "no-store" });
  if (!res.ok) throw new Error("failed to fetch strategy derives");
  const data = await res.json();
  return data.derives || [];
}

export async function setAutoTrading(enabled: boolean) {
  const res = await fetch(`${API_BASE}/v1/strategy/auto-trade`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ enabled })
  });
  if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || "toggle auto trade failed");
  return res.json();
}

export async function setBias(bias: "Long" | "Short" | "Hybrid") {
  const res = await fetch(`${API_BASE}/v1/control/bias`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ bias })
  });
  if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || "set bias failed");
  return res.json();
}

export async function placeOrder(payload: {
  symbol: string;
  side: "Buy" | "Sell";
  orderType: "market" | "limit";
  size: number;
  entryPrice: number;
  stopLoss: number;
  takeProfit: number;
  execution: "paper" | "live";
  clientTag?: string;
}) {
  const res = await fetch(`${API_BASE}/v1/trade/order`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload)
  });
  if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || "place order failed");
  return res.json();
}

export async function fetchOrders(limit = 20): Promise<Order[]> {
  const res = await fetch(`${API_BASE}/v1/trade/orders?limit=${limit}`, { cache: "no-store" });
  if (!res.ok) throw new Error("failed to fetch orders");
  const data = await res.json();
  return data.orders || [];
}
