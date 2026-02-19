package view

import (
	"activity-bot/internal/helpers"
	"fmt"
)

func FormatHelpText(ownerID int64) string {
	return fmt.Sprintf(`
📌 <b>Справка по боту</b>

📋 Все доступные команды можно открыть через кнопку <b>"Команды бота"</b>.

💬 Если нашли баг или есть предложения — %s
`, helpers.Mention(ownerID, "напишите разработчику"))
}

func FormatStartMessage() string {
	return `👋 Привет!

Я чат-менеджер. Считаю сообщения и помогаю контролировать еженедельную активность

Добавь меня в группу или открой список команд ниже 👇`
}
