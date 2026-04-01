package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"activity-bot/internal/rest"
	"activity-bot/internal/rest/view"
	"activity-bot/internal/session"
	"activity-bot/internal/user"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/hibiken/asynq"
)

type Handler struct {
	service        *rest.Service
	userService    *user.Service
	memberService  *member.Service
	chatService    *chat.Service
	adminService   *admin.Service
	dateParser     *helpers.DateParser
	sessionService *session.Service
	asyncClient    *asynq.Client
}

func New(service *rest.Service, userService *user.Service, memberService *member.Service, chatService *chat.Service, adminService *admin.Service, dateParser *helpers.DateParser, sessionService *session.Service, asyncClient *asynq.Client) *Handler {
	return &Handler{service, userService, memberService, chatService, adminService, dateParser, sessionService, asyncClient}
}

func (h *Handler) SetRest(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	u, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	sender, err := ctx.Sender()
	if err != nil {
		return err
	}
	date, err := ctx.Date()
	if err != nil {
		return err
	}
	reason := ctx.TextOrDefault("")
	if sender.Status < ctx.RequiredStatus() {
		return h.createRequest(b, ctx, u, date, reason)
	}

	if err := h.service.SetMemberRestWithHistory(ctx.StdContext(), c.ID, u.User.ID, ctx.EffectiveMessage.MessageId, date, reason); err != nil {
		_ = ctx.Reply(b, "Не удалось создать рест", nil)
		return err
	}

	text := view.FormatRestSet(*u, date, reason)
	return ctx.ReplyHTML(b, text)
}

func (h *Handler) createRequest(b *gotgbot.Bot, ctx *command.Context, u *model.ChatMember, date time.Time, reason string) error {

	kb := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "Одобрить", CallbackData: fmt.Sprintf("approve:%d", u.User.ID), Style: "success", IconCustomEmojiId: helpers.SuccessEmojiGray},
				{Text: "Отклонить", CallbackData: fmt.Sprintf("reject:%d", u.User.ID), Style: "danger", IconCustomEmojiId: helpers.DangerEmojiGray},
			},
		},
	}

	msg, err := ctx.EffectiveMessage.Reply(b, view.FormatRestRequest(*u, date, reason), &gotgbot.SendMessageOpts{
		ParseMode:   gotgbot.ParseModeHTML,
		ReplyMarkup: kb,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})
	if err != nil {
		return err
	}

	slog.Info("rest requested", "message_id", msg.MessageId)
	if err := h.service.CreateRestRequest(ctx.StdContext(), u.ChatID, u.User.ID, msg.MessageId, date); err != nil {
		_ = ctx.Reply(b, "Не удалось создать заявку", nil)

		return err
	}

	return err
}

func (h *Handler) ShowRest(b *gotgbot.Bot, ctx *command.Context) error {
	u, err := ctx.AnyUser()
	if err != nil {
		return err
	}

	return ctx.ReplyHTML(b, view.FormatRestShow(*u))

}

func (h *Handler) AllUserRests(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	u, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	if u == nil {
		return cmd.ErrNoUser
	}
	var requests []model.ApprovedRestRequest

	requests, err = h.service.GetRequests(ctx.StdContext(), c.ID, u.User.ID)
	if err != nil {
		_ = ctx.Reply(b, "Не удалось получить список рестов", nil)
		return err
	}

	return ctx.ReplyHTML(b, view.FormatRestRequests(requests))
}

func (h *Handler) EndRest(b *gotgbot.Bot, ctx *command.Context) error {
	m, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	mod, err := ctx.Sender()
	if err != nil {
		return err
	}
	if m.User.ID != mod.User.ID && !mod.StatusGranted(ctx.RequiredStatus()) {
		return ctx.Reply(b, "Вы можете удалить из реста только себя", nil)
	}

	if !m.IsRestActive(time.Now()) {
		isSelf := m.User.ID == ctx.EffectiveUser.Id
		return ctx.ReplyHTML(b, view.FormatRestNotInRest(*m, isSelf))
	}

	if err := h.service.EndMemberRest(ctx.StdContext(), m.ChatID, m.User.ID); err != nil {
		_ = ctx.Reply(b, "Не удалось удалить пользователя из реста", nil)
		return err
	}

	isSelf := m.User.ID == ctx.EffectiveUser.Id
	return ctx.ReplyHTML(b, view.FormatRestEnded(*m, isSelf))
}

