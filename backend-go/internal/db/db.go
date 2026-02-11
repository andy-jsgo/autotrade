package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS fills (
			id BIGSERIAL PRIMARY KEY,
			symbol TEXT NOT NULL,
			side TEXT NOT NULL,
			price DOUBLE PRECISION NOT NULL,
			size DOUBLE PRECISION NOT NULL,
			realized_pnl DOUBLE PRECISION NOT NULL DEFAULT 0,
			status TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
		`CREATE TABLE IF NOT EXISTS reviews (
			id BIGSERIAL PRIMARY KEY,
			fill_id BIGINT NOT NULL REFERENCES fills(id) ON DELETE CASCADE,
			verdict TEXT NOT NULL,
			tags TEXT[] NOT NULL DEFAULT '{}',
			notes TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
		`CREATE TABLE IF NOT EXISTS strategy_derives (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			base_strategy TEXT NOT NULL,
			win_rate DOUBLE PRECISION NOT NULL,
			pnl_ratio DOUBLE PRECISION NOT NULL,
			condition TEXT NOT NULL,
			recommendation TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
		`CREATE TABLE IF NOT EXISTS market_snapshots (
			id BIGSERIAL PRIMARY KEY,
			symbol TEXT NOT NULL,
			timeframe TEXT NOT NULL,
			ohlcv JSONB NOT NULL,
			open_interest DOUBLE PRECISION,
			funding_rate DOUBLE PRECISION,
			exec_context JSONB NOT NULL DEFAULT '{}'::jsonb,
			captured_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
		`CREATE TABLE IF NOT EXISTS control_state (
			id SMALLINT PRIMARY KEY,
			bias TEXT NOT NULL,
			auto_trading BOOLEAN NOT NULL DEFAULT false,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
		`ALTER TABLE control_state ADD COLUMN IF NOT EXISTS auto_trading BOOLEAN NOT NULL DEFAULT false;`,
		`CREATE TABLE IF NOT EXISTS wallet_sessions (
			id BIGSERIAL PRIMARY KEY,
			address TEXT NOT NULL,
			connected BOOLEAN NOT NULL DEFAULT true,
			agent_approved BOOLEAN NOT NULL DEFAULT false,
			agent_pub_key TEXT NOT NULL DEFAULT '',
			signature TEXT NOT NULL DEFAULT '',
			message TEXT NOT NULL DEFAULT '',
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
		`CREATE TABLE IF NOT EXISTS strategy_runtime (
			id SMALLINT PRIMARY KEY,
			runtime_status TEXT NOT NULL DEFAULT 'idle',
			last_signal TEXT NOT NULL DEFAULT '',
			last_error TEXT NOT NULL DEFAULT '',
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
		`CREATE TABLE IF NOT EXISTS orders (
			id BIGSERIAL PRIMARY KEY,
			symbol TEXT NOT NULL,
			side TEXT NOT NULL,
			order_type TEXT NOT NULL,
			size DOUBLE PRECISION NOT NULL,
			entry_price DOUBLE PRECISION NOT NULL,
			stop_loss DOUBLE PRECISION NOT NULL,
			take_profit DOUBLE PRECISION NOT NULL,
			status TEXT NOT NULL DEFAULT 'open',
			execution TEXT NOT NULL DEFAULT 'paper',
			client_tag TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
		`INSERT INTO control_state (id, bias) VALUES (1, 'Hybrid') ON CONFLICT (id) DO NOTHING;`,
		`INSERT INTO strategy_runtime (id) VALUES (1) ON CONFLICT (id) DO NOTHING;`,
	}

	for _, statement := range statements {
		if _, err := pool.Exec(ctx, statement); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	if _, err := pool.Exec(ctx, `
		INSERT INTO fills (symbol, side, price, size, realized_pnl, status)
		SELECT 'BTC', 'Buy', 102345.5, 0.01, 12.7, 'closed'
		WHERE NOT EXISTS (SELECT 1 FROM fills);
	`); err != nil {
		return err
	}

	if _, err := pool.Exec(ctx, `
		INSERT INTO strategy_derives (name, base_strategy, win_rate, pnl_ratio, condition, recommendation)
		SELECT 'Strategy_V1_HighVolatilityFilter', 'Strategy_V1', 0.58, 1.41, 'volatility > p75 AND open_interest falling', 'watch'
		WHERE NOT EXISTS (SELECT 1 FROM strategy_derives);
	`); err != nil {
		return err
	}

	return nil
}
