package chat

import (
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"context"
	"time"
)

type Service struct {
	repo              Repository
	userRepo          user.Repository
	defaultWeeklyNorm int32
}

func NewService(repo Repository, userRepo user.Repository, defaultWeeklyNorm int32) *Service {
	return &Service{repo, userRepo, defaultWeeklyNorm}
}

func (s *Service) GetNorm(ctx context.Context, chatID int64) (int32, error) {
	c, err := s.repo.GetOrCreate(ctx, model.NewChat(chatID, s.defaultWeeklyNorm))
	if err != nil {
		return 0, err
	}

	return c.WeeklyNorm, nil
}

func (s *Service) SetNorm(ctx context.Context, chatID int64, norm int) error {
	return s.repo.SetNorm(ctx, chatID, norm)
}

func (s *Service) GetMemberStats(ctx context.Context, chatID int64) ([]model.WeeklyMessageReportMember, error) {
	return s.repo.GetWeeklyReport(ctx, chatID)
}

func (s *Service) GetExemptMembers(ctx context.Context, chatID int64) ([]model.ExemptMember, error) {
	return s.repo.GetChatExemptUsers(ctx, chatID)
}

func (s *Service) ExemptUser(ctx context.Context, chatID int64, userID int64, exemptUntil time.Time) error {
	return s.repo.ExemptMember(ctx, chatID, userID, exemptUntil)
}

func (s *Service) GetMemberExempt(ctx context.Context, chatID int64, userID int64) (*time.Time, error) {
	member, err := s.repo.GetMember(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	return member.ExemptUntil, nil
}

func (s *Service) EndMemberExempt(ctx context.Context, chatID int64, userID int64) error {
	return s.repo.RemoveMemberExempt(ctx, chatID, userID)
}

func (s *Service) CreateExemptRequest(ctx context.Context, chatID, userID, messageID int64, exemptUntil time.Time) error {
	return s.repo.AddExemptRequest(ctx, model.ExemptRequest{
		ChatID:      chatID,
		UserID:      userID,
		RequestedAt: time.Now(),
		ExemptUntil: exemptUntil,
		Status:      "pending",
		MessageID:   messageID,
	})
}

func (s *Service) GetExemptRequest(ctx context.Context, chatID int64, userID, messageID int) (model.ExemptRequest, error) {
	return s.repo.GetExemptRequest(ctx, chatID, int64(userID), messageID)
}

func (s *Service) ApproveExemptRequest(ctx context.Context, chatID, userID, messageID int64, exemptUntil time.Time) error {
	return s.repo.ApproveExemptWithTx(ctx, model.ExemptRequest{
		ChatID:      chatID,
		UserID:      userID,
		ExemptUntil: exemptUntil,
		MessageID:   messageID,
	})
}

func (s *Service) RejectExemptRequest(ctx context.Context, chatID, userID int64, messageID int) error {
	return s.repo.RejectExemptRequest(ctx, chatID, userID, messageID)
}

func (s *Service) AddAdmin(ctx context.Context, chatID int64, userID int64) error {
	return s.repo.AddAdmin(ctx, chatID, userID)
}

func (s *Service) RemoveAdmin(ctx context.Context, chatID int64, userID int64) error {
	return s.repo.RemoveAdmin(ctx, chatID, userID)
}

func (s *Service) GetAdmins(ctx context.Context, chatID int64) ([]model.ChatAdmin, error) {
	return s.repo.GetAdmins(ctx, chatID)
}

func (s *Service) IsAdmin(ctx context.Context, chatID int64, userID int64) (bool, error) {
	return s.repo.IsAdmin(ctx, chatID, userID)
}

func (s *Service) UpdateChatMembers(ctx context.Context, chatID int64, admins []model.User) error {
	if err := s.userRepo.UpsertUsers(ctx, admins); err != nil {
		return err
	}
	userIDs := make([]int64, len(admins))
	for i, u := range admins {
		userIDs[i] = u.ID
	}
	return s.repo.UpsertChatMembers(ctx, chatID, userIDs)
}
