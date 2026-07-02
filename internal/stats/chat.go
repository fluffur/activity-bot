package stats

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/chat"
	"fmt"
	"time"

	"github.com/gotd/botapi"
)

func (h *Handler) Chat(c *botapi.Context) error {
	args := cctx.MustArgs(c)
	ch := cctx.MustChat(c)
	fromDate, toDate, err := ParseTimeRange(ch, args)
	if err != nil {
		return fmt.Errorf("stats time range: %w", err)
	}

	calculatedData, err := h.service.GetChatStats(c, ch.ID, fromDate, toDate, ch.NewbieThresholdDays)
	if err != nil {
		return fmt.Errorf("failed to calculate stats: %w", err)
	}

	loc := cctx.MustLocalizer(c)
	htmlMessage := RenderStats(loc, ch, calculatedData)

	_, err = c.Reply(
		htmlMessage,
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)
	return err
}

func ParseTimeRange(ch chat.Chat, args cctx.ParsedArgs) (time.Time, time.Time, error) {
	now := time.Now()

	switch {
	case len(args.Durations) == 1:
		return now.Add(-args.Durations[0]), now, nil

	case len(args.DateTimes) == 2:
		return args.DateTimes[0], args.DateTimes[1], nil

	default:
		from, to, err := currentChatWeekRange(ch.WeekStartDay, ch.WeekStartTimeMicros)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("current week range: %w", err)
		}
		return from, to, nil
	}
}

func currentChatWeekRange(
	weekStartDay int16,
	weekStartTimeMicros int64,
) (time.Time, time.Time, error) {
	now := time.Now()

	startTime := time.Duration(weekStartTimeMicros) * time.Microsecond

	hours := int(startTime / time.Hour)
	minutes := int((startTime % time.Hour) / time.Minute)
	seconds := int((startTime % time.Minute) / time.Second)

	var weekDay time.Weekday

	if weekStartDay == 7 {
		weekDay = time.Sunday
	} else {
		weekDay = time.Weekday(weekStartDay)
	}

	daysBack := (7 + int(now.Weekday()) - int(weekDay)) % 7

	start := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		hours,
		minutes,
		seconds,
		0,
		now.Location(),
	).AddDate(0, 0, -daysBack)

	if start.After(now) {
		start = start.AddDate(0, 0, -7)
	}

	return start, now, nil
}
