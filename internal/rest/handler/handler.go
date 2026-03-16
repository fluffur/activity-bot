package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/cmd"
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
	adminService   *admin.Service
	dateParser     *helpers.DateParser
	sessionService *session.Service
	asyncClient    *asynq.Client
}

func New(service *rest.Service, userService *user.Service, memberService *member.Service, adminService *admin.Service, dateParser *helpers.DateParser, sessionService *session.Service, asyncClient *asynq.Client) *Handler {
	return &Handler{service, userService, memberService, adminService, dateParser, sessionService, asyncClient}
}

func (h *Handler) Set(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()
	if targetUser == nil {
		return cmd.ErrNoUser
	}

	firstArgument := ctx.FirstArgument()

	date := time.Time{}
	ok := false
	if len(ctx.ParsedDates()) > 0 {
		date = ctx.ParsedDates()[0]
		ok = true
	} else if firstArgument != "" {
		date, ok = h.dateParser.Parse(firstArgument)
	}

	if !ok {
		return ctx.Reply(b, "Не понял формат. Примеры:\n+рест 12.01\n+рест 2 недели\n+рест месяц", nil)
	}
	if date.Before(time.Now()) {
		return ctx.Reply(b, "Нельзя указывать прошедшую дату", nil)
	}

	if !h.adminService.CheckIsAdmin(ctx.StdContext(), ctx.TargetChatID(), ctx.EffectiveSender.Id()) {
		return h.createRequest(b, ctx, targetUser, date)
	}

	if err := h.service.SetMemberRest(ctx.StdContext(), ctx.TargetChatID(), targetUser.ID, date, ctx.SecondArgument()); err != nil {
		_ = ctx.Reply(b, "Не удалось создать рест", nil)
		return err
	}

	text := view.FormatRestSet(*targetUser, date, ctx.SecondArgument())
	return ctx.ReplyHTML(b, text)
}

func (h *Handler) createRequest(b *gotgbot.Bot, ctx *cmd.Context, targetUser *model.User, date time.Time) error {

	kb := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "Одобрить", CallbackData: fmt.Sprintf("approve:%d", targetUser.ID), Style: "success", IconCustomEmojiId: helpers.SuccessEmojiGray},
				{Text: "Отклонить", CallbackData: fmt.Sprintf("reject:%d", targetUser.ID), Style: "danger", IconCustomEmojiId: helpers.DangerEmojiGray},
			},
		},
	}

	msg, err := ctx.EffectiveMessage.Reply(b, view.FormatRestRequest(*targetUser, date, ctx.SecondArgument()), &gotgbot.SendMessageOpts{
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
	if err := h.service.CreateRestRequest(ctx.StdContext(), ctx.TargetChatID(), targetUser.ID, msg.MessageId, date); err != nil {
		_ = ctx.Reply(b, "Не удалось создать заявку", nil)

		return err
	}

	return err
}

func (h *Handler) Show(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()

	if targetUser == nil {
		return cmd.ErrNoUser
	}

	m, err := h.memberService.GetChatMember(ctx.StdContext(), ctx.TargetChatID(), targetUser.ID)
	if err != nil {
		return err
	}

	return ctx.ReplyHTML(b, view.FormatRestShow(m))

}

func (h *Handler) List(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()

	var requests []model.ApprovedRestRequest
	var err error

	if targetUser != nil {
		requests, err = h.service.GetUserApprovedRequests(ctx.StdContext(), targetUser.ID)
	} else {
		requests, err = h.service.GetApprovedRequests(ctx.StdContext())
	}

	if err != nil {
		_ = ctx.Reply(b, "Не удалось получить список рестов", nil)
		return err
	}

	return ctx.ReplyHTML(b, view.FormatRestRequests(requests))
}

func (h *Handler) End(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()

	if targetUser == nil {
		return cmd.ErrNoUser
	}

	if targetUser.ID != ctx.EffectiveUser.Id && !h.adminService.CheckIsAdmin(ctx.StdContext(), ctx.TargetChatID(), ctx.EffectiveSender.Id()) {
		return ctx.Reply(b, "Вы можете удалить из реста только себя", nil)
	}

	m, err := h.memberService.GetChatMember(ctx.StdContext(), ctx.TargetChatID(), targetUser.ID)
	if err != nil {
		return ctx.Reply(b, "Не удалось проверить рест пользователя", nil)
	}
	if !m.RestUntil.IsZero() {
		isSelf := targetUser.ID == ctx.EffectiveUser.Id
		return ctx.ReplyHTML(b, view.FormatRestNotInRest(*targetUser, isSelf))
	}

	if err := h.service.EndMemberRest(ctx.StdContext(), ctx.TargetChatID(), targetUser.ID); err != nil {
		_ = ctx.Reply(b, "Не удалось удалить пользователя из реста", nil)
		return err
	}

	isSelf := targetUser.ID == ctx.EffectiveUser.Id
	return ctx.ReplyHTML(b, view.FormatRestEnded(*targetUser, isSelf))
}

func (h *Handler) ApproveRestRequest(b *gotgbot.Bot, ctx *cmd.Context) error {
	chatID := ctx.TargetChatID()

	fromID, err := parseRequestCallbackData(ctx.CallbackQuery.Data)
	restRequest, err := h.service.GetRestRequest(ctx.StdContext(), chatID, fromID, ctx.EffectiveMessage.MessageId)
	if err != nil {
		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Не найден запрос на рест",
		})
		return err
	}

	if !h.adminService.CheckIsAdmin(ctx.StdContext(), chatID, ctx.EffectiveSender.Id()) {
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
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

	u, err := h.userService.GetUser(ctx.StdContext(), fromID)
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

func (h *Handler) RejectRestRequest(b *gotgbot.Bot, ctx *cmd.Context) error {
	cctx := ctx.StdContext()

	chatID := ctx.TargetChatID()

	fromID, err := parseRequestCallbackData(ctx.CallbackQuery.Data)
	restRequest, err := h.service.GetRestRequest(cctx, chatID, fromID, ctx.EffectiveMessage.MessageId)
	if err != nil {
		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Не найден запрос на рест",
		})
		return err
	}

	if restRequest.UserID != ctx.EffectiveSender.Id() && !h.adminService.CheckIsAdmin(cctx, chatID, ctx.EffectiveSender.Id()) {
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
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

	u, err := h.userService.GetUser(cctx, fromID)
	if err != nil {
		_, _, err = b.EditMessageText(view.FormatRestRequestRejected(nil),
			&gotgbot.EditMessageTextOpts{
				ChatId:    ctx.EffectiveChat.Id,
				MessageId: ctx.EffectiveMessage.MessageId,
				ParseMode: gotgbot.ParseModeHTML,
			},
		)
		return err
	}

	_, _, err = b.EditMessageText(view.FormatRestRequestRejected(&u),
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
