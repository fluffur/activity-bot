package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"activity-bot/internal/rest"
	"activity-bot/internal/user"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	service      *rest.Service
	userService  *user.Service
	adminService *admin.Service
	dateParser   *rest.DateParser
}

func New(service *rest.Service, userService *user.Service, adminService *admin.Service, dateParser *rest.DateParser) *Handler {
	return &Handler{service, userService, adminService, dateParser}
}

func (h *Handler) Set(b *gotgbot.Bot, ctx *ext.Context, cctx *cmd.Context) error {
	targetUser := cctx.FirstUser()
	if targetUser == nil {
		slog.Error("Failed to find target user in Set SetRest")
		return nil
	}

	firstArgument := cctx.FirstArgument()

	if firstArgument == "" {
		_, err := ctx.EffectiveMessage.Reply(b, "Вы забыли указать срок реста, попробуйте написать +рест 2 недели в ответ пользователю", nil)

		return err
	}

	date, ok := h.dateParser.Parse(firstArgument)
	if !ok {
		_, err := ctx.EffectiveMessage.Reply(b, "Не понял формат. Примеры:\n+рест 12.01\n+рест 2 недели\n+рест месяц", nil)

		return err
	}
	if date.Before(time.Now()) {
		_, err := ctx.EffectiveMessage.Reply(b, "Нельзя указывать прошедшую дату", nil)

		return err
	}

	if !h.adminService.CheckIsAdmin(ctx.EffectiveChat.Id, ctx.EffectiveSender.Id()) {
		return h.createRequest(b, ctx, targetUser, date)
	}

	if err := h.service.SetMemberRest(ctx.EffectiveChat.Id, targetUser.ID, date); err != nil {
		slog.Error("failed to add member to rest", "chat_id", ctx.EffectiveChat.Id, "user_id", targetUser.ID, "error", err)
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

func (h *Handler) createRequest(b *gotgbot.Bot, ctx *ext.Context, targetUser *model.User, date time.Time) error {

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

	slog.Info("rest requested", "message_id", msg.MessageId)
	if err := h.service.CreateRestRequest(ctx.EffectiveChat.Id, targetUser.ID, msg.MessageId, date); err != nil {
		slog.Error("failed to create rest request", "chat_id", ctx.EffectiveChat.Id, "user_id", targetUser.ID, "message_id", msg.MessageId, "error", err)
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось создать заявку", nil)

		return err
	}

	return err
}

func (h *Handler) Show(b *gotgbot.Bot, ctx *ext.Context, cctx *cmd.Context) error {
	targetUser := cctx.FirstUser()

	if targetUser == nil {
		slog.Error("Failed to find target user in Show SetRest")
		return nil
	}

	e, err := h.service.GetMemberRest(ctx.EffectiveChat.Id, targetUser.ID)
	if err != nil || e == nil {
		_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Пользователь %s не находится в ресте", helpers.Link(*targetUser)), &gotgbot.SendMessageOpts{
			ParseMode: gotgbot.ParseModeHTML,
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
		})

		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Пользователь %s находится в ресте до %s", helpers.Link(*targetUser), helpers.FormatToHumanDate(*e)), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})

	return err

}

func (h *Handler) End(b *gotgbot.Bot, ctx *ext.Context, cctx *cmd.Context) error {
	targetUser := cctx.FirstUser()

	if targetUser == nil {
		slog.Error("failed to find target user in End SetRest")
		return nil
	}

	if targetUser.ID != ctx.EffectiveUser.Id && !h.adminService.CheckIsAdmin(ctx.EffectiveChat.Id, ctx.EffectiveSender.Id()) {
		_, err := ctx.EffectiveMessage.Reply(b, "Вы можете удалить из реста только себя", nil)
		return err
	}

	e, err := h.service.GetMemberRest(ctx.EffectiveChat.Id, targetUser.ID)
	if err != nil {
		slog.Error("failed to check member rest status", "chat_id", ctx.EffectiveChat.Id, "user_id", targetUser.ID, "error", err)
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось проверить рест пользователя", nil)
		return err
	}
	if e == nil {
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

	if err := h.service.EndMemberRest(ctx.EffectiveChat.Id, targetUser.ID); err != nil {
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

func (h *Handler) ApproveRestRequest(b *gotgbot.Bot, ctx *ext.Context) error {
	fromID, err := parseRequestCallbackData(ctx.CallbackQuery.Data)
	restRequest, err := h.service.GetRestRequest(ctx.EffectiveChat.Id, fromID, ctx.EffectiveMessage.MessageId)
	if err != nil {
		slog.Error("rest request not found during approval", "chat_id", ctx.EffectiveChat.Id, "user_id", fromID, "message_id", ctx.EffectiveMessage.MessageId, "error", err)
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Не найден запрос на рест",
		})
		return err
	}

	if !h.adminService.CheckIsAdmin(ctx.EffectiveChat.Id, ctx.EffectiveSender.Id()) {
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Подтвердить запрос может только администратор",
		})
		return err

	}

	if err := h.service.ApproveRestRequest(ctx.EffectiveChat.Id, fromID, ctx.EffectiveMessage.MessageId, restRequest.RestUntil); err != nil {
		slog.Error("failed to approve rest request", "chat_id", ctx.EffectiveChat.Id, "user_id", fromID, "message_id", ctx.EffectiveMessage.MessageId, "error", err)
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Не удалось одобрить запрос",
		})
		return err
	}
	u, err := h.userService.GetUser(fromID)
	if err != nil {
		slog.Error("failed to get user during rest approval", "user_id", fromID, "error", err)
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Не удалось найти пользователя",
		})
		return err
	}

	_, _, err = b.EditMessageText(fmt.Sprintf(`Запрос одобрен. У %s рест до %s`,
		helpers.Link(u),
		helpers.FormatToHumanDate(restRequest.RestUntil),
	), &gotgbot.EditMessageTextOpts{
		ChatId:    ctx.EffectiveChat.Id,
		MessageId: ctx.EffectiveMessage.MessageId,
		ParseMode: gotgbot.ParseModeHTML,
	})

	return err
}

func (h *Handler) RejectRestRequest(b *gotgbot.Bot, ctx *ext.Context) error {
	fromID, err := parseRequestCallbackData(ctx.CallbackQuery.Data)
	restRequest, err := h.service.GetRestRequest(ctx.EffectiveChat.Id, fromID, ctx.EffectiveMessage.MessageId)
	if err != nil {
		slog.Error("rest request not found during rejection", "chat_id", ctx.EffectiveChat.Id, "user_id", fromID, "message_id", ctx.EffectiveMessage.MessageId, "error", err)
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Не найден запрос на рест",
		})
		return err
	}

	if restRequest.UserID != ctx.EffectiveSender.Id() && !h.adminService.CheckIsAdmin(ctx.EffectiveChat.Id, ctx.EffectiveSender.Id()) {
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Отклонить запрос может только администратор или заявитель реста",
		})
		return err

	}
	slog.Info("rejecting rest request", "message_id", ctx.EffectiveMessage.MessageId)
	if err := h.service.RejectRestRequest(ctx.EffectiveChat.Id, ctx.EffectiveSender.Id(), ctx.EffectiveMessage.MessageId); err != nil {
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Не удалось отклонить запрос",
		})
		return err
	}

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

func parseRequestCallbackData(callbackData string) (int64, error) {
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
