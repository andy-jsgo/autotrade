package service

import (
	"context"
	"strings"

	"autotrade/backend-go/internal/model"
	"autotrade/backend-go/internal/repo"
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

type badRequestErr struct{ msg string }

func (e badRequestErr) Error() string { return e.msg }

func ErrBadRequest(msg string) error { return badRequestErr{msg: msg} }

func IsBadRequest(err error) bool {
	_, ok := err.(badRequestErr)
	return ok
}
