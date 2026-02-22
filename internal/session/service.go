package session

import (
	"context"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo}
}

func (s *Service) SetActiveChat(ctx context.Context, userID int64, chatID int64) error {
	return s.repo.SetSession(ctx, userID, chatID)
}

func (s *Service) GetActiveChat(ctx context.Context, userID int64) (int64, error) {
	return s.repo.GetSession(ctx, userID)
}

func (s *Service) ClearSession(ctx context.Context, userID int64) error {
	return s.repo.DeleteSession(ctx, userID)
}
