package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/user"
	"errors"
	"fmt"
	"log"
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

	reason := getReason(ctx.FirstArgument(), ctx.SecondArgument())

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

	until := parseUntil(
		h.dateParser,
		ctx.FirstArgument(),
		0,
		true,
	)

	reason := getReason(ctx.FirstArgument(), ctx.SecondArgument())

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
	} else {
		text += " навсегда"
	}
	if reason != "" {
		text += fmt.Sprintf("\nПричина: %s", reason)
	}

	_, err := b.SendMessage(ctx.EffectiveChat.Id, text, &gotgbot.SendMessageOpts{
		ParseMode: "HTML",
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})
	return err
}

func (h *Handler) Mute(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()
	if targetUser == nil {
		return cmd.ErrNoUser
	}
	until := parseUntil(
		h.dateParser,
		ctx.FirstArgument(),
		7*24*time.Hour,
		true,
	)

	reason := getReason(ctx.FirstArgument(), ctx.SecondArgument())

	if err := h.service.Mute(ctx.StdContext(), ctx.EffectiveChat.Id, targetUser.ID, ctx.EffectiveSender.Id(), until, reason); err != nil {
		if errors.Is(err, admin.ErrUserIsProtected) {
			_, err := ctx.EffectiveMessage.Reply(b, "Нельзя замутить администратора или создателя", nil)
			return err
		}
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось замутить пользователя", nil)
		return err
	}
	log.Println(reason)
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
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})
	return err
}

func (h *Handler) ShowWarns(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()
	if targetUser == nil {
		return cmd.ErrNoUser
	}

	count, err := h.service.GetWarnsCount(ctx.StdContext(), ctx.EffectiveChat.Id, targetUser.ID)
	if err != nil {
		return err
	}
	maxWarns, err := h.service.GetMaxWarns(ctx.StdContext(), ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}
	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("У пользователя %s %d/%d предупреждений", helpers.Link(*targetUser), count, maxWarns), &gotgbot.SendMessageOpts{
		ParseMode: "HTML",
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})

	return err
}

func (h *Handler) Warn(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()
	if targetUser == nil {
		return cmd.ErrNoUser
	}

	defaultPeriod := 14 * 24 * time.Hour
	defaultTime := time.Now().Add(defaultPeriod)
	until := parseUntil(
		h.dateParser,
		ctx.FirstArgument(),
		defaultPeriod,
		false,
	)
	if until == nil {
		until = &defaultTime
	}

	reason := getReason(ctx.FirstArgument(), ctx.SecondArgument())

	count, banned, err := h.service.Warn(ctx.StdContext(), ctx.EffectiveChat.Id, targetUser.ID, ctx.EffectiveSender.Id(), reason, until)
	if err != nil {
		if errors.Is(err, admin.ErrUserIsProtected) {
			_, err := ctx.EffectiveMessage.Reply(b, "Нельзя выдать предупреждение администратору или создателю", nil)
			return err
		}
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось выдать предупреждение", nil)
		return err
	}

	maxWarns, _ := h.service.GetMaxWarns(ctx.StdContext(), ctx.EffectiveChat.Id)

	text := fmt.Sprintf("Пользователю %s выдано предупреждение (%d/%d) до %s", helpers.Link(*targetUser), count, maxWarns, helpers.FormatToHumanDate(*until))
	if reason != "" {
		text += fmt.Sprintf("\nПричина: %s", reason)
	}

	if banned {
		text += "\n\nПользователь забанен за превышение лимита предупреждений."
	}

	_, err = b.SendMessage(ctx.EffectiveChat.Id, text, &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})
	return err
}

func (h *Handler) ShowMaxWarns(b *gotgbot.Bot, ctx *cmd.Context) error {
	maxWarns, err := h.service.GetMaxWarns(ctx.StdContext(), ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}
	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Текущий лимит предупреждений: %d", maxWarns), nil)
	return err
}

