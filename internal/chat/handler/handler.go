package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/rest"
	"fmt"
	"strconv"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	service      *chat.Service
	adminService *admin.Service
	dateParser   *rest.DateParser
}

func New(service *chat.Service, adminService *admin.Service, dateParser *rest.DateParser) *Handler {
	return &Handler{service, adminService, dateParser}
}

func (h *Handler) ShowNorm(b *gotgbot.Bot, ctx *ext.Context, _ *cmd.Context) error {
	norm, err := h.service.GetNorm(ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Норма чата: %d сообщений", norm), nil)

	return err
}

func (h *Handler) SetNorm(b *gotgbot.Bot, ctx *ext.Context, cctx *cmd.Context) error {
	norm, err := strconv.Atoi(cctx.FirstArgument())
	if err != nil {
		return err
	}

	if err := h.service.SetNorm(ctx.EffectiveChat.Id, norm); err != nil {
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Установлена новая норма чата: %d", norm), nil)

	return err
}

func (h *Handler) ShowNewbieThreshold(b *gotgbot.Bot, ctx *ext.Context, _ *cmd.Context) error {
	threshold, err := h.service.GetNewbieThreshold(ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Пользователи считаются новичками первые %d %s", threshold, helpers.PluralizeDays(threshold)), nil)
	return err
}

func (h *Handler) SetNewbieThreshold(b *gotgbot.Bot, ctx *ext.Context, cctx *cmd.Context) error {
	days, ok := h.dateParser.ParseDays(cctx.FirstArgument())
	if !ok {
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось распознать срок. Используйте формат: 3 дня, неделя, 14 дней или просто число.", nil)
		return err
	}

	if err := h.service.SetNewbieThreshold(ctx.EffectiveChat.Id, days); err != nil {
		return err
	}

	_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Теперь пользователи считаются новичками первые %d %s", days, helpers.PluralizeDays(days)), nil)
	return err
}

func (h *Handler) SetPrompt(b *gotgbot.Bot, ctx *ext.Context, cctx *cmd.Context) error {
	if err := h.service.SetChatPrompt(ctx.EffectiveChat.Id, cctx.FirstArgument()); err != nil {
		return err
	}

	_, err := ctx.EffectiveMessage.Reply(b, "Промпт установлен успешно", nil)
	return err
}

func (h *Handler) ShowPrompt(b *gotgbot.Bot, ctx *ext.Context, _ *cmd.Context) error {
	c, err := h.service.GetChat(ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Промпт: \"%s\"", c.GeminiSystemPrompt), nil)
	return err

}
