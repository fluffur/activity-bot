package chat

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/command"
	"fmt"
	"log"
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
