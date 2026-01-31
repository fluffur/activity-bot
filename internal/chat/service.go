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

func (s *Service) GetNorm(chatID int64) (int, error) {
	ctx := context.Background()
	return s.repo.GetNorm(ctx, chatID, s.defaultWeeklyNorm)
}

func (s *Service) SetNorm(chatID int64, norm int) error {
	ctx := context.Background()
	return s.repo.SetNorm(ctx, chatID, int32(norm))
}

func (s *Service) GetChat(chatID int64) (model.Chat, error) {
	ctx := context.Background()
	return s.repo.GetChat(ctx, chatID)
}

func (s *Service) SetChatPrompt(chatID int64, prompt string) error {
	ctx := context.Background()
	return s.repo.SetChatPrompt(ctx, chatID, prompt)
}

func (s *Service) SetNewbieThreshold(chatID int64, threshold int) error {
	ctx := context.Background()
	return s.repo.SetNewbieThreshold(ctx, chatID, int32(threshold))
}

func (s *Service) GetNewbieThreshold(chatID int64) (int, error) {
	ctx := context.Background()
	return s.repo.GetNewbieThreshold(ctx, chatID)
}
