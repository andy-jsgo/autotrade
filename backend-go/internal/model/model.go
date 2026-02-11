package model

import "time"

type AccountState struct {
	Equity    float64   `json:"equity"`
	Leverage  float64   `json:"leverage"`
	OpenPnL   float64   `json:"openPnl"`
	UpdatedAt time.Time `json:"updatedAt"`
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

type WalletSession struct {
	Address       string    `json:"address"`
	Connected     bool      `json:"connected"`
	AgentApproved bool      `json:"agentApproved"`
	AgentPubKey   string    `json:"agentPubKey"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type StrategyStatus struct {
	Bias          string    `json:"bias"`
	AutoTrading   bool      `json:"autoTrading"`
	RuntimeStatus string    `json:"runtimeStatus"`
	LastSignal    string    `json:"lastSignal"`
	LastError     string    `json:"lastError"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type OrderInput struct {
	Symbol     string  `json:"symbol"`
	Side       string  `json:"side"`
	OrderType  string  `json:"orderType"`
	Size       float64 `json:"size"`
	EntryPrice float64 `json:"entryPrice"`
	StopLoss   float64 `json:"stopLoss"`
	TakeProfit float64 `json:"takeProfit"`
	ClientTag  string  `json:"clientTag"`
	Execution  string  `json:"execution"`
}

type Order struct {
	ID         int64     `json:"id"`
	Symbol     string    `json:"symbol"`
	Side       string    `json:"side"`
	OrderType  string    `json:"orderType"`
	Size       float64   `json:"size"`
	EntryPrice float64   `json:"entryPrice"`
	StopLoss   float64   `json:"stopLoss"`
	TakeProfit float64   `json:"takeProfit"`
	Status     string    `json:"status"`
	Execution  string    `json:"execution"`
	CreatedAt  time.Time `json:"createdAt"`
}
