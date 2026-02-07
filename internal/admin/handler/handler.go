package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/user"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Handler struct {
	service       *admin.Service
	userService   *user.Service
	memberService *member.Service
	dateParser    *helpers.DateParser
}

func New(service *admin.Service, userService *user.Service, memberService *member.Service, dateParser *helpers.DateParser) *Handler {
	return &Handler{service, userService, memberService, dateParser}
}

func (h *Handler) IsAdmin(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()

	if targetUser == nil {
		return cmd.ErrNoUser
	}

	if h.service.CheckIsAdmin(ctx.StdContext(), ctx.EffectiveChat.Id, targetUser.ID) {
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

	if err := h.service.AddAdmin(ctx.StdContext(), ctx.EffectiveChat.Id, targetUser.ID); err != nil {
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

	if err := h.service.RemoveAdmin(ctx.StdContext(), ctx.EffectiveChat.Id, targetUser.ID); err != nil {
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
	admins, err := h.service.GetAdminsEnsured(ctx.StdContext(), ctx.EffectiveChat.Id, h.memberService.SyncChatMembers)
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

func (h *Handler) Kick(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()
	if targetUser == nil {
		return cmd.ErrNoUser
	}

	reason := strings.Join(ctx.Args(), " ")

	if err := h.service.Kick(ctx.StdContext(), ctx.EffectiveChat.Id, targetUser.ID, ctx.EffectiveSender.Id(), reason); err != nil {
		if errors.Is(err, admin.ErrUserIsProtected) {
			_, err := ctx.EffectiveMessage.Reply(b, "Нельзя кикнуть администратора или создателя", nil)
			return err
		}
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось кикнуть пользователя", nil)
		return err
	}

	text := fmt.Sprintf("Пользователь %s был кикнут из чата", helpers.Link(*targetUser))
	if reason != "" {
		text += fmt.Sprintf("\nПричина: %s", reason)
	}

	_, err := b.SendMessage(ctx.EffectiveChat.Id, text, &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
	})
	return err
}

func (h *Handler) Ban(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()
	if targetUser == nil {
		return cmd.ErrNoUser
	}

	var until *time.Time
	reason := strings.Join(ctx.Args(), " ")

	if len(ctx.Args()) > 0 {
		if t, ok := h.dateParser.Parse(ctx.Args()[0]); ok {
			until = &t
			reason = strings.Join(ctx.Args()[1:], " ")
		}
	}

	if err := h.service.Ban(ctx.StdContext(), ctx.EffectiveChat.Id, targetUser.ID, ctx.EffectiveSender.Id(), until, reason); err != nil {
		if errors.Is(err, admin.ErrUserIsProtected) {
			_, err := ctx.EffectiveMessage.Reply(b, "Нельзя забанить администратора или создателя", nil)
			return err
		}
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось забанить пользователя", nil)
		return err
	}

	text := fmt.Sprintf("Пользователь %s забанен", helpers.Link(*targetUser))
	if until != nil {
		text += fmt.Sprintf(" до %s", helpers.FormatToHumanDate(*until))
	}
	if reason != "" {
		text += fmt.Sprintf("\nПричина: %s", reason)
	}

	_, err := b.SendMessage(ctx.EffectiveChat.Id, text, &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
	})
	return err
}

func (h *Handler) Mute(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()
	if targetUser == nil {
		return cmd.ErrNoUser
	}

	var until *time.Time
	reason := strings.Join(ctx.Args(), " ")

	if len(ctx.Args()) > 0 {
		if t, ok := h.dateParser.Parse(ctx.Args()[0]); ok {
			until = &t
			reason = strings.Join(ctx.Args()[1:], " ")
		}
	}

	if err := h.service.Mute(ctx.StdContext(), ctx.EffectiveChat.Id, targetUser.ID, ctx.EffectiveSender.Id(), until, reason); err != nil {
		if errors.Is(err, admin.ErrUserIsProtected) {
			_, err := ctx.EffectiveMessage.Reply(b, "Нельзя замутить администратора или создателя", nil)
			return err
		}
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось замутить пользователя", nil)
		return err
	}

	text := fmt.Sprintf("Пользователь %s замучен", helpers.Link(*targetUser))
	if until != nil {
		text += fmt.Sprintf(" до %s", helpers.FormatToHumanDate(*until))
	} else {
		text += " навсегда"
	}
	if reason != "" {
		text += fmt.Sprintf("\nПричина: %s", reason)
	}

	_, err := b.SendMessage(ctx.EffectiveChat.Id, text, &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
	})
	return err
}

func (h *Handler) Warn(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()
	if targetUser == nil {
		return cmd.ErrNoUser
	}

	reason := strings.Join(ctx.Args(), " ")
	count, banned, err := h.service.Warn(ctx.StdContext(), ctx.EffectiveChat.Id, targetUser.ID, ctx.EffectiveSender.Id(), reason)
	if err != nil {
		if errors.Is(err, admin.ErrUserIsProtected) {
			_, err := ctx.EffectiveMessage.Reply(b, "Нельзя выдать предупреждение администратору или создателю", nil)
			return err
		}
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось выдать предупреждение", nil)
		return err
	}

	maxWarns, _ := h.service.GetMaxWarns(ctx.StdContext(), ctx.EffectiveChat.Id)

	text := fmt.Sprintf("Пользователю %s выдано предупреждение (%d/%d)", helpers.Link(*targetUser), count, maxWarns)
	if reason != "" {
		text += fmt.Sprintf("\nПричина: %s", reason)
	}

	if banned {
		text += "\n\nПользователь забанен за превышение лимита предупреждений."
	}

	_, err = b.SendMessage(ctx.EffectiveChat.Id, text, &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
	})
	return err
}

