package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/call"
	"activity-bot/internal/call/view"
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"context"
	"fmt"
	"strconv"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	service      *call.Service
	chatService  *chat.Service
	adminService *admin.Service
}

func New(service *call.Service, chatService *chat.Service, adminService *admin.Service) *Handler {
	return &Handler{service, chatService, adminService}
}

func (h *Handler) Call(b *gotgbot.Bot, ctx *cmd.Context) error {
	return h.service.Call(ctx.StdContext(), b, ctx.Context, ctx.FirstArgument())
}

func (h *Handler) SetMentionsPerMessage(b *gotgbot.Bot, ctx *cmd.Context) error {
	countStr := ctx.FirstArgument()
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 || count > 100 {
		return ctx.Reply(b, "Укажите число от 1 до 100", nil)
	}

	if err := h.service.SetMentionsPerMessage(ctx.StdContext(), ctx.EffectiveChat.Id, int32(count)); err != nil {
		return err
	}

	return ctx.Reply(b, fmt.Sprintf("Лимит упоминаний в одном сообщении изменен на %d", count), nil)
}

func (h *Handler) ShowCallTypes(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.chatService.GetChat(ctx.StdContext(), ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	return ctx.Reply(b, "Выберите типы упоминаний для команды call:", &gotgbot.SendMessageOpts{
		ReplyMarkup: h.getCallTypesKeyboard(int32(c.MentionTypes)),
	})
}

func (h *Handler) CallbackCallType(b *gotgbot.Bot, ctx *ext.Context) error {
	isAdmin, err := h.adminService.IsAdmin(context.Background(), ctx.EffectiveChat.Id, ctx.EffectiveSender.Id())
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

	c, err := h.chatService.GetChat(context.Background(), ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	newTypes := int32(c.MentionTypes) ^ bit
	if err := h.service.SetMentionTypes(context.Background(), ctx.EffectiveChat.Id, newTypes); err != nil {
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
		{"Эмодзи", call.MentionTypeEmoji},
		{"Имя", call.MentionTypeName},
		{"ТГ Роль", call.MentionTypeRole},
	}

	var rows [][]gotgbot.InlineKeyboardButton
	for _, t := range types {
		status := ""
		if currentTypes&t.bit > 0 {
			status = "primary"
		}
		rows = append(rows, []gotgbot.InlineKeyboardButton{{
			Text:         fmt.Sprintf("%s", t.name),
			CallbackData: fmt.Sprintf("call_type:%d", t.bit),
			Style:        status,
		}})
	}

	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func (h *Handler) SetWelcomeCallMessage(b *gotgbot.Bot, ctx *cmd.Context) error {
	message := ctx.FirstArgument()
	if err := h.service.SetWelcomeCallMessage(ctx.StdContext(), ctx.EffectiveChat.Id, message); err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatWelcomeCallMessageSet(), nil)
}

func (h *Handler) EnableCallOnJoin(b *gotgbot.Bot, ctx *cmd.Context) error {
	if err := h.service.EnableCallOnJoin(ctx.StdContext(), ctx.EffectiveChat.Id); err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatCallOnJoinEnabled(), nil)
}

func (h *Handler) DisableCallOnJoin(b *gotgbot.Bot, ctx *cmd.Context) error {
	if err := h.service.DisableCallOnJoin(ctx.StdContext(), ctx.EffectiveChat.Id); err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatCallOnJoinDisabled(), nil)
}

func (h *Handler) ShowWelcomeCallMessage(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.chatService.GetChat(ctx.StdContext(), ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}
	return ctx.Reply(b, view.FormatWelcomeCallMessage(c.WelcomeCallMessage), nil)
}
