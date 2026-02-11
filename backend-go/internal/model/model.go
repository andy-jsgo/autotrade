package model

import "time"

type AccountState struct {
	Equity      float64   `json:"equity"`
	Leverage    float64   `json:"leverage"`
	OpenPnL     float64   `json:"openPnl"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Fill struct {
	ID          int64     `json:"id"`
	Symbol      string    `json:"symbol"`
	Side        string    `json:"side"`
	Price       float64   `json:"price"`
	Size        float64   `json:"size"`
	RealizedPnL float64   `json:"realizedPnl"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ReviewInput struct {
	FillID  int64    `json:"fillId"`
	Verdict string   `json:"verdict"`
	Tags    []string `json:"tags"`
	Notes   string   `json:"notes"`
}

type StrategyDerive struct {
	Name           string  `json:"name"`
	BaseStrategy   string  `json:"baseStrategy"`
	WinRate        float64 `json:"winRate"`
	PnLRatio       float64 `json:"pnlRatio"`
	Condition      string  `json:"condition"`
	Recommendation string  `json:"recommendation"`
}
