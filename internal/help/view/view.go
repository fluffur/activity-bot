package view

import (
	"activity-bot/internal/helpers"
	"fmt"
)

func FormatHelpText(ownerID int64) string {
	return fmt.Sprintf(`
📌 <b>Помощь по боту</b>

Все команды доступны через кнопку "Команды бота" в стартовом сообщении.

📬 По вопросам и багам — %s
`, helpers.Mention(ownerID, "напишите разработчику"))
}

func FormatStartMessage() string {
	return "Привет! Я могу следить за еженедельной нормой сообщений в группе. " +
		"Добавь меня в группу или воспользуйся кнопкой ниже, чтобы увидеть команды."
}
