package chat

import "context"

type Service struct {
	repository Repository
}

func NewService(r Repository) *Service {
	return &Service{
		repository: r,
	}
}

func (s *Service) Create(ctx context.Context, chat Chat) error {
	return s.repository.Create(ctx, chat)
}

func (s *Service) Get(ctx context.Context, chatID int64) (Chat, error) {
	return s.repository.Get(ctx, chatID)
}

func (s *Service) SetMentionTypes(ctx context.Context, chatID int64, mentionTypes MentionTypes) error {
	return s.repository.SetMentionTypes(ctx, chatID, mentionTypes)
}

func (s *Service) SetSkipSummonConfirmation(ctx context.Context, chatID int64, skipSummonConfirmation bool) error {
	return s.repository.SetSkipSummonConfirmation(ctx, chatID, skipSummonConfirmation)
}
