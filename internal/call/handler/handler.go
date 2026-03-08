package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/call"
	"activity-bot/internal/call/view"
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/logger"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"activity-bot/internal/session"
	"activity-bot/internal/stats"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
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

func (h *Handler) callMembers(
	b *gotgbot.Bot,
	ctx *cmd.Context,
	getMembers func() ([]model.ChatMember, error),
	emptyMsg string,
) error {

	members, err := getMembers()
	if err != nil {
		return err
	}

	if len(members) == 0 {
		return ctx.Reply(b, emptyMsg, nil)
	}

	return h.handleCall(b, ctx, members)
}

func (h *Handler) adminCallback(
	b *gotgbot.Bot,
	ctx *cmd.Context,
	handler func(*gotgbot.Bot, *cmd.Context) error,
) error {

	isAdmin, err := h.adminService.IsAdmin(
		ctx.StdContext(),
		ctx.EffectiveChat.Id,
		ctx.EffectiveSender.Id(),
	)
	if err != nil {
		return err
	}

	if !isAdmin {
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Требуются права администратора",
		})
		return err
	}

	_, _ = ctx.CallbackQuery.Answer(b, nil)

	return handler(b, ctx)
}

func (h *Handler) Call(b *gotgbot.Bot, ctx *cmd.Context) error {
	return h.callMembers(
		b,
		ctx,
		func() ([]model.ChatMember, error) {
			return h.memberService.GetChatMembers(ctx.StdContext(), ctx.TargetChatID())
		},
		"Не найдено пользователей для созыва, скорее всего бот был добавлен недавно и понадобится время, чтобы он успел познакомиться со всеми участниками!",
	)
}

func (h *Handler) CallInactive(b *gotgbot.Bot, ctx *cmd.Context) error {
	return h.callMembers(
		b,
		ctx,
		func() ([]model.ChatMember, error) {
			return h.service.GetInactiveMembers(ctx.StdContext(), ctx.TargetChatID())
		},
		"Нет участников, не писавших более суток",
	)
}

func (h *Handler) CallNoNorm(b *gotgbot.Bot, ctx *cmd.Context) error {
	chatID := ctx.TargetChatID()

	c, err := h.chatService.GetChat(ctx.StdContext(), chatID)
	if err != nil {
		return err
	}

	from, to := stats.ResolvePeriod(
		stats.PeriodWeek,
		time.Now().In(helpers.MoscowLocation),
		c.WeekStartDay,
		c.WeekStartTime,
	)

	return h.callMembers(
		b,
		ctx,
		func() ([]model.ChatMember, error) {
			return h.memberService.GetNoNormMembers(ctx.StdContext(), chatID, from, to)
		},
		"✅ Все участники выполнили норму!",
	)
}

func (h *Handler) CallNoNormWarn(b *gotgbot.Bot, ctx *cmd.Context) error {
	chatID := ctx.TargetChatID()

	c, err := h.chatService.GetChat(ctx.StdContext(), chatID)
	if err != nil {
		return err
	}

	from, to := stats.ResolvePeriod(
		stats.PeriodWeek,
		time.Now().In(helpers.MoscowLocation),
		c.WeekStartDay,
		c.WeekStartTime,
	)

	return h.callMembers(
		b,
		ctx,
		func() ([]model.ChatMember, error) {
			return h.memberService.GetNoNormWarnMembers(ctx.StdContext(), chatID, from, to)
		},
		"✅ Все участники выполнили норму предупреждения!",
	)
}

func (h *Handler) CallInactiveCallback(b *gotgbot.Bot, ctx *cmd.Context) error {
	return h.adminCallback(b, ctx, h.CallInactive)
}

func (h *Handler) CallNoNormCallback(b *gotgbot.Bot, ctx *cmd.Context) error {
	return h.adminCallback(b, ctx, h.CallNoNorm)
}

func (h *Handler) CallNoNormWarnCallback(b *gotgbot.Bot, ctx *cmd.Context) error {
	return h.adminCallback(b, ctx, h.CallNoNormWarn)
}

