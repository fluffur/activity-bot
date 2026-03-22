package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"time"
)

func FormatProfile(m model.MemberStats, full bool) string {
	now := time.Now().UTC()

	isNewbie := false
	if m.NewbieThreshold > 0 {
		newbieUntil := m.JoinedAt.AddDate(0, 0, m.NewbieThreshold)
		if newbieUntil.After(now) {
			isNewbie = true
		}
	}

	var newbieEmoji string
	if isNewbie {
		newbieEmoji = " " + helpers.NewbieEmoji()
	}

	status := helpers.TranslateMemberStatus(m.Status, m.LeftAt)
	leftAtInfo := ""
	if !m.LeftAt.IsZero() {
		leftAtInfo = ", покинул чат " + helpers.FormatLastSeen(m.LeftAt)
	}

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
			helpers.CustomEmoji(5337121636992690373, "⏰"),
			m.DayRollingCount,
			m.WeekRollingCount,
			m.MonthRollingCount,
			helpers.Line(),
		)
	}

	text := fmt.Sprintf(
		`%s Информация о %s
• %s %s (%s %s%s%s)
%s`,
		helpers.CustomEmoji(5316727448644103237, "👤"),
		helpers.RoleEmojiLink(m.ChatMember),
		status,
		helpers.StatusEmoji(m.Status),
		helpers.Gendered(m.ChatMember.User.Gender, "зашел", "зашла"),
		helpers.FormatLastSeen(m.JoinedAt),
		newbieEmoji,
		leftAtInfo,
		activityBlock,
	)
	isRestActive := m.RestUntil.After(time.Now())

	if m.NormWarn > 0 || m.NormBan > 0 {
		normStatus := getNormStatusEmoji(m.WeekCount, m.NormWarn, m.NormBan, isRestActive, isNewbie)
		text += fmt.Sprintf("\n%s", normStatus)
	}

	daysSinceRestEnd := int(now.Sub(m.RestUntil).Hours() / 24)

	if isRestActive && m.LeftAt.IsZero() {
		text += fmt.Sprintf(
			"<blockquote>%s Рест до %s</blockquote>",
			helpers.RestEmoji(),
			helpers.FormatToHumanDateTime(m.RestUntil),
		)
	} else if daysSinceRestEnd >= 0 && daysSinceRestEnd <= 3 {
		text += fmt.Sprintf(
			"\n<blockquote>%s Последний рест был завершен %s</blockquote>",
			helpers.RestEmoji(),
			helpers.FormatToHumanDate(m.RestUntil),
		)
	}

	return text
}

func getNormStatusEmoji(weekCount, normWarn, normBan int, isRestActive, isNewbie bool) string {
	if isRestActive || isNewbie {
		return fmt.Sprintf("%s Освобождение от нормы", helpers.CustomEmoji(5456648248968121823, "🕊"))
	}
	if normBan > 0 && weekCount < normBan {
		return fmt.Sprintf("%s Норма не набрана (%d/%d), <b>бан</b>", helpers.CustomEmoji(5224340348465073584, "🚫"), weekCount, normBan)
	}
	if normWarn > 0 && weekCount < normWarn {
		return fmt.Sprintf("%s Норма не набрана (%d/%d), <b>варн</b>", helpers.CustomEmoji(5224340348465073584, "⚠️"), weekCount, normWarn)
	}
	return fmt.Sprintf("%s Норма набрана", helpers.CustomEmoji(5224694451338759997, "✅"))
}
