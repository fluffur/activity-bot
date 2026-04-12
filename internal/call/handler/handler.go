package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/call"
	"activity-bot/internal/call/view"
	"activity-bot/internal/chat"
	"activity-bot/internal/command"
	"activity-bot/internal/conversation"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"activity-bot/internal/options"
	"activity-bot/internal/session"
	"activity-bot/internal/stats"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
)

type Handler struct {
	service        *call.Service
	memberService  *member.Service
	chatService    *chat.Service
	adminService   *admin.Service
	sessionService *session.Service
	storage        conversation.Storage
	mu             sync.Mutex
	// promptMessages[chatID][userID] = messageID вопроса "ответьте на это сообщение..."
	promptMessages map[int64]map[int64]int64
}

const (
	CallStateInactive   = "call_inactive_msg"
	CallStateNoNorm     = "call_no_norm_msg"
	CallStateNoNormWarn = "call_no_norm_warn_msg"
	CallStateNoNormBan  = "call_no_norm_ban_msg"
)

func New(
	service *call.Service,
	memberService *member.Service,
	chatService *chat.Service,
	adminService *admin.Service,
	sessionService *session.Service,
	storage conversation.Storage,
) *Handler {
	return &Handler{
		service:        service,
		memberService:  memberService,
		chatService:    chatService,
		adminService:   adminService,
		sessionService: sessionService,
		storage:        storage,
		promptMessages: make(map[int64]map[int64]int64),
	}
}

// func (h *Handler) callMembers(
//
//	b *gotgbot.Bot,
//	ctx *command.Context,
//	getMembers func() ([]model.ChatMember, error),
//	emptyMsg string,
//
// ) error {
//
//		members, err := getMembers()
//		if err != nil {
//			return err
//		}
//
//		if len(members) == 0 {
//			return ctx.Reply(b, emptyMsg, nil)
//		}
//
//		return h.doCall(ctx, u, msg, entities, members)
//	}
//
// func (h *Handler) adminCallback(
//
//	b *gotgbot.Bot,
//	ctx *command.Context,
//	handler func(*gotgbot.Bot, *command.Context) error,
//
// ) error {
//
//		m, err := h.memberService.GetChatMember(
//			ctx.StdContext(),
//			ctx.EffectiveChat.Id,
//			ctx.EffectiveSender.Id(),
//		)
//		if err != nil {
//			return err
//		}
//callINactive
//		if !m.StatusGranted(model.StatusModerator) {
//			_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
//				Text: fmt.Sprintf("%d Требуются права: %s", m.Status, m.Status.String()),
//			})
//			return err
//		}
//
//		_, _ = ctx.CallbackQuery.Answer(b, nil)
//
//		return handler(b, ctx)
//	}

func (h *Handler) Call(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	members, err := h.memberService.GetChatMembers(ctx.StdContext(), c.ID)
	if err != nil {
		return err
	}

	if len(members) == 0 {
		return nil
	}

	msg := c.WelcomeCallMessage
	var entities []tg.MessageEntityClass
	if ctx.RawArgs != "" {
		msg = ctx.RawArgs
		entities = ctx.RawArgsEntities
	}

	return h.doCall(ctx, u, msg, entities, members)
}

func (h *Handler) CallInactive(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	members, err := h.service.GetInactiveMembers(ctx.StdContext(), c.ID)
	if err != nil {
		return err
	}
	if len(members) == 0 {
		return ctx.ReplyOnly(u, options.WithText("Нет участников, не писавших более суток"))
	}
	return h.doCall(ctx, u, ctx.RawArgs, ctx.RawArgsEntities, members)
}

func (h *Handler) CallNoNorm(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	from, to := stats.ResolvePeriod(
		stats.PeriodWeek,
		time.Now().In(helpers.MoscowLocation),
		c.WeekStartDay,
		c.WeekStartTime,
	)

	members, err := h.memberService.GetNoNormMembers(ctx.StdContext(), c.ID, from, to)
	if err != nil {
		return err
	}
	if len(members) == 0 {
		eb := &entity.Builder{}
		helpers.WriteSuccessEmoji(eb)
		eb.Plain(" Все участники выполнили норму")
		return ctx.ReplyOnly(u, options.WithBuilder(eb))
	}

	return h.doCall(ctx, u, ctx.RawArgs, nil, members)
}

