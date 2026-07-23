package stats

import (
	"activity-bot/internal/cctx"
	"fmt"
	"time"

	"github.com/gotd/botapi"
)

func (h *Handler) Profile(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	cm := cctx.MustChatMember(c)
	args, err := cctx.Args(c)
	if err == nil {
		if u, ok := args.User(); ok {
			cm = u
		}
	}

	statsRange := buildProfileStatsRange(ch.WeekStartDay, ch.WeekStartTimeMicros)

	profile, err := h.service.GetProfileStats(c, ch.ID, cm.User.ID, statsRange)
	if err != nil {
		return fmt.Errorf("profile stats: %w", err)
	}

	loc := cctx.MustLocalizer(c)
	htmlMessage := RenderProfile(loc, ch, profile, statsRange.WeekStart)

	_, err = c.Reply(
		htmlMessage,
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)

	return err
}

func buildProfileStatsRange(
	weekStartDay int16,
	weekStartTimeMicros int64,
) ProfileStatsRange {
	loc, _ := time.LoadLocation("Europe/Moscow")
	now := time.Now().In(loc)

	weekStart, _ := currentChatWeekRange(
		now,
		weekStartDay,
		weekStartTimeMicros,
	)

	dayStart := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0, 0, 0, 0,
		now.Location(),
	)

	monthStart := time.Date(
		now.Year(),
		now.Month(),
		1,
		0, 0, 0, 0,
		now.Location(),
	)

	return ProfileStatsRange{
		DayStart:          dayStart,
		DayRollingStart:   now.Add(-24 * time.Hour),
		WeekStart:         weekStart,
		WeekRollingStart:  now.Add(-7 * 24 * time.Hour),
		MonthStart:        monthStart,
		MonthRollingStart: now.Add(-30 * 24 * time.Hour),
	}
}
