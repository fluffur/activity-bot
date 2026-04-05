package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"strings"

	"github.com/gotd/td/telegram/message/entity"
)

func FormatInactiveMembers(members []model.InactiveMember) string {
	var sb strings.Builder
	sb.WriteString("<b>😴 Неактивные участники (более 1 суток)</b>\n\n")

	for i, m := range members {
		userTitle := m.Member.Tag
		if userTitle == "" {
			userTitle = m.Member.User.FirstName
		}
		sb.WriteString(fmt.Sprintf(
			"%d. %s",
			i+1,
			helpers.RoleEmojiLink(m.Member),
		))

		if !m.LastActivity.IsZero() {
			sb.WriteString(fmt.Sprintf(
				" — %s",
				helpers.FormatLastSeen(m.LastActivity),
			))
		} else {
			sb.WriteString(" — не писал ни разу")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

func WriteInactiveMembers(eb *entity.Builder, members []model.InactiveMember) {
	eb.Bold("😴 Неактивные участники (более 1 суток)\n\n")

	for i, m := range members {
		userTitle := m.Member.Tag
		if userTitle == "" {
			userTitle = m.Member.User.FirstName
		}
		eb.Plain(fmt.Sprintf("%d. ", i+1))
		helpers.WriteRoleEmojiLink(eb, m.Member)
		eb.Plain(" — ")

		if !m.LastActivity.IsZero() {
			eb.FormattedDate("date", true, false, false, false, false, false, int(m.LastActivity.Unix()))
		} else {
			eb.Plain("не писал ни разу")
		}

		eb.Plain("\n")
	}
}
