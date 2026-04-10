package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/call"
	"activity-bot/internal/call/view"
	"activity-bot/internal/chat"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"activity-bot/internal/session"
	"activity-bot/internal/stats"
	"fmt"
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

func New(service *call.Service, memberService *member.Service, chatService *chat.Service, adminService *admin.Service, sessionService *session.Service) *Handler {
	return &Handler{
		service:        service,
		memberService:  memberService,
		chatService:    chatService,
		adminService:   adminService,
		sessionService: sessionService,
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
//
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
		_, err = ctx.Reply(u, ext.ReplyTextString("Нет участников, не писавших более суток"), nil)
		return err
	}

	return h.doCall(ctx, u, "", nil, members)
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
		_, err = ctx.Reply(u, ext.ReplyTextString(fmt.Sprintf("%s Все участники выполнили норму!", helpers.SuccessEmoji())), nil)
		return err
	}

	return h.doCall(ctx, u, "", nil, members)
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
		_, err = ctx.Reply(u, ext.ReplyTextString(fmt.Sprintf("%s Все участники выполнили норму предупреждения!", helpers.SuccessEmoji())), nil)
		return err
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
		_, err = ctx.Reply(u, ext.ReplyTextString(fmt.Sprintf("%s Все участники выполнили норму бана!", helpers.SuccessEmoji())), nil)
		return err
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

		_, err = ctx.SendMessage(u.EffectiveChat().GetID(), &tg.MessagesSendMessageRequest{
			ReplyTo:  &tg.InputReplyToMessage{ReplyToMsgID: u.EffectiveMessage.GetID()},
			Entities: finalEntities,
			Message:  finalText,
		})
		if err != nil {
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
		_, err = ctx.Reply(u, ext.ReplyTextString("Укажите число от 1 до 50"), nil)
		return err
	}

	if err := h.service.SetMentionsPerMessage(
		ctx.StdContext(),
		c.ID,
		int32(count),
	); err != nil {
		return err
	}

	_, err = ctx.Reply(
		u,
		ext.ReplyTextString(fmt.Sprintf("Лимит упоминаний в одном сообщении изменен на %d", count)),
		nil,
	)
	return err
}

func (h *Handler) ShowMentionsPerMessage(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	_, err = ctx.Reply(
		u,
		ext.ReplyTextString(fmt.Sprintf("Лимит упоминаний в одном сообщении: %d", c.MentionsPerMessage)),
		nil,
	)
	return err
}

func (h *Handler) ShowCallTypes(ctx *command.Context, u *ext.Update) error {
	if u.CallbackQuery != nil {
		_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{QueryID: u.CallbackQuery.QueryID})
	}
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	_, err = ctx.Reply(
		u,
		ext.ReplyTextString("Настройте стиль упоминаний:"),
		&ext.ReplyOpts{
			Markup: h.getCallTypesKeyboard(c.MentionTypes),
		},
	)
	return err
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

	_, err = ctx.Reply(u, ext.ReplyTextString(view.FormatWelcomeCallMessageSet()), nil)
	return err
}

func (h *Handler) DeleteWelcomeCallMessage(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	if c.WelcomeCallMessage == "" {
		_, err := ctx.Reply(u, ext.ReplyTextString("Сообщение ещё не было установлено"), nil)

		return err
	}

	if err := h.service.SetWelcomeCallMessage(ctx.StdContext(), c.ID, ""); err != nil {
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextString("Сообщение удалено"), nil)
	return err

}

func (h *Handler) EnableCallOnJoin(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	if err := h.service.EnableCallOnJoin(ctx.StdContext(), c.ID); err != nil {
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextString(view.FormatCallOnJoinEnabled()), nil)
	return err
}

func (h *Handler) DisableCallOnJoin(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	if err := h.service.DisableCallOnJoin(ctx.StdContext(), c.ID); err != nil {
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextString(view.FormatCallOnJoinDisabled()), nil)
	return err
}

func (h *Handler) ShowWelcomeCallMessage(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextString(view.FormatWelcomeCallMessage(c.WelcomeCallMessage)), nil)
	return err
}

