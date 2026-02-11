package repo

import (
	"context"

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
