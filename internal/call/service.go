package call

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"activity-bot/internal/stats"
	"context"
)

type Service struct {
	repo          chat.Repository
	memberService *member.Service
	statsService  *stats.Service
}

func NewService(repo chat.Repository, memberService *member.Service, statsService *stats.Service) *Service {
	return &Service{repo, memberService, statsService}
}

func (s *Service) GetAllMembers(ctx context.Context, chatID int64) ([]model.ChatMember, error) {
	return s.memberService.GetChatMembers(ctx, chatID)
}

func (s *Service) GetChatSettings(ctx context.Context, chatID int64) (model.Chat, error) {
	return s.repo.GetChat(ctx, chatID)
}

func (s *Service) SetWelcomeCallMessage(ctx context.Context, chatID int64, message string) error {
	return s.repo.SetWelcomeCallMessage(ctx, chatID, message)
}

func (s *Service) EnableCallOnJoin(ctx context.Context, chatID int64) error {
	return s.repo.UpdateCallOnJoin(ctx, chatID, true)
}

func (s *Service) DisableCallOnJoin(ctx context.Context, chatID int64) error {
	return s.repo.UpdateCallOnJoin(ctx, chatID, false)
}

func (s *Service) SetMentionsPerMessage(ctx context.Context, chatID int64, count int32) error {
	return s.repo.SetMentionsPerMessage(ctx, chatID, count)
}

func (s *Service) SetMentionTypes(ctx context.Context, chatID int64, types int32) error {
	return s.repo.SetMentionTypes(ctx, chatID, types)
}

func (s *Service) GetInactiveMembers(ctx context.Context, chatID int64) ([]model.ChatMember, error) {
	inactive, err := s.statsService.GetInactiveMembers(ctx, chatID)
	if err != nil {
		return nil, err
	}
	members := make([]model.ChatMember, len(inactive))
	for i, m := range inactive {
		members[i] = m.Member
	}
	return members, nil
}
