package member

import (
	"activity-bot/internal/adapter"
	"activity-bot/internal/chat"
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"context"
	"fmt"
	"time"
)

type ChatMembersProvider interface {
	GetChatMembers(ctx context.Context, chatID int64) ([]model.ChatMemberUpdate, error)
}

type Service struct {
	repo             Repository
	chatRepo         chat.Repository
	userRepo         user.Repository
	adminsProvider   ChatMembersProvider
	memberTagAdapter *adapter.MemberTagAdapter
}

func NewService(repo Repository, chatRepo chat.Repository, userRepo user.Repository, adminsProvider ChatMembersProvider, memberTagAdapter *adapter.MemberTagAdapter) *Service {
	return &Service{repo, chatRepo, userRepo, adminsProvider, memberTagAdapter}
}

func (s *Service) SetMemberTitle(ctx context.Context, chatID int64, userID int64, title string) error {
	return s.repo.UpdateCustomTitle(ctx, chatID, userID, title)
}
func (s *Service) GetMembersWithTitle(ctx context.Context, chatID int64) ([]model.ChatMember, error) {
	return s.repo.GetWithCustomTitles(ctx, chatID)
}

func (s *Service) GetAnyMembersWithTitle(ctx context.Context, chatID int64) ([]model.ChatMember, error) {
	return s.repo.GetAnyWithCustomTitles(ctx, chatID)
}

func (s *Service) GetMemberTitle(ctx context.Context, chatID int64, userID int64) (string, error) {
	return s.repo.GetTag(ctx, chatID, userID)
}

func (s *Service) GetChatMember(ctx context.Context, chatID int64, userID int64) (model.ChatMember, error) {
	return s.repo.GetChatMember(ctx, chatID, userID)
}

func (s *Service) GetChatMembers(ctx context.Context, chatID int64) ([]model.ChatMember, error) {
	return s.repo.FindByChatID(ctx, chatID)
}

func (s *Service) UpdateChatMembers(ctx context.Context, chatID int64, members []model.ChatMemberUpdate) error {
	if _, err := s.chatRepo.Ensure(ctx, model.Chat{ID: chatID}); err != nil {
		return fmt.Errorf("failed to ensure chat: %w", err)
	}

	users := make([]model.User, len(members))
	for i, m := range members {
		users[i] = m.User
	}

	if err := s.userRepo.UpsertUsers(ctx, users); err != nil {
		return err
	}

	if err := s.repo.ResetCreators(ctx, chatID); err != nil {
		return fmt.Errorf("failed to reset creators: %w", err)
	}
	if err := s.repo.UpsertChatMembers(ctx, chatID, members); err != nil {
		return fmt.Errorf("failed to upsert members: %w", err)
	}

	return nil
}

func (s *Service) ProcessLeftMember(ctx context.Context, chatID int64, userID int64) (model.ChatMember, error) {
	member, err := s.repo.GetChatMember(ctx, chatID, userID)
	if err != nil {
		return model.ChatMember{}, err
	}

	if err := s.repo.Remove(ctx, chatID, userID); err != nil {
		return model.ChatMember{}, err
	}

	return member, nil
}

func (s *Service) EnsureMemberExists(ctx context.Context, chatID int64, userID int64, username, firstName, lastName, role string) (model.ChatMember, error) {
	return s.repo.EnsureFull(ctx, chatID, userID, role, firstName, lastName, username)
}

func (s *Service) SyncChatMembers(ctx context.Context, chatID int64) (int, error) {
	members, err := s.adminsProvider.GetChatMembers(ctx, chatID)
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

func (s *Service) GetNoNormMembers(ctx context.Context, chatID int64, from, to *time.Time) ([]model.ChatMember, error) {
	return s.repo.GetNoNormMembers(ctx, chatID, from, to)
}

func (s *Service) GetNoNormWarnMembers(ctx context.Context, chatID int64, from *time.Time, to *time.Time) ([]model.ChatMember, error) {
	return s.repo.GetNoNormWarnMembers(ctx, chatID, from, to)
}

func (s *Service) GetNoNormBanMembers(ctx context.Context, chatID int64, from *time.Time, to *time.Time) ([]model.ChatMember, error) {
	return s.repo.GetNoNormBanMembers(ctx, chatID, from, to)
}

func (s *Service) FindChatMembersByTag(ctx context.Context, chatID int64, customTitle string) ([]model.ChatMember, error) {
	return s.repo.FindByTag(ctx, chatID, customTitle)
}

func (s *Service) GetChatMemberByUsername(ctx context.Context, chatID int64, username string) (model.ChatMember, error) {
	return s.repo.GetByUsername(ctx, chatID, username)
}

func (s *Service) SetChatMemberEmoji(ctx context.Context, chatID int64, userID int64, emojis model.Emojis) error {
	return s.repo.SetEmoji(ctx, chatID, userID, emojis)
}