func (h *Handler) CallNoNormWarn(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	from, to := stats.ResolvePeriod(
		stats.PeriodWeek,
		time.Now(),
		c.WeekStartDay,
		c.WeekStartTime,
	)

	members, err := h.memberService.GetNoNormWarnMembers(ctx.StdContext(), c.ID, from, to)
	if err != nil {
		return err
	}
	if len(members) == 0 {
		return ctx.ReplyOnly(u, options.WithText("Все участники выполнили норму предупреждения"))
	}

	return h.doCall(ctx, u, "", nil, members)
}

func (h *Handler) CallNoNormBan(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	from, to := stats.ResolvePeriod(
		stats.PeriodWeek,
		time.Now(),
		c.WeekStartDay,
		c.WeekStartTime,
	)

	members, err := h.memberService.GetNoNormBanMembers(ctx.StdContext(), c.ID, from, to)
	if err != nil {
		return err
	}
	if len(members) == 0 {
		return ctx.ReplyOnly(u, options.WithText("Все участники выполнили норму бана"))
	}

	return h.doCall(ctx, u, "", nil, members)
}

func (h *Handler) doCall(
	ctx *command.Context,
	u *ext.Update,
	message string,
	entities []tg.MessageEntityClass,
	members []model.ChatMember,
) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	mentionsLimit := int(c.MentionsPerMessage)
	if mentionsLimit <= 0 {
		mentionsLimit = 5
	}

	if message == "" {
		message = c.WelcomeCallMessage
	}

	for i := 0; i < len(members); i += mentionsLimit {
		end := i + mentionsLimit
		if end > len(members) {
			end = len(members)
		}

		eb := &entity.Builder{}
		view.FormatCallChunkBuilder(eb, message, members[i:end], c.MentionTypes)

		finalText, chunkEntities := eb.Complete()
		finalEntities := append(entities, chunkEntities...)

		if err := ctx.ReplyOnly(u, options.WithText(finalText), options.WithEntities(finalEntities)); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) SetMentionsPerMessage(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	count, err := ctx.Number()
	if err != nil {
		return err
	}
	if count <= 0 || count > 50 {
		return ctx.ReplyOnly(u, options.WithText("Укажите число от 1 до 50"))
	}

	if err := h.service.SetMentionsPerMessage(
		ctx.StdContext(),
		c.ID,
		int32(count),
	); err != nil {
		return err
	}

	return ctx.ReplyOnly(
		u,
		options.WithText(fmt.Sprintf("Лимит упоминаний в одном сообщении изменен на %d", count)),
	)
}

func (h *Handler) ShowMentionsPerMessage(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	return ctx.ReplyOnly(
		u,
		options.WithText(fmt.Sprintf("Лимит упоминаний в одном сообщении: %d", c.MentionsPerMessage)),
	)
}

func (h *Handler) ShowCallTypes(ctx *command.Context, u *ext.Update) error {
	if u.CallbackQuery != nil {
		_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{QueryID: u.CallbackQuery.QueryID})
	}
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	return ctx.ReplyOnly(
		u,
		options.WithText("Настройте стиль упоминаний:"),
		options.WithMarkup(h.getCallTypesKeyboard(c.MentionTypes)),
	)
}

func (h *Handler) CallbackCallType(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	data, _ := u.CallbackQuery.GetData()
	var bit int32
	if _, err := fmt.Sscanf(string(data), "call_type:%d", &bit); err != nil {
		return err
	}

	current := c.MentionTypes
	var newTypes int32

	if bit == view.MentionTypeNWSP {
		newTypes = view.MentionTypeNWSP
	} else {
		current &^= view.MentionTypeNWSP
		newTypes = current ^ bit
	}

	if newTypes == c.MentionTypes {
		_, err = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			Message: "Настройки не изменились",
			QueryID: u.CallbackQuery.QueryID,
		})
		return err
	}

	if err := h.service.SetMentionTypes(ctx.StdContext(), c.ID, newTypes); err != nil {
		return err
	}

	_, err = ctx.EditMessage(u.EffectiveChat().GetID(), &tg.MessagesEditMessageRequest{
		ID:          u.CallbackQuery.GetMsgID(),
		ReplyMarkup: h.getCallTypesKeyboard(newTypes),
	})
	if err != nil {
		return err
	}

	_, err = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
		Message: "Настройки обновлены",
		QueryID: u.CallbackQuery.QueryID,
	})
	return err
}

