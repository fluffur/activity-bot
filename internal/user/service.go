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

func (s *Service) GetUserByUsername(ctx context.Context, username string) (model.User, error) {
	return s.repo.GetByUsername(ctx, username)
}

func (s *Service) EnsureUserExists(ctx context.Context, id int64, username, firstName, lastName string) (model.User, error) {
	return s.repo.Ensure(ctx, id, username, firstName, lastName)
}

func (s *Service) GetByCustomTitle(ctx context.Context, chatID int64, title string) ([]model.ChatMember, error) {
	return s.repo.GetByCustomTitle(ctx, chatID, title)
}

func (s *Service) SetGender(ctx context.Context, userID int64, gender string) error {
	return s.repo.SetGender(ctx, userID, gender)
}
