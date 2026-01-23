package exempt

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"context"
	"fmt"
	"html"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Handler struct {
	service      *Service
	userService  *user.Service
	adminService *admin.Service
	dateParser   *DateParser
	setExemptRe  *regexp.Regexp
}

func NewHandler(service *Service, userService *user.Service, adminService *admin.Service, dateParser *DateParser, setExemptRe *regexp.Regexp) *Handler {
	return &Handler{service, userService, adminService, dateParser, setExemptRe}
}

func (h *Handler) ExemptMember(ctx context.Context, b *bot.Bot, update *models.Update) {
	senderMember, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
		ChatID: update.Message.Chat.ID,
		UserID: update.Message.From.ID,
	})
	if err != nil {
		helpers.AnswerMessage(ctx, b, update, "Не удалось проверить статус пользователя")
		return
	}

	text := update.Message.Text
	matches := h.setExemptRe.FindStringSubmatch(text)
	if len(matches) < 2 {
		helpers.AnswerMessage(ctx, b, update, "Неверный формат команды")
		return
	}
	args := strings.TrimSpace(matches[2])

	targetUserID, restArg, found, err := helpers.ExtractTargetUser(ctx, h.userService, update, args)
	if err != nil {
		helpers.AnswerMessage(ctx, b, update, "Не удалось найти пользователя")
		return
	}
	if !found {
		targetUserID = update.Message.From.ID
		restArg = args
	}

	date, ok := h.dateParser.Parse(restArg)
	if !ok {
		if restArg == "" {
			h.ShowMemberExempt(ctx, b, update)
			return
		}
		helpers.AnswerMessage(ctx, b, update,
			"Не понял формат. Примеры:\n+рест 12.01\n+рест 2 недели\n+рест месяц",
		)
		return
	}
	if date.Before(time.Now()) {
		helpers.AnswerMessage(ctx, b, update, "Нельзя указывать прошедшую дату")
		return
	}

	u, err := h.userService.GetUser(ctx, targetUserID)
	if err != nil {
		helpers.AnswerMessage(ctx, b, update, "Не удалось получить пользователя")
		return
	}

	isAdmin, err := h.adminService.IsAdmin(ctx, update.Message.Chat.ID, update.Message.From.ID)
	if err != nil {
		log.Println("Check admin error", err)
	}

	if senderMember.Owner == nil && !isAdmin {
		if err := h.createExemptRequest(ctx, b, update, targetUserID, u, date); err != nil {
			log.Println("Failed to create exempt request", err)
			helpers.AnswerMessage(ctx, b, update, "Не удалось создать заявку")
		}
		return
	}

	if err := h.service.ExemptMember(ctx, update.Message.Chat.ID, targetUserID, date); err != nil {
		helpers.AnswerMessage(ctx, b, update, "Не удалось создать рест")
		return
	}

	if targetUserID == update.Message.From.ID {
		helpers.AnswerMessage(ctx, b, update, fmt.Sprintf("Вы добавлены в рест до %s", date.Format("02.01.2006")))
	} else {
		helpers.AnswerMessage(ctx, b, update, fmt.Sprintf(`Пользователь <a href="tg://user?id=%d">%s</a> добавлен в рест до %s`, targetUserID, html.EscapeString(u.FirstName), date.Format("02.01.2006")))
	}
}

func (h *Handler) createExemptRequest(ctx context.Context, b *bot.Bot, update *models.Update, targetUserID int64, targetUser model.User, date time.Time) error {
	if err := h.service.CreateExemptRequest(ctx, update.Message.Chat.ID, targetUserID, int64(update.Message.ID), date); err != nil {
		return err
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "✅ Одобрить", CallbackData: fmt.Sprintf("approve:%d:%d", targetUserID, update.Message.ID)},
				{Text: "❌ Отклонить", CallbackData: fmt.Sprintf("reject:%d:%d", targetUserID, update.Message.ID)},
			},
		},
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: fmt.Sprintf(
			"Для пользователя <a href=\"tg://user?id=%d\">%s</a> запрошен рест до %s",
			targetUserID,
			html.EscapeString(targetUser.FirstName),
			date.Format("02.01.2006"),
		),
		ParseMode:   "HTML",
		ReplyMarkup: kb,
	})
	return err
}

func (h *Handler) ShowMemberExempt(ctx context.Context, b *bot.Bot, update *models.Update) {
	targetUserID, _, found, err := helpers.ExtractTargetUser(ctx, h.userService, update, "")
	if err != nil {
		helpers.AnswerMessage(ctx, b, update, "Не удалось найти пользователя")
		return
	}
	if !found {
		targetUserID = update.Message.From.ID
	}

	u, err := h.userService.GetUser(ctx, targetUserID)
	if err != nil {
		helpers.AnswerMessage(ctx, b, update, "Не удалось получить пользователя для информации о ресте")
		return
	}

	exempt, err := h.service.GetMemberExempt(ctx, update.Message.Chat.ID, targetUserID)
	if err != nil || exempt == nil {
		helpers.AnswerMessage(ctx, b, update, "Пользователь не находится в ресте")
		return
	}

	helpers.AnswerMessage(ctx, b, update,
		fmt.Sprintf(`Пользователь <a href="tg://user?id=%d">%s</a> находится в ресте до %s`,
			targetUserID,
			html.EscapeString(u.FirstName),
			exempt.Format("02.01.2006"),
		),
	)
}