func (h *Handler) getCallTypesKeyboard(currentTypes int32) tg.ReplyMarkupClass {
	types := []struct {
		name string
		bit  int32
	}{
		{"Пустота", view.MentionTypeNWSP},
		{"Эмодзи", view.MentionTypeEmoji},
		{"Имя", view.MentionTypeName},
		{"Роль", view.MentionTypeRole},
	}

	var rows []tg.KeyboardButtonRow
	var row []tg.KeyboardButtonClass

	for i, t := range types {
		text := t.name
		if (t.bit == view.MentionTypeNWSP && currentTypes == view.MentionTypeNWSP) ||
			(t.bit != view.MentionTypeNWSP && currentTypes&t.bit > 0) {
			text = "✅ " + text
		}

		row = append(row, &tg.KeyboardButtonCallback{
			Text: text,
			Data: []byte(fmt.Sprintf("call_type:%d", t.bit)),
		})

		if (i+1)%2 == 0 {
			rows = append(rows, tg.KeyboardButtonRow{Buttons: row})
			row = nil
		}
	}

	if len(row) > 0 {
		rows = append(rows, tg.KeyboardButtonRow{Buttons: row})
	}

	return &tg.ReplyInlineMarkup{Rows: rows}
}

func (h *Handler) SetWelcomeCallMessage(ctx *command.Context, u *ext.Update) error {
	message := ctx.RawArgs
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	if err := h.service.SetWelcomeCallMessage(ctx.StdContext(), c.ID, message); err != nil {
		return err
	}

	return ctx.ReplyOnly(u, options.WithText("Новое сообщение созыва установлено"))
}

func (h *Handler) DeleteWelcomeCallMessage(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	if c.WelcomeCallMessage == "" {
		return ctx.ReplyOnly(u, options.WithText("Сообщение ещё не было установлено"))
	}

	if err := h.service.SetWelcomeCallMessage(ctx.StdContext(), c.ID, ""); err != nil {
		return err
	}

	return ctx.ReplyOnly(u, options.WithText("Сообщение созыва удалено"))
}

func (h *Handler) EnableCallOnJoin(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	if err := h.service.EnableCallOnJoin(ctx.StdContext(), c.ID); err != nil {
		return err
	}

	return ctx.ReplyOnly(u, options.WithText("Теперь при вступлении новых участников будет выполняться созыв"))
}

func (h *Handler) DisableCallOnJoin(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	if err := h.service.DisableCallOnJoin(ctx.StdContext(), c.ID); err != nil {
		return err
	}

	return ctx.ReplyOnly(u, options.WithText("Теперь при инвайте новых участников не будет выполняться созыв"))
}

func (h *Handler) ShowWelcomeCallMessage(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	return ctx.ReplyOnly(u, options.WithText(view.FormatWelcomeCallMessage(c.WelcomeCallMessage)))
}

func (h *Handler) startCallConversation(
	ctx *command.Context,
	u *ext.Update,
	nextState string,
) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	chatID := c.ID
	m, err := h.memberService.GetChatMember(ctx.StdContext(), chatID, u.EffectiveUser().GetID())
	if err != nil {
		return err
	}

	if !m.StatusGranted(ctx.RequiredStatus()) {
		if u.CallbackQuery != nil {
			_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
				Message: "Требуются права администратора",
				Alert:   true,
				QueryID: u.CallbackQuery.QueryID,
			})
		}
		return nil
	}

	callType := "всех"
	switch nextState {
	case CallStateInactive:
		callType = "неактивных"
	case CallStateNoNorm:
		callType = "без нормы"
	case CallStateNoNormWarn:
		callType = "без нормы (предупреждение)"
	case CallStateNoNormBan:
		callType = "без нормы (бан)"
	}
	sender, err := ctx.Sender()
	if err != nil {
		return err
	}

	eb := &entity.Builder{}
	helpers.WriteRoleEmojiLink(eb, *sender)

	eb.Plain(fmt.Sprintf(
		", введите сообщение созыва %s: ",
		callType,
	))

	markup := &tg.ReplyInlineMarkup{
		Rows: []tg.KeyboardButtonRow{
			{
				Buttons: []tg.KeyboardButtonClass{
					&tg.KeyboardButtonCallback{
						Text: "Без сообщения",
						Data: []byte(fmt.Sprintf("call_nomsg:%s", nextState)),
						Style: tg.KeyboardButtonStyle{
							BgPrimary: true,
						},
					},
				},
			},
			{
				Buttons: []tg.KeyboardButtonClass{
					&tg.KeyboardButtonCallback{
						Text: "Отменить",
						Data: []byte("call_cancel"),
					},
				},
			},
		},
	}

	text, entities := eb.Complete()
	promptMsg, err := ctx.SendMessage(u.EffectiveChat().GetID(), &tg.MessagesSendMessageRequest{
		ReplyMarkup: markup,
		ReplyTo:     &tg.InputReplyToMessage{TopMsgID: u.CallbackQuery.GetMsgID()},
		Message:     text,
		Entities:    entities,
	})
	if err != nil {
		return err
	}

	uid := u.EffectiveUser().GetID()
	h.mu.Lock()
	if h.promptMessages[chatID] == nil {
		h.promptMessages[chatID] = make(map[int64]int64)
	}
	h.promptMessages[chatID][uid] = int64(promptMsg.ID)
	h.mu.Unlock()

	if u.CallbackQuery != nil {
		_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{QueryID: u.CallbackQuery.QueryID})
	}

	log.Println("next state", nextState)
	return conversation.SetState(ctx.StdContext(), h.storage, chatID, uid, nextState, time.Hour)
}

