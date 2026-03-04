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

func (s *Service) EnsureChatExists(ctx context.Context, chatID int64, title string) (model.Chat, error) {
	return s.repo.Ensure(ctx, model.Chat{
		ID:       chatID,
		Title:    title,
		NormWarn: s.defaultWeeklyNorm,
	})
}

func (s *Service) GetNorm(ctx context.Context, chatID int64) (int, error) {
	return s.repo.GetNorm(ctx, chatID, s.defaultWeeklyNorm)
}

func (s *Service) SetNorm(ctx context.Context, chatID int64, norm int) error {
	return s.repo.SetWarnNorm(ctx, chatID, int32(norm))
}

func (s *Service) SetBanNorm(ctx context.Context, chatID int64, norm int) error {
	return s.repo.SetBanNorm(ctx, chatID, int32(norm))
}

func (s *Service) GetChat(ctx context.Context, chatID int64) (model.Chat, error) {
	return s.repo.GetChat(ctx, chatID)
}

func (s *Service) SetTitle(ctx context.Context, chatID int64, title string) error {
	return s.repo.SetTitle(ctx, chatID, title)
}

func (s *Service) SetChatPrompt(ctx context.Context, chatID int64, prompt string) error {
	return s.repo.SetChatPrompt(ctx, chatID, prompt)
}

func (s *Service) SetNewbieThreshold(ctx context.Context, chatID int64, threshold int) error {
	return s.repo.SetNewbieThreshold(ctx, chatID, int32(threshold))
}

func (s *Service) GetNewbieThreshold(ctx context.Context, chatID int64) (int, error) {
	return s.repo.GetNewbieThreshold(ctx, chatID)
}

func (s *Service) SetMaxLadder(ctx context.Context, chatID int64, maxLadder int32) error {
	return s.repo.SetMaxLadder(ctx, chatID, maxLadder)
}

func (s *Service) GetMaxLadder(ctx context.Context, chatID int64) (int32, error) {
	c, err := s.repo.GetChat(ctx, chatID)
	if err != nil {
		return 0, err
	}
	return c.MaxLadder, nil
}

func (s *Service) SetWeekStartDay(ctx context.Context, chatID int64, day int) error {

	return s.repo.SetWeekStartDay(ctx, chatID, day)
}

func (s *Service) SetCommandPrefix(ctx context.Context, chatID int64, prefix string) error {
	return s.repo.SetCommandPrefix(ctx, chatID, prefix)
}

func (s *Service) SetAllowPrefixless(ctx context.Context, chatID int64, allow bool) error {
	return s.repo.SetAllowPrefixless(ctx, chatID, allow)
}

func (s *Service) ListChatsWithoutNorm(ctx context.Context, userID int64) ([]model.ChatWithoutNorm, error) {
	return s.repo.GetChatsWithoutNorm(ctx, userID)
}

func (s *Service) EnableTags(ctx context.Context, chatID int64, enabled bool) error {
	return s.repo.SetTagsEnabled(ctx, chatID, enabled)
}

func (s *Service) SetWeekStartTime(ctx context.Context, chatID int64, time string) error {
	return s.repo.SetWeekStartTime(ctx, chatID, time)
}
