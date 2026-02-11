"use client";

import { useEffect, useState } from "react";
import {
  fetchStrategyDerives,
  fetchStrategyStatus,
  fetchWalletSession,
  setAutoTrading,
  setBias,
  StrategyDerive,
  StrategyStatus,
  WalletSession
} from "../../lib/api";

const EMPTY_STATUS: StrategyStatus = {
  bias: "Hybrid",
  autoTrading: false,
  runtimeStatus: "idle",
  lastSignal: "",
  lastError: "",
  updatedAt: new Date().toISOString()
};

const EMPTY_WALLET: WalletSession = {
  address: "",
  connected: false,
  agentApproved: false,
  agentPubKey: ""
};

export default function StrategyPage() {
  const [status, setStatus] = useState<StrategyStatus>(EMPTY_STATUS);
  const [wallet, setWallet] = useState<WalletSession>(EMPTY_WALLET);
  const [derives, setDerives] = useState<StrategyDerive[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const refresh = async () => {
    const [st, ds, ws] = await Promise.all([
      fetchStrategyStatus(),
      fetchStrategyDerives(),
      fetchWalletSession()
    ]);
    setStatus(st);
    setDerives(ds);
    setWallet(ws);
  };

  useEffect(() => {
    refresh().catch(() => undefined);
    const timer = setInterval(() => refresh().catch(() => undefined), 4000);
    return () => clearInterval(timer);
  }, []);

  const onSetBias = async (bias: "Long" | "Short" | "Hybrid") => {
    setBusy(true);
    setError("");
    try {
      await setBias(bias);
      await refresh();
    } catch (e: any) {
      setError(e.message || "设置失败");
    } finally {
      setBusy(false);
    }
  };

  const onToggleAuto = async () => {
    setBusy(true);
    setError("");
    try {
      await setAutoTrading(!status.autoTrading);
      await refresh();
    } catch (e: any) {
      setError(e.message || "切换失败");
    } finally {
      setBusy(false);
    }
  };

  return (
    <main className="container page-content">
      <header>
        <p className="kicker">HyperClaw V2.5</p>
        <h1>策略</h1>
      </header>

      <section className="card block">
        <h3>运行控制</h3>
        <p>状态: {status.runtimeStatus}</p>
        <p>最近信号: {status.lastSignal || "-"}</p>
        <p className="errorline">{status.lastError || ""}</p>
        <div className="actions triple">
          <button disabled={busy} onClick={() => onSetBias("Long")}>Long</button>
          <button disabled={busy} onClick={() => onSetBias("Short")}>Short</button>
          <button disabled={busy} onClick={() => onSetBias("Hybrid")}>Hybrid</button>
        </div>
        <button
          className={status.autoTrading ? "good" : "bad"}
          disabled={busy || !wallet.agentApproved}
          onClick={onToggleAuto}
        >
          {status.autoTrading ? "停止自动交易" : "启动自动交易"}
        </button>
        {!wallet.agentApproved ? <p className="errorline">需要先完成 Agent 授权</p> : null}
        {error ? <p className="errorline">{error}</p> : null}
      </section>

      <section className="card block">
        <h3>派生策略</h3>
        <div className="list">
          {derives.map((d, idx) => (
            <div key={`${d.name}-${idx}`} className="item">
              <strong>{d.name}</strong>
              <span>Base: {d.baseStrategy}</span>
              <span>WinRate: {(d.winRate * 100).toFixed(2)}%</span>
              <span>PnL Ratio: {d.pnlRatio.toFixed(2)}</span>
              <span>{d.condition}</span>
              <span>建议: {d.recommendation}</span>
            </div>
          ))}
          {derives.length === 0 ? <p>暂无派生策略</p> : null}
        </div>
      </section>
    </main>
  );
}
