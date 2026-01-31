package stats

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

func (s *Service) GetMemberStats(chatID int64, from, to *time.Time) ([]model.MessageReportMember, error) {
	ctx := context.Background()
	return s.repo.GetReport(ctx, chatID, from, to)
}

func (s *Service) ImportActivity(chatID int64, period ReportPeriod, userIDs []int64, counts []int32) error {
	from, to := ResolvePeriod(period, time.Now())
	if from == nil || to == nil {
		return nil
	}

	return s.repo.ImportActivityBulk(context.Background(), chatID, *from, *to, userIDs, counts)
}

type ReportPeriod string

const (
	PeriodDay        ReportPeriod = "day"
	PeriodWeek       ReportPeriod = "week"
	PeriodMonth      ReportPeriod = "month"
	PeriodSevenDays  ReportPeriod = "seven_days"
	PeriodThirtyDays ReportPeriod = "thirty_days"
	PeriodAll        ReportPeriod = "all"
)

func ResolvePeriod(period ReportPeriod, now time.Time) (from *time.Time, to *time.Time) {
	switch period {
	case PeriodDay:
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 0, 1).Add(-time.Second)
		return &start, &end
	case PeriodWeek:
		weekday := int(now.Weekday())
		daysSinceMonday := (weekday + 6) % 7
		monday := now.AddDate(0, 0, -daysSinceMonday)
		monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
		sunday := monday.AddDate(0, 0, 6)
		sunday = time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 0, sunday.Location())

		return &monday, &sunday

	case PeriodMonth:
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 1, 0).Add(-time.Second)
		return &start, &end

	case PeriodSevenDays:
		start := now.AddDate(0, 0, -7)
		return &start, &now

	case PeriodThirtyDays:
		start := now.AddDate(0, 0, -30)
		return &start, &now

	case PeriodAll:
		return nil, nil
	}

	return nil, nil
}
