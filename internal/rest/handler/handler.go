package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/chat"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"activity-bot/internal/options"
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

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
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

func (h *Handler) SetRest(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return fmt.Errorf("set rest: get chat: %w", err)
	}
	target, err := ctx.AnyUser()
	if err != nil {
		return fmt.Errorf("set rest: resolve target: %w", err)
	}
	sender, err := ctx.Sender()
	if err != nil {
		return fmt.Errorf("set rest: resolve sender: %w", err)
	}
	date, err := ctx.Date()
	if err != nil {
		return fmt.Errorf("set rest: parse date: %w", err)
	}
	reason := ctx.TextOrDefault("")

	if sender.Status < ctx.RequiredStatus() {
		return h.createRequest(ctx, u, target, date, reason)
	}

	if err := h.service.SetMemberRestWithHistory(ctx.StdContext(), c.ID, target.User.ID, int64(u.EffectiveMessage.GetID()), date, reason); err != nil {
		return ctx.ReplyOnly(u, options.WithText("Не удалось создать рест"))
	}

	eb := &entity.Builder{}
	view.WriteRestSet(eb, *target, date, reason)

	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) createRequest(ctx *command.Context, u *ext.Update, target *model.ChatMember, date time.Time, reason string) error {
	kb := &tg.ReplyInlineMarkup{
		Rows: []tg.KeyboardButtonRow{
			{
				Buttons: []tg.KeyboardButtonClass{
					&tg.KeyboardButtonCallback{
						Text: "Одобрить",
						Data: []byte(fmt.Sprintf("approve:%d", target.User.ID)),
					},
					&tg.KeyboardButtonCallback{
						Text: "Отклонить",
						Data: []byte(fmt.Sprintf("reject:%d", target.User.ID)),
					},
				},
			},
		},
	}

	eb := &entity.Builder{}
	view.WriteRestRequest(eb, *target, date, reason)

	msg, err := ctx.Reply(u, options.WithBuilder(eb), options.WithMarkup(kb))
	if err != nil {
		return fmt.Errorf("create rest request: send request message: %w", err)
	}

	slog.Info("rest requested", "message_id", msg.GetID())
	if err := h.service.CreateRestRequest(ctx.StdContext(), target.ChatID, target.User.ID, int64(msg.GetID()), date); err != nil {
		return ctx.ReplyOnly(u, options.WithText("Не удалось создать заявку"))
	}

	return nil
}

