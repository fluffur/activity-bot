package view

import (
	"activity-bot/internal/helpers"
	"fmt"
)

func FormatHelpText(ownerUsername, commandsLink string) string {
	return fmt.Sprintf(`
📋 %s

💬 %s
`, helpers.AnyLink(commandsLink, "Посмотреть список команд"), helpers.Link(ownerUsername, "Связаться с разработчиком"))
}

func FormatStartMessage(commandsLink string) string {
	return fmt.Sprintf(`👋 Привет!

Я чат-менеджер. Считаю сообщения и помогаю контролировать еженедельную активность

Добавь меня в группу или %s`, helpers.AnyLink(commandsLink, "открой список команд"))
}
