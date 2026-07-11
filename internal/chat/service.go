package chat

import "context"

type Service struct {
	repository Repository
	ownerID    int64
}

func NewService(r Repository, ownerID int64) *Service {
	return &Service{
		repository: r,
		ownerID:    ownerID,
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

func (s *Service) GetUserManagedChats(ctx context.Context, userID int64, search string) ([]Chat, error) {
	if userID == s.ownerID {
		return s.repository.GetAllChats(ctx, search)
	}
	return s.repository.GetUserManagedChats(ctx, userID, search)
}
