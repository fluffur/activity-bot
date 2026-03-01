package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"log/slog"
	"time"
)

func FormatProfile(m model.MemberStats) string {
	customTitle := "—"
	if m.CustomTitle != nil && *m.CustomTitle != "" {
		customTitle = *m.CustomTitle
	}
	lastName := ""
	slog.Info("lastname", "lastname", m.User.LastName)
	if m.User.LastName != "" {
		lastName = " " + m.User.LastName
	}
	name := helpers.LinkWithContent(
		m.User,
		fmt.Sprintf("%s (%s)", m.User.FirstName+lastName, customTitle),
	)

	status := helpers.TranslateMemberStatus(m.Status, m.LeftAt)

	text := fmt.Sprintf(
		`👤 <b>Информация о:</b> %s
👑 <b>Статус:</b> %s (с %s)
	
📊 <b>Активность</b>
└ Сегодня: <b>%d</b>
└ Неделя: <b>%d</b>
└ Всего: <b>%d</b>

⏰ <b>Динамика</b>
└ 24ч: <b>%d</b> | 7д: <b>%d</b> | 30д: <b>%d</b>`,
		name,
		status,
		helpers.FormatToHumanDateTime(m.JoinedAt),
		m.DayCount,
		m.WeekCount,
		m.AllTime,
		m.DayRollingCount,
		m.WeekRollingCount,
		m.MonthRollingCount,
	)

	if m.RestUntil != nil {
		if m.RestUntil.After(time.Now()) {
			text += fmt.Sprintf(
				"\n\n💤 <b>Рест до:</b> %s",
				helpers.FormatToHumanDateTime(*m.RestUntil),
			)
		} else {
			text += fmt.Sprintf(
				"\n\n💤 <b>Последний рест был завершен:</b> %s",
				helpers.FormatToHumanDate(*m.RestUntil),
			)
		}
	}

	if m.NormWarn > 0 || m.NormBan > 0 {
		normStatus := getNormStatusEmoji(m.WeekCount, m.NormWarn, m.NormBan)
		text += fmt.Sprintf("\n\n%s", normStatus)
	}

	return text
}

func getNormStatusEmoji(weekCount, normWarn, normBan int32) string {
	if normBan > 0 && weekCount < normBan {
		return fmt.Sprintf("🚫 Норма не набрана (%d/%d), бан", weekCount, normBan)
	}
	if normWarn > 0 && weekCount < normWarn {
		return fmt.Sprintf("⚠️ Норма не набрана (%d/%d), варн", weekCount, normWarn)
	}
	return "✅ Норма набрана"
}