func (h *Handler) ShowRest(ctx *command.Context, u *ext.Update) error {
	target, err := ctx.AnyUser()
	if err != nil {
		return fmt.Errorf("show rest: resolve user: %w", err)
	}

	eb := &entity.Builder{}
	view.WriteRestShow(eb, *target)

	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) AllUserRests(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return fmt.Errorf("all user rests: get chat: %w", err)
	}

	target, err := ctx.AnyUser()
	if err != nil {
		return fmt.Errorf("all user rests: resolve user: %w", err)
	}
	var requests []model.ApprovedRestRequest

	requests, err = h.service.GetRequests(ctx.StdContext(), c.ID, target.User.ID)
	if err != nil {
		return ctx.ReplyOnly(u, options.WithText("Не удалось получить список рестов"))
	}

	eb := &entity.Builder{}
	view.WriteRestRequests(eb, requests)

	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) EndRest(ctx *command.Context, u *ext.Update) error {
	m, err := ctx.AnyUser()
	if err != nil {
		return fmt.Errorf("end rest: resolve user: %w", err)
	}
	mod, err := ctx.Sender()
	if err != nil {
		return fmt.Errorf("end rest: resolve sender: %w", err)
	}
	if m.User.ID != mod.User.ID && !mod.StatusGranted(ctx.RequiredStatus()) {
		return ctx.ReplyOnly(u, options.WithText("Вы можете удалить из реста только себя"))
	}

	if !m.IsRestActive(time.Now()) {
		isSelf := m.User.ID == u.EffectiveUser().GetID()
		eb := &entity.Builder{}
		view.WriteRestNotInRest(eb, *m, isSelf)
		return ctx.ReplyOnly(u, options.WithBuilder(eb))
	}

	if err := h.service.EndMemberRest(ctx.StdContext(), m.ChatID, m.User.ID); err != nil {
		return ctx.ReplyOnly(u, options.WithText("Не удалось удалить пользователя из реста"))
	}

	isSelf := m.User.ID == u.EffectiveUser().GetID()
	eb := &entity.Builder{}
	view.WriteRestEnded(eb, *m, isSelf)
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) ApproveRestRequest(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return fmt.Errorf("approve rest request: get chat: %w", err)
	}
	chatID := c.ID
	cq := u.CallbackQuery
	if cq == nil {
		return nil
	}
	data, _ := cq.GetData()
	fromID, err := parseRequestCallbackData(string(data))
	if err != nil {
		return fmt.Errorf("approve rest request: parse callback: %w", err)
	}
	restRequest, err := h.service.GetRestRequest(ctx.StdContext(), chatID, fromID, int64(cq.GetMsgID()))
	if err != nil {
		_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: cq.QueryID,
			Message: "Не найден запрос на рест",
			Alert:   true,
		})
		return fmt.Errorf("approve rest request: load request: %w", err)
	}
	moderator, err := h.memberService.GetChatMember(ctx.StdContext(), chatID, u.EffectiveUser().GetID())
	if err != nil {
		return fmt.Errorf("approve rest request: get moderator: %w", err)
	}

	requiredStatus := model.StatusAdmin
	if h.chatService != nil {
		if s, err := h.chatService.GetCommandPermission(ctx.StdContext(), chatID, "rests"); err == nil {
			requiredStatus = s
		}
	}

	if moderator.Status < requiredStatus {
		_, err = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: cq.QueryID,
			Message: "Подтвердить запрос может только администратор",
			Alert:   true,
		})
		return fmt.Errorf("approve rest request: insufficient rights callback: %w", err)
	}

	if err := h.service.ApproveRestRequest(ctx.StdContext(), chatID, fromID, int64(cq.GetMsgID()), restRequest.RestUntil); err != nil {
		_, err := ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: cq.QueryID,
			Message: "Не удалось одобрить запрос",
			Alert:   true,
		})
		return fmt.Errorf("approve rest request: approve request callback: %w", err)
	}

	target, err := h.memberService.GetChatMember(ctx.StdContext(), chatID, fromID)
	if err != nil {
		_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: cq.QueryID,
			Message: "Не удалось найти пользователя",
		})
		return fmt.Errorf("approve rest request: get target member: %w", err)
	}

	eb := &entity.Builder{}
	view.WriteRestRequestApproved(eb, target, restRequest.RestUntil)
	result, entities := eb.Complete()

	_, err = ctx.EditMessage(u.EffectiveChat().GetID(), &tg.MessagesEditMessageRequest{
		ID:       cq.GetMsgID(),
		Message:  result,
		Entities: entities,
	})
	if err != nil {
		return fmt.Errorf("approve rest request: edit message: %w", err)
	}

	_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
		QueryID: cq.QueryID,
		Message: "Запрос одобрен",
	})
	return nil
}

