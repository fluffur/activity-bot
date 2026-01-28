package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/user"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	service       *admin.Service
	userService   *user.Service
	memberService *member.Service
}

func New(service *admin.Service, userService *user.Service, memberService *member.Service) *Handler {
	return &Handler{service, userService, memberService}
}

func (h *Handler) IsAdmin(b *gotgbot.Bot, ctx *ext.Context, cctx *cmd.Context) error {
	targetUser := cctx.FirstUser()

	if targetUser == nil {
		slog.Error("No user in IsAdmin")
		return nil
	}

	if h.service.CheckIsAdmin(ctx.EffectiveChat.Id, targetUser.ID) {
		_, err := ctx.EffectiveMessage.Reply(b, "Пользователь является администратором чата", nil)
		return err
	}

	_, err := ctx.EffectiveMessage.Reply(b, "Пользователь не является администратором чата", nil)
	return err
}

func (h *Handler) AddAdmin(b *gotgbot.Bot, ctx *ext.Context, cctx *cmd.Context) error {
	targetUser := cctx.FirstUser()

	if targetUser == nil {
		_, err := ctx.EffectiveMessage.Reply(b, "Вы забыли указать пользователя, которого хотите сделать админом, либо он был не найден в чате", nil)
		return err
	}

	if err := h.service.AddAdmin(ctx.EffectiveChat.Id, targetUser.ID); err != nil {
		if errors.Is(err, admin.ErrUserIsAlreadyAdmin) {
			_, err := ctx.EffectiveMessage.Reply(b, "Пользователь уже является администратором", nil)
			return err
		}

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

func (h *Handler) RemoveAdmin(b *gotgbot.Bot, ctx *ext.Context, cctx *cmd.Context) error {
	targetUser := cctx.FirstUser()

	if targetUser == nil {
		slog.Error("No user in RemoveAdmin")
		return nil
	}

	if err := h.service.RemoveAdmin(ctx.EffectiveChat.Id, targetUser.ID); err != nil {
		if errors.Is(err, admin.ErrUserIsNotAdmin) {
			_, err := ctx.EffectiveMessage.Reply(b, "Пользователь не является администратором", nil)

			return err
		}

		if errors.Is(err, admin.ErrUserIsCreator) {
			_, err := ctx.EffectiveMessage.Reply(b, "Нельзя удалить создателя из списка администраторов", nil)

			return err
		}

		slog.Error("failed to remove admin", "chat_id", ctx.EffectiveChat.Id, "user_id", targetUser.ID, "error", err)
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось удалить администратора", nil)

		return err
	}

	_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Пользователь %s удалён из администраторов бота", helpers.Link(*targetUser)), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})

	return err
}

func (h *Handler) ListAdmins(b *gotgbot.Bot, ctx *ext.Context, _ *cmd.Context) error {
	admins, err := h.service.GetAdminsEnsured(ctx.EffectiveChat.Id, h.memberService.SyncChatMembers)
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
	for i, a := range admins {
		sb.WriteString(fmt.Sprintf("\n%d. %s", i+1, helpers.Link(a)))
	}
	_, err = ctx.EffectiveMessage.Reply(b, sb.String(), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})

	return err
}
