package stats

import (
	"activity-bot/internal/model"
	"context"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo}
}

func (s *Service) GetMemberStats(chatID int64) ([]model.WeeklyMessageReportMember, error) {
	ctx := context.Background()
	return s.repo.GetWeeklyReport(ctx, chatID)
}
