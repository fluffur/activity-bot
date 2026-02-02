package stats

import (
	"activity-bot/internal/model"
	"context"
	"time"
)

type Repository interface {
	GetReport(ctx context.Context, chatID int64, from, to *time.Time) ([]model.MessageReportMember, error)
	GetReportOne(ctx context.Context, chatID int64, userID int64) (model.MemberStats, error)
}
