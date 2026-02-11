import ReviewDeck from "../components/review-deck";
import { fetchFills, fetchState } from "../lib/api";

export default async function Home() {
  const [state, fills] = await Promise.all([fetchState(), fetchFills(20)]);

  return (
    <main className="container">
      <header>
        <p className="kicker">HyperClaw V2.5</p>
        <h1>移动复盘台</h1>
        <div className="statline">
          <span>净值 {state.state.equity.toFixed(2)}</span>
          <span>杠杆 {state.state.leverage.toFixed(1)}x</span>
          <span>Bias {state.bias}</span>
        </div>
      </header>
      <ReviewDeck fills={fills} />
    </main>
  );
}
