package member

import (
	"activity-bot/internal/adapter"
	"activity-bot/internal/chat"
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"context"
)

type ChatAdminsProvider interface {
	GetChatAdmins(chatID int64) ([]model.ChatMemberUpdate, error)
}

type Service struct {
	repo             Repository
	chatRepo         chat.Repository
	userRepo         user.Repository
	adminsProvider   ChatAdminsProvider
	defaultNormWarn  int32
	memberTagAdapter *adapter.MemberTagAdapter
}

func NewService(repo Repository, chatRepo chat.Repository, userRepo user.Repository, adminsProvider ChatAdminsProvider, defaultWeeklyNorm int32, memberTagAdapter *adapter.MemberTagAdapter) *Service {
	return &Service{repo, chatRepo, userRepo, adminsProvider, defaultWeeklyNorm, memberTagAdapter}
}

func (s *Service) SetMemberTitle(ctx context.Context, chatID int64, userID int64, title string) error {
	if err := s.memberTagAdapter.SetMemberTag(ctx, chatID, userID, title); err != nil {
		return err
	}
	return s.repo.UpdateCustomTitle(ctx, chatID, userID, &title)
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

func (s *Service) GetChatMember(ctx context.Context, chatID int64, userID int64) (model.ChatMember, error) {
	return s.repo.Get(ctx, chatID, userID)
}

func (s *Service) GetChatMembers(ctx context.Context, chatID int64) ([]model.ChatMember, error) {
	return s.repo.FindByChatID(ctx, chatID)
}

func (s *Service) UpdateChatMembers(ctx context.Context, chatID int64, members []model.ChatMemberUpdate) error {
	if _, err := s.chatRepo.Ensure(ctx, model.Chat{ID: chatID, NormWarn: s.defaultNormWarn}); err != nil {
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

func (s *Service) ProcessLeftMember(ctx context.Context, chatID int64, userID int64) (model.ChatMember, error) {
	member, err := s.repo.Get(ctx, chatID, userID)
	if err != nil {
		return model.ChatMember{}, err
	}

	if err := s.repo.Remove(ctx, chatID, userID); err != nil {
		return model.ChatMember{}, err
	}

	return member, nil
}

func (s *Service) EnsureMemberExists(ctx context.Context, chatID int64, userID int64, username, firstName, lastName, role string) (model.ChatMember, error) {
	return s.repo.EnsureFull(ctx, chatID, userID, role, firstName, lastName, username, s.defaultNormWarn)
}

func (s *Service) SyncChatMembers(ctx context.Context, chatID int64) (int, error) {
	members, err := s.adminsProvider.GetChatAdmins(chatID)
	if err != nil {
		return 0, err
	}

	if err := s.UpdateChatMembers(ctx, chatID, members); err != nil {
		return 0, err
	}

	userIDs := make([]int64, len(members))
	for i, m := range members {
		userIDs[i] = m.User.ID
	}

	return len(members), nil
}

func (s *Service) SetOnlyNewbies(ctx context.Context, chatID int64, users []*model.User) error {
	return s.repo.SetOnlyNewbies(ctx, chatID, users)
}

func (s *Service) SetNewbies(ctx context.Context, chatID int64, users []*model.User) error {
	return s.repo.SetNewbies(ctx, chatID, users)
}

func (s *Service) GetCommandLevels(ctx context.Context, chatID int64) (map[string]int16, error) {
	return s.repo.GetCommandLevels(ctx, chatID)
}

func (s *Service) SetCommandLevel(ctx context.Context, chatID int64, commandID string, level int16) error {
	return s.repo.SetCommandLevel(ctx, chatID, commandID, level)
}
