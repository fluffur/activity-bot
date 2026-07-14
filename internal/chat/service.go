package chat

import (
	"context"
	"time"
)

type Service struct {
	repo    Repository
	ownerID int64
}

func NewService(r Repository, ownerID int64) *Service {
	return &Service{
		repo:    r,
		ownerID: ownerID,
	}
}

func (s *Service) Create(ctx context.Context, chat Chat) error {
	return s.repo.Create(ctx, chat)
}

func (s *Service) Get(ctx context.Context, chatID int64) (Chat, error) {
	return s.repo.Get(ctx, chatID)
}

func (s *Service) SetMentionTypes(ctx context.Context, chatID int64, mentionTypes MentionTypes) error {
	return s.repo.SetMentionTypes(ctx, chatID, mentionTypes)
}

func (s *Service) SetSkipSummonConfirmation(ctx context.Context, chatID int64, skipSummonConfirmation bool) error {
	return s.repo.SetSkipSummonConfirmation(ctx, chatID, skipSummonConfirmation)
}

func (s *Service) GetUserManagedChats(ctx context.Context, userID int64, search string) ([]Chat, error) {
	if userID == s.ownerID {
		return s.repo.GetAllChats(ctx, search)
	}
	return s.repo.GetUserManagedChats(ctx, userID, search)
}

func (s *Service) SetNewbieThreshold(ctx context.Context, chatID int64, threshold int32) error {
	return s.repo.SetNewbieThreshold(ctx, chatID, threshold)
}

func (s *Service) SetTitle(ctx context.Context, chatID int64, title string) error {
	return s.repo.SetTitle(ctx, chatID, title)
}

func (s *Service) SetChatPrompt(ctx context.Context, chatID int64, prompt string) error {
	return s.repo.SetChatPrompt(ctx, chatID, prompt)
}

func (s *Service) SetWeekStartDay(ctx context.Context, chatID int64, day int) error {
	return s.repo.SetWeekStartDay(ctx, chatID, day)
}

func timeToMicroseconds(s string) (int64, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return 0, err
	}

	return int64(t.Hour())*int64(time.Hour/time.Microsecond) +
		int64(t.Minute())*int64(time.Minute/time.Microsecond) +
		int64(t.Second())*int64(time.Second/time.Microsecond) +
		int64(t.Nanosecond()/1000), nil
}

func (s *Service) SetWeekStartTime(ctx context.Context, chatID int64, time string) error {
	t, err := timeToMicroseconds(time)
	if err != nil {
		return err
	}
	return s.repo.SetWeekStartTime(ctx, chatID, t)
}

func (s *Service) SetCommandPrefix(ctx context.Context, chatID int64, prefix string) error {
	return s.repo.SetCommandPrefix(ctx, chatID, prefix)
}

func (s *Service) SetAllowPrefixless(ctx context.Context, chatID int64, allow bool) error {
	return s.repo.SetAllowPrefixless(ctx, chatID, allow)
}

func (s *Service) SetEmojisEnabled(ctx context.Context, chatID int64, enabled bool) error {
	return s.repo.SetEmojisEnabled(ctx, chatID, enabled)
}

func (s *Service) GetChatsWithoutTitle(ctx context.Context) ([]Chat, error) {
	return s.repo.ListWithoutTitle(ctx)
}

func (s *Service) RemoveChat(ctx context.Context, chatID int64) error {
	return s.repo.Remove(ctx, chatID)
}
