package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
)

func FormatProfile(m model.MemberStats) string {
	customTitle := "—"
	if m.CustomTitle != nil && *m.CustomTitle != "" {
		customTitle = *m.CustomTitle
	}
	extraText := ""

	if m.RestUntil != nil {
		extraText = fmt.Sprintf("💤 Рест до %s", helpers.FormatToHumanDate(*m.RestUntil))
	} else {
		if m.NormWarn > 0 {
			warnEmoji := "❌"
			if m.WeekCount >= m.NormWarn {
				warnEmoji = "✅"
			}
			extraText += fmt.Sprintf(
				"%s <b>Норма (варн)</b>: %d/%d за эту неделю",
				warnEmoji,
				m.WeekCount,
				m.NormWarn,
			)

		}

		if m.NormBan > 0 {
			banEmoji := "❌"
			if m.WeekCount >= m.NormBan {
				banEmoji = "✅"
			}
			extraText += fmt.Sprintf(
				"\n%s <b>Норма (бан)</b>: %d/%d за эту неделю",
				banEmoji,
				m.WeekCount,
				m.NormBan,
			)
		}
	}

	return fmt.Sprintf(
		`<b>📊 Информация о %s</b>

🌟 Статус: <b>%s</b> | Присоединился: <b>%s</b>

<b>📅 Активность</b>: сегодня <b>%d</b> | неделя <b>%d</b> | месяц <b>%d</b>

<b>🔄 Активность в последние</b>: 24ч <b>%d</b> | 7д <b>%d</b> | 30д <b>%d</b>

📝 <b>Всего сообщений:</b> %d

%s
`,
		helpers.LinkWithContent(m.User, fmt.Sprintf("%s (%s)", m.User.FirstName, customTitle)),
		helpers.TranslateMemberStatus(m.Status),
		helpers.FormatToHumanDate(m.JoinedAt),
		m.DayCount,
		m.WeekCount,
		m.MonthCount,
		m.DayRollingCount,
		m.WeekRollingCount,
		m.MonthRollingCount,
		m.AllTime,
		extraText,
	)
}
