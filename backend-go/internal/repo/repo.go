package repo

import (
	"context"
	"time"

	"autotrade/backend-go/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) GetState(ctx context.Context) (model.AccountState, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(1000 + realized_pnl), 1000) AS equity,
			5.0 AS leverage,
			COALESCE(SUM(CASE WHEN status = 'open' THEN realized_pnl ELSE 0 END), 0) AS open_pnl,
			now()
		FROM fills;
	`)

	var s model.AccountState
	err := row.Scan(&s.Equity, &s.Leverage, &s.OpenPnL, &s.UpdatedAt)
	return s, err
}

func (r *Repo) GetFills(ctx context.Context, limit int) ([]model.Fill, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, symbol, side, price, size, realized_pnl, status, created_at
		FROM fills
		ORDER BY created_at DESC
		LIMIT $1;
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fills := make([]model.Fill, 0, limit)
	for rows.Next() {
		var f model.Fill
		if err := rows.Scan(&f.ID, &f.Symbol, &f.Side, &f.Price, &f.Size, &f.RealizedPnL, &f.Status, &f.CreatedAt); err != nil {
			return nil, err
		}
		fills = append(fills, f)
	}
	return fills, rows.Err()
}

func (r *Repo) SaveReview(ctx context.Context, in model.ReviewInput) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO reviews (fill_id, verdict, tags, notes)
		VALUES ($1, $2, $3, $4);
	`, in.FillID, in.Verdict, in.Tags, in.Notes)
	return err
}

func (r *Repo) GetDerives(ctx context.Context) ([]model.StrategyDerive, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT name, base_strategy, win_rate, pnl_ratio, condition, recommendation
		FROM strategy_derives
		ORDER BY created_at DESC;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StrategyDerive, 0)
	for rows.Next() {
		var d model.StrategyDerive
		if err := rows.Scan(&d.Name, &d.BaseStrategy, &d.WinRate, &d.PnLRatio, &d.Condition, &d.Recommendation); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func (r *Repo) SetBias(ctx context.Context, bias string) error {
	_, err := r.pool.Exec(ctx, `UPDATE control_state SET bias = $1, updated_at = now() WHERE id = 1;`, bias)
	return err
}

func (r *Repo) GetBias(ctx context.Context) (string, error) {
	var bias string
	err := r.pool.QueryRow(ctx, `SELECT bias FROM control_state WHERE id = 1`).Scan(&bias)
	return bias, err
}

func (r *Repo) SaveWalletSession(ctx context.Context, address, signature, message string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO wallet_sessions (address, connected, signature, message, updated_at)
		VALUES ($1, true, $2, $3, now());
	`, address, signature, message)
	return err
}

func (r *Repo) GetLatestWalletSession(ctx context.Context) (model.WalletSession, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT address, connected, agent_approved, agent_pub_key, updated_at
		FROM wallet_sessions
		ORDER BY updated_at DESC
		LIMIT 1;
	`)
	var out model.WalletSession
	err := row.Scan(&out.Address, &out.Connected, &out.AgentApproved, &out.AgentPubKey, &out.UpdatedAt)
	return out, err
}

func (r *Repo) ApproveAgent(ctx context.Context, agentPubKey string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE wallet_sessions
		SET agent_approved = true, agent_pub_key = $1, updated_at = now()
		WHERE id = (
			SELECT id FROM wallet_sessions ORDER BY updated_at DESC LIMIT 1
		);
	`, agentPubKey)
	return err
}

func (r *Repo) StrategyStatus(ctx context.Context) (model.StrategyStatus, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT c.bias, c.auto_trading, sr.runtime_status, sr.last_signal, sr.last_error, GREATEST(c.updated_at, sr.updated_at)
		FROM control_state c
		JOIN strategy_runtime sr ON sr.id = 1
		WHERE c.id = 1;
	`)
	var out model.StrategyStatus
	err := row.Scan(&out.Bias, &out.AutoTrading, &out.RuntimeStatus, &out.LastSignal, &out.LastError, &out.UpdatedAt)
	return out, err
}

func (r *Repo) SetAutoTrading(ctx context.Context, enabled bool) error {
	runtimeStatus := "paused"
	if enabled {
		runtimeStatus = "running"
	}
	if _, err := r.pool.Exec(ctx, `
		UPDATE control_state SET auto_trading = $1, updated_at = now() WHERE id = 1;
	`, enabled); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE strategy_runtime SET runtime_status = $1, updated_at = now() WHERE id = 1;
	`, runtimeStatus)
	return err
}

func (r *Repo) UpdateRuntimeSignal(ctx context.Context, signal, errMsg string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE strategy_runtime
		SET last_signal = $1, last_error = $2, updated_at = now()
		WHERE id = 1;
	`, signal, errMsg)
	return err
}

func (r *Repo) CreateOrder(ctx context.Context, in model.OrderInput) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO orders
		(symbol, side, order_type, size, entry_price, stop_loss, take_profit, execution, client_tag, status)
		VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9, 'open')
		RETURNING id;
	`, in.Symbol, in.Side, in.OrderType, in.Size, in.EntryPrice, in.StopLoss, in.TakeProfit, in.Execution, in.ClientTag).Scan(&id)
	return id, err
}

func (r *Repo) CreateFillFromOrder(ctx context.Context, in model.OrderInput) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO fills (symbol, side, price, size, realized_pnl, status, created_at)
		VALUES ($1, $2, $3, $4, 0, 'open', now());
	`, in.Symbol, in.Side, in.EntryPrice, in.Size)
	return err
}

func (r *Repo) GetOrders(ctx context.Context, limit int) ([]model.Order, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, symbol, side, order_type, size, entry_price, stop_loss, take_profit, status, execution, created_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT $1;
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.Order, 0, limit)
	for rows.Next() {
		var it model.Order
		if err := rows.Scan(&it.ID, &it.Symbol, &it.Side, &it.OrderType, &it.Size, &it.EntryPrice, &it.StopLoss, &it.TakeProfit, &it.Status, &it.Execution, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *Repo) IsAgentApproved(ctx context.Context) (bool, error) {
	var approved bool
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(agent_approved, false)
		FROM wallet_sessions
		ORDER BY updated_at DESC
		LIMIT 1;
	`).Scan(&approved)
	return approved, err
}

func (r *Repo) TouchSignal(ctx context.Context, signal string) {
	_, _ = r.pool.Exec(ctx, `
		UPDATE strategy_runtime SET last_signal = $1, updated_at = $2 WHERE id = 1;
	`, signal, time.Now().UTC())
}
