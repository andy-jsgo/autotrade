"use client";

import { useEffect, useMemo, useState } from "react";
import { fetchOrders, fetchWalletSession, Order, placeOrder, WalletSession } from "../../lib/api";

const EMPTY_WALLET: WalletSession = {
  address: "",
  connected: false,
  agentApproved: false,
  agentPubKey: ""
};

export default function TradePage() {
  const [wallet, setWallet] = useState<WalletSession>(EMPTY_WALLET);
  const [orders, setOrders] = useState<Order[]>([]);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const [symbol, setSymbol] = useState("BTC");
  const [side, setSide] = useState<"Buy" | "Sell">("Buy");
  const [size, setSize] = useState(0.002);
  const [entryPrice, setEntryPrice] = useState(100000);
  const [stopLoss, setStopLoss] = useState(99500);
  const [takeProfit, setTakeProfit] = useState(100800);

  const canTrade = useMemo(() => wallet.connected && wallet.agentApproved, [wallet]);

  const refresh = async () => {
    const [session, list] = await Promise.all([fetchWalletSession(), fetchOrders(20)]);
    setWallet(session);
    setOrders(list);
  };

  useEffect(() => {
    refresh().catch(() => undefined);
    const timer = setInterval(() => refresh().catch(() => undefined), 4000);
    return () => clearInterval(timer);
  }, []);

  const onPlaceOrder = async () => {
    setBusy(true);
    setError("");
    try {
      await placeOrder({
        symbol,
        side,
        orderType: "market",
        size,
        entryPrice,
        stopLoss,
        takeProfit,
        execution: "paper"
      });
      await refresh();
    } catch (e: any) {
      setError(e.message || "下单失败");
    } finally {
      setBusy(false);
    }
  };

  return (
    <main className="container page-content">
      <header>
        <p className="kicker">HyperClaw V2.5</p>
        <h1>交易</h1>
      </header>

      <section className="card block">
        <h3>手动下单（原子 SL/TP）</h3>
        <p className="mono">状态: {canTrade ? "可交易" : "请先在我的页面绑定并授权"}</p>
        <div className="grid2">
          <label>Symbol<input value={symbol} onChange={(e) => setSymbol(e.target.value.toUpperCase())} /></label>
          <label>Side
            <select value={side} onChange={(e) => setSide(e.target.value as "Buy" | "Sell")}>
              <option>Buy</option>
              <option>Sell</option>
            </select>
          </label>
          <label>Size<input type="number" value={size} onChange={(e) => setSize(Number(e.target.value))} /></label>
          <label>Entry<input type="number" value={entryPrice} onChange={(e) => setEntryPrice(Number(e.target.value))} /></label>
          <label>StopLoss<input type="number" value={stopLoss} onChange={(e) => setStopLoss(Number(e.target.value))} /></label>
          <label>TakeProfit<input type="number" value={takeProfit} onChange={(e) => setTakeProfit(Number(e.target.value))} /></label>
        </div>
        <button disabled={busy || !canTrade} onClick={onPlaceOrder}>提交订单</button>
        {error ? <p className="errorline">{error}</p> : null}
      </section>

      <section className="card block">
        <h3>订单流</h3>
        <div className="list">
          {orders.map((o) => (
            <div key={o.id} className="item">
              <strong>{o.symbol} {o.side}</strong>
              <span>{o.execution} | {o.size} @ {o.entryPrice}</span>
              <span>SL {o.stopLoss} / TP {o.takeProfit}</span>
              <span>{new Date(o.createdAt).toLocaleString()}</span>
            </div>
          ))}
          {orders.length === 0 ? <p>暂无订单</p> : null}
        </div>
      </section>
    </main>
  );
}
