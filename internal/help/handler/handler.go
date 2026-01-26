package handler

import (
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	ownerID int64
}

func New(ownerID int64) *Handler {
	return &Handler{ownerID}
}

func (h *Handler) Start(b *gotgbot.Bot, ctx *ext.Context, _ *command.Context) error {
	_, err := ctx.EffectiveMessage.Reply(b, "Привет! Я могу следить за еженедельной нормой сообщений в группе. Добавь меня в группу или введи команду /help.", nil)

	return err
}

func (h *Handler) Help(b *gotgbot.Bot, ctx *ext.Context, _ *command.Context) error {
	helpText := fmt.Sprintf(`
📌 <b>Команды бота</b>

💬 <b>Норма чата</b>
/норма — показать текущую норму сообщений
/норма (число) — установить новую норму (например: /норма 50)

🛌 <b>Рест</b>
/рест — показать текущий рест
/рест (период) — поставить рест (например: /рест неделя, /рест 2 недели, /рест месяц)
/-рест — завершить рест досрочно

🛡 <b>Роли и администрация</b>
/роль — показать вашу роль
/роль @user — показать роль пользователя
/роль @user (название) — установить роль (только для админов)
/роли — список всех участников с ролями
/админы — список администраторов
/админ @user — добавить администратора
/-админ @user — удалить администратора
/обновить чат — синхронизировать участников и титулы

📊 <b>Отчёты</b>
/отчёт — получить еженедельный отчёт по активности

❓ <b>Вопросы</b>
<i>Как считаются сообщения?</i>
Сообщения считаются один к одному, только среди пользователей, не находящихся в ресте.

💡 Команды поддерживают префиксы !, /, ., + (например: !рест, /норма, .роль)

📬 По техническим вопросам, багам и предложениям пишите %s
`, helpers.Mention(h.ownerID, "в личные сообщения разработчику"))

	_, err := ctx.EffectiveMessage.Reply(b, helpText, &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
	})

	return err
}
