package admin

import (
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/user"
	"fmt"
	"log"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	service     *Service
	userService *user.Service
}

func NewHandler(service *Service, userService *user.Service) *Handler {
	return &Handler{service, userService}
}

func (h *Handler) AddAdmin(b *gotgbot.Bot, ctx *ext.Context, cctx *command.Context) error {
	if len(cctx.Users) == 0 {
		_, err := ctx.EffectiveMessage.Reply(b, "Пользователь не найден в базе данных бота. Попробуйте упомянуть его через ответ на сообщение или дождитесь, пока он напишет что-нибудь.", nil)
		return err
	}

	targetUser := cctx.Users[0]
	if err := h.service.AddAdmin(ctx.EffectiveChat.Id, targetUser.ID); err != nil {
		log.Println("AddAdmin AddAdmin", err)
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
		log.Println("RemoveAdmin IsAdmin", err)
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось проверить статус пользователя", nil)

		return err
	}

	if !isAdmin {
		_, err := ctx.EffectiveMessage.Reply(b, "Пользователь не является администратором", nil)

		return err
	}

	if err := h.service.RemoveAdmin(ctx.EffectiveChat.Id, targetUser.ID); err != nil {
		log.Println("RemoveAdmin", err)
		_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Пользователь %s удалён из администраторов бота", helpers.Link(*targetUser)), nil)

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
	admins, err := h.service.GetAdmins(ctx.EffectiveChat.Id)
	if err != nil {
		log.Println("ListAdmins GetAdmins", err)
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
