package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/chat"
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

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
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
		return err
	}
	target, err := ctx.AnyUser()
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
		return h.createRequest(ctx, u, target, date, reason)
	}

	if err := h.service.SetMemberRestWithHistory(ctx.StdContext(), c.ID, target.User.ID, int64(u.EffectiveMessage.GetID()), date, reason); err != nil {
		_, _ = ctx.Reply(u, ext.ReplyTextString("Не удалось создать рест"), nil)
		return err
	}

	eb := &entity.Builder{}
	view.WriteRestSet(eb, *target, date, reason)
	finalText, finalEntities := eb.Complete()

	_, err = ctx.SendMessage(u.EffectiveChat().GetID(), &tg.MessagesSendMessageRequest{
		ReplyTo:  &tg.InputReplyToMessage{ReplyToMsgID: u.EffectiveMessage.GetID()},
		Message:  finalText,
		Entities: finalEntities,
	})
	return err
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
	finalText, finalEntities := eb.Complete()

	msg, err := ctx.SendMessage(u.EffectiveChat().GetID(), &tg.MessagesSendMessageRequest{
		ReplyTo:     &tg.InputReplyToMessage{ReplyToMsgID: u.EffectiveMessage.GetID()},
		Message:     finalText,
		Entities:    finalEntities,
		ReplyMarkup: kb,
	})
	if err != nil {
		return err
	}

	slog.Info("rest requested", "message_id", msg.GetID())
	if err := h.service.CreateRestRequest(ctx.StdContext(), target.ChatID, target.User.ID, int64(msg.GetID()), date); err != nil {
		_, _ = ctx.Reply(u, ext.ReplyTextString("Не удалось создать заявку"), nil)
		return err
	}

	return nil
}

func (h *Handler) ShowRest(ctx *command.Context, u *ext.Update) error {
	target, err := ctx.AnyUser()
	if err != nil {
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		view.WriteRestShow(eb, *target)
		return nil
	})), nil)
	return err
}

func (h *Handler) AllUserRests(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	target, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	var requests []model.ApprovedRestRequest

	requests, err = h.service.GetRequests(ctx.StdContext(), c.ID, target.User.ID)
	if err != nil {
		_, _ = ctx.Reply(u, ext.ReplyTextString("Не удалось получить список рестов"), nil)
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		view.WriteRestRequests(eb, requests)
		return nil
	})), nil)
	return err
}

func (h *Handler) EndRest(ctx *command.Context, u *ext.Update) error {
	m, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	mod, err := ctx.Sender()
	if err != nil {
		return err
	}
	if m.User.ID != mod.User.ID && !mod.StatusGranted(ctx.RequiredStatus()) {
		_, err = ctx.Reply(u, ext.ReplyTextString("Вы можете удалить из реста только себя"), nil)
		return err
	}

	if !m.IsRestActive(time.Now()) {
		isSelf := m.User.ID == u.EffectiveUser().GetID()
		_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
			view.WriteRestNotInRest(eb, *m, isSelf)
			return nil
		})), nil)
		return err
	}

	if err := h.service.EndMemberRest(ctx.StdContext(), m.ChatID, m.User.ID); err != nil {
		_, _ = ctx.Reply(u, ext.ReplyTextString("Не удалось удалить пользователя из реста"), nil)
		return err
	}

	isSelf := m.User.ID == u.EffectiveUser().GetID()
	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		view.WriteRestEnded(eb, *m, isSelf)
		return nil
	})), nil)
	return err
}

