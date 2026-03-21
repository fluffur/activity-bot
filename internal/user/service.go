package user

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

func (s *Service) GetUser(ctx context.Context, id int64) (model.User, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) EnsureUserExists(ctx context.Context, id int64, username, firstName, lastName string) (model.User, error) {
	return s.repo.Ensure(ctx, id, username, firstName, lastName)
}

func (s *Service) GetByTag(ctx context.Context, chatID int64, tag string) ([]model.ChatMember, error) {
	return s.repo.GetByTag(ctx, chatID, tag)
}

func (s *Service) SetGender(ctx context.Context, userID int64, gender string) error {
	return s.repo.SetGender(ctx, userID, gender)
}

func (s *Service) SetEmoji(ctx context.Context, userID int64, emoji string) error {
	return s.repo.SetEmoji(ctx, userID, emoji)
}

func (s *Service) SetCustomEmojiID(ctx context.Context, userID int64, emojiID string) error {
	return s.repo.SetCustomEmojiID(ctx, userID, emojiID)
}
