package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"time"
)

func FormatProfile(m model.ChatMemberStats, full bool) string {
	now := time.Now().UTC()

	isNewbie := false
	if m.Chat.NewbieThresholdDays > 0 {
		newbieUntil := m.ChatMember.JoinedAt.AddDate(0, 0, int(m.Chat.NewbieThresholdDays))
		if newbieUntil.After(now) {
			isNewbie = true
		}
	}

	status := m.ChatMember.Status.String()

	var activityBlock string

	if full {
		activityBlock = fmt.Sprintf(`%s
%s Актив<blockquote>▸ сегодня: <code>%d</code>
▸ эта неделя: <code>%d</code>
▸ этот месяц: <code>%d</code>
▸ всего: <code>%d</code></blockquote>
%s
%s За последние<blockquote>▸ сутки: <code>%d</code> 
▸ 7 дней: <code>%d</code>
▸ 30 дней: <code>%d</code></blockquote>
%s`,
			helpers.Line(),
			helpers.StatsEmoji(),
			m.DayCount,
			m.WeekCount,
			m.MonthCount,
			m.AllTime,
			helpers.Line(),
			helpers.CustomEmoji("5258419835922030550", "⏰"),
			m.DayRollingCount,
			m.WeekRollingCount,
			m.MonthRollingCount,
			helpers.Line(),
		)
	}

	var info string
	if m.ChatMember.LeftAt.IsZero() {
		info = fmt.Sprintf("• В чате c %s (%s)", helpers.FormatToHumanDateTime(m.ChatMember.JoinedAt), helpers.FormatLastSeenPlain(m.ChatMember.JoinedAt))
	} else {
		info = fmt.Sprintf("❌ %s чат (%s — %s)", helpers.Gendered(m.ChatMember.User.Gender, "Покинул", "Покинула"), helpers.FormatToHumanDateTime(m.ChatMember.JoinedAt), helpers.FormatToHumanDateTime(m.ChatMember.LeftAt))
	}
	text := fmt.Sprintf(
		`%s Информация о %s
• Ранг: %s – %s
%s 
%s`,
		helpers.CustomEmoji("5316727448644103237", "👤"),
		helpers.RoleEmojiLink(m.ChatMember),
		helpers.StatusEmoji(m.ChatMember.Status),
		status,
		info,
		activityBlock,
	)
	isRestActive := m.ChatMember.RestUntil.After(time.Now())

	if m.Chat.NormWarn > 0 || m.Chat.NormBan > 0 {
		normStatus := getNormStatusEmoji(m.WeekCount, int64(m.Chat.NormWarn), int64(m.Chat.NormBan), isRestActive, isNewbie)
		text += fmt.Sprintf("\n%s", normStatus)
	}

	daysSinceRestEnd := int(now.Sub(m.ChatMember.RestUntil).Hours() / 24)

	if isRestActive && m.ChatMember.LeftAt.IsZero() {
		text += fmt.Sprintf(
			"<blockquote>%s Рест до %s</blockquote>",
			helpers.RestEmoji(),
			helpers.FormatToHumanDateTime(m.ChatMember.RestUntil),
		)
	} else if daysSinceRestEnd >= 0 && daysSinceRestEnd <= 3 {
		text += fmt.Sprintf(
			"\n<blockquote>%s Последний рест был завершен %s</blockquote>",
			helpers.RestEmoji(),
			helpers.FormatToHumanDateTime(m.ChatMember.RestUntil),
		)
	}

	return text
}

func getNormStatusEmoji(weekCount, normWarn, normBan int64, isRestActive, isNewbie bool) string {
	if isRestActive || isNewbie {
		return fmt.Sprintf("%s Освобождение от нормы", helpers.CustomEmoji("5456648248968121823", "🕊"))
	}
	if normBan > 0 && weekCount < normBan {
		return fmt.Sprintf("%s Норма не набрана (%d/%d), <b>бан</b>", helpers.CustomEmoji("5260342697075416641", "🚫"), weekCount, normBan)
	}
	if normWarn > 0 && weekCount < normWarn {
		return fmt.Sprintf("%s Норма не набрана (%d/%d), <b>варн</b>", helpers.CustomEmoji("5258474669769497337", "⚠️"), weekCount, normWarn)
	}
	return fmt.Sprintf("%s Норма набрана", helpers.CustomEmoji("5260416304224936047", "✅"))
}
