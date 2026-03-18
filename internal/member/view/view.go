package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"html"
	"strings"
)

func FormatRolesList(members []model.ChatMember) string {
	var sb strings.Builder
	sb.WriteString("🎭 Роли всех участников:\n\n")
	sb.WriteString("<blockquote expandable>")

	for i, m := range members {
		sb.WriteString(fmt.Sprintf("%d. %s — %s\n", i+1, helpers.RoleLink(m), helpers.UserLink(m.User)))
	}
	sb.WriteString("</blockquote>")
	sb.WriteString("\nЧтобы изменить роль участника введите <code>!роль @участник название</code>")

	return sb.String()
}

func FormatRoleUpdated(user model.ChatMember, role string) string {
	return fmt.Sprintf("Роль %s обновлена на \"%s\"", helpers.RoleLink(user), html.EscapeString(role))
}

func FormatMemberRole(user model.User, title string) string {
	return fmt.Sprintf("Роль участника %s — %s", helpers.UserLink(user), html.EscapeString(title))
}

func FormatSyncResult(count int) string {
	return fmt.Sprintf("Чат обновлён. Найдено %d участников", count)
}