func (h *Handler) handleCall(b *gotgbot.Bot, ctx *cmd.Context, members []model.ChatMember) error {

	return h.doCall(ctx.StdContext(), b, ctx.TargetChatID(), ctx.EffectiveMessage, ctx.HTML(), members)
}

func (h *Handler) doCall(
	stdCtx context.Context,
	b *gotgbot.Bot,
	chatID int64,
	srcMsg *gotgbot.Message,
	htmlMessage string,
	members []model.ChatMember,
) error {

	var replyParams *gotgbot.ReplyParameters
	if srcMsg != nil {
		replyParams = &gotgbot.ReplyParameters{
			ChatId:    chatID,
			MessageId: srcMsg.MessageId,
		}
	}

	chatSettings, err := h.service.GetChatSettings(stdCtx, chatID)
	if err != nil {
		return err
	}

	mentionsLimit := int(chatSettings.MentionsPerMessage)
	if mentionsLimit <= 0 {
		mentionsLimit = 5
	}

	message := htmlMessage
	if message == "" {
		message = chatSettings.WelcomeCallMessage
	}

	if message != "" {
		message = view.ReplaceMentionsWithLinks(message)
	}

	for i := 0; i < len(members); i += mentionsLimit {

		end := i + mentionsLimit
		if end > len(members) {
			end = len(members)
		}

		chunkText := view.FormatCallChunk(message, members[i:end], chatSettings.MentionTypes)

		if srcMsg != nil && len(srcMsg.Photo) > 0 {

			lastPhoto := srcMsg.Photo[len(srcMsg.Photo)-1]

			if _, err := b.SendPhoto(
				chatID,
				gotgbot.InputFileByID(lastPhoto.FileId),
				&gotgbot.SendPhotoOpts{
					ParseMode:       gotgbot.ParseModeHTML,
					Caption:         chunkText,
					HasSpoiler:      srcMsg.HasMediaSpoiler,
					ReplyParameters: replyParams,
				},
			); err != nil {
				return err
			}

		} else {

			if _, err := b.SendMessage(
				chatID,
				chunkText,
				&gotgbot.SendMessageOpts{
					ParseMode: gotgbot.ParseModeHTML,
					LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
						IsDisabled: true,
					},
					ReplyParameters: replyParams,
				},
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *Handler) SetMentionsPerMessage(b *gotgbot.Bot, ctx *cmd.Context) error {

	countStr := ctx.FirstArgument()

	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 || count > 100 {
		return ctx.Reply(b, "Укажите число от 1 до 100", nil)
	}

	if err := h.service.SetMentionsPerMessage(
		ctx.StdContext(),
		ctx.TargetChatID(),
		int32(count),
	); err != nil {
		return err
	}

	return ctx.Reply(
		b,
		fmt.Sprintf("Лимит упоминаний в одном сообщении изменен на %d", count),
		nil,
	)
}

func (h *Handler) ShowCallTypes(b *gotgbot.Bot, ctx *cmd.Context) error {

	c, err := h.chatService.GetChat(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	return ctx.Reply(
		b,
		"Выберите типы упоминаний для команды call:",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: h.getCallTypesKeyboard(int32(c.MentionTypes)),
		},
	)
}

func (h *Handler) CallbackCallType(b *gotgbot.Bot, ctx *cmd.Context) error {

	chatID := ctx.TargetChatID()

	isAdmin, err := h.adminService.IsAdmin(
		ctx.StdContext(),
		chatID,
		ctx.EffectiveSender.Id(),
	)
	if err != nil {
		return err
	}

	if !isAdmin {
		_, err = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "У вас нет прав администратора для выполнения действия",
		})
		return err
	}

	var bit int32
	if _, err := fmt.Sscanf(ctx.CallbackQuery.Data, "call_type:%d", &bit); err != nil {
		return err
	}

	c, err := h.chatService.GetChat(ctx.StdContext(), chatID)
	if err != nil {
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
		_, err = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Настройки не изменились",
		})
		return err
	}

	if err := h.service.SetMentionTypes(ctx.StdContext(), chatID, newTypes); err != nil {
		return err
	}

	_, _, err = ctx.EffectiveMessage.EditReplyMarkup(
		b,
		&gotgbot.EditMessageReplyMarkupOpts{
			ReplyMarkup: h.getCallTypesKeyboard(newTypes),
		},
	)
	if err != nil {
		return err
	}

	_, err = ctx.CallbackQuery.Answer(
		b,
		&gotgbot.AnswerCallbackQueryOpts{
			Text: "Настройки обновлены",
		},
	)

	return err
}

