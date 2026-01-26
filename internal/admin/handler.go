package admin

import (
	"activity-bot/internal/command"
	"activity-bot/internal/common"
	"activity-bot/internal/helpers"
	"activity-bot/internal/user"
	"fmt"
	"log/slog"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	service           *Service
	userService       *user.Service
	permissionChecker *common.PermissionChecker
	chatUpdater       *common.ChatUpdater
}

func NewHandler(service *Service, userService *user.Service, permissionChecker *common.PermissionChecker, chatUpdater *common.ChatUpdater) *Handler {
	return &Handler{service, userService, permissionChecker, chatUpdater}
}

func (h *Handler) IsAdmin(b *gotgbot.Bot, ctx *ext.Context, cctx *command.Context) error {
	if len(cctx.Users) == 0 {
		_, err := ctx.EffectiveMessage.Reply(b, "Пользователь не найден в базе данных бота. Попробуйте упомянуть его через ответ на сообщение или дождитесь, пока он напишет что-нибудь.", nil)
		return err
	}

	targetUser := cctx.Users[0]

	if h.permissionChecker.IsAdmin(b, ctx.EffectiveChat.Id, targetUser.ID) {
		_, err := ctx.EffectiveMessage.Reply(b, "Пользователь является администратором чата", nil)
		return err
	}

	_, err := ctx.EffectiveMessage.Reply(b, "Пользователь не является администратором чата", nil)
	return err
}

func (h *Handler) AddAdmin(b *gotgbot.Bot, ctx *ext.Context, cctx *command.Context) error {
	if len(cctx.Users) == 0 {
		_, err := ctx.EffectiveMessage.Reply(b, "Пользователь не найден в базе данных бота. Попробуйте упомянуть его через ответ на сообщение или дождитесь, пока он напишет что-нибудь.", nil)
		return err
	}

	targetUser := cctx.Users[0]
	if err := h.service.AddAdmin(ctx.EffectiveChat.Id, targetUser.ID); err != nil {
		slog.Error("failed to add admin", "chat_id", ctx.EffectiveChat.Id, "user_id", targetUser.ID, "error", err)
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось добавить администратора", nil)
		return err
	}

	_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Пользователь %s назначен администратором бота", helpers.Link(*targetUser)), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})

	return err
}

func (h *Handler) RemoveAdmin(b *gotgbot.Bot, ctx *ext.Context, cctx *command.Context) error {
	if len(cctx.Users) == 0 {
		_, err := ctx.EffectiveMessage.Reply(b, "Пользователь не найден в базе данных бота.", nil)
		return err
	}

	targetUser := cctx.Users[0]

	isAdmin, err := h.service.IsAdmin(ctx.EffectiveChat.Id, targetUser.ID)
	if err != nil {
		slog.Error("failed to check admin status", "chat_id", ctx.EffectiveChat.Id, "user_id", targetUser.ID, "error", err)
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось проверить статус пользователя", nil)

		return err
	}

	if !isAdmin {
		_, err := ctx.EffectiveMessage.Reply(b, "Пользователь не является администратором", nil)

		return err
	}

	if err := h.service.RemoveAdmin(ctx.EffectiveChat.Id, targetUser.ID); err != nil {
		slog.Error("failed to remove admin", "chat_id", ctx.EffectiveChat.Id, "user_id", targetUser.ID, "error", err)
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось удалить администратора", nil)
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Пользователь %s удалён из администраторов бота", helpers.Link(*targetUser)), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})

	return err
}

func (h *Handler) ListAdmins(b *gotgbot.Bot, ctx *ext.Context, _ *command.Context) error {
	admins, err := h.service.GetAdminsEnsured(ctx.EffectiveChat.Id, h.chatUpdater.UpdateChatMembers)
	if err != nil {
		slog.Error("failed to list admins", "chat_id", ctx.EffectiveChat.Id, "error", err)
		_, err = ctx.EffectiveMessage.Reply(b, "Не удалось получить список администраторов", nil)
		return err
	}

	if len(admins) == 0 {
		_, err = ctx.EffectiveMessage.Reply(b, "Список администраторов пуст", nil)
		return err
	}

	var sb strings.Builder
	sb.WriteString("👮 Администраторы бота:\n")
	for i, admin := range admins {
		sb.WriteString(fmt.Sprintf("\n%d. %s", i+1, helpers.Link(admin)))
	}
	_, err = ctx.EffectiveMessage.Reply(b, sb.String(), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})

	return err
}
