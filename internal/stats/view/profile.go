package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"strings"
	"time"
)

func FormatProfile(m model.MemberStats) string {
	var displayName string
	if m.CustomTitle != nil && *m.CustomTitle != "" {
		displayName = *m.CustomTitle
	} else {
		fullName := strings.TrimSpace(m.User.FirstName + " " + m.User.LastName)
		if fullName == "" {
			fullName = "—"
		}
		displayName = fullName
	}

	name := helpers.LinkWithContent(
		m.User,
		displayName,
	)

	status := helpers.TranslateMemberStatus(m.Status, m.LeftAt)
	leftAtInfo := ""
	if m.LeftAt != nil {
		leftAtInfo = ", покинул чат" + helpers.FormatToHumanDate(*m.LeftAt)
	}
	text := fmt.Sprintf(
		`> Информация о %s
> %s (в чате с %s%s)
────────────────
📊 Актив<blockquote expandable>▸ сегодня: %d
▸ эта неделя: %d
▸ этот месяц: %d</blockquote>
────────────────
📝 Всего сообщений: %d
────────────────
⏰ За последние<blockquote expandable>▸ сутки: %d 
▸ 7 дней: %d
▸ 30 дней: %d</blockquote>
────────────────`,
		name,
		status,
		helpers.FormatToHumanDateTime(m.JoinedAt),
		leftAtInfo,
		m.DayCount,
		m.WeekCount,
		m.MonthCount,
		m.AllTime,
		m.DayRollingCount,
		m.WeekRollingCount,
		m.MonthRollingCount,
	)
	isRestActive := m.RestUntil != nil && m.RestUntil.After(time.Now())

	if m.NormWarn > 0 || m.NormBan > 0 {
		normStatus := getNormStatusEmoji(m.WeekCount, m.NormWarn, m.NormBan, isRestActive)
		text += fmt.Sprintf("\n%s", normStatus)
	}

	if m.RestUntil != nil {
		if m.RestUntil.After(time.Now()) {
			text += fmt.Sprintf(
				"<blockquote>💤 Рест до %s</blockquote>",
				helpers.FormatToHumanDateTime(*m.RestUntil),
			)
		} else {
			text += fmt.Sprintf(
				"\n<blockquote>💤 Последний рест был завершен %s</blockquote>",
				helpers.FormatToHumanDate(*m.RestUntil),
			)
		}
	}

	return text
}

func getNormStatusEmoji(weekCount, normWarn, normBan int32, isRestActive bool) string {
	if isRestActive {
		return "🕊 Освобождение от нормы"
	}
	if normBan > 0 && weekCount < normBan {
		return fmt.Sprintf("🚫 Норма не набрана (%d/%d), <b>бан</b>", weekCount, normBan)
	}
	if normWarn > 0 && weekCount < normWarn {
		return fmt.Sprintf("⚠️ Норма не набрана (%d/%d), <b>варн</b>", weekCount, normWarn)
	}
	return "✅ Норма набрана"
}