func (h *Handler) EndMemberExempt(ctx context.Context, b *bot.Bot, update *models.Update) {
	targetUserID, _, found, err := helpers.ExtractTargetUser(ctx, h.userService, update, "")
	if err != nil {
		helpers.AnswerMessage(ctx, b, update, "Не удалось найти пользователя")
		return
	}
	if !found {
		targetUserID = update.Message.From.ID
	}

	member, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
		ChatID: update.Message.Chat.ID,
		UserID: update.Message.From.ID,
	})
	if err != nil {
		helpers.AnswerMessage(ctx, b, update, "Не удалось проверить вашу роль в чате")
		return
	}

	if targetUserID != update.Message.From.ID && member.Owner == nil {
		helpers.AnswerMessage(ctx, b, update, "Вы можете удалить из реста только себя")
		return
	}

	u, err := h.userService.GetUser(ctx, targetUserID)
	if err != nil {
		helpers.AnswerMessage(ctx, b, update, "Не удалось получить пользователя")
		return
	}

	exempt, err := h.service.GetMemberExempt(ctx, update.Message.Chat.ID, targetUserID)

	if err != nil {
		log.Println("Get member exempt", err)
		helpers.AnswerMessage(ctx, b, update, "Ошибка, не удалось проверить рест пользователя")
		return
	}
	if exempt == nil {
		if targetUserID == update.Message.From.ID {
			helpers.AnswerMessage(ctx, b, update, "У вас нет реста")
		} else {
			helpers.AnswerMessage(ctx, b, update, "У пользователя нет реста")

		}
		return
	}

	if err := h.service.EndMemberExempt(ctx, update.Message.Chat.ID, targetUserID); err != nil {
		helpers.AnswerMessage(ctx, b, update, "Не удалось удалить пользователя из реста")
		return
	}

	if targetUserID == update.Message.From.ID {
		helpers.AnswerMessage(ctx, b, update, "Вы удалены из реста")
	} else {
		helpers.AnswerMessage(ctx, b, update,
			fmt.Sprintf(`Пользователь <a href="tg://user?id=%d">%s</a> удалён из реста`,
				targetUserID,
				html.EscapeString(u.FirstName),
			),
		)
	}
}

func (h *Handler) ApproveExemptRequest(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !helpers.CheckOwnerOrAdmin(ctx, b, h.adminService, update.CallbackQuery.Message.Message.Chat.ID, update.CallbackQuery.From.ID) {
		helpers.AnswerCallback(ctx, b, update, "Кнопка доступна только создателю или пользователю с ролью администратор в боте")
		return
	}

	callbackData := update.CallbackQuery.Data
	parts := strings.SplitN(callbackData, ":", 3)
	if len(parts) != 3 {
		helpers.AnswerCallback(ctx, b, update, "Некорректная структура запроса")
		return
	}
	fromID, err := strconv.Atoi(parts[1])
	if err != nil {
		helpers.AnswerCallback(ctx, b, update, "Некорректный запрос, не найден айди пользователя")
		return
	}
	messageID, err := strconv.Atoi(parts[2])
	if err != nil {
		helpers.AnswerCallback(ctx, b, update, "Некорректный запрос, не найден айди сообщения")
		return
	}

	exemptRequest, err := h.service.GetExemptRequest(ctx, update.CallbackQuery.Message.Message.Chat.ID, fromID, messageID)
	if err != nil {
		log.Println("Exempt request not found", err)
		helpers.AnswerCallback(ctx, b, update, "Не найден запрос на рест")
		return
	}

	if err := h.service.ApproveExemptRequest(ctx, update.CallbackQuery.Message.Message.Chat.ID, int64(fromID), int64(messageID), exemptRequest.ExemptUntil); err != nil {
		log.Println("Failed to approve exempt request", err)
		helpers.AnswerCallback(ctx, b, update, "Не удалось одобрить запрос")
		return
	}
	u, err := h.userService.GetUser(ctx, int64(fromID))
	if err != nil {
		log.Println(err)
		helpers.AnswerCallback(ctx, b, update, "Не удалось найти пользователя")
		return
	}

	helpers.EditMessage(ctx, b, update,
		fmt.Sprintf(`Запрос одобрен. У <a href="tg://user?id=%d">%s</a> рест до %s`,
			fromID,
			html.EscapeString(u.FirstName),
			exemptRequest.ExemptUntil.Format("02.01.2006"),
		),
	)

}

func (h *Handler) RejectExemptRequest(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !helpers.CheckOwnerOrAdmin(ctx, b, h.adminService, update.CallbackQuery.Message.Message.Chat.ID, update.CallbackQuery.From.ID) {
		helpers.AnswerCallback(ctx, b, update, "Кнопка доступна только создателю или пользователю с ролью администратор в боте")
		return
	}

	if err := h.service.RejectExemptRequest(ctx, update.CallbackQuery.Message.Message.Chat.ID, update.CallbackQuery.From.ID, update.CallbackQuery.Message.Message.ID); err != nil {
		helpers.AnswerCallback(ctx, b, update, "Не удалось отклонить запрос")
		return
	}

	helpers.EditMessage(ctx, b, update, fmt.Sprintf("Запрос на рест для  <a href=\"tg://user?id=%d\">%s</a> отклонён", update.CallbackQuery.Message.Message.From.ID, update.CallbackQuery.Message.Message.From.FirstName))
}