//
//func (h *Handler) startCallConversation(
//	b *gotgbot.Bot,
//	ctx *command.Context,
//	nextState string,
//) error {
//	c, err := ctx.Chat()
//	if err != nil {
//		return err
//	}
//	chatID := c.ID
//	m, err := ctx.Sender()
//	if err != nil {
//		return err
//	}
//	if !m.StatusGranted(ctx.RequiredStatus()) {
//		if ctx.CallbackQuery != nil {
//			_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
//				Text:      "Требуются права администратора",
//				ShowAlert: true,
//			})
//		}
//		return handlers.EndConversation()
//	}
//
//	callType := "всех"
//	switch nextState {
//	case CallStateInactive:
//		callType = "неактивных"
//	case CallStateNoNorm:
//		callType = "без нормы"
//	case CallStateNoNormWarn:
//		callType = "без нормы (предупреждение)"
//	case CallStateNoNormBan:
//		callType = "без нормы (бан)"
//	}
//
//	userMention := "Пользователь"
//	if ctx.EffectiveUser != nil {
//		userMention = helpers.Mention(ctx.EffectiveUser.Id, ctx.EffectiveUser.FirstName)
//	}
//
//	text := fmt.Sprintf(
//		"%s, введите сообщение созыва %s: ",
//		userMention,
//		callType,
//	)
//	promptMsg, err := ctx.EffectiveMessage.Reply(
//		b,
//		text,
//		&gotgbot.SendMessageOpts{
//			ParseMode: gotgbot.ParseModeHTML,
//			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
//				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
//					{
//						{
//							Text:         "Без сообщения",
//							Style:        "primary",
//							CallbackData: fmt.Sprintf("call_nomsg:%s", nextState),
//						}},
//					{{
//						Text:         "Отменить",
//						Style:        "danger",
//						CallbackData: "call_cancel",
//					},
//					},
//				},
//			},
//		},
//	)
//	if err != nil {
//		return err
//	}
//
//	if ctx.EffectiveSender != nil {
//		uid := ctx.EffectiveSender.Id()
//		h.mu.Lock()
//		if h.promptMessages[chatID] == nil {
//			h.promptMessages[chatID] = make(map[int64]int64)
//		}
//		h.promptMessages[chatID][uid] = promptMsg.MessageId
//		h.mu.Unlock()
//	}
//
//	if ctx.CallbackQuery != nil {
//		_, _ = ctx.CallbackQuery.Answer(b, nil)
//	}
//	log.Println("next state", nextState)
//	return handlers.NextConversationState(nextState)
//}
//
//func (h *Handler) StartCallInactiveConversation(b *gotgbot.Bot, ctx *command.Context) error {
//	return h.startCallConversation(b, ctx, CallStateInactive)
//}
//
//func (h *Handler) StartCallNoNormConversation(b *gotgbot.Bot, ctx *command.Context) error {
//	return h.startCallConversation(b, ctx, CallStateNoNorm)
//}
//
//func (h *Handler) StartCallNoNormWarnConversation(b *gotgbot.Bot, ctx *command.Context) error {
//	return h.startCallConversation(b, ctx, CallStateNoNormWarn)
//}
//
//func (h *Handler) StartCallNoNormBanConversation(b *gotgbot.Bot, ctx *command.Context) error {
//	return h.startCallConversation(b, ctx, CallStateNoNormBan)
//}
//
//func (h *Handler) handleCallWithMessage(
//	b *gotgbot.Bot,
//	ctx *command.Context,
//	getMembers func(stdCtx context.Context, chatID int64) ([]model.ChatMember, error),
//) error {
//	stdCtx := ctx.StdContext()
//
//	c, err := ctx.Chat()
//	if err != nil {
//		return err
//	}
//
//	if ctx.EffectiveSender != nil {
//		uid := ctx.EffectiveSender.Id()
//
//		var promptID int64
//		h.mu.Lock()
//		if byUser, ok := h.promptMessages[c.ID]; ok {
//			if mid, ok2 := byUser[uid]; ok2 {
//				promptID = mid
//				delete(byUser, uid)
//				if len(byUser) == 0 {
//					delete(h.promptMessages, c.ID)
//				}
//			}
//		}
//		h.mu.Unlock()
//
//		if promptID != 0 {
//			m := &gotgbot.Message{MessageId: promptID, Chat: gotgbot.Chat{Id: c.ID}}
//			if _, ok, errEdit := m.EditReplyMarkup(b, nil); errEdit != nil || !ok {
//				logger.L.Warn(
//					"failed to clear stored call prompt keyboard",
//					"error", errEdit,
//					"edited", ok,
//					"chat_id", c.ID,
//					"message_id", promptID,
//				)
//			}
//		}
//	}
//
//	members, err := getMembers(stdCtx, c.ID)
//	if err != nil {
//		return err
//	}
//
//	if len(members) == 0 {
//		_, err = ctx.EffectiveMessage.Reply(b, "Не найдено пользователей для созыва.", nil)
//		if err != nil {
//			return err
//		}
//		return handlers.EndConversation()
//	}
//
//	html := ctx.EffectiveMessage.OriginalHTML()
//
//	if err := h.doCall(stdCtx, b, c.ID, ctx.EffectiveMessage, html, members); err != nil {
//		return err
//	}
//
//	return handlers.EndConversation()
//}
//
//func (h *Handler) HandleCallInactiveMessage(b *gotgbot.Bot, ctx *command.Context) error {
//	return h.handleCallWithMessage(b, ctx, func(stdCtx context.Context, chatID int64) ([]model.ChatMember, error) {
//		return h.service.GetInactiveMembers(stdCtx, chatID)
//	})
//}
//
//func (h *Handler) HandleCallNoNormMessage(b *gotgbot.Bot, ctx *command.Context) error {
//	return h.handleCallWithMessage(
//		b,
//		ctx,
//		func(stdCtx context.Context, chatID int64) ([]model.ChatMember, error) {
//			c, err := h.chatService.GetChat(stdCtx, chatID)
//			if err != nil {
//				return nil, err
//			}
//
//			from, to := stats.ResolvePeriod(
//				stats.PeriodWeek,
//				time.Now(),
//				c.WeekStartDay,
//				c.WeekStartTime,
//			)
//
//			return h.memberService.GetNoNormMembers(stdCtx, chatID, from, to)
//		},
//	)
//}
//
//func (h *Handler) HandleCallNoNormWarnMessage(b *gotgbot.Bot, ctx *command.Context) error {
//	return h.handleCallWithMessage(
//		b,
//		ctx,
//		func(stdCtx context.Context, chatID int64) ([]model.ChatMember, error) {
//			c, err := h.chatService.GetChat(stdCtx, chatID)
//			if err != nil {
//				return nil, err
//			}
//
//			from, to := stats.ResolvePeriod(
//				stats.PeriodWeek,
//				time.Now(),
//				c.WeekStartDay,
//				c.WeekStartTime,
//			)
//
//			return h.memberService.GetNoNormWarnMembers(stdCtx, chatID, from, to)
//		},
//	)
//}
//
//func (h *Handler) HandleCallNoNormBanMessage(b *gotgbot.Bot, ctx *command.Context) error {
//	return h.handleCallWithMessage(
//		b,
//		ctx,
//		func(stdCtx context.Context, chatID int64) ([]model.ChatMember, error) {
//			c, err := h.chatService.GetChat(stdCtx, chatID)
//			if err != nil {
//				return nil, err
//			}
//
//			from, to := stats.ResolvePeriod(
//				stats.PeriodWeek,
//				time.Now(),
//				c.WeekStartDay,
//				c.WeekStartTime,
//			)
//
//			return h.memberService.GetNoNormBanMembers(stdCtx, chatID, from, to)
//		},
//	)
//}
//
//func (h *Handler) CancelCallConversation(b *gotgbot.Bot, ctx *command.Context) error {
//	if ctx.CallbackQuery != nil && ctx.CallbackQuery.Message != nil {
//		if _, _, err := ctx.CallbackQuery.Message.EditText(
//			b,
//			"❌ Операция созыва отменена.", nil,
//		); err != nil {
//			logger.L.Error("Failed to edit cancel call prompt", "error", err)
//		}
//	}
//
//	if ctx.CallbackQuery != nil {
//		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
//			Text: "Созыв отменён",
//		})
//	}
//	return handlers.EndConversation()
//}
//
//func (h *Handler) NoMessageCallConversation(b *gotgbot.Bot, ctx *command.Context) error {
//	stdCtx := context.Background()
//
//	c, err := ctx.Chat()
//	if err != nil {
//		return err
//	}
//
//	state := ""
//	if ctx.CallbackQuery != nil {
//		data := ctx.CallbackQuery.Data
//		const prefix = "call_nomsg:"
//		if len(data) > len(prefix) && data[:len(prefix)] == prefix {
//			state = data[len(prefix):]
//		}
//	}
//
//	var members []model.ChatMember
//	switch state {
//	case CallStateInactive:
//		members, err = h.service.GetInactiveMembers(stdCtx, c.ID)
//	case CallStateNoNorm:
//		c, gErr := h.chatService.GetChat(stdCtx, c.ID)
//		if gErr != nil {
//			return gErr
//		}
//		from, to := stats.ResolvePeriod(
//			stats.PeriodWeek,
//			time.Now(),
//			c.WeekStartDay,
//			c.WeekStartTime,
//		)
//		members, err = h.memberService.GetNoNormMembers(stdCtx, c.ID, from, to)
//	case CallStateNoNormWarn, CallStateNoNormBan:
//		c, gErr := h.chatService.GetChat(stdCtx, c.ID)
//		if gErr != nil {
//			return gErr
//		}
//		from, to := stats.ResolvePeriod(
//			stats.PeriodWeek,
//			time.Now(),
//			c.WeekStartDay,
//			c.WeekStartTime,
//		)
//		members, err = h.memberService.GetNoNormWarnMembers(stdCtx, c.ID, from, to)
//	default:
//		return handlers.EndConversation()
//	}
//
//	if err != nil {
//		return err
//	}
//
//	if len(members) == 0 {
//		_, err = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
//			Text: "Нет участников для созыва.",
//		})
//		if err != nil {
//			return err
//		}
//		return handlers.EndConversation()
//	}
//
//	if err := h.doCall(stdCtx, b, c.ID, ctx.EffectiveMessage, "", members); err != nil {
//		return err
//	}
//
//	if ctx.CallbackQuery != nil {
//		if ctx.CallbackQuery.Message != nil {
//			if _, _, err := ctx.CallbackQuery.Message.EditReplyMarkup(
//				b,
//				&gotgbot.EditMessageReplyMarkupOpts{},
//			); err != nil {
//				logger.L.Warn("failed to clear keyboard on no-message call", "error", err)
//			}
//		}
//
//		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
//			Text: "Созыв отправлен без сообщения.",
//		})
//	}
//
//	return handlers.EndConversation()
//}