func (h *Handler) SetMaxWarns(b *gotgbot.Bot, ctx *cmd.Context) error {
	if ctx.FirstArgument() == "" {
		max, _ := h.service.GetMaxWarns(ctx.StdContext(), ctx.EffectiveChat.Id)
		_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Текущий лимит предупреждений: %d", max), nil)
		return err
	}

	max, err := strconv.Atoi(ctx.FirstArgument())
	if err != nil || max <= 0 {
		_, err := ctx.EffectiveMessage.Reply(b, "Лимит предупреждений должен быть положительным числом", nil)
		return err
	}

	if err := h.service.SetMaxWarns(ctx.StdContext(), ctx.EffectiveChat.Id, max); err != nil {
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось обновить лимит предупреждений", nil)
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Лимит предупреждений изменен на %d", max), nil)
	return err
}

func (h *Handler) Unban(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()
	if targetUser == nil {
		return cmd.ErrNoUser
	}

	if err := h.service.Unban(ctx.StdContext(), ctx.EffectiveChat.Id, targetUser.ID); err != nil {
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось разбанить пользователя", nil)
		return err
	}

	_, err := b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("Пользователь %s разбанен", helpers.Link(*targetUser)), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
	})
	return err
}

func (h *Handler) Unmute(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()
	if targetUser == nil {
		return cmd.ErrNoUser
	}

	if err := h.service.Unmute(ctx.StdContext(), ctx.EffectiveChat.Id, targetUser.ID); err != nil {
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось размутить пользователя", nil)
		return err
	}

	_, err := b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("Пользователь %s размучен", helpers.Link(*targetUser)), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
	})
	return err
}

func (h *Handler) Unwarn(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()
	if targetUser == nil {
		return cmd.ErrNoUser
	}

	count, err := h.service.Unwarn(ctx.StdContext(), ctx.EffectiveChat.Id, targetUser.ID)
	if err != nil {
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось снять предупреждение", nil)
		return err
	}

	maxWarns, _ := h.service.GetMaxWarns(ctx.StdContext(), ctx.EffectiveChat.Id)

	_, err = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("С пользователя %s снято предупреждение (%d/%d)", helpers.Link(*targetUser), count, maxWarns), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
	})
	return err
}

func (h *Handler) ClearWarns(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()
	if targetUser == nil {
		return cmd.ErrNoUser
	}

	if err := h.service.ClearWarns(ctx.StdContext(), ctx.EffectiveChat.Id, targetUser.ID); err != nil {
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось очистить предупреждения", nil)
		return err
	}

	_, err := b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("Все предупреждения пользователя %s были аннулированы", helpers.Link(*targetUser)), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
	})
	return err
}
