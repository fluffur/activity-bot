package stats

import (
	"activity-bot/internal/model"
	"context"
	"time"
)

type Repository interface {
	GetReport(ctx context.Context, chatID int64, from, to *time.Time) ([]model.MessageReportMember, error)
	GetReportOne(ctx context.Context, chatID int64, userID int64) (model.MemberStats, error)
	GetMessageActivityByDay(ctx context.Context, chatID int64, userID int64) ([]model.MessageActivity, error)
	GetInactiveMembers(ctx context.Context, chatID int64) ([]model.InactiveMember, error)
	GetMessageActivityByDayAll(ctx context.Context, chatID int64, from *time.Time, to *time.Time) ([]model.MessageActivity, error)
}
