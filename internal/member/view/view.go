package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"html"

	"github.com/gotd/td/telegram/message/entity"
)

func FormatRolesList(members []model.ChatMember) string {
	eb := &entity.Builder{}
	WriteRolesList(eb, members)
	res, _ := eb.Complete()
	return res
}

func FormatRoleUpdated(user model.ChatMember, role string) string {
	eb := &entity.Builder{}
	WriteRoleUpdated(eb, user, role)
	res, _ := eb.Complete()
	return res
}

func WriteRolesList(eb *entity.Builder, members []model.ChatMember) {
	eb.Plain("🎭 Роли всех участников:\n\n")
	for i, m := range members {
		eb.Plain(fmt.Sprintf("%d. ", i+1))
		helpers.WriteRoleEmojiLink(eb, m)
		eb.Plain(" @")
		eb.Plain(m.User.DisplayName())
		eb.Plain("\n")
	}
	eb.Plain("\nЧтобы изменить роль участника введите ")
	eb.Code("!роль @участник название")
}

func WriteRoleUpdated(eb *entity.Builder, user model.ChatMember, role string) {
	eb.Plain("Роль ")
	helpers.WriteRoleEmojiLink(eb, user)
	eb.Plain(" обновлена на \"")
	eb.Plain(html.EscapeString(role))
	eb.Plain("\"")
}

func FormatMemberRole(title string) string {
	return fmt.Sprintf("Роль участника: %s", title)
}

func FormatSyncResult(count int) string {
	return fmt.Sprintf("Чат обновлён. Найдено %d участников", count)
}
