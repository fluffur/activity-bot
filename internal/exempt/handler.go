package exempt

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Handler struct {
	service      *Service
	userService  *user.Service
	adminService *admin.Service
	dateParser   *DateParser
}

func NewHandler(service *Service, userService *user.Service, adminService *admin.Service, dateParser *DateParser) *Handler {
	return &Handler{service, userService, adminService, dateParser}
}

func (h *Handler) Set(b *gotgbot.Bot, ctx *ext.Context, cctx *command.Context) error {
	targetUser := cctx.Users[0]

	if len(cctx.Args) < 1 {
		_, err := ctx.EffectiveMessage.Reply(b, "Вы забыли указать срок реста, попробуйте написать +рест 2 недели в ответ пользователю", nil)

		return err
	}

	date, ok := h.dateParser.Parse(cctx.Args[0])
	if !ok {
		_, err := ctx.EffectiveMessage.Reply(b, "Не понял формат. Примеры:\n+рест 12.01\n+рест 2 недели\n+рест месяц", nil)

		return err
	}
	if date.Before(time.Now()) {
		_, err := ctx.EffectiveMessage.Reply(b, "Нельзя указывать прошедшую дату", nil)

		return err
	}

	if helpers.IsSenderAdmin(b, ctx, h.adminService) {
		return h.createExemptRequest(b, ctx, targetUser, date)
	}

	if err := h.service.ExemptMember(ctx.EffectiveChat.Id, targetUser.ID, date); err != nil {
		log.Println("Set", err)
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось создать рест", nil)
		return err
	}

	var text string
	if targetUser.ID == ctx.EffectiveUser.Id {
		text = fmt.Sprintf("Вы добавлены в рест до %s", helpers.FormatToHumanDate(date))
	} else {
		text = fmt.Sprintf("Пользователь %s добавлен в рест до %s", helpers.Link(*targetUser), helpers.FormatToHumanDate(date))
	}

	_, err := ctx.EffectiveMessage.Reply(b, text, &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
	})
	return err
}

func (h *Handler) createExemptRequest(b *gotgbot.Bot, ctx *ext.Context, targetUser *model.User, date time.Time) error {
	if err := h.service.CreateExemptRequest(ctx.EffectiveChat.Id, targetUser.ID, ctx.EffectiveMessage.MessageId, date); err != nil {
		log.Println("CreateExemptRequest", err)
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось создать заявку", nil)

		return err
	}

	kb := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "✅ Одобрить", CallbackData: fmt.Sprintf("approve:%d:%d", targetUser.ID, ctx.EffectiveMessage.MessageId)},
				{Text: "❌ Отклонить", CallbackData: fmt.Sprintf("reject:%d:%d", targetUser.ID, ctx.EffectiveMessage.MessageId)},
			},
		},
	}

	_, err := b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf(
		"Для пользователя %s запрошен рест до %s",
		helpers.Link(*targetUser),
		helpers.FormatToHumanDate(date),
	), &gotgbot.SendMessageOpts{
		ParseMode:   gotgbot.ParseModeHTML,
		ReplyMarkup: kb,
	})
	return err
}

func (h *Handler) Show(b *gotgbot.Bot, ctx *ext.Context, cctx *command.Context) error {
	targetUser := cctx.Users[0]

	exempt, err := h.service.GetMemberExempt(ctx.EffectiveChat.Id, targetUser.ID)
	if err != nil || exempt == nil {
		_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Пользователь %s не находится в ресте", helpers.Link(*targetUser)), &gotgbot.SendMessageOpts{
			ParseMode: gotgbot.ParseModeHTML,
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
		})

		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Пользователь %s находится в ресте до %s", helpers.Link(*targetUser), helpers.FormatToHumanDate(*exempt)), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})

	return err

}

func (h *Handler) End(b *gotgbot.Bot, ctx *ext.Context, cctx *command.Context) error {
	targetUser := cctx.Users[0]

	if targetUser.ID != ctx.EffectiveUser.Id && !helpers.IsSenderAdmin(b, ctx, h.adminService) {
		_, err := ctx.EffectiveMessage.Reply(b, "Вы можете удалить из реста только себя", nil)
		return err
	}

	exempt, err := h.service.GetMemberExempt(ctx.EffectiveChat.Id, targetUser.ID)
	if err != nil {
		log.Println("Get member exempt", err)
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось проверить рест пользователя", nil)
		return err
	}
	if exempt == nil {
		if targetUser.ID == ctx.EffectiveUser.Id {
			_, err := ctx.EffectiveMessage.Reply(b, "Вы не находитесь в ресте", nil)
			return err
		}

		_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Пользователь %s не находится в ресте", helpers.Link(*targetUser)), &gotgbot.SendMessageOpts{
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
			ParseMode: gotgbot.ParseModeHTML,
		})
		return err
	}

	if err := h.service.EndMemberExempt(ctx.EffectiveChat.Id, targetUser.ID); err != nil {
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось удалить пользователя из реста", nil)
		return err
	}

	if targetUser.ID == ctx.EffectiveUser.Id {
		_, err := ctx.EffectiveMessage.Reply(b, "Вы успешно удалены из реста", nil)
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Пользователь %s успешно удалён из реста", helpers.Link(*targetUser)), &gotgbot.SendMessageOpts{
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
		ParseMode: gotgbot.ParseModeHTML,
	})
	return err
}

func (h *Handler) ApproveExemptRequest(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !helpers.CheckOwnerOrAdmin(ctx, b, h.adminService, update.CallbackQuery.Message.Message.Chat.ID, update.CallbackQuery.From.ID) {
		helpers.AnswerCallback(ctx, b, update, "Кнопка доступна только создателю или пользователю с ролью администратор в боте")
		return
	}

	fromID, messageID, err := parseExemptRequestCallbackData(update)

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
	u, err := h.userService.GetUser(int64(fromID))
	if err != nil {
		log.Println(err)
		helpers.AnswerCallback(ctx, b, update, "Не удалось найти пользователя")
		return
	}

	helpers.EditMessage(ctx, b, update,
		fmt.Sprintf(`Запрос одобрен. У %s рест до %s`,
			helpers.Link(u),
			helpers.FormatToHumanDate(exemptRequest.ExemptUntil),
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

	fromID, _, err := parseExemptRequestCallbackData(update)
	u, err := h.userService.GetUser(int64(fromID))
	if err != nil {
		helpers.EditMessage(ctx, b, update, fmt.Sprintf("Запрос на рест отклонён"))
		return
	}

	helpers.EditMessage(ctx, b, update, fmt.Sprintf("Запрос на рест для %s отклонён", helpers.Link(u)))
}

func parseExemptRequestCallbackData(update *models.Update) (fromID int, messageID int, err error) {
	callbackData := update.CallbackQuery.Data

	parts := strings.SplitN(callbackData, ":", 3)
	if len(parts) != 3 {
		return
	}
	fromID, err = strconv.Atoi(parts[1])
	if err != nil {
		return
	}
	messageID, err = strconv.Atoi(parts[2])
	if err != nil {
		return
	}

	return
}
