"use client";

import { ConnectButton } from "@rainbow-me/rainbowkit";
import { useEffect, useState } from "react";
import { useAccount, useSignMessage } from "wagmi";
import { approveAgent, connectWallet, fetchWalletSession, WalletSession } from "../../lib/api";

const EMPTY_WALLET: WalletSession = {
  address: "",
  connected: false,
  agentApproved: false,
  agentPubKey: ""
};

export default function MePage() {
  const { address, isConnected } = useAccount();
  const { signMessageAsync } = useSignMessage();

  const [wallet, setWallet] = useState<WalletSession>(EMPTY_WALLET);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const refresh = async () => {
    const ws = await fetchWalletSession();
    setWallet(ws);
  };

  useEffect(() => {
    refresh().catch(() => undefined);
  }, []);

  const onBindWallet = async () => {
    setBusy(true);
    setError("");
    try {
      if (!isConnected || !address) throw new Error("请先连接钱包");
      const message = `HyperClaw login nonce ${Date.now()}`;
      const signature = await signMessageAsync({ message });
      await connectWallet({ address, signature, message });
      await refresh();
    } catch (e: any) {
      setError(e.message || "绑定失败");
    } finally {
      setBusy(false);
    }
  };

  const onApproveAgent = async () => {
    setBusy(true);
    setError("");
    try {
      await approveAgent();
      await refresh();
    } catch (e: any) {
      setError(e.message || "授权失败");
    } finally {
      setBusy(false);
    }
  };

  return (
    <main className="container page-content">
      <header>
        <p className="kicker">HyperClaw V2.5</p>
        <h1>我的</h1>
      </header>

      <section className="card block">
        <h3>钱包接入</h3>
        <ConnectButton />
        <p className="mono">Rainbow 地址: {address || "未连接"}</p>
        <p className="mono">会话地址: {wallet.address || "未绑定"}</p>
        <p>会话状态: {wallet.connected ? "已绑定" : "未绑定"}</p>
        <p>Agent 状态: {wallet.agentApproved ? "已授权" : "未授权"}</p>
        <div className="actions">
          <button disabled={busy || !isConnected} onClick={onBindWallet}>签名绑定会话</button>
          <button disabled={busy || !wallet.connected} onClick={onApproveAgent}>授权 Agent</button>
        </div>
        {error ? <p className="errorline">{error}</p> : null}
      </section>
    </main>
  );
}
