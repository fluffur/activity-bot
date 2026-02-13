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

func (s *Service) Save(ctx context.Context, chatID, userID, msgID int64) error {
	return s.repo.Save(ctx, model.NewMessage(chatID, userID, msgID))
}
