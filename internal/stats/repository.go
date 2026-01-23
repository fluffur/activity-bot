package stats

import (
	"activity-bot/internal/model"
	"context"
)

type Repository interface {
	GetWeeklyReport(ctx context.Context, chatID int64) ([]model.WeeklyMessageReportMember, error)
}
