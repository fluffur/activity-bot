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

func (s *Service) GetExemptMembers(ctx context.Context, chatID int64) ([]model.ExemptMember, error) {
	return s.repo.GetFromChat(ctx, chatID)
}

func (s *Service) ExemptMember(ctx context.Context, chatID int64, userID int64, exemptUntil time.Time) error {
	return s.repo.Exempt(ctx, chatID, userID, exemptUntil)
}

func (s *Service) GetMemberExempt(ctx context.Context, chatID int64, userID int64) (*time.Time, error) {
	return s.repo.Get(ctx, chatID, userID)
}

func (s *Service) EndMemberExempt(ctx context.Context, chatID int64, userID int64) error {
	return s.repo.Remove(ctx, chatID, userID)
}

func (s *Service) CreateExemptRequest(ctx context.Context, chatID, userID, messageID int64, exemptUntil time.Time) error {
	return s.repo.AddRequest(ctx, model.ExemptRequest{
		ChatID:      chatID,
		UserID:      userID,
		RequestedAt: time.Now(),
		ExemptUntil: exemptUntil,
		Status:      "pending",
		MessageID:   messageID,
	})
}

func (s *Service) GetExemptRequest(ctx context.Context, chatID int64, userID, messageID int) (model.ExemptRequest, error) {
	return s.repo.GetRequest(ctx, chatID, int64(userID), messageID)
}

func (s *Service) ApproveExemptRequest(ctx context.Context, chatID, userID, messageID int64, exemptUntil time.Time) error {
	return s.repo.ApproveRequestWithTx(ctx, model.ExemptRequest{
		ChatID:      chatID,
		UserID:      userID,
		ExemptUntil: exemptUntil,
		MessageID:   messageID,
	})
}

func (s *Service) RejectExemptRequest(ctx context.Context, chatID, userID int64, messageID int) error {
	return s.repo.RejectRequest(ctx, chatID, userID, messageID)
}