func (h *Handler) getCallTypesKeyboard(currentTypes int32) gotgbot.InlineKeyboardMarkup {

	types := []struct {
		name string
		bit  int32
	}{
		{"Пустота", view.MentionTypeNWSP},
		{"Эмодзи", view.MentionTypeEmoji},
		{"Имя", view.MentionTypeName},
		{"Роль", view.MentionTypeRole},
	}

	var rows [][]gotgbot.InlineKeyboardButton
	var row []gotgbot.InlineKeyboardButton

	for i, t := range types {

		status := ""
		checkMark := ""

		if currentTypes&t.bit > 0 {
			status = "primary"
			checkMark = "✅ "
		}

		btn := gotgbot.InlineKeyboardButton{
			Text:         fmt.Sprintf("%s%s", checkMark, t.name),
			CallbackData: fmt.Sprintf("call_type:%d", t.bit),
			Style:        status,
		}

		row = append(row, btn)

		if (i+1)%2 == 0 {
			rows = append(rows, row)
			row = nil
		}
	}

	if len(row) > 0 {
		rows = append(rows, row)
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}

func (h *Handler) SetWelcomeCallMessage(b *gotgbot.Bot, ctx *cmd.Context) error {
	message := ctx.FirstArgument()
	if err := h.service.SetWelcomeCallMessage(ctx.StdContext(), ctx.TargetChatID(), message); err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatWelcomeCallMessageSet(), nil)
}

func (h *Handler) EnableCallOnJoin(b *gotgbot.Bot, ctx *cmd.Context) error {
	if err := h.service.EnableCallOnJoin(ctx.StdContext(), ctx.TargetChatID()); err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatCallOnJoinEnabled(), nil)
}

func (h *Handler) DisableCallOnJoin(b *gotgbot.Bot, ctx *cmd.Context) error {
	if err := h.service.DisableCallOnJoin(ctx.StdContext(), ctx.TargetChatID()); err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatCallOnJoinDisabled(), nil)
}

func (h *Handler) ShowWelcomeCallMessage(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.chatService.GetChat(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}
	return ctx.Reply(b, view.FormatWelcomeCallMessage(c.WelcomeCallMessage), nil)
}

func (h *Handler) startCallConversation(
	b *gotgbot.Bot,
	ctx *ext.Context,
	nextState string,
) error {
	stdCtx := context.Background()

	chatID, err := cmd.GetChatID(h.sessionService, ctx, stdCtx)
	if err != nil {
		chatID = ctx.EffectiveChat.Id
	}

	isAdmin, err := h.adminService.IsAdmin(stdCtx, chatID, ctx.EffectiveSender.Id())
	if err != nil {
		return err
	}

	if !isAdmin {
		if ctx.CallbackQuery != nil {
			_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
				Text:      "Требуются права администратора",
				ShowAlert: true,
			})
		}
		return handlers.EndConversation()
	}

	// Человекочитаемое имя типа созыва.
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

	userMention := "Пользователь"
	if ctx.EffectiveUser != nil {
		userMention = helpers.Mention(ctx.EffectiveUser.Id, ctx.EffectiveUser.FirstName)
	}

	text := fmt.Sprintf(
		"%s, введите сообщение созыва типа %s: ",
		userMention,
		callType,
	)
	promptMsg, err := ctx.EffectiveMessage.Reply(
		b,
		text,
		&gotgbot.SendMessageOpts{
			ParseMode: gotgbot.ParseModeHTML,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{
							Text:         "Без сообщения",
							Style:        "primary",
							CallbackData: fmt.Sprintf("call_nomsg:%s", nextState),
						},
						{
							Text:         "Отменить",
							Style:        "danger",
							CallbackData: "call_cancel",
						},
					},
				},
			},
		},
	)
	if err != nil {
		return err
	}

	if ctx.EffectiveSender != nil {
		uid := ctx.EffectiveSender.Id()
		h.mu.Lock()
		if h.promptMessages[chatID] == nil {
			h.promptMessages[chatID] = make(map[int64]int64)
		}
		h.promptMessages[chatID][uid] = promptMsg.MessageId
		h.mu.Unlock()
	}

	if ctx.CallbackQuery != nil {
		_, _ = ctx.CallbackQuery.Answer(b, nil)
	}

	return handlers.NextConversationState(nextState)
}

