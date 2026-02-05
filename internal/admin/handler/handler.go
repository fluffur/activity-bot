package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/user"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Handler struct {
	service       *admin.Service
	userService   *user.Service
	memberService *member.Service
}

func New(service *admin.Service, userService *user.Service, memberService *member.Service) *Handler {
	return &Handler{service, userService, memberService}
}

func (h *Handler) IsAdmin(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()

	if targetUser == nil {
		return cmd.ErrNoUser
	}

	if h.service.CheckIsAdmin(context.Background(), ctx.EffectiveChat.Id, targetUser.ID) {
		_, err := ctx.EffectiveMessage.Reply(b, "Пользователь является администратором чата", nil)
		return err
	}

	_, err := ctx.EffectiveMessage.Reply(b, "Пользователь не является администратором чата", nil)
	return err
}

func (h *Handler) AddAdmin(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()

	if targetUser == nil {
		_, err := ctx.EffectiveMessage.Reply(b, "Вы забыли указать пользователя, которого хотите сделать админом, либо он был не найден в чате", nil)
		return err
	}

	if err := h.service.AddAdmin(context.Background(), ctx.EffectiveChat.Id, targetUser.ID); err != nil {
		if errors.Is(err, admin.ErrUserIsAlreadyAdmin) {
			_, err := ctx.EffectiveMessage.Reply(b, "Пользователь уже является администратором", nil)
			return err
		}

		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось добавить администратора", nil)
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

func (h *Handler) RemoveAdmin(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()

	if targetUser == nil {
		return cmd.ErrNoUser
	}

	if err := h.service.RemoveAdmin(context.Background(), ctx.EffectiveChat.Id, targetUser.ID); err != nil {
		if errors.Is(err, admin.ErrUserIsNotAdmin) {
			_, err := ctx.EffectiveMessage.Reply(b, "Пользователь не является администратором", nil)

			return err
		}

		if errors.Is(err, admin.ErrUserIsCreator) {
			_, err := ctx.EffectiveMessage.Reply(b, "Нельзя удалить создателя из списка администраторов", nil)

			return err
		}

		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось удалить администратора", nil)

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

func (h *Handler) ListAdmins(b *gotgbot.Bot, ctx *cmd.Context) error {
	admins, err := h.service.GetAdminsEnsured(context.Background(), ctx.EffectiveChat.Id, h.memberService.SyncChatMembers)
	if err != nil {
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось получить список администраторов", nil)
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
