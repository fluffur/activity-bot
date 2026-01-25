package message

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

func (s *Service) Save(chatID int64, userID int64) error {
	ctx := context.Background()
	return s.repo.Save(ctx, model.NewMessage(chatID, userID))
}
