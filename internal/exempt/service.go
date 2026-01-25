package exempt

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

func (s *Service) GetExemptMembers(chatID int64) ([]model.ExemptMember, error) {
	ctx := context.Background()
	return s.repo.GetFromChat(ctx, chatID)
}

func (s *Service) ExemptMember(chatID int64, userID int64, exemptUntil time.Time) error {
	ctx := context.Background()
	return s.repo.Exempt(ctx, chatID, userID, exemptUntil)
}

func (s *Service) GetMemberExempt(chatID int64, userID int64) (*time.Time, error) {
	ctx := context.Background()
	return s.repo.Get(ctx, chatID, userID)
}

func (s *Service) EndMemberExempt(chatID int64, userID int64) error {
	ctx := context.Background()
	return s.repo.Remove(ctx, chatID, userID)
}

func (s *Service) CreateExemptRequest(chatID, userID, messageID int64, exemptUntil time.Time) error {
	ctx := context.Background()
	return s.repo.AddRequest(ctx, model.ExemptRequest{
		ChatID:      chatID,
		UserID:      userID,
		RequestedAt: time.Now(),
		ExemptUntil: exemptUntil,
		Status:      "pending",
		MessageID:   messageID,
	})
}

func (s *Service) GetExemptRequest(chatID int64, userID, messageID int64) (model.ExemptRequest, error) {
	ctx := context.Background()
	return s.repo.GetRequest(ctx, chatID, userID, messageID)
}

func (s *Service) ApproveExemptRequest(chatID, userID, messageID int64, exemptUntil time.Time) error {
	ctx := context.Background()
	return s.repo.ApproveRequestWithTx(ctx, model.ExemptRequest{
		ChatID:      chatID,
		UserID:      userID,
		ExemptUntil: exemptUntil,
		MessageID:   messageID,
	})
}

func (s *Service) RejectExemptRequest(chatID, userID, messageID int64) error {
	ctx := context.Background()
	return s.repo.RejectRequest(ctx, chatID, userID, messageID)
}