func (h *Handler) StartCallInactiveConversation(ctx *command.Context, u *ext.Update) error {
	return h.startCallConversation(ctx, u, CallStateInactive)
}

func (h *Handler) StartCallNoNormConversation(ctx *command.Context, u *ext.Update) error {
	return h.startCallConversation(ctx, u, CallStateNoNorm)
}

func (h *Handler) StartCallNoNormWarnConversation(ctx *command.Context, u *ext.Update) error {
	return h.startCallConversation(ctx, u, CallStateNoNormWarn)
}

func (h *Handler) StartCallNoNormBanConversation(ctx *command.Context, u *ext.Update) error {
	return h.startCallConversation(ctx, u, CallStateNoNormBan)
}

func (h *Handler) handleCallWithMessage(
	ctx *command.Context,
	u *ext.Update,
	getMembers func(stdCtx context.Context, chatID int64) ([]model.ChatMember, error),
) error {
	stdCtx := ctx.StdContext()

	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	uid := u.EffectiveUser().GetID()
	var promptID int64
	h.mu.Lock()
	if byUser, ok := h.promptMessages[c.ID]; ok {
		if mid, ok2 := byUser[uid]; ok2 {
			promptID = mid
			delete(byUser, uid)
			if len(byUser) == 0 {
				delete(h.promptMessages, c.ID)
			}
		}
	}
	h.mu.Unlock()

	if promptID != 0 {
		_, _ = ctx.EditMessage(c.ID, &tg.MessagesEditMessageRequest{
			ID:          int(promptID),
			ReplyMarkup: &tg.ReplyInlineMarkup{},
		})
	}

	members, err := getMembers(stdCtx, c.ID)
	if err != nil {
		return err
	}

	if len(members) == 0 {
		if err := ctx.ReplyOnly(u, options.WithText("Не найдено пользователей для созыва")); err != nil {
			return err
		}
		return conversation.StopConversation(stdCtx, h.storage, c.ID, uid)
	}

	var entities []tg.MessageEntityClass
	text := ""
	if u.EffectiveMessage != nil {
		text = u.EffectiveMessage.Text
		entities = u.EffectiveMessage.Entities
	}

	if err := h.doCall(ctx, u, text, entities, members); err != nil {
		return err
	}

	return conversation.StopConversation(stdCtx, h.storage, c.ID, uid)
}

func (h *Handler) HandleCallInactiveMessage(ctx *command.Context, u *ext.Update) error {
	return h.handleCallWithMessage(ctx, u, func(stdCtx context.Context, chatID int64) ([]model.ChatMember, error) {
		return h.service.GetInactiveMembers(stdCtx, chatID)
	})
}

func (h *Handler) HandleCallNoNormMessage(ctx *command.Context, u *ext.Update) error {
	return h.handleCallWithMessage(
		ctx,
		u,
		func(stdCtx context.Context, chatID int64) ([]model.ChatMember, error) {
			c, err := h.chatService.GetChat(stdCtx, chatID)
			if err != nil {
				return nil, err
			}

			from, to := stats.ResolvePeriod(
				stats.PeriodWeek,
				time.Now(),
				c.WeekStartDay,
				c.WeekStartTime,
			)

			return h.memberService.GetNoNormMembers(stdCtx, chatID, from, to)
		},
	)
}

