package admin

import (
	"activity-bot/internal/model"
	"context"
	"errors"
	"log/slog"
)

type ChatMemberStatusProvider interface {
	GetChatMemberStatus(chatID, userID int64) (string, error)
}

var (
	ErrUserIsNotAdmin     = errors.New("user is not admin")
	ErrUserIsAlreadyAdmin = errors.New("user is already admin")
	ErrUserIsCreator      = errors.New("user is creator")
)

type Service struct {
	repo         Repository
	memberStatus ChatMemberStatusProvider
	ownerID      int64
}

func NewService(repo Repository, statusProvider ChatMemberStatusProvider, ownerID int64) *Service {
	return &Service{repo, statusProvider, ownerID}
}

func (s *Service) AddAdmin(ctx context.Context, chatID int64, userID int64) error {
	isCreator, err := s.IsCreator(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if isCreator {
		return ErrUserIsAlreadyAdmin
	}

	isAdmin, err := s.IsAdmin(ctx, chatID, userID)
	if err != nil {
		return err
	}

	if isAdmin {
		return ErrUserIsAlreadyAdmin
	}

	return s.repo.Add(ctx, chatID, userID)
}

func (s *Service) RemoveAdmin(ctx context.Context, chatID int64, userID int64) error {
	isCreator, err := s.IsCreator(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if isCreator {
		return ErrUserIsCreator
	}

	isAdmin, err := s.IsAdmin(ctx, chatID, userID)
	if err != nil {
		return err
	}

	if !isAdmin {
		return ErrUserIsNotAdmin
	}

	return s.repo.Remove(ctx, chatID, userID)
}

func (s *Service) GetAdminsEnsured(
	ctx context.Context,
	chatID int64,
	sync func(ctx context.Context, chatID int64) (int, error),
) ([]model.User, error) {
	admins, err := s.repo.GetFromChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	if len(admins) > 0 {
		return admins, nil
	}

	if _, err := sync(ctx, chatID); err != nil {
		return nil, err
	}

	return s.repo.GetFromChat(ctx, chatID)
}

func (s *Service) IsCreator(ctx context.Context, chatID int64, userID int64) (bool, error) {
	if userID == s.ownerID {
		return true, nil
	}

	isCreator, err := s.repo.IsCreator(ctx, chatID, userID)
	if err == nil {
		return isCreator, nil
	}

	status, err := s.memberStatus.GetChatMemberStatus(chatID, userID)
	if err != nil {
		return false, err
	}

	return status == "creator", nil
}

func (s *Service) GetRole(ctx context.Context, chatID int64, userID int64) (string, error) {
	return s.repo.GetRole(ctx, chatID, userID)
}

func (s *Service) CheckIsAdmin(ctx context.Context, chatID, userID int64) bool {
	isAdmin, err := s.IsAdmin(ctx, chatID, userID)
	if err != nil {
		slog.Error("failed to check admin", "chat_id", chatID, "user_id", userID, "error", err)
		return false
	}
	return isAdmin
}

func (s *Service) CheckIsCreator(ctx context.Context, chatID, userID int64) bool {
	isCreator, err := s.IsCreator(ctx, chatID, userID)
	if err != nil {
		slog.Error("failed to check creator", "chat_id", chatID, "user_id", userID, "error", err)
		return false
	}
	return isCreator
}

func (s *Service) IsAdmin(ctx context.Context, chatID, userID int64) (bool, error) {
	if userID == s.ownerID {
		return true, nil
	}

	isAdmin, err := s.repo.IsAdmin(ctx, chatID, userID)
	if err != nil {
		return false, err
	}
	if isAdmin {
		return true, nil
	}

	status, err := s.memberStatus.GetChatMemberStatus(chatID, userID)
	if err != nil {
		return false, err
	}

	return status == "creator", nil
}
