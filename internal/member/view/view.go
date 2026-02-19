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
	sb.WriteString("🎭 Роли всех участников:\n")
	for i, m := range members {
		sb.WriteString(fmt.Sprintf("\n%d. %s — %s", i+1, helpers.Link(m.User), html.EscapeString(m.CustomTitle)))
	}
	return sb.String()
}

func FormatRoleUpdated(user model.User, role string) string {
	return fmt.Sprintf("Роль участника %s обновлена на \"%s\"", helpers.Link(user), html.EscapeString(role))
}

func FormatMemberRole(user model.User, title string) string {
	return fmt.Sprintf("Роль участника %s — %s", helpers.Link(user), html.EscapeString(title))
}

func FormatSyncResult(count int) string {
	return fmt.Sprintf("Чат обновлён. Найдено %d участников", count)
}
