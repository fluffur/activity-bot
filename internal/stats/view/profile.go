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
		leftAtInfo = ", покинул чат " + helpers.FormatToHumanDate(*m.LeftAt)
	}

	text := fmt.Sprintf(
		`> Информация о %s
> %s %s (в чате с %s%s)
───────────────
📊 Актив<blockquote>▸ сегодня: %d
▸ эта неделя: %d
▸ этот месяц: %d
▸ всего: %d</blockquote>
───────────────
⏰ За последние<blockquote>▸ сутки: %d 
▸ 7 дней: %d
▸ 30 дней: %d</blockquote>───────────────`,
		name,
		status,
		getStatusEmoji(m.Status),
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

		now := time.Now()
		daysSinceRestEnd := int(now.Sub(*m.RestUntil).Hours() / 24)

		if m.RestUntil.After(time.Now()) {
			text += fmt.Sprintf(
				"<blockquote>💤 Рест до %s</blockquote>",
				helpers.FormatToHumanDateTime(*m.RestUntil),
			)
		} else if daysSinceRestEnd >= 0 && daysSinceRestEnd <= 3 {
			text += fmt.Sprintf(
				"\n<blockquote>💤 Последний рест был завершен %s</blockquote>",
				helpers.FormatToHumanDate(*m.RestUntil),
			)
		}
	}

	return text
}

func getStatusEmoji(status string) string {

	switch status {
	case "member":
		return helpers.CustomEmoji(5298673841378191838, "🤓")
	case "administrator":
		return helpers.CustomEmoji(5296305123964773541, "💀")
	case "creator":
		return helpers.CustomEmoji(5298512178809168321, "😎")
	}
	return helpers.CustomEmoji(5298532506889382420, "😭")
}

func getNormStatusEmoji(weekCount, normWarn, normBan int32, isRestActive bool) string {
	if isRestActive {
		return fmt.Sprintf("%s Освобождение от нормы", helpers.CustomEmoji(5456648248968121823, "🕊"))
	}
	if normBan > 0 && weekCount < normBan {
		return fmt.Sprintf("🚫 Норма не набрана (%d/%d), <b>бан</b>", weekCount, normBan)
	}
	if normWarn > 0 && weekCount < normWarn {
		return fmt.Sprintf("⚠️ Норма не набрана (%d/%d), <b>варн</b>", weekCount, normWarn)
	}
	return "✅ Норма набрана"
}