func (h *Handler) RejectRestRequest(ctx *command.Context, u *ext.Update) error {
	cctx := ctx.StdContext()

	c, err := ctx.Chat()
	if err != nil {
		return fmt.Errorf("reject rest request: get chat: %w", err)
	}
	chatID := c.ID

	cq := u.CallbackQuery
	if cq == nil {
		return nil
	}
	data, _ := cq.GetData()
	fromID, err := parseRequestCallbackData(string(data))
	if err != nil {
		return fmt.Errorf("reject rest request: parse callback: %w", err)
	}
	restRequest, err := h.service.GetRestRequest(cctx, chatID, fromID, int64(cq.GetMsgID()))
	if err != nil {
		_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: cq.QueryID,
			Message: "Не найден запрос на рест",
			Alert:   true,
		})
		return fmt.Errorf("reject rest request: load request: %w", err)
	}
	moderator, err := h.memberService.GetChatMember(ctx.StdContext(), chatID, u.EffectiveUser().GetID())
	if err != nil {
		return fmt.Errorf("reject rest request: get moderator: %w", err)
	}

	requiredStatus := model.StatusAdmin
	if h.chatService != nil {
		if s, err := h.chatService.GetCommandPermission(ctx.StdContext(), chatID, "rests"); err == nil {
			requiredStatus = s
		}
	}

	if restRequest.UserID != u.EffectiveUser().GetID() && moderator.Status < requiredStatus {
		_, err = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: cq.QueryID,
			Message: "Отклонить запрос может только администратор или заявитель реста",
			Alert:   true,
		})
		return fmt.Errorf("reject rest request: insufficient rights callback: %w", err)
	}
	slog.Info("rejecting rest request", "message_id", cq.GetMsgID())
	if err := h.service.RejectRestRequest(cctx, chatID, u.EffectiveUser().GetID(), int64(cq.GetMsgID())); err != nil {
		_, err := ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: cq.QueryID,
			Message: "Не удалось отклонить запрос",
			Alert:   true,
		})
		return fmt.Errorf("reject rest request: reject callback: %w", err)
	}

	target, err := h.memberService.GetChatMember(cctx, chatID, fromID)
	eb := &entity.Builder{}
	if err != nil {
		view.WriteRestRequestRejected(eb, nil)
	} else {
		view.WriteRestRequestRejected(eb, &target)
	}
	result, entities := eb.Complete()

	_, err = ctx.EditMessage(u.EffectiveChat().GetID(), &tg.MessagesEditMessageRequest{
		ID:       cq.GetMsgID(),
		Message:  result,
		Entities: entities,
	})
	if err != nil {
		return fmt.Errorf("reject rest request: edit message: %w", err)
	}

	_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
		QueryID: cq.QueryID,
		Message: "Запрос отклонён",
	})
	return nil
}

func parseRequestCallbackData(callbackData string) (int64, error) {
	parts := strings.SplitN(callbackData, ":", 2)
	if len(parts) != 2 {
		return 0, errors.New("invalid callback data")
	}
	fromID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse callback user id: %w", err)
	}
	return fromID, nil
}

func (h *Handler) RemoveRestRequest(ctx *command.Context, u *ext.Update) error {
	target, err := ctx.AnyUser()
	if err != nil {
		return fmt.Errorf("remove rest request: resolve user: %w", err)
	}
	number, err := ctx.Number()
	if err != nil {
		return ctx.ReplyOnly(u, options.WithText("Укажите корректный номер рест запроса"))
	}

	requests, err := h.service.GetRequests(ctx.StdContext(), target.ChatID, target.User.ID)
	if err != nil {
		return fmt.Errorf("remove rest request: get requests: %w", err)
	}
	if number > len(requests) {
		return ctx.ReplyOnly(u, options.WithText("Не найден запрос с этим номером"))
	}

	request := requests[number-1]

	if request.RestUntil.Equal(target.RestUntil) {
		if err := h.service.DeleteRestRequestAndEndRest(ctx.StdContext(), request.ChatID, request.UserID, request.ID); err != nil {
			return fmt.Errorf("remove rest request: delete and end rest: %w", err)
		}
		return ctx.ReplyOnly(u, options.WithText("Удалён действительный запрос на рест, вместе с этим удалён и статус реста у участника"))
	}

	if err := h.service.DeleteRestRequest(ctx.StdContext(), request.ID); err != nil {
		return fmt.Errorf("remove rest request: delete request: %w", err)
	}
	return ctx.ReplyOnly(u, options.WithText("Запрос на рест успешно удален"))
}