func (h *Handler) HandleCallNoNormWarnMessage(ctx *command.Context, u *ext.Update) error {
	return h.handleCallWithMessage(
		ctx,
		u,
		func(stdCtx context.Context, chatID int64) ([]model.ChatMember, error) {
			c, err := h.chatService.GetChat(stdCtx, chatID)
			if err != nil {
				return nil, err
			}

			from, to := stats.ResolvePeriod(
				stats.PeriodWeek,
				time.Now(),
				c.WeekStartDay,
				c.WeekStartTime,
			)

			return h.memberService.GetNoNormWarnMembers(stdCtx, chatID, from, to)
		},
	)
}

func (h *Handler) HandleCallNoNormBanMessage(ctx *command.Context, u *ext.Update) error {
	return h.handleCallWithMessage(
		ctx,
		u,
		func(stdCtx context.Context, chatID int64) ([]model.ChatMember, error) {
			c, err := h.chatService.GetChat(stdCtx, chatID)
			if err != nil {
				return nil, err
			}

			from, to := stats.ResolvePeriod(
				stats.PeriodWeek,
				time.Now(),
				c.WeekStartDay,
				c.WeekStartTime,
			)

			return h.memberService.GetNoNormBanMembers(stdCtx, chatID, from, to)
		},
	)
}

func (h *Handler) CancelCallConversation(ctx *command.Context, u *ext.Update) error {
	log.Println("call")
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	uid := u.EffectiveUser().GetID()

	if u.CallbackQuery != nil {
		_, _ = ctx.EditMessage(c.ID, &tg.MessagesEditMessageRequest{
			ID:      u.CallbackQuery.GetMsgID(),
			Message: "❌ Операция созыва отменена.",
		})
		_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			Message: "Созыв отменён",
			QueryID: u.CallbackQuery.QueryID,
		})
	}

	return conversation.StopConversation(ctx.StdContext(), h.storage, c.ID, uid)
}

func (h *Handler) NoMessageCallConversation(ctx *command.Context, u *ext.Update) error {
	stdCtx := ctx.StdContext()

	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	uid := u.EffectiveUser().GetID()

	state := ""
	if u.CallbackQuery != nil {
		data, _ := u.CallbackQuery.GetData()
		const prefix = "call_nomsg"
		if strings.HasPrefix(string(data), prefix) {
			parts := strings.Split(string(data), ":")
			if len(parts) > 1 {
				state = parts[1]
			}
		}
		_, _ = ctx.EditMessage(c.ID, &tg.MessagesEditMessageRequest{
			ID: u.CallbackQuery.GetMsgID(),
		})
	}

	var members []model.ChatMember
	switch state {
	case CallStateInactive:
		members, err = h.service.GetInactiveMembers(stdCtx, c.ID)
	case CallStateNoNorm:
		from, to := stats.ResolvePeriod(
			stats.PeriodWeek,
			time.Now(),
			c.WeekStartDay,
			c.WeekStartTime,
		)
		members, err = h.memberService.GetNoNormMembers(stdCtx, c.ID, from, to)
	case CallStateNoNormWarn, CallStateNoNormBan:
		from, to := stats.ResolvePeriod(
			stats.PeriodWeek,
			time.Now(),
			c.WeekStartDay,
			c.WeekStartTime,
		)
		members, err = h.memberService.GetNoNormWarnMembers(stdCtx, c.ID, from, to)
	default:
		return conversation.StopConversation(stdCtx, h.storage, c.ID, uid)
	}

	if err != nil {
		return err
	}

	if len(members) == 0 {
		if u.CallbackQuery != nil {
			_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
				Message: "Нет участников для созыва.",
				Alert:   true,
				QueryID: u.CallbackQuery.QueryID,
			})
		}
		return conversation.StopConversation(stdCtx, h.storage, c.ID, uid)
	}

	if err := h.doCall(ctx, u, "", nil, members); err != nil {
		return err
	}

	if u.CallbackQuery != nil {
		_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			Message: "Созыв отправлен без сообщения.",
			QueryID: u.CallbackQuery.QueryID,
		})
		_, _ = ctx.EditMessage(c.ID, &tg.MessagesEditMessageRequest{
			ID:          u.CallbackQuery.GetMsgID(),
			ReplyMarkup: &tg.ReplyInlineMarkup{},
		})
	}

	return conversation.StopConversation(stdCtx, h.storage, c.ID, uid)
}
