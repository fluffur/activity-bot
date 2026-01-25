package chat

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"fmt"
	"log"
	"log/slog"
	"strconv"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	service      *Service
	adminService *admin.Service
}

func NewHandler(service *Service, adminService *admin.Service) *Handler {
	return &Handler{service, adminService}
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
	days, err := strconv.Atoi(cctx.Args[0])
	if err != nil {
		slog.Error("failed to parse newbie days", "error", err)
		_, err = ctx.EffectiveMessage.Reply(b, "Количество дней должно быть числом", nil)
		return err
	}

	if err := h.service.SetNewbieThreshold(ctx.EffectiveChat.Id, days); err != nil {
		slog.Error("failed to set newbie threshold", "chat_id", ctx.EffectiveChat.Id, "error", err)
		_, err = ctx.EffectiveMessage.Reply(b, "Не удалось установить срок для новичков", nil)
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Теперь пользователи считаются новичками первые %d %s", days, helpers.PluralizeDays(days)), nil)
	return err
}
