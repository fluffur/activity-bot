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
		sb.WriteString(fmt.Sprintf(
			"%d. %s (%s)",
			i+1,
			helpers.LinkWithContent(
				m.Member.User,
				fmt.Sprintf("%s", m.Member.User.FirstName),
			),
			m.Member.CustomTitle,
		))

		if m.LastActivity != nil {
			sb.WriteString(fmt.Sprintf(
				" — %s (%s)",
				helpers.FormatToHumanDate(*m.LastActivity),
				helpers.FormatLastSeen(*m.LastActivity),
			))
		} else {
			sb.WriteString(" — не писал ни разу")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
