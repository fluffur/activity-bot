package admin

import (
	"activity-bot/internal/model"
	"context"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) AddAdmin(chatID int64, userID int64) error {
	ctx := context.Background()
	return s.repo.Add(ctx, chatID, userID)
}

func (s *Service) RemoveAdmin(chatID int64, userID int64) error {
	ctx := context.Background()
	return s.repo.Remove(ctx, chatID, userID)
}

func (s *Service) GetAdmins(chatID int64) ([]model.User, error) {
	ctx := context.Background()
	return s.repo.GetFromChat(ctx, chatID)
}

func (s *Service) IsAdmin(chatID int64, userID int64) (bool, error) {
	ctx := context.Background()
	return s.repo.IsAdmin(ctx, chatID, userID)
}