func (h *Handler) ApproveRestRequest(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	chatID := c.ID
	fromID, err := parseRequestCallbackData(ctx.CallbackQuery.Data)
	restRequest, err := h.service.GetRestRequest(ctx.StdContext(), chatID, fromID, ctx.EffectiveMessage.MessageId)
	if err != nil {
		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Не найден запрос на рест",
		})
		return err
	}
	moderator, err := h.memberService.GetChatMember(ctx.StdContext(), chatID, ctx.EffectiveSender.Id())
	if err != nil {
		return err
	}

	requiredStatus := model.StatusAdmin
	if h.chatService != nil {
		if s, err := h.chatService.GetCommandPermission(ctx.StdContext(), chatID, "rests"); err == nil {
			requiredStatus = s
		}
	}

	if moderator.Status < requiredStatus {
		_, err = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Подтвердить запрос может только администратор",
		})
		return err
	}

	if err := h.service.ApproveRestRequest(ctx.StdContext(), chatID, fromID, ctx.EffectiveMessage.MessageId, restRequest.RestUntil); err != nil {
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Не удалось одобрить запрос",
		})
		return err
	}

	u, err := h.memberService.GetChatMember(ctx.StdContext(), chatID, fromID)
	if err != nil {
		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Не удалось найти пользователя",
		})
		return err
	}

	_, _, err = b.EditMessageText(view.FormatRestRequestApproved(u, restRequest.RestUntil), &gotgbot.EditMessageTextOpts{
		ChatId:    ctx.EffectiveChat.Id,
		MessageId: ctx.EffectiveMessage.MessageId,
		ParseMode: gotgbot.ParseModeHTML,
	})
	return err
}

func (h *Handler) RejectRestRequest(b *gotgbot.Bot, ctx *command.Context) error {
	cctx := ctx.StdContext()

	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	chatID := c.ID

	fromID, err := parseRequestCallbackData(ctx.CallbackQuery.Data)
	restRequest, err := h.service.GetRestRequest(cctx, chatID, fromID, ctx.EffectiveMessage.MessageId)
	if err != nil {
		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Не найден запрос на рест",
		})
		return err
	}
	moderator, err := h.memberService.GetChatMember(ctx.StdContext(), chatID, ctx.EffectiveSender.Id())
	if err != nil {
		return err
	}

	requiredStatus := model.StatusAdmin
	if h.chatService != nil {
		if s, err := h.chatService.GetCommandPermission(ctx.StdContext(), chatID, "rests"); err == nil {
			requiredStatus = s
		}
	}

	if restRequest.UserID != ctx.EffectiveSender.Id() && moderator.Status < requiredStatus {
		_, err = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Отклонить запрос может только администратор или заявитель реста",
		})
		return err
	}
	slog.Info("rejecting rest request", "message_id", ctx.EffectiveMessage.MessageId)
	if err := h.service.RejectRestRequest(cctx, chatID, ctx.EffectiveSender.Id(), ctx.EffectiveMessage.MessageId); err != nil {
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Не удалось отклонить запрос",
		})
		return err
	}

	u, err := h.memberService.GetChatMember(cctx, chatID, fromID)
	if err != nil {
		_, _, err = b.EditMessageText(view.FormatRestRequestRejected(nil),
			&gotgbot.EditMessageTextOpts{
				ChatId:      ctx.EffectiveChat.Id,
				MessageId:   ctx.EffectiveMessage.MessageId,
				ParseMode:   gotgbot.ParseModeHTML,
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{},
			},
		)
		return err
	}

	_, _, err = b.EditMessageText(view.FormatRestRequestRejected(&u),
		&gotgbot.EditMessageTextOpts{
			ChatId:      ctx.EffectiveChat.Id,
			MessageId:   ctx.EffectiveMessage.MessageId,
			ParseMode:   gotgbot.ParseModeHTML,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{},
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

func (h *Handler) RemoveRestRequest(b *gotgbot.Bot, ctx *command.Context) error {
	u, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	number, err := ctx.Number()
	if err != nil {
		_ = ctx.Reply(b, "Укажите корректный номер рест запроса", nil)
		return err
	}

	requests, err := h.service.GetRequests(ctx.StdContext(), u.ChatID, u.User.ID)
	if err != nil {
		return err
	}
	if number > len(requests) {
		return ctx.Reply(b, "Не найден запрос с этим номером", nil)
	}

	request := requests[number-1]

	if request.RestUntil.Equal(u.RestUntil) {

		if err := h.service.DeleteRestRequestAndEndRest(ctx.StdContext(), request.ChatID, request.UserID, request.ID); err != nil {
			return err
		}
		return ctx.Reply(b, "Удалён действительный запрос на рест, вместе с этим удалён и статус реста у участника", nil)
	}

	if err := h.service.DeleteRestRequest(ctx.StdContext(), request.ID); err != nil {
		return err
	}
	return ctx.Reply(b, "Запрос на рест успешно удален", nil)

}
