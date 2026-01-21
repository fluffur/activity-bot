package message

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"context"

	"github.com/go-telegram/bot/models"
)

type Service struct {
	repo              Repository
	userRepo          user.Repository
	chatRepo          chat.Repository
	defaultWeeklyNorm int32
}

func NewService(repo Repository, userRepo user.Repository, chatRepo chat.Repository, defaultWeeklyNorm int32) *Service {
	return &Service{repo, userRepo, chatRepo, defaultWeeklyNorm}
}

func (s *Service) Save(ctx context.Context, chatID int64, user *models.User) error {
	return s.repo.Save(ctx, model.NewMessage(chatID, user.ID))
}
