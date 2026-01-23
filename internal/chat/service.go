package chat

import (
	"activity-bot/internal/model"
	"context"
)

type Service struct {
	repo              Repository
	defaultWeeklyNorm int32
}

func NewService(repo Repository, defaultWeeklyNorm int32) *Service {
	return &Service{repo, defaultWeeklyNorm}
}

func (s *Service) EnsureChatExists(ctx context.Context, chatID int64) (model.Chat, error) {
	return s.repo.Ensure(ctx, model.Chat{
		ID:         chatID,
		WeeklyNorm: s.defaultWeeklyNorm,
	})
}

func (s *Service) GetNorm(ctx context.Context, chatID int64) (int, error) {
	return s.repo.GetNorm(ctx, chatID, s.defaultWeeklyNorm)
}

func (s *Service) SetNorm(ctx context.Context, chatID int64, norm int) error {
	return s.repo.SetNorm(ctx, chatID, int32(norm))
}
