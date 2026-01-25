package member

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"context"
)

type Service struct {
	repo     Repository
	chatRepo chat.Repository
	userRepo user.Repository
}

func NewService(repo Repository, chatRepo chat.Repository, userRepo user.Repository) *Service {
	return &Service{repo, chatRepo, userRepo}
}

func (s *Service) SetMemberTitle(chatID int64, userID int64, title string) error {
	ctx := context.Background()
	return s.repo.UpdateCustomTitle(ctx, chatID, userID, title)
}

func (s *Service) GetMembersWithTitle(chatID int64) ([]model.ChatMember, error) {
	ctx := context.Background()
	return s.repo.GetWithCustomTitles(ctx, chatID)
}

func (s *Service) GetMemberTitle(chatID int64, userID int64) (string, error) {
	ctx := context.Background()
	return s.repo.GetCustomTitle(ctx, chatID, userID)
}

func (s *Service) UpdateChatMembers(chatID int64, members []model.ChatMemberUpdate) error {
	ctx := context.Background()

	users := make([]model.User, len(members))
	for i, m := range members {
		users[i] = m.User
	}

	if _, err := s.chatRepo.Ensure(ctx, model.Chat{chatID, 100}); err != nil {
		return err
	}

	if err := s.userRepo.UpsertUsers(ctx, users); err != nil {
		return err
	}

	return s.repo.UpsertChatMembers(ctx, chatID, members)
}

func (s *Service) ProcessLeftMember(chatID int64, userID int64) (string, error) {
	ctx := context.Background()
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
