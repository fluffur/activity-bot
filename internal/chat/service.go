package chat

import (
	"activity-bot/internal/model"
	"context"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo}
}

func (s *Service) EnsureChatExists(ctx context.Context, chatID int64, title string) (model.Chat, error) {
	return s.repo.Ensure(ctx, model.Chat{
		ID:    chatID,
		Title: title,
	})
}

func (s *Service) SetWarnNorm(ctx context.Context, chatID int64, norm int) error {
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

func (s *Service) SetNewbieThreshold(ctx context.Context, chatID int64, threshold int32) error {
	return s.repo.SetNewbieThreshold(ctx, chatID, threshold)
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

func (s *Service) GetChatsWithoutTitle(ctx context.Context) ([]model.Chat, error) {
	return s.repo.GetChatsWithoutTitle(ctx)
}

func (s *Service) GetChatsWithEnabledBroadcast(ctx context.Context) ([]model.Chat, error) {
	return s.repo.GetChatsWithEnabledBroadcast(ctx)
}

func (s *Service) GetUserManagedChats(ctx context.Context, userID int64, ownerID int64, text string) ([]model.Chat, error) {
	if userID == ownerID {
		return s.repo.GetAllChats(ctx, text)
	}
	return s.repo.GetUserManagedChats(ctx, userID, text)
}

func (s *Service) DisableBroadcast(ctx context.Context, chatID int64) error {
	return s.repo.SetChatBroadcast(ctx, chatID, false)
}

func (s *Service) EnableBroadcast(ctx context.Context, chatID int64) error {
	return s.repo.SetChatBroadcast(ctx, chatID, true)
}

func (s *Service) RemoveWarnNorm(ctx context.Context, chatID int64) error {
	return s.repo.SetWarnNorm(ctx, chatID, 0)
}

func (s *Service) RemoveBanNorm(ctx context.Context, chatID int64) error {
	return s.repo.SetBanNorm(ctx, chatID, 0)
}

func (s *Service) GetCommandPermissions(ctx context.Context, chatID int64) (map[string]model.Status, error) {
	return s.repo.GetCommandPermissions(ctx, chatID)
}

func (s *Service) GetCommandPermission(ctx context.Context, chatID int64, key string) (model.Status, error) {
	return s.repo.GetCommandPermission(ctx, chatID, key)
}

func (s *Service) SetCommandPermission(ctx context.Context, chatID int64, key string, status model.Status) error {
	return s.repo.SetCommandPermission(ctx, chatID, key, status)
}

func (s *Service) RemoveChat(ctx context.Context, chatID int64) error {
	return s.repo.Remove(ctx, chatID)
}