func (h *Handler) SetMaxWarns(b *gotgbot.Bot, ctx *cmd.Context) error {
	maxWarns, err := strconv.Atoi(ctx.FirstArgument())
	if err != nil || maxWarns <= 0 {
		_, err := ctx.EffectiveMessage.Reply(b, "Лимит предупреждений должен быть положительным числом", nil)
		return err
	}

	if err := h.service.SetMaxWarns(ctx.StdContext(), ctx.EffectiveChat.Id, maxWarns); err != nil {
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось обновить лимит предупреждений", nil)
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Лимит предупреждений изменен на %d", maxWarns), nil)
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
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
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
	title, err := h.memberService.GetMemberTitle(ctx.StdContext(), ctx.EffectiveChat.Id, targetUser.ID)
	if err != nil {
		return err
	}
	if title != "" {
		if ok, err := b.PromoteChatMember(ctx.EffectiveChat.Id, targetUser.ID, &gotgbot.PromoteChatMemberOpts{
			CanPinMessages:  true,
			CanPostMessages: true,
			CanEditMessages: true,
		}); err != nil || !ok {
			_, _ = ctx.EffectiveMessage.Reply(b, "Пользователь размучен, но не удалось вернуть роль", nil)
			return err
		}

		if _, err := b.SetChatAdministratorCustomTitle(ctx.EffectiveChat.Id, targetUser.ID, title, nil); err != nil {
			_, _ = ctx.EffectiveMessage.Reply(b, "Пользователь размучет, но роль уже назначена кем-то другим", nil)

			return err
		}
	}

	_, err = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("Пользователь %s размучен", helpers.Link(*targetUser)), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
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
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})
	return err
}

func (h *Handler) ToggleRights(b *gotgbot.Bot, ctx *cmd.Context) error {
	arg := ctx.FirstArgument()
	if arg == "" {
		role, _ := h.service.GetDevRole(ctx.StdContext(), ctx.EffectiveSender.Id())
		mapping := map[string]string{
			admin.DevRoleMember:  "участник",
			admin.DevRoleAdmin:   "администратор",
			admin.DevRoleCreator: "создатель",
		}
		_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Текущая роль разработчика: %s\n\nИспользуйте: !права [участник|админ|создатель]", mapping[role]), nil)
		return err
	}

	var targetRole string
	switch strings.ToLower(arg) {
	case "участник", "member":
		targetRole = admin.DevRoleMember
	case "админ", "администратор", "admin":
		targetRole = admin.DevRoleAdmin
	case "создатель", "creator":
		targetRole = admin.DevRoleCreator
	default:
		_, err := ctx.EffectiveMessage.Reply(b, "Неизвестная роль. Используйте: участник, админ или создатель", nil)
		return err
	}

	if err := h.service.SetDevRole(ctx.StdContext(), ctx.EffectiveSender.Id(), targetRole); err != nil {
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось сохранить права", nil)
		return err
	}

	mapping := map[string]string{
		admin.DevRoleMember:  "участник",
		admin.DevRoleAdmin:   "администратор",
		admin.DevRoleCreator: "создатель",
	}
	_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Роль разработчика изменена на: %s", mapping[targetRole]), nil)
	return err
}

func (h *Handler) AddDeveloper(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()
	if targetUser == nil {
		return cmd.ErrNoUser
	}

	role := admin.DevRoleCreator
	if arg := ctx.SecondArgument(); arg != "" {
		switch strings.ToLower(arg) {
		case "участник", "member":
			role = admin.DevRoleMember
		case "админ", "admin":
			role = admin.DevRoleAdmin
		case "создатель", "creator":
			role = admin.DevRoleCreator
		}
	}

	if err := h.service.SetDevRole(ctx.StdContext(), targetUser.ID, role); err != nil {
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось добавить разработчика", nil)
		return err
	}

	_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Пользователь %s назначен разработчиком бота с ролью %s", helpers.Link(*targetUser), role), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
	})
	return err
}

func (h *Handler) RemoveDeveloper(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()
	if targetUser == nil {
		return cmd.ErrNoUser
	}

	if err := h.service.RemoveDeveloper(ctx.StdContext(), targetUser.ID); err != nil {
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось удалить разработчика", nil)
		return err
	}

	_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Пользователь %s удален из списка разработчиков", helpers.Link(*targetUser)), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
	})
	return err
}

func (h *Handler) ListDevelopers(b *gotgbot.Bot, ctx *cmd.Context) error {
	users, roles, err := h.service.GetAllDevelopers(ctx.StdContext())
	if err != nil {
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось получить список разработчиков", nil)
		return err
	}

	var sb strings.Builder
	sb.WriteString("🛠 Разработчики бота:\n")
	for i, u := range users {
		sb.WriteString(fmt.Sprintf("\n%d. %s (%s)", i+1, helpers.Link(u), roles[i]))
	}

	_, err = ctx.EffectiveMessage.Reply(b, sb.String(), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
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
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})
	return err
}

func parseUntil(
	parser *helpers.DateParser,
	arg string,
	defaultDuration time.Duration,
	allowForever bool,
) *time.Time {
	if allowForever && arg == "навсегда" {
		return nil
	}

	if t, ok := parser.Parse(arg); ok {
		return &t
	}

	if defaultDuration > 0 {
		t := time.Now().Add(defaultDuration)
		return &t
	}

	return nil
}

func getReason(firstArgument, secondArgument string) string {
	if secondArgument != "" {
		return secondArgument
	}

	return firstArgument
}
