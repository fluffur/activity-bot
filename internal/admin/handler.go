package admin

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/user"
	"context"
	"fmt"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Handler struct {
	service     *Service
	userService *user.Service
}

func NewHandler(service *Service, userService *user.Service) *Handler {
	return &Handler{service, userService}
}

func (h *Handler) AddAdmin(ctx context.Context, b *bot.Bot, update *models.Update) {
	member, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
		ChatID: update.Message.Chat.ID,
		UserID: update.Message.From.ID,
	})
	if err != nil {
		helpers.SendMessage(ctx, b, update, "Не удалось проверить ваши права")
		return
	}
	if member.Owner == nil {
		helpers.SendMessage(ctx, b, update, "Только создатель чата может добавлять администраторов бота")
		return
	}

	targetUser, _, err := helpers.ExtractTargetUser(h.userService, update, "")
	if err != nil {
		helpers.SendMessage(ctx, b, update, "Не удалось найти пользователя")
		return
	}
	if err := h.service.AddAdmin(ctx, update.Message.Chat.ID, targetUser.ID); err != nil {
		helpers.SendMessage(ctx, b, update, "Не удалось добавить администратора")
		return
	}

	helpers.SendMessage(ctx, b, update, fmt.Sprintf("Пользователь %s назначен администратором бота", helpers.Link(targetUser)))
}

func (h *Handler) RemoveAdmin(ctx context.Context, b *bot.Bot, update *models.Update) {
	member, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
		ChatID: update.Message.Chat.ID,
		UserID: update.Message.From.ID,
	})
	if err != nil {
		helpers.SendMessage(ctx, b, update, "Не удалось проверить ваши права")
		return
	}
	if member.Owner == nil {
		helpers.SendMessage(ctx, b, update, "Только создатель чата может удалять администраторов бота")
		return
	}

	targetUser, _, err := helpers.ExtractTargetUser(h.userService, update, "")
	if err != nil {
		helpers.SendMessage(ctx, b, update, "Не удалось найти пользователя")
		return
	}

	if err := h.service.RemoveAdmin(ctx, update.Message.Chat.ID, targetUser.ID); err != nil {
		helpers.SendMessage(ctx, b, update, "Не удалось удалить администратора")
		return
	}

	helpers.SendMessage(ctx, b, update, fmt.Sprintf("Пользователь %s удалён из администраторов бота", helpers.Link(targetUser)))
}

func (h *Handler) ListAdmins(ctx context.Context, b *bot.Bot, update *models.Update) {
	admins, err := h.service.GetAdmins(ctx, update.Message.Chat.ID)
	if err != nil {
		helpers.SendMessage(ctx, b, update, "Не удалось получить список администраторов")
		return
	}

	if len(admins) == 0 {
		helpers.SendMessage(ctx, b, update, "Список администраторов пуст")
		return
	}

	var sb strings.Builder
	sb.WriteString("👮 Администраторы бота:\n")
	for i, admin := range admins {
		sb.WriteString(fmt.Sprintf("\n%d. %s", i+1, helpers.Link(admin)))
	}
	helpers.SendMessage(ctx, b, update, sb.String())
}
