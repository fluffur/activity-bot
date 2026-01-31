package stats

import (
	"activity-bot/internal/model"
	"context"
	"time"
)

type Repository interface {
	GetReport(ctx context.Context, chatID int64, from, to *time.Time) ([]model.MessageReportMember, error)
	ImportActivityBulk(ctx context.Context, chatID int64, from, to time.Time, userIDs []int64, counts []int32) error
}
