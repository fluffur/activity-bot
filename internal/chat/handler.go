package chat

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/helpers"
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Handler struct {
	service      *Service
	adminService *admin.Service
	setNormRe    *regexp.Regexp
}

func NewHandler(service *Service, adminService *admin.Service, setNormRe *regexp.Regexp) *Handler {
	return &Handler{service, adminService, setNormRe}
}

func (h *Handler) ShowNorm(ctx context.Context, b *bot.Bot, update *models.Update) {
	norm, err := h.service.GetNorm(ctx, update.Message.Chat.ID)
	if err != nil {
		helpers.SendMessage(ctx, b, update, "Не удалось отправить норму чата")
		log.Println("Failed to show chat norm", err)
		return
	}

	helpers.SendMessage(ctx, b, update, fmt.Sprintf("Норма чата: %d сообщений", norm))
}

func (h *Handler) SetNorm(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !helpers.CheckOwnerOrAdmin(ctx, b, h.adminService, update.Message.Chat.ID, update.Message.From.ID) {
		helpers.SendMessage(ctx, b, update, "Команда установки нормы доступна только создателю чата и администраторам бота")
		return
	}

	matches := h.setNormRe.FindStringSubmatch(update.Message.Text)
	if len(matches) < 3 {
		helpers.SendMessage(ctx, b, update, "Неверный формат команды")
		return
	}
	norm, err := strconv.Atoi(matches[2])
	if err != nil {
		helpers.SendMessage(ctx, b, update, "Норма должна быть числом")
		return
	}

	if err := h.service.SetNorm(ctx, update.Message.Chat.ID, norm); err != nil {
		helpers.SendMessage(ctx, b, update, "Не удалось установить норму чата")
		log.Println("Failed to set chat norm", err)
		return
	}

	helpers.SendMessage(ctx, b, update, "Новая норма чата установлена")
}
