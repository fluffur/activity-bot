package help

import (
	"activity-bot/internal/base"
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Handler struct {
	base.Handler
}

func NewHandler() *Handler {
	return &Handler{base.Handler{}}
}

func (h *Handler) Start(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.AnswerMessage(ctx, b, update, "Привет! Я могу следить за еженедельной нормой сообщений в группе. Добавь меня в группу или введи команду /help.")
}

func (h *Handler) Help(ctx context.Context, b *bot.Bot, update *models.Update) {
	helpText := `📌 Команды бота для работы с нормой сообщений и рестами

Норма чата:
• /норма — показать текущую норму сообщений в группе
• /норма (число) — установить новую норму сообщений (например: /норма 50)

Рест (освобождение от нормы):
• /рест — показать ваш текущий рест
• /рест период — поставить себе рест (например: /рест неделя, +рест 2 недели, !рест месяц)
• -рест — завершить рест раньше времени

Отчёты:
• /отчёт — получить еженедельный отчёт по норме сообщений в группе

💡 Примеры использования:
• +рест 2 недели
• !рест неделя
• /рест месяц
• -рест

Надеюсь, это поможет следить за активностью группы и вовремя отдыхать от нормы! 🟢`

	h.AnswerMessage(ctx, b, update, helpText)
}
