package admin

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/user"
	"context"
	"fmt"
	"html"
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
		helpers.AnswerMessage(ctx, b, update, "Не удалось проверить ваши права")
		return
	}
	if member.Owner == nil {
		helpers.AnswerMessage(ctx, b, update, "Только создатель чата может добавлять администраторов бота")
		return
	}

	targetUserID, _, found, err := helpers.ExtractTargetUser(ctx, h.userService, update, "")
	if err != nil {
		helpers.AnswerMessage(ctx, b, update, "Не удалось найти пользователя")
		return
	}
	if !found {
		helpers.AnswerMessage(ctx, b, update, "Пользователь не найден")
		return
	}

	if err := h.service.AddAdmin(ctx, update.Message.Chat.ID, targetUserID); err != nil {
		helpers.AnswerMessage(ctx, b, update, "Не удалось добавить администратора")
		return
	}

	u, err := h.userService.GetUser(ctx, targetUserID)
	name := "пользователя"
	if err == nil {
		name = html.EscapeString(u.FirstName)
	}

	helpers.AnswerMessage(ctx, b, update, fmt.Sprintf("Пользователь <a href=\"tg://user?id=%d\">%s</a> назначен администратором бота", targetUserID, name))
}

func (h *Handler) RemoveAdmin(ctx context.Context, b *bot.Bot, update *models.Update) {
	member, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
		ChatID: update.Message.Chat.ID,
		UserID: update.Message.From.ID,
	})
	if err != nil {
		helpers.AnswerMessage(ctx, b, update, "Не удалось проверить ваши права")
		return
	}
	if member.Owner == nil {
		helpers.AnswerMessage(ctx, b, update, "Только создатель чата может удалять администраторов бота")
		return
	}

	targetUserID, _, found, err := helpers.ExtractTargetUser(ctx, h.userService, update, "")
	if err != nil {
		helpers.AnswerMessage(ctx, b, update, "Не удалось найти пользователя")
		return
	}
	if !found {
		helpers.AnswerMessage(ctx, b, update, "Пользователь не найден")
		return
	}

	if err := h.service.RemoveAdmin(ctx, update.Message.Chat.ID, targetUserID); err != nil {
		helpers.AnswerMessage(ctx, b, update, "Не удалось удалить администратора")
		return
	}

	u, err := h.userService.GetUser(ctx, targetUserID)
	name := "пользователя"
	if err == nil {
		name = html.EscapeString(u.FirstName)
	}

	helpers.AnswerMessage(ctx, b, update, fmt.Sprintf("Пользователь <a href=\"tg://user?id=%d\">%s</a> удалён из администраторов бота", targetUserID, name))
}

func (h *Handler) ListAdmins(ctx context.Context, b *bot.Bot, update *models.Update) {
	admins, err := h.service.GetAdmins(ctx, update.Message.Chat.ID)
	if err != nil {
		helpers.AnswerMessage(ctx, b, update, "Не удалось получить список администраторов")
		return
	}

	if len(admins) == 0 {
		helpers.AnswerMessage(ctx, b, update, "Список администраторов пуст")
		return
	}

	var sb strings.Builder
	sb.WriteString("👮 Администраторы бота:\n")
	for _, admin := range admins {
		sb.WriteString(fmt.Sprintf("\n<a href=\"tg://user?id=%d\">%s</a> (с %s)", admin.UserID, html.EscapeString(admin.DisplayName), admin.CreatedAt.Format("02.01.2006")))
	}
	helpers.AnswerMessage(ctx, b, update, sb.String())
}
