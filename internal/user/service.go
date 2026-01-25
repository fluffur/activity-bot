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

func (s *Service) GetUser(id int64) (model.User, error) {
	ctx := context.Background()
	return s.repo.Get(ctx, id)
}

func (s *Service) GetUserByUsername(username string) (model.User, error) {
	ctx := context.Background()
	return s.repo.GetByUsername(ctx, username)
}

func (s *Service) EnsureUserExists(id int64, username, firstName, lastName string) (model.User, error) {
	ctx := context.Background()
	return s.repo.Ensure(ctx, id, username, firstName, lastName)
}
