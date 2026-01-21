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
