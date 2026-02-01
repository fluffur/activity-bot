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

func (s *Service) GetRestMembers(chatID int64) ([]model.RestMember, error) {
	ctx := context.Background()
	return s.repo.GetFromChat(ctx, chatID)
}

func (s *Service) SetMemberRest(chatID int64, userID int64, until time.Time) error {
	ctx := context.Background()
	return s.repo.SetRest(ctx, chatID, userID, until)
}

func (s *Service) GetMemberRest(chatID int64, userID int64) (*time.Time, error) {
	ctx := context.Background()
	return s.repo.GetRestUntil(ctx, chatID, userID)
}

func (s *Service) EndMemberRest(chatID int64, userID int64) error {
	ctx := context.Background()
	return s.repo.EndMemberRest(ctx, chatID, userID)
}

func (s *Service) CreateRestRequest(chatID, userID, messageID int64, until time.Time) error {
	ctx := context.Background()
	return s.repo.AddRequest(ctx, model.RestRequest{
		ChatID:      chatID,
		UserID:      userID,
		RequestedAt: time.Now(),
		RestUntil:   until,
		Status:      "pending",
		MessageID:   messageID,
	})
}

func (s *Service) GetRestRequest(chatID int64, userID, messageID int64) (model.RestRequest, error) {
	ctx := context.Background()
	return s.repo.GetRequest(ctx, chatID, userID, messageID)
}

func (s *Service) ApproveRestRequest(chatID, userID, messageID int64, until time.Time) error {
	ctx := context.Background()
	return s.repo.ApproveRequestWithTx(ctx, model.RestRequest{
		ChatID:    chatID,
		UserID:    userID,
		RestUntil: until,
		MessageID: messageID,
	})
}

func (s *Service) RejectRestRequest(chatID, userID, messageID int64) error {
	ctx := context.Background()
	return s.repo.RejectRequest(ctx, chatID, userID, messageID)
}
