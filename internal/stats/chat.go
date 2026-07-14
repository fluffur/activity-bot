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

	fromDate, toDate := ParseTimeRange(ch, args)

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

func ParseTimeRange(ch chat.Chat, args cctx.ParsedArgs) (from, to time.Time) {
	loc, _ := time.LoadLocation("Europe/Moscow")
	now := time.Now().In(loc)

	switch {
	case len(args.Durations) == 1:
		return now.Add(-args.Durations[0]), now

	case len(args.DateTimes) == 2:
		return args.DateTimes[0], args.DateTimes[1]

	default:
		from, to = currentChatWeekRange(now, ch.WeekStartDay, ch.WeekStartTimeMicros)

		return from, to
	}
}

func currentChatWeekRange(
	now time.Time,
	weekStartDay int16,
	weekStartTimeMicros int64,
) (from, to time.Time) {
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

	return start, now
}