func (h *Handler) StartCallInactiveConversation(b *gotgbot.Bot, ctx *ext.Context) error {
	return h.startCallConversation(b, ctx, CallStateInactive)
}

func (h *Handler) StartCallNoNormConversation(b *gotgbot.Bot, ctx *ext.Context) error {
	return h.startCallConversation(b, ctx, CallStateNoNorm)
}

func (h *Handler) StartCallNoNormWarnConversation(b *gotgbot.Bot, ctx *ext.Context) error {
	return h.startCallConversation(b, ctx, CallStateNoNormWarn)
}

func (h *Handler) StartCallNoNormBanConversation(b *gotgbot.Bot, ctx *ext.Context) error {
	return h.startCallConversation(b, ctx, CallStateNoNormBan)
}

func (h *Handler) handleCallWithMessage(
	b *gotgbot.Bot,
	ctx *ext.Context,
	getMembers func(stdCtx context.Context, chatID int64) ([]model.ChatMember, error),
) error {
	stdCtx := context.Background()

	chatID, err := cmd.GetChatID(h.sessionService, ctx, stdCtx)
	if err != nil {
		chatID = ctx.EffectiveChat.Id
	}

	if ctx.EffectiveSender != nil {
		uid := ctx.EffectiveSender.Id()

		var promptID int64
		h.mu.Lock()
		if byUser, ok := h.promptMessages[chatID]; ok {
			if mid, ok2 := byUser[uid]; ok2 {
				promptID = mid
				delete(byUser, uid)
				if len(byUser) == 0 {
					delete(h.promptMessages, chatID)
				}
			}
		}
		h.mu.Unlock()

		if promptID != 0 {
			m := &gotgbot.Message{MessageId: promptID, Chat: gotgbot.Chat{Id: chatID}}
			if _, ok, errEdit := m.EditReplyMarkup(b, nil); errEdit != nil || !ok {
				logger.L.Warn(
					"failed to clear stored call prompt keyboard",
					"error", errEdit,
					"edited", ok,
					"chat_id", chatID,
					"message_id", promptID,
				)
			}
		}
	}

	members, err := getMembers(stdCtx, chatID)
	if err != nil {
		return err
	}

	if len(members) == 0 {
		_, err = ctx.EffectiveMessage.Reply(b, "Не найдено пользователей для созыва.", nil)
		if err != nil {
			return err
		}
		return handlers.EndConversation()
	}

	html := ctx.EffectiveMessage.OriginalHTML()

	if err := h.doCall(stdCtx, b, chatID, ctx.EffectiveMessage, html, members); err != nil {
		return err
	}

	return handlers.EndConversation()
}

func (h *Handler) HandleCallInactiveMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	return h.handleCallWithMessage(b, ctx, func(stdCtx context.Context, chatID int64) ([]model.ChatMember, error) {
		return h.service.GetInactiveMembers(stdCtx, chatID)
	})
}

func (h *Handler) HandleCallNoNormMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	return h.handleCallWithMessage(
		b,
		ctx,
		func(stdCtx context.Context, chatID int64) ([]model.ChatMember, error) {
			c, err := h.chatService.GetChat(stdCtx, chatID)
			if err != nil {
				return nil, err
			}

			from, to := stats.ResolvePeriod(
				stats.PeriodWeek,
				time.Now().In(helpers.MoscowLocation),
				c.WeekStartDay,
				c.WeekStartTime,
			)

			return h.memberService.GetNoNormMembers(stdCtx, chatID, from, to)
		},
	)
}

func (h *Handler) HandleCallNoNormWarnMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	return h.handleCallWithMessage(
		b,
		ctx,
		func(stdCtx context.Context, chatID int64) ([]model.ChatMember, error) {
			c, err := h.chatService.GetChat(stdCtx, chatID)
			if err != nil {
				return nil, err
			}

			from, to := stats.ResolvePeriod(
				stats.PeriodWeek,
				time.Now().In(helpers.MoscowLocation),
				c.WeekStartDay,
				c.WeekStartTime,
			)

			return h.memberService.GetNoNormWarnMembers(stdCtx, chatID, from, to)
		},
	)
}

func (h *Handler) HandleCallNoNormBanMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	return h.handleCallWithMessage(
		b,
		ctx,
		func(stdCtx context.Context, chatID int64) ([]model.ChatMember, error) {
			c, err := h.chatService.GetChat(stdCtx, chatID)
			if err != nil {
				return nil, err
			}

			from, to := stats.ResolvePeriod(
				stats.PeriodWeek,
				time.Now().In(helpers.MoscowLocation),
				c.WeekStartDay,
				c.WeekStartTime,
			)

			return h.memberService.GetNoNormBanMembers(stdCtx, chatID, from, to)
		},
	)
}

func (h *Handler) CancelCallConversation(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.CallbackQuery != nil && ctx.CallbackQuery.Message != nil {
		if _, _, err := ctx.CallbackQuery.Message.EditText(
			b,
			"❌ Операция созыва отменена.", nil,
		); err != nil {
			logger.L.Error("Failed to edit cancel call prompt", "error", err)
		}
	}

	if ctx.CallbackQuery != nil {
		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Созыв отменён",
		})
	}
	return handlers.EndConversation()
}

func (h *Handler) NoMessageCallConversation(b *gotgbot.Bot, ctx *ext.Context) error {
	stdCtx := context.Background()

	chatID, err := cmd.GetChatID(h.sessionService, ctx, stdCtx)
	if err != nil {
		chatID = ctx.EffectiveChat.Id
	}

	state := ""
	if ctx.CallbackQuery != nil {
		data := ctx.CallbackQuery.Data
		const prefix = "call_nomsg:"
		if len(data) > len(prefix) && data[:len(prefix)] == prefix {
			state = data[len(prefix):]
		}
	}

	var members []model.ChatMember
	switch state {
	case CallStateInactive:
		members, err = h.service.GetInactiveMembers(stdCtx, chatID)
	case CallStateNoNorm:
		c, gErr := h.chatService.GetChat(stdCtx, chatID)
		if gErr != nil {
			return gErr
		}
		from, to := stats.ResolvePeriod(
			stats.PeriodWeek,
			time.Now().In(helpers.MoscowLocation),
			c.WeekStartDay,
			c.WeekStartTime,
		)
		members, err = h.memberService.GetNoNormMembers(stdCtx, chatID, from, to)
	case CallStateNoNormWarn, CallStateNoNormBan:
		c, gErr := h.chatService.GetChat(stdCtx, chatID)
		if gErr != nil {
			return gErr
		}
		from, to := stats.ResolvePeriod(
			stats.PeriodWeek,
			time.Now().In(helpers.MoscowLocation),
			c.WeekStartDay,
			c.WeekStartTime,
		)
		members, err = h.memberService.GetNoNormWarnMembers(stdCtx, chatID, from, to)
	default:
		return handlers.EndConversation()
	}

	if err != nil {
		return err
	}

	if len(members) == 0 {
		_, err = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Нет участников для созыва.",
		})
		if err != nil {
			return err
		}
		return handlers.EndConversation()
	}

	if err := h.doCall(stdCtx, b, chatID, ctx.EffectiveMessage, "", members); err != nil {
		return err
	}

	if ctx.CallbackQuery != nil {
		if ctx.CallbackQuery.Message != nil {
			if _, _, err := ctx.CallbackQuery.Message.EditReplyMarkup(
				b,
				&gotgbot.EditMessageReplyMarkupOpts{},
			); err != nil {
				logger.L.Warn("failed to clear keyboard on no-message call", "error", err)
			}
		}

		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Созыв отправлен без сообщения.",
		})
	}

	return handlers.EndConversation()
}
