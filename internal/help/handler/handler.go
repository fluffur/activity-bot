package handler

import (
	"activity-bot/internal/cmd"
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

func (h *Handler) Start(b *gotgbot.Bot, ctx *ext.Context, _ *cmd.Context) error {
	_, err := ctx.EffectiveMessage.Reply(b, "Привет! Я могу следить за еженедельной нормой сообщений в группе. Добавь меня в группу или введи команду /help.", nil)

	return err
}

func (h *Handler) Help(b *gotgbot.Bot, ctx *ext.Context, _ *cmd.Context) error {
	helpText := fmt.Sprintf(`
📌 <b>Команды бота</b>

📊 <b>Активность</b>
/отчёт — недельный отчёт по активности
/норма — показать норму сообщений
/норма <i>число</i> — установить норму (только админы)

🛌 <b>Рест</b>
/рест — показать текущий рест
/рест <i>период</i> — поставить рест (например: неделя, 2 недели, месяц)
/-рест — завершить рест

🎭 <b>Роли</b>
/роль — показать свою роль
/роль @user — показать роль пользователя
/роль @user <i>название</i> — установить роль (админы)
/-роль @user — удалить роль
/роли — список всех ролей в чате

👮 <b>Администрация</b>
/админы — список администраторов бота
/админ @user — добавить администратора
/-админ @user — удалить администратора (только создатель)
/обновить чат — синхронизировать участников и роли

📞 <b>Прочее</b>
/call <i>текст</i> — вызвать участников (админы)

💡 Команды поддерживают префиксы: <code>/ ! . +</code>

📬 По вопросам и багам — %s
`, helpers.Mention(h.ownerID, "напишите разработчику"))

	_, err := ctx.EffectiveMessage.Reply(b, helpText, &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
	})

	return err
}
