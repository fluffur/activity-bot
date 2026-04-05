package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"time"

	"github.com/gotd/td/telegram/message/entity"
)

func WriteProfile(eb *entity.Builder, m model.ChatMemberStats, full bool) {
	now := time.Now().UTC()

	isNewbie := false
	if m.Chat.NewbieThresholdDays > 0 {
		newbieUntil := m.ChatMember.JoinedAt.AddDate(0, 0, int(m.Chat.NewbieThresholdDays))
		if newbieUntil.After(now) {
			isNewbie = true
		}
	}

	status := m.ChatMember.Status.String()

	helpers.WriteCustomEmoji(eb, "5316727448644103237", "👤")
	eb.Plain(" Информация о ")
	helpers.WriteRoleEmojiLink(eb, m.ChatMember)
	eb.Plain("\n")

	eb.Plain("• Ранг ")
	helpers.WriteStatusEmoji(eb, m.ChatMember.Status)
	eb.Plain("– ")
	eb.Plain(status)
	eb.Plain("\n")

	if m.ChatMember.LeftAt.IsZero() {
		eb.Plain(fmt.Sprintf("• В чате c %s (%s)",
			helpers.FormatToHumanDateTime(m.ChatMember.JoinedAt),
			helpers.FormatLastSeenPlain(m.ChatMember.JoinedAt),
		))
	} else {
		eb.Plain(fmt.Sprintf("❌ %s чат (%s — %s)",
			helpers.Gendered(m.ChatMember.User.Gender, "Покинул", "Покинула"),
			helpers.FormatToHumanDateTime(m.ChatMember.JoinedAt),
			helpers.FormatToHumanDateTime(m.ChatMember.LeftAt),
		))
	}

	if full {
		eb.Plain("\n\n")

		helpers.WriteStatsEmoji(eb)
		eb.Plain(" Актив\n")

		token := eb.Token()
		eb.Plain(fmt.Sprintf(
			"▸ сегодня: %d\n▸ эта неделя: %d\n▸ этот месяц: %d\n▸ всего: %d\n",
			m.DayCount,
			m.WeekCount,
			m.MonthCount,
			m.AllTime,
		))
		token.Apply(eb, entity.Blockquote(true))

		eb.Plain("\n")
		helpers.WriteCustomEmoji(eb, "5258419835922030550", "⏰")
		eb.Plain(" За последние\n")

		token = eb.Token()
		eb.Plain(fmt.Sprintf(
			"▸ сутки: %d\n▸ 7 дней: %d\n▸ 30 дней: %d\n",
			m.DayRollingCount,
			m.WeekRollingCount,
			m.MonthRollingCount,
		))
		token.Apply(eb, entity.Blockquote(true))
	}

	isRestActive := m.ChatMember.RestUntil.After(time.Now())

	if m.Chat.NormWarn > 0 || m.Chat.NormBan > 0 {
		eb.Plain("\n")
		writeNormStatus(eb, m.WeekCount, int64(m.Chat.NormWarn), int64(m.Chat.NormBan), isRestActive, isNewbie)
	}

	daysSinceRestEnd := int(now.Sub(m.ChatMember.RestUntil).Hours() / 24)

	if isRestActive && m.ChatMember.LeftAt.IsZero() {
		eb.Plain("\n")
		token := eb.Token()
		helpers.WriteRestEmoji(eb)
		eb.Plain(" Рест до ")
		eb.Plain(helpers.FormatToHumanDateTime(m.ChatMember.RestUntil))
		token.Apply(eb, entity.Blockquote(true))

	} else if daysSinceRestEnd >= 0 && daysSinceRestEnd <= 3 {
		eb.Plain("\n")
		token := eb.Token()
		helpers.WriteRestEmoji(eb)
		eb.Plain(" Последний рест был завершен ")
		eb.Plain(helpers.FormatToHumanDateTime(m.ChatMember.RestUntil))
		token.Apply(eb, entity.Blockquote(true))
	}
}

func writeNormStatus(eb *entity.Builder, weekCount, normWarn, normBan int64, isRestActive, isNewbie bool) {
	if isRestActive || isNewbie {
		helpers.WriteCustomEmoji(eb, "5456648248968121823", "🕊")
		eb.Plain(" Освобождение от нормы")
		return
	}
	eb.Plain("\n")

	if normBan > 0 && weekCount < normBan {
		helpers.WriteCustomEmoji(eb, "5260342697075416641", "🚫")
		eb.Plain(fmt.Sprintf(" Норма не набрана (%d/%d), ", weekCount, normBan))
		eb.Bold("бан")
		return
	}

	if normWarn > 0 && weekCount < normWarn {
		helpers.WriteCustomEmoji(eb, "5258474669769497337", "⚠️")
		eb.Plain(fmt.Sprintf(" Норма не набрана (%d/%d), ", weekCount, normWarn))
		eb.Bold("варн")
		return
	}

	helpers.WriteCustomEmoji(eb, "5260416304224936047", "✅")
	eb.Plain(" Норма набрана")
}
