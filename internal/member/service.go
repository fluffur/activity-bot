package member

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"context"
)

type ChatAdminsProvider interface {
	GetChatAdmins(chatID int64) ([]model.ChatMemberUpdate, error)
}

type Service struct {
	repo              Repository
	chatRepo          chat.Repository
	userRepo          user.Repository
	adminsProvider    ChatAdminsProvider
	defaultWeeklyNorm int32
}

func NewService(repo Repository, chatRepo chat.Repository, userRepo user.Repository, adminsProvider ChatAdminsProvider, defaultWeeklyNorm int32) *Service {
	return &Service{repo, chatRepo, userRepo, adminsProvider, defaultWeeklyNorm}
}

func (s *Service) SetMemberTitle(ctx context.Context, chatID int64, userID int64, title *string) error {
	return s.repo.UpdateCustomTitle(ctx, chatID, userID, title)
}

func (s *Service) SetMemberRole(ctx context.Context, chatID int64, userID int64, role string) error {
	return s.repo.UpdateStatus(ctx, chatID, userID, role)
}

func (s *Service) GetMembersWithTitle(ctx context.Context, chatID int64) ([]model.ChatMember, error) {
	return s.repo.GetWithCustomTitles(ctx, chatID)
}

func (s *Service) GetAnyMembersWithTitle(ctx context.Context, chatID int64) ([]model.ChatMember, error) {
	return s.repo.GetAnyWithCustomTitles(ctx, chatID)
}

func (s *Service) GetMemberTitle(ctx context.Context, chatID int64, userID int64) (string, error) {
	return s.repo.GetCustomTitle(ctx, chatID, userID)
}

func (s *Service) GetChatMembers(ctx context.Context, chatID int64) ([]model.ChatMember, error) {
	return s.repo.FindByChatID(ctx, chatID)
}

func (s *Service) UpdateChatMembers(ctx context.Context, chatID int64, members []model.ChatMemberUpdate) error {
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

func (s *Service) EnsureMemberExists(ctx context.Context, chatID int64, userID int64, username, firstName, lastName, role string) (model.ChatMember, error) {
	return s.repo.EnsureFull(ctx, chatID, userID, role, firstName, lastName, username, s.defaultWeeklyNorm)
}

func (s *Service) SyncChatMembers(ctx context.Context, chatID int64) (int, error) {
	members, err := s.adminsProvider.GetChatAdmins(chatID)
	if err != nil {
		return 0, err
	}

	if err := s.UpdateChatMembers(ctx, chatID, members); err != nil {
		return 0, err
	}

	return len(members), nil
}

func (s *Service) SetOnlyNewbies(ctx context.Context, chatID int64, users []*model.User) error {
	return s.repo.SetOnlyNewbies(ctx, chatID, users)
}

func (s *Service) SetNewbies(ctx context.Context, chatID int64, users []*model.User) error {
	return s.repo.SetNewbies(ctx, chatID, users)
}
