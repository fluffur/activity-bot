package exempt

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
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

	if !helpers.IsSenderAdmin(b, ctx, h.adminService) {
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

	kb := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "✅ Одобрить", CallbackData: fmt.Sprintf("approve:%d", targetUser.ID)},
				{Text: "❌ Отклонить", CallbackData: fmt.Sprintf("reject:%d", targetUser.ID)},
			},
		},
	}

	msg, err := b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf(
		"Для пользователя %s запрошен рест до %s",
		helpers.Link(*targetUser),
		helpers.FormatToHumanDate(date),
	), &gotgbot.SendMessageOpts{
		ParseMode:   gotgbot.ParseModeHTML,
		ReplyMarkup: kb,
	})
	if err != nil {
		return err
	}

	log.Println(msg.MessageId)
	if err := h.service.CreateExemptRequest(ctx.EffectiveChat.Id, targetUser.ID, msg.MessageId, date); err != nil {
		log.Println("CreateExemptRequest", err)
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось создать заявку", nil)

		return err
	}

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

func (h *Handler) ApproveExemptRequest(b *gotgbot.Bot, ctx *ext.Context) error {
	fromID, err := parseExemptRequestCallbackData(ctx.CallbackQuery.Data)
	exemptRequest, err := h.service.GetExemptRequest(ctx.EffectiveChat.Id, fromID, ctx.EffectiveMessage.MessageId)
	if err != nil {
		log.Println("Exempt request not found", err)
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Не найден запрос на рест",
		})
		return err
	}

	if err := h.service.ApproveExemptRequest(ctx.EffectiveChat.Id, fromID, ctx.EffectiveMessage.MessageId, exemptRequest.ExemptUntil); err != nil {
		log.Println("Failed to approve exempt request", err)
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Не удалось одобрить запрос",
		})
		return err
	}
	u, err := h.userService.GetUser(fromID)
	if err != nil {
		log.Println("ApproveExemptRequest GetUser", err)
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Не удалось найти пользователя",
		})
		return err
	}

	_, _, err = b.EditMessageText(fmt.Sprintf(`Запрос одобрен. У %s рест до %s`,
		helpers.Link(u),
		helpers.FormatToHumanDate(exemptRequest.ExemptUntil),
	), &gotgbot.EditMessageTextOpts{
		ChatId:    ctx.EffectiveChat.Id,
		MessageId: ctx.EffectiveMessage.MessageId,
		ParseMode: gotgbot.ParseModeHTML,
	})

	return err
}

func (h *Handler) RejectExemptRequest(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Println(ctx.EffectiveMessage.MessageId)
	if err := h.service.RejectExemptRequest(ctx.EffectiveChat.Id, ctx.EffectiveSender.Id(), ctx.EffectiveMessage.MessageId); err != nil {
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Не удалось отклонить запрос",
		})
		return err
	}

	fromID, err := parseExemptRequestCallbackData(ctx.CallbackQuery.Data)
	u, err := h.userService.GetUser(fromID)
	if err != nil {
		_, _, err = b.EditMessageText("Запрос на рест отклонён",
			&gotgbot.EditMessageTextOpts{
				ChatId:    ctx.EffectiveChat.Id,
				MessageId: ctx.EffectiveMessage.MessageId,
				ParseMode: gotgbot.ParseModeHTML,
			},
		)

		return err
	}

	_, _, err = b.EditMessageText(fmt.Sprintf("Запрос на рест для %s отклонён", helpers.Link(u)),
		&gotgbot.EditMessageTextOpts{
			ChatId:    ctx.EffectiveChat.Id,
			MessageId: ctx.EffectiveMessage.MessageId,
			ParseMode: gotgbot.ParseModeHTML,
		},
	)

	return err
}

func parseExemptRequestCallbackData(callbackData string) (int64, error) {
	parts := strings.SplitN(callbackData, ":", 2)
	if len(parts) != 2 {
		return 0, errors.New("invalid callback data")
	}
	fromID, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	return int64(fromID), nil
}
