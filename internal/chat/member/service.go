package member

import (
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"context"
)

type Service struct {
	repo     Repository
	userRepo user.Repository
}

func NewService(repo Repository, userRepo user.Repository) *Service {
	return &Service{repo, userRepo}
}

func (s *Service) SetMemberTitle(ctx context.Context, chatID int64, userID int64, title string) error {
	return s.repo.UpdateCustomTitle(ctx, chatID, userID, title)
}

func (s *Service) GetMembersWithTitle(ctx context.Context, chatID int64) ([]model.ChatMember, error) {
	return s.repo.GetWithCustomTitles(ctx, chatID)
}

func (s *Service) GetMemberTitle(ctx context.Context, chatID int64, userID int64) (string, error) {
	return s.repo.GetCustomTitle(ctx, chatID, userID)
}

func (s *Service) UpdateChatMembers(ctx context.Context, chatID int64, members []model.ChatMemberUpdate) error {

	users := make([]model.User, len(members))
	for i, m := range members {
		users[i] = m.User
	}

	if err := s.userRepo.UpsertUsers(ctx, users); err != nil {
		return err
	}

	return s.repo.UpsertChatMembers(ctx, chatID, members)
}

func (s *Service) ProcessLeftMember(ctx context.Context, chatID int64, userID int64) (string, error) {
	member, err := s.repo.Get(ctx, chatID, userID)
	if err != nil {
		return "", err
	}

	if err := s.repo.Remove(ctx, chatID, userID); err != nil {
		return "", err
	}

	return member.CustomTitle, nil
}

func (s *Service) EnsureMemberExists(ctx context.Context, chatID int64, userID int64) (model.ChatMember, error) {
	return s.repo.EnsureExists(ctx, chatID, userID)
}