func (h *Handler) ApproveRestRequest(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	chatID := c.ID
	data, _ := u.CallbackQuery.GetData()
	fromID, err := parseRequestCallbackData(string(data))
	if err != nil {
		return err
	}
	restRequest, err := h.service.GetRestRequest(ctx.StdContext(), chatID, fromID, int64(u.EffectiveMessage.GetID()))
	if err != nil {
		_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: u.CallbackQuery.QueryID,
			Message: "Не найден запрос на рест",
			Alert:   true,
		})
		return err
	}
	moderator, err := h.memberService.GetChatMember(ctx.StdContext(), chatID, u.EffectiveUser().GetID())
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
		_, err = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: u.CallbackQuery.QueryID,
			Message: "Подтвердить запрос может только администратор",
			Alert:   true,
		})
		return err
	}

	if err := h.service.ApproveRestRequest(ctx.StdContext(), chatID, fromID, int64(u.EffectiveMessage.GetID()), restRequest.RestUntil); err != nil {
		_, err := ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: u.CallbackQuery.QueryID,
			Message: "Не удалось одобрить запрос",
			Alert:   true,
		})
		return err
	}

	target, err := h.memberService.GetChatMember(ctx.StdContext(), chatID, fromID)
	if err != nil {
		_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: u.CallbackQuery.QueryID,
			Message: "Не удалось найти пользователя",
		})
		return err
	}

	eb := &entity.Builder{}
	view.WriteRestRequestApproved(eb, target, restRequest.RestUntil)
	result, entities := eb.Complete()

	_, err = ctx.EditMessage(u.EffectiveChat().GetID(), &tg.MessagesEditMessageRequest{
		ID:       u.EffectiveMessage.GetID(),
		Message:  result,
		Entities: entities,
	})
	if err != nil {
		return err
	}

	_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
		QueryID: u.CallbackQuery.QueryID,
		Message: "Запрос одобрен",
	})
	return nil
}

func (h *Handler) RejectRestRequest(ctx *command.Context, u *ext.Update) error {
	cctx := ctx.StdContext()

	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	chatID := c.ID

	data, _ := u.CallbackQuery.GetData()
	fromID, err := parseRequestCallbackData(string(data))
	if err != nil {
		return err
	}
	restRequest, err := h.service.GetRestRequest(cctx, chatID, fromID, int64(u.EffectiveMessage.GetID()))
	if err != nil {
		_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: u.CallbackQuery.QueryID,
			Message: "Не найден запрос на рест",
			Alert:   true,
		})
		return err
	}
	moderator, err := h.memberService.GetChatMember(ctx.StdContext(), chatID, u.EffectiveUser().GetID())
	if err != nil {
		return err
	}

	requiredStatus := model.StatusAdmin
	if h.chatService != nil {
		if s, err := h.chatService.GetCommandPermission(ctx.StdContext(), chatID, "rests"); err == nil {
			requiredStatus = s
		}
	}

	if restRequest.UserID != u.EffectiveUser().GetID() && moderator.Status < requiredStatus {
		_, err = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: u.CallbackQuery.QueryID,
			Message: "Отклонить запрос может только администратор или заявитель реста",
			Alert:   true,
		})
		return err
	}
	slog.Info("rejecting rest request", "message_id", u.EffectiveMessage.GetID())
	if err := h.service.RejectRestRequest(cctx, chatID, u.EffectiveUser().GetID(), int64(u.EffectiveMessage.GetID())); err != nil {
		_, err := ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: u.CallbackQuery.QueryID,
			Message: "Не удалось отклонить запрос",
			Alert:   true,
		})
		return err
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
		ID:       u.EffectiveMessage.GetID(),
		Message:  result,
		Entities: entities,
	})
	if err != nil {
		return err
	}

	_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
		QueryID: u.CallbackQuery.QueryID,
		Message: "Запрос отклонён",
	})
	return nil
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

func (h *Handler) RemoveRestRequest(ctx *command.Context, u *ext.Update) error {
	target, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	number, err := ctx.Number()
	if err != nil {
		_, _ = ctx.Reply(u, ext.ReplyTextString("Укажите корректный номер рест запроса"), nil)
		return err
	}

	requests, err := h.service.GetRequests(ctx.StdContext(), target.ChatID, target.User.ID)
	if err != nil {
		return err
	}
	if number > len(requests) {
		_, err = ctx.Reply(u, ext.ReplyTextString("Не найден запрос с этим номером"), nil)
		return err
	}

	request := requests[number-1]

	if request.RestUntil.Equal(target.RestUntil) {
		if err := h.service.DeleteRestRequestAndEndRest(ctx.StdContext(), request.ChatID, request.UserID, request.ID); err != nil {
			return err
		}
		_, err = ctx.Reply(u, ext.ReplyTextString("Удалён действительный запрос на рест, вместе с этим удалён и статус реста у участника"), nil)
		return err
	}

	if err := h.service.DeleteRestRequest(ctx.StdContext(), request.ID); err != nil {
		return err
	}
	_, err = ctx.Reply(u, ext.ReplyTextString("Запрос на рест успешно удален"), nil)
	return err
}
