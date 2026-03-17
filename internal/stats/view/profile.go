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
	if m.CustomTitle != "" {
		displayName = m.CustomTitle
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

	text := fmt.Sprintf(
		`%s Информация о %s
• %s %s (%s %s%s%s)
%s
%s Актив<blockquote>▸ сегодня: <code>%d</code>
▸ эта неделя: <code>%d</code>
▸ этот месяц: <code>%d</code>
▸ всего: <code>%d</code></blockquote>
%s
%s За последние<blockquote>▸ сутки: <code>%d</code> 
▸ 7 дней: <code>%d</code>
▸ 30 дней: <code>%d</code></blockquote>
%s`,
		helpers.CustomEmoji(5316727448644103237, "👤"),
		name,
		status,
		getStatusEmoji(m.Status),
		helpers.Gendered(m.User.Gender, "зашел", "зашла"),
		helpers.FormatLastSeen(m.JoinedAt),
		newbieEmoji,
		leftAtInfo,
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
