package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/call"
	"activity-bot/internal/call/view"
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/session"
	"fmt"
	"strconv"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Handler struct {
	service        *call.Service
	chatService    *chat.Service
	adminService   *admin.Service
	sessionService *session.Service
}

func New(service *call.Service, chatService *chat.Service, adminService *admin.Service, sessionService *session.Service) *Handler {
	return &Handler{service, chatService, adminService, sessionService}
}

func (h *Handler) Call(b *gotgbot.Bot, ctx *cmd.Context) error {
	return h.service.CallAll(ctx, b, ctx.HTML())
}

func (h *Handler) CallInactive(b *gotgbot.Bot, ctx *cmd.Context) error {
	return h.service.CallInactive(ctx, b, ctx.HTML())
}

func (h *Handler) SetMentionsPerMessage(b *gotgbot.Bot, ctx *cmd.Context) error {
	countStr := ctx.FirstArgument()
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 || count > 100 {
		return ctx.Reply(b, "Укажите число от 1 до 100", nil)
	}

	if err := h.service.SetMentionsPerMessage(ctx.StdContext(), ctx.TargetChatID(), int32(count)); err != nil {
		return err
	}

	return ctx.Reply(b, fmt.Sprintf("Лимит упоминаний в одном сообщении изменен на %d", count), nil)
}

func (h *Handler) ShowCallTypes(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.chatService.GetChat(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	return ctx.Reply(b, "Выберите типы упоминаний для команды call:", &gotgbot.SendMessageOpts{
		ReplyMarkup: h.getCallTypesKeyboard(int32(c.MentionTypes)),
	})
}

func (h *Handler) CallbackCallType(b *gotgbot.Bot, ctx *cmd.Context) error {
	chatID := ctx.TargetChatID()
	isAdmin, err := h.adminService.IsAdmin(ctx.StdContext(), chatID, ctx.EffectiveSender.Id())
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

	if bit == call.MentionTypeNWSP {
		newTypes = call.MentionTypeNWSP
	} else {
		current &^= call.MentionTypeNWSP
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

	_, _, err = ctx.EffectiveMessage.EditReplyMarkup(b, &gotgbot.EditMessageReplyMarkupOpts{
		ReplyMarkup: h.getCallTypesKeyboard(newTypes),
	})
	if err != nil {
		return err
	}

	_, err = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Настройки обновлены"})
	return err
}

func (h *Handler) getCallTypesKeyboard(currentTypes int32) gotgbot.InlineKeyboardMarkup {
	types := []struct {
		name string
		bit  int32
	}{
		{"Пустота", call.MentionTypeNWSP},
		{"Эмодзи", call.MentionTypeEmoji},
		{"Имя", call.MentionTypeName},
		{"Роль", call.MentionTypeRole},
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
