package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/chat"
	"activity-bot/internal/command"
	"activity-bot/internal/exempt"
	"activity-bot/internal/helpers"
	"fmt"
	"log"
	"log/slog"
	"strconv"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	service      *chat.Service
	adminService *admin.Service
	dateParser   *exempt.DateParser
}

func New(service *chat.Service, adminService *admin.Service, dateParser *exempt.DateParser) *Handler {
	return &Handler{service, adminService, dateParser}
}

func (h *Handler) ShowNorm(b *gotgbot.Bot, ctx *ext.Context, _ *command.Context) error {
	norm, err := h.service.GetNorm(ctx.EffectiveChat.Id)
	if err != nil {
		log.Println("Failed to show chat norm", err)
		_, err = ctx.EffectiveMessage.Reply(b, "Не удалось отправить норму чата", nil)

		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Норма чата: %d сообщений", norm), nil)

	return err
}

func (h *Handler) SetNorm(b *gotgbot.Bot, ctx *ext.Context, cctx *command.Context) error {
	norm, err := strconv.Atoi(cctx.Args[0])
	if err != nil {
		log.Println("Failed to set chat norm", err)
		_, err = ctx.EffectiveMessage.Reply(b, "Норма должна быть числом", nil)

		return err
	}

	if err := h.service.SetNorm(ctx.EffectiveChat.Id, norm); err != nil {
		log.Println("Failed to set chat norm", err)
		_, err = ctx.EffectiveMessage.Reply(b, "Не удалось установить норму чата", nil)

		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Установлена новая норма чата: %d", norm), nil)

	return err
}

func (h *Handler) ShowNewbieThreshold(b *gotgbot.Bot, ctx *ext.Context, _ *command.Context) error {
	threshold, err := h.service.GetNewbieThreshold(ctx.EffectiveChat.Id)
	if err != nil {
		slog.Error("failed to set newbie threshold", "chat_id", ctx.EffectiveChat.Id, "error", err)
		_, err = ctx.EffectiveMessage.Reply(b, "Не удалось установить срок для новичков", nil)
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Пользователи считаются новичками первые %d %s", threshold, helpers.PluralizeDays(threshold)), nil)
	return err
}

func (h *Handler) SetNewbieThreshold(b *gotgbot.Bot, ctx *ext.Context, cctx *command.Context) error {
	days, ok := h.dateParser.ParseDays(cctx.Args[0])
	if !ok {
		slog.Warn("failed to parse newbie days", "arg", cctx.Args[0])
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось распознать срок. Используйте формат: 3 дня, неделя, 14 дней или просто число.", nil)
		return err
	}

	if err := h.service.SetNewbieThreshold(ctx.EffectiveChat.Id, days); err != nil {
		slog.Error("failed to set newbie threshold", "chat_id", ctx.EffectiveChat.Id, "error", err)
		_, err = ctx.EffectiveMessage.Reply(b, "Не удалось установить срок для новичков", nil)
		return err
	}

	_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Теперь пользователи считаются новичками первые %d %s", days, helpers.PluralizeDays(days)), nil)
	return err
}

func (h *Handler) SetOnlyNewbies(b *gotgbot.Bot, ctx *ext.Context, cctx *command.Context) error {
	if len(cctx.Users) == 0 {
		_, err := ctx.EffectiveMessage.Reply(b, "Укажите хотя бы одного юзера", nil)

		return err
	}
	if err := h.service.SetOnlyNewbies(ctx.EffectiveChat.Id, cctx.Users); err != nil {
		log.Println("failed to set only-newbies", err)
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось установить олдов", nil)
		return err
	}

	_, err := ctx.EffectiveMessage.Reply(b, "Олды установлены", nil)

	return err
}
