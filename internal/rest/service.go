package rest

import (
	"activity-bot/internal/model"
	"context"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo}
}

func (s *Service) GetRestMembers(ctx context.Context, chatID int64) ([]model.RestMember, error) {
	return s.repo.GetFromChat(ctx, chatID)
}

func (s *Service) SetMemberRest(ctx context.Context, chatID int64, userID int64, until time.Time) error {
	return s.repo.SetRest(ctx, chatID, userID, until)
}

func (s *Service) GetMemberRest(ctx context.Context, chatID int64, userID int64) (*time.Time, error) {
	return s.repo.GetRestUntil(ctx, chatID, userID)
}

func (s *Service) EndMemberRest(ctx context.Context, chatID int64, userID int64) error {
	return s.repo.EndMemberRest(ctx, chatID, userID)
}

func (s *Service) CreateRestRequest(ctx context.Context, chatID, userID, messageID int64, until time.Time) error {
	return s.repo.AddRequest(ctx, model.RestRequest{
		ChatID:      chatID,
		UserID:      userID,
		RequestedAt: time.Now(),
		RestUntil:   until,
		Status:      "pending",
		MessageID:   messageID,
	})
}

func (s *Service) GetRestRequest(ctx context.Context, chatID int64, userID, messageID int64) (model.RestRequest, error) {
	return s.repo.GetRequest(ctx, chatID, userID, messageID)
}

func (s *Service) ApproveRestRequest(ctx context.Context, chatID, userID, messageID int64, until time.Time) error {
	return s.repo.ApproveRequestWithTx(ctx, model.RestRequest{
		ChatID:    chatID,
		UserID:    userID,
		RestUntil: until,
		MessageID: messageID,
	})
}

func (s *Service) RejectRestRequest(ctx context.Context, chatID, userID, messageID int64) error {
	return s.repo.RejectRequest(ctx, chatID, userID, messageID)
}
