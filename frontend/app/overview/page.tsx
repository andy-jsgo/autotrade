"use client";

import { useEffect, useState } from "react";
import { fetchState, fetchStrategyStatus, fetchWalletSession, StrategyStatus, WalletSession } from "../../lib/api";

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

export default function OverviewPage() {
  const [equity, setEquity] = useState(0);
  const [leverage, setLeverage] = useState(0);
  const [status, setStatus] = useState<StrategyStatus>(EMPTY_STATUS);
  const [wallet, setWallet] = useState<WalletSession>(EMPTY_WALLET);

  useEffect(() => {
    const load = async () => {
      const [state, strategy, session] = await Promise.all([
        fetchState(),
        fetchStrategyStatus(),
        fetchWalletSession()
      ]);
      setEquity(state.state.equity ?? 0);
      setLeverage(state.state.leverage ?? 0);
      setStatus(strategy);
      setWallet(session);
    };

    load().catch(() => undefined);
    const timer = setInterval(() => load().catch(() => undefined), 4000);
    return () => clearInterval(timer);
  }, []);

  return (
    <main className="container page-content">
      <header>
        <p className="kicker">HyperClaw V2.5</p>
        <h1>总览</h1>
        <div className="statline">
          <span>净值 {equity.toFixed(2)}</span>
          <span>杠杆 {leverage.toFixed(1)}x</span>
          <span>Bias {status.bias}</span>
        </div>
      </header>

      <section className="card block">
        <h3>策略运行</h3>
        <p>状态: {status.runtimeStatus}</p>
        <p>自动交易: {status.autoTrading ? "开启" : "关闭"}</p>
        <p>最近信号: {status.lastSignal || "-"}</p>
        <p className="errorline">{status.lastError || ""}</p>
      </section>

      <section className="card block">
        <h3>账户与权限</h3>
        <p className="mono">钱包: {wallet.address || "未绑定"}</p>
        <p>会话: {wallet.connected ? "已绑定" : "未绑定"}</p>
        <p>Agent: {wallet.agentApproved ? "已授权" : "未授权"}</p>
      </section>
    </main>
  );
}
