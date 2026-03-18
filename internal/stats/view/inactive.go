package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"strings"
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
			helpers.RoleLink(m.Member),
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
