package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"time"
)

func FormatProfile(m model.MemberStats) string {
	customTitle := "—"
	if m.CustomTitle != nil && *m.CustomTitle != "" {
		customTitle = *m.CustomTitle
	}

	name := helpers.LinkWithContent(
		m.User,
		fmt.Sprintf("%s (%s)", m.User.FirstName, customTitle),
	)

	status := helpers.TranslateMemberStatus(m.Status, m.LeftAt)

	text := fmt.Sprintf(
		`👤 <b>Информация о:</b> %s
👑 <b>Статус:</b> %s (с %s)
	
📊 <b>Активность</b>
└ Всего: <b>%d</b>
└ Сегодня: <b>%d</b>
└ Неделя: <b>%d</b>

⏰ <b>Динамика</b>
└ 24ч: <b>%d</b> | 7д: <b>%d</b> | 30д: <b>%d</b>`,
		name,
		status,
		helpers.FormatToHumanDate(m.JoinedAt),
		m.AllTime,
		m.DayCount,
		m.WeekCount,
		m.DayRollingCount,
		m.WeekRollingCount,
		m.MonthRollingCount,
	)

	if m.RestUntil != nil {
		if m.RestUntil.After(time.Now()) {
			text += fmt.Sprintf(
				"\n\n💤 <b>Рест до:</b> %s",
				helpers.FormatToHumanDate(*m.RestUntil),
			)
		} else {
			text += fmt.Sprintf(
				"\n\n💤 <b>Последний рест был завершен:</b> %s",
				helpers.FormatToHumanDate(*m.RestUntil),
			)
		}
	}

	if m.NormWarn > 0 || m.NormBan > 0 {
		normEmoji := getNormEmoji(
			m.WeekCount,
			m.NormBan,
			m.NormWarn,
			m.RestUntil != nil && m.RestUntil.After(time.Now()),
		)

		text += fmt.Sprintf("\n\n%s <b>Норма (неделя)</b>\n", normEmoji)

		if m.NormWarn > 0 {
			text += fmt.Sprintf(
				"└ Варн: <b>%d</b> / %d\n",
				m.WeekCount,
				m.NormWarn,
			)
		}

		if m.NormBan > 0 {
			text += fmt.Sprintf(
				"└ Бан: <b>%d</b> / %d",
				m.WeekCount,
				m.NormBan,
			)
		}
	}

	return text
}

func getNormEmoji(weekCount, normBan, normWarn int32, isInRest bool) string {
	if isInRest {
		return "😴"
	}
	if normBan == 0 && normWarn == 0 {
		return "⚪"
	}

	if weekCount < normBan {
		return "❌"
	}

	if weekCount < normWarn {
		return "⚠️"
	}

	return "✅"
}
