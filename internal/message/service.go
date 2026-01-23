package message

import (
	"activity-bot/internal/model"
	"context"

	"github.com/go-telegram/bot/models"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo}
}

func (s *Service) Save(ctx context.Context, chatID int64, user *models.User) error {
	return s.repo.Save(ctx, model.NewMessage(chatID, user.ID))
}
