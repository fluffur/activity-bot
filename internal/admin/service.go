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

func (s *Service) GetAdminsEnsured(
	chatID int64,
	sync func(chatID int64) (int, error),
) ([]model.User, error) {

	admins, err := s.GetAdmins(chatID)
	if err != nil {
		return nil, err
	}

	if len(admins) > 0 {
		return admins, nil
	}

	if _, err := sync(chatID); err != nil {
		return nil, err
	}

	return s.GetAdmins(chatID)
}

func (s *Service) IsAdmin(chatID int64, userID int64) (bool, error) {
	ctx := context.Background()
	return s.repo.IsAdmin(ctx, chatID, userID)
}

func (s *Service) IsCreator(chatID int64, userID int64) (bool, error) {
	ctx := context.Background()
	return s.repo.IsCreator(ctx, chatID, userID)
}

func (s *Service) GetRole(chatID int64, userID int64) (string, error) {
	ctx := context.Background()
	return s.repo.GetRole(ctx, chatID, userID)
}
