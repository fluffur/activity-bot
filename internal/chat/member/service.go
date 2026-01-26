package member

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"context"
)

type Service struct {
	repo              Repository
	chatRepo          chat.Repository
	userRepo          user.Repository
	defaultWeeklyNorm int32
}

func NewService(repo Repository, chatRepo chat.Repository, userRepo user.Repository, defaultWeeklyNorm int32) *Service {
	return &Service{repo, chatRepo, userRepo, defaultWeeklyNorm}
}

func (s *Service) SetMemberTitle(chatID int64, userID int64, title string) error {
	ctx := context.Background()
	return s.repo.UpdateCustomTitle(ctx, chatID, userID, title)
}

func (s *Service) SetMemberRole(chatID int64, userID int64, role string) error {
	ctx := context.Background()
	return s.repo.UpdateRole(ctx, chatID, userID, role)
}

func (s *Service) GetMembersWithTitle(chatID int64) ([]model.ChatMember, error) {
	ctx := context.Background()
	return s.repo.GetWithCustomTitles(ctx, chatID)
}

func (s *Service) GetMemberTitle(chatID int64, userID int64) (string, error) {
	ctx := context.Background()
	return s.repo.GetCustomTitle(ctx, chatID, userID)
}

func (s *Service) GetChatMembers(chatID int64) ([]model.ChatMember, error) {
	ctx := context.Background()

	return s.repo.FindByChatID(ctx, chatID)
}

func (s *Service) UpdateChatMembers(chatID int64, members []model.ChatMemberUpdate) error {
	ctx := context.Background()

	if _, err := s.chatRepo.Ensure(ctx, model.Chat{ID: chatID, WeeklyNorm: s.defaultWeeklyNorm}); err != nil {
		return err
	}

	users := make([]model.User, len(members))
	for i, m := range members {
		users[i] = m.User
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

func (s *Service) EnsureMemberExists(chatID int64, userID int64, username, firstName, lastName, role string) (model.ChatMember, error) {
	ctx := context.Background()

	return s.repo.EnsureFull(ctx, chatID, userID, role, firstName, lastName, username, s.defaultWeeklyNorm)
}
