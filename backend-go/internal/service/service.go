package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"autotrade/backend-go/internal/model"
	"autotrade/backend-go/internal/repo"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	repo *repo.Repo
}

func New(r *repo.Repo) *Service {
	return &Service{repo: r}
}

func (s *Service) State(ctx context.Context) (map[string]any, error) {
	st, err := s.repo.GetState(ctx)
	if err != nil {
		return nil, err
	}
	bias, err := s.repo.GetBias(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{"state": st, "bias": bias}, nil
}

func (s *Service) Fills(ctx context.Context, limit int) ([]model.Fill, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.GetFills(ctx, limit)
}

func (s *Service) Review(ctx context.Context, in model.ReviewInput) error {
	if in.FillID <= 0 {
		return ErrBadRequest("fillId is required")
	}
	in.Verdict = strings.ToLower(strings.TrimSpace(in.Verdict))
	if in.Verdict != "good" && in.Verdict != "bad" {
		return ErrBadRequest("verdict must be good or bad")
	}
	return s.repo.SaveReview(ctx, in)
}

func (s *Service) Derives(ctx context.Context) ([]model.StrategyDerive, error) {
	return s.repo.GetDerives(ctx)
}

func (s *Service) SetBias(ctx context.Context, bias string) error {
	normalized := strings.ToLower(strings.TrimSpace(bias))
	var out string
	switch normalized {
	case "long":
		out = "Long"
	case "short":
		out = "Short"
	case "hybrid":
		out = "Hybrid"
	default:
		return ErrBadRequest("bias must be Long, Short, or Hybrid")
	}
	return s.repo.SetBias(ctx, out)
}

func (s *Service) WalletSession(ctx context.Context) (model.WalletSession, error) {
	out, err := s.repo.GetLatestWalletSession(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.WalletSession{}, nil
		}
		return model.WalletSession{}, err
	}
	return out, nil
}

func (s *Service) ConnectWallet(ctx context.Context, address, signature, message string) error {
	if !strings.HasPrefix(strings.ToLower(address), "0x") || len(address) < 10 {
		return ErrBadRequest("invalid wallet address")
	}
	if strings.TrimSpace(signature) == "" {
		return ErrBadRequest("signature is required")
	}
	return s.repo.SaveWalletSession(ctx, address, signature, message)
}

func (s *Service) ApproveAgent(ctx context.Context) (string, error) {
	ws, err := s.repo.GetLatestWalletSession(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrBadRequest("connect wallet first")
		}
		return "", err
	}
	agent := fmt.Sprintf("agent_%s", strings.ToLower(ws.Address[len(ws.Address)-8:]))
	if err := s.repo.ApproveAgent(ctx, agent); err != nil {
		return "", err
	}
	return agent, nil
}

func (s *Service) StrategyStatus(ctx context.Context) (model.StrategyStatus, error) {
	return s.repo.StrategyStatus(ctx)
}

func (s *Service) SetAutoTrading(ctx context.Context, enabled bool) error {
	return s.repo.SetAutoTrading(ctx, enabled)
}

func (s *Service) CreateOrder(ctx context.Context, in model.OrderInput) (int64, error) {
	if strings.TrimSpace(in.Symbol) == "" || strings.TrimSpace(in.Side) == "" {
		return 0, ErrBadRequest("symbol and side are required")
	}
	if in.Size <= 0 || in.EntryPrice <= 0 {
		return 0, ErrBadRequest("size and entryPrice must be positive")
	}
	if in.StopLoss <= 0 || in.TakeProfit <= 0 {
		return 0, ErrBadRequest("stopLoss and takeProfit are required")
	}
	if strings.TrimSpace(in.Execution) == "" {
		in.Execution = "paper"
	}
	approved, err := s.repo.IsAgentApproved(ctx)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}
	if !approved {
		return 0, ErrBadRequest("agent is not approved")
	}
	if strings.TrimSpace(in.OrderType) == "" {
		in.OrderType = "market"
	}
	id, err := s.repo.CreateOrder(ctx, in)
	if err != nil {
		return 0, err
	}
	_ = s.repo.CreateFillFromOrder(ctx, in)
	prettySide := strings.ToLower(in.Side)
	if len(prettySide) > 0 {
		prettySide = strings.ToUpper(prettySide[:1]) + prettySide[1:]
	}
	s.repo.TouchSignal(ctx, fmt.Sprintf("%s %s %.4f", strings.ToUpper(in.Symbol), prettySide, in.Size))
	return id, nil
}

func (s *Service) Orders(ctx context.Context, limit int) ([]model.Order, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.GetOrders(ctx, limit)
}

type badRequestErr struct{ msg string }

func (e badRequestErr) Error() string { return e.msg }

func ErrBadRequest(msg string) error { return badRequestErr{msg: msg} }

func IsBadRequest(err error) bool {
	_, ok := err.(badRequestErr)
	return ok
}
