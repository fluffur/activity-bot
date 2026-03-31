package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/admin/view"
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/logger"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/hibiken/asynq"
	"golang.org/x/time/rate"
)

type Handler struct {
	service       *admin.Service
	memberService *member.Service
	chatService   *chat.Service
	dateParser    *helpers.DateParser
	asyncClient   *asynq.Client
	factory       *cmd.Factory
}

func New(service *admin.Service, memberService *member.Service, chatService *chat.Service, dateParser *helpers.DateParser, asyncClient *asynq.Client, factory *cmd.Factory) *Handler {
	return &Handler{service, memberService, chatService, dateParser, asyncClient, factory}
}

func (h *Handler) IsAdmin(b *gotgbot.Bot, ctx *command.Context) error {
	u, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	return ctx.ReplyHTML(b, fmt.Sprintf("Ранг участника: %d", u.Status))
}

func (h *Handler) SetStatus(b *gotgbot.Bot, ctx *command.Context) error {
	m, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	sender, err := ctx.Sender()
	if err != nil {
		return err
	}
	if !sender.CanModerate(*m) {
		return ctx.Reply(b, "Недостаточно прав для изменения статуса", nil)
	}

	s, err := ctx.Number()
	if err != nil {
		return err
	}
	status := model.Status(s)
	if status >= sender.Status {
		return ctx.Reply(b, "Нельзя установить участнику статус равный своему или выше", nil)
	}
	if err := h.service.SetStatus(ctx.StdContext(), *m, status); err != nil {
		if errors.Is(err, admin.ErrUserIsAlreadyAdmin) {
			return ctx.Reply(b, "Участник %s ", nil)
		}
		if errors.Is(err, admin.ErrUserIsCreator) {
			return ctx.Reply(b, "Нельзя изменить статус владельца", nil)
		}

		_ = ctx.Reply(b, "Не удалось добавить администратора", nil)
		return err
	}

	return ctx.ReplyHTML(b, view.FormatAdminAdded(*m, status))

}

func (h *Handler) RemoveAdmin(b *gotgbot.Bot, ctx *command.Context) error {
	u, err := ctx.User()
	if err != nil {
		return err
	}
	if err := h.service.SetStatus(ctx.StdContext(), *u, 0); err != nil {
		if errors.Is(err, admin.ErrUserIsNotAdmin) {
			return ctx.Reply(b, "Пользователь не является администратором", nil)
		}

		if errors.Is(err, admin.ErrUserIsCreator) {
			return ctx.Reply(b, "Нельзя удалить создателя из списка администраторов", nil)
		}

		_ = ctx.Reply(b, "Не удалось удалить администратора", nil)

		return err
	}

	return ctx.ReplyHTML(b, view.FormatAdminRemoved(*u))
}

func (h *Handler) ListAdmins(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	admins, err := h.service.GetAdminsEnsured(ctx.StdContext(), c.ID, h.memberService.SyncChatMembers)
	if err != nil {
		_ = ctx.Reply(b, "Не удалось получить список администраторов", nil)
		return err
	}

	if len(admins) == 0 {
		return ctx.Reply(b, "Список администраторов пуст", nil)
	}

	return ctx.ReplyHTML(b, view.FormatAdminsList(admins))
}

func (h *Handler) Kick(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstMember()
	if targetUser == nil {
		return cmd.ErrNoUser
	}
	mod, err := h.memberService.GetChatMember(ctx.StdContext(), ctx.TargetChatID(), ctx.EffectiveSender.Id())
	if err != nil {
		return err
	}

	reason := getReason(ctx.FirstArgument(), ctx.SecondArgument(), nil)

	title := ctx.EffectiveChat.Title
	if ctx.EffectiveChat.Type == "private" {
		c, err := b.GetChat(ctx.TargetChatID(), nil)
		if err == nil {
			title = c.Title
		}
	}
	dmText := view.FormatDirectModerationAction(*targetUser, title, "kick", nil, reason)
	if _, err := b.SendMessage(targetUser.User.ID, dmText, &gotgbot.SendMessageOpts{ParseMode: gotgbot.ParseModeHTML}); err != nil {
		slog.Warn("Failed to send kick DM notification", "user_id", targetUser.User.ID, "error", err)
	}

	if err := h.service.Kick(ctx.StdContext(), *targetUser, mod, reason); err != nil {
		if errors.Is(err, admin.ErrUserIsProtected) {
			return ctx.Reply(b, "Нельзя кикнуть администратора или создателя", nil)
		}
		if _, err := h.memberService.ProcessLeftMember(ctx.StdContext(), ctx.TargetChatID(), targetUser.User.ID); err != nil {
			return err
		}
		return fmt.Errorf("failed to kick: %w", err)
	}

	if _, err := h.memberService.ProcessLeftMember(ctx.StdContext(), ctx.TargetChatID(), targetUser.User.ID); err != nil {
		return err
	}

	return ctx.ReplyHTML(b, view.FormatModerationAction(*targetUser, "kick", nil, reason))
}

func (h *Handler) Ban(b *gotgbot.Bot, ctx *cmd.Context) error {
	m := ctx.FirstMember()
	if m == nil {
		return cmd.ErrNoUser
	}
	mod, err := h.memberService.GetChatMember(ctx.StdContext(), ctx.TargetChatID(), ctx.EffectiveSender.Id())
	if err != nil {
		return err
	}

	until := parseUntil(
		h.dateParser,
		ctx.FirstArgument(),
		0,
		true,
	)

	reason := getReason(ctx.FirstArgument(), ctx.SecondArgument(), until)

	title := ctx.EffectiveChat.Title
	if ctx.EffectiveChat.Type == "private" {
		c, err := b.GetChat(ctx.TargetChatID(), nil)
		if err == nil {
			title = c.Title
		}
	}

	if err := h.service.Ban(ctx.StdContext(), *m, mod, until, reason); err != nil {
		if errors.Is(err, admin.ErrUserIsProtected) {
			return ctx.Reply(b, "Нельзя забанить администратора или создателя", nil)
		}
		if _, err := h.memberService.ProcessLeftMember(ctx.StdContext(), ctx.TargetChatID(), m.User.ID); err != nil {
			return err
		}
		return fmt.Errorf("failed to ban: %w", err)
	}
	dmText := view.FormatDirectModerationAction(*m, title, "ban", until, reason)
	if _, err := b.SendMessage(m.User.ID, dmText, &gotgbot.SendMessageOpts{ParseMode: gotgbot.ParseModeHTML}); err != nil {
		slog.Warn("Failed to send ban DM notification", "user_id", m.User.ID, "error", err)
	}
	if _, err := h.memberService.ProcessLeftMember(ctx.StdContext(), ctx.TargetChatID(), m.User.ID); err != nil {
		return err
	}

	return ctx.ReplyHTML(b, view.FormatModerationAction(*m, "ban", until, reason))
}

func (h *Handler) Mute(b *gotgbot.Bot, ctx *cmd.Context) error {
	m := ctx.FirstMember()
	if m == nil {
		return cmd.ErrNoUser
	}

	until := parseUntil(
		h.dateParser,
		ctx.FirstArgument(),
		7*24*time.Hour,
		true,
	)

	reason := getReason(ctx.FirstArgument(), ctx.SecondArgument(), until)
	mod, err := h.memberService.GetChatMember(ctx.StdContext(), ctx.TargetChatID(), ctx.EffectiveSender.Id())
	if err != nil {
		return err
	}
	if err := h.service.Mute(ctx.StdContext(), *m, mod, until, reason); err != nil {
		if errors.Is(err, admin.ErrUserIsProtected) {
			return ctx.Reply(b, "Нельзя замутить администратора или создателя", nil)
		}
		if errors.Is(err, admin.ErrInvalidRange) {
			return ctx.Reply(b, "Срок ограничения должен быть от 30 секунд до 366 дней", nil)
		}
		_ = ctx.Reply(b, "Не удалось замутить пользователя", nil)
		return err
	}

	if until != nil {
		payload, _ := json.Marshal(model.RestoreRolePayload{
			ChatID: ctx.TargetChatID(),
			UserID: m.User.ID,
		})
		task := asynq.NewTask("role:restore", payload)
		taskID := fmt.Sprintf("role:restore:%d:%d", ctx.TargetChatID(), m.User.ID)
		_, err := h.asyncClient.Enqueue(task, asynq.ProcessAt(*until), asynq.TaskID(taskID))
		if err != nil {
			slog.Error("Failed to enqueue restore task", "error", err)
		}
	}

	return ctx.ReplyHTML(b, view.FormatModerationAction(*m, "mute", until, reason))
}

func (h *Handler) ShowWarns(b *gotgbot.Bot, ctx *cmd.Context) error {
	m := ctx.FirstMember()
	if m == nil {
		return cmd.ErrNoUser
	}

	warns, err := h.service.GetWarns(ctx.StdContext(), ctx.TargetChatID(), m.User.ID)
	if err != nil {
		return err
	}

	maxWarns, err := h.service.GetMaxWarns(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	now := time.Now()
	activeWarns := make([]model.Warn, 0, len(warns))
	for _, w := range warns {
		if w.ExpiresAt.IsZero() || w.ExpiresAt.After(now) {
			activeWarns = append(activeWarns, w)
		}
	}

	if len(activeWarns) == 0 {
		return ctx.ReplyHTML(b, fmt.Sprintf("%s У %s нет активных варнов", helpers.SuccessEmoji(), helpers.RoleEmojiLink(*m)))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("⚠️ Варны пользователя %s (активные: %d/%d):\n\n",
		helpers.RoleEmojiLink(*m), len(activeWarns), maxWarns))

	for i, w := range activeWarns {
		createdStr := helpers.FormatToHumanDateTime(w.CreatedAt)
		expireStr := ""
		if !w.ExpiresAt.IsZero() {
			expireStr = fmt.Sprintf(", истекает %s", helpers.FormatToHumanDateTime(w.ExpiresAt))
		}

		modName := helpers.RoleEmojiLink(w.Moderator)

		reasonStr := ""
		if w.Reason != "" {
			reasonStr = fmt.Sprintf(", причина: %s", w.Reason)
		}

		sb.WriteString(fmt.Sprintf("%d. Выдан %s модератором %s%s%s\n",
			i+1,
			createdStr,
			modName,
			expireStr,
			reasonStr,
		))
	}

	return ctx.ReplyHTML(b, sb.String())
}

func (h *Handler) Warnlist(b *gotgbot.Bot, ctx *cmd.Context) error {
	warns, err := h.service.GetWarnsByChat(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		_ = ctx.Reply(b, "Не удалось получить список предупреждений", nil)
		return err
	}

	maxWarns, err := h.service.GetMaxWarns(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	now := time.Now()
	activeWarns := make([]model.Warn, 0, len(warns))
	for _, w := range warns {
		if w.ExpiresAt.IsZero() || w.ExpiresAt.After(now) {
			activeWarns = append(activeWarns, w)
		}
	}

	return ctx.ReplyHTML(b, view.FormatWarnlist(activeWarns, maxWarns))
}

func (h *Handler) Warn(b *gotgbot.Bot, ctx *cmd.Context) error {
	m := ctx.FirstMember()
	if m == nil {
		return cmd.ErrNoUser
	}
	mod, err := h.memberService.GetChatMember(ctx.StdContext(), ctx.TargetChatID(), ctx.EffectiveSender.Id())
	if err != nil {
		return err
	}
	arg := ctx.FirstArgument()
	secondArg := ctx.SecondArgument()

	var until *time.Time
	var reason string

	if strings.ToLower(arg) == "навсегда" {
		until = nil
		if secondArg != "" {
			reason = secondArg
		} else {
			reason = ""
		}
	} else if t, ok := h.dateParser.Parse(arg); ok {
		until = &t
		reason = secondArg
	} else {
		defaultPeriod := 14 * 24 * time.Hour
		tt := time.Now().Add(defaultPeriod)
		until = &tt
		reason = arg
	}

	count, banned, err := h.service.Warn(ctx.StdContext(), *m, mod, reason, until)
	if err != nil {
		if errors.Is(err, admin.ErrUserIsProtected) {
			return ctx.Reply(b, "Нельзя выдать предупреждение администратору или создателю", nil)
		}
		if banned {
			if _, err := h.memberService.ProcessLeftMember(ctx.StdContext(), ctx.TargetChatID(), m.User.ID); err != nil {
				return err
			}
		} else {
			_ = ctx.Reply(b, "Не удалось выдать предупреждение", nil)
		}
		return fmt.Errorf("failed to give warn: %w", err)
	}
	if banned {
		if _, err := h.memberService.ProcessLeftMember(ctx.StdContext(), ctx.TargetChatID(), m.User.ID); err != nil {
			return err
		}
	}
	maxWarns, _ := h.service.GetMaxWarns(ctx.StdContext(), ctx.TargetChatID())

	return ctx.ReplyHTML(b, view.FormatWarnInfo(*m, count, maxWarns, until, reason, banned))
}

func (h *Handler) ShowMaxWarns(b *gotgbot.Bot, ctx *cmd.Context) error {
	maxWarns, err := h.service.GetMaxWarns(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}
	return ctx.Reply(b, fmt.Sprintf("Текущий лимит предупреждений: %d", maxWarns), nil)
}

func (h *Handler) SetMaxWarns(b *gotgbot.Bot, ctx *cmd.Context) error {
	maxWarns, err := strconv.Atoi(ctx.FirstArgument())
	if err != nil || maxWarns <= 0 {
		return ctx.Reply(b, "Лимит предупреждений должен быть положительным числом", nil)
	}

	if err := h.service.SetMaxWarns(ctx.StdContext(), ctx.TargetChatID(), maxWarns); err != nil {
		_ = ctx.Reply(b, "Не удалось обновить лимит предупреждений", nil)
		return err
	}

	return ctx.Reply(b, fmt.Sprintf("Лимит предупреждений изменен на %d", maxWarns), nil)
}

func (h *Handler) Unban(b *gotgbot.Bot, ctx *cmd.Context) error {
	m := ctx.FirstMember()
	if m == nil {
		return cmd.ErrNoUser
	}

	if err := h.service.Unban(ctx.StdContext(), ctx.TargetChatID(), m.User.ID); err != nil {
		_ = ctx.Reply(b, "Не удалось разбанить пользователя", nil)
		return err
	}

	return ctx.ReplyHTML(b, fmt.Sprintf("Пользователь %s %s",
		helpers.RoleEmojiLink(*m),
		helpers.Gendered(m.User.Gender, "разбанен", "разбанена"),
	))
}

func (h *Handler) Unmute(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstMember()
	if targetUser == nil {
		return cmd.ErrNoUser
	}

	if err := h.service.Unmute(ctx.StdContext(), ctx.TargetChatID(), targetUser.User.ID); err != nil {
		_ = ctx.Reply(b, "Не удалось размутить пользователя", nil)
		return err
	}
	m := ctx.FirstMember()
	if m == nil {
		return ctx.Reply(b, "Пользователь размучен, но не удалось вернуть роль", nil)
	}
	title := m.Tag
	if title != "" {
		if ok, err := b.PromoteChatMember(ctx.TargetChatID(), targetUser.User.ID, &gotgbot.PromoteChatMemberOpts{
			CanManageChat:   true,
			CanPostMessages: true,
			CanEditMessages: true,
		}); err != nil || !ok {
			_ = ctx.Reply(b, "Пользователь размучен, но не удалось вернуть роль", nil)
			return err
		}

	}

	return ctx.ReplyHTML(b, view.FormatUnmuteInfo(*targetUser))
}

func (h *Handler) Unwarn(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstMember()
	if targetUser == nil {
		return cmd.ErrNoUser
	}

	count, err := h.service.Unwarn(ctx.StdContext(), ctx.TargetChatID(), targetUser.User.ID)
	if err != nil {
		_ = ctx.Reply(b, "Не удалось снять предупреждение", nil)
		return err
	}

	maxWarns, _ := h.service.GetMaxWarns(ctx.StdContext(), ctx.TargetChatID())

	return ctx.ReplyHTML(b, view.FormatUnwarnInfo(*targetUser, count, maxWarns))
}

func (h *Handler) ToggleRights(b *gotgbot.Bot, ctx *cmd.Context) error {
	m := ctx.FirstMember()
	if m == nil {
		return cmd.ErrNoUser
	}
	arg := ctx.FirstArgument()

	if arg == "" {
		return ctx.ReplyHTML(b,
			fmt.Sprintf("Права разработчика в этой группе: %s", m.Status.String()),
		)
	}

	s, err := strconv.Atoi(arg)
	if err != nil {
		return err
	}

	status := model.Status(s)
	if err := h.service.SetStatus(ctx.StdContext(), *m, status); err != nil {
		return fmt.Errorf("failed to set dev status: %w", err)
	}

	return ctx.ReplyHTML(b,
		fmt.Sprintf("Права разработчика изменены на: %s", status.String()),
	)
}

func (h *Handler) UpdateChats(b *gotgbot.Bot, ctx *cmd.Context) error {
	chats, err := h.chatService.GetChatsWithoutTitle(ctx.StdContext())
	if err != nil {
		return err
	}

	limiter := rate.NewLimiter(rate.Every(1000*time.Millisecond), 2)

	for _, c := range chats {
		if err := limiter.Wait(ctx.StdContext()); err != nil {
			return err
		}

		ch, err := b.GetChat(c.ID, nil)
		if err != nil {
			slog.Error("failed to get chat", "chat", c, "err", err)
			continue
		}
		logger.L.Info("found chat title", "title", ch.Title, "id", ch.Id)
		if err := h.chatService.SetTitle(ctx.StdContext(), c.ID, ch.Title); err != nil {
			return err
		}
	}
	return ctx.Reply(b, "Чаты обновлены", nil)
}
func (h *Handler) ClearWarns(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstMember()
	if targetUser == nil {
		return cmd.ErrNoUser
	}

	if err := h.service.ClearWarns(ctx.StdContext(), ctx.TargetChatID(), targetUser.User.ID); err != nil {
		_ = ctx.Reply(b, "Не удалось очистить предупреждения", nil)
		return err
	}

	return ctx.ReplyHTML(b, view.FormatWarnsCleared(*targetUser))
}

func (h *Handler) FakeLeave(b *gotgbot.Bot, ctx *cmd.Context) error {
	m := ctx.FirstMember()
	if m == nil {
		return cmd.ErrNoUser
	}
	u := m.User
	_, err := b.SendMessage(ctx.TargetChatID(), fmt.Sprintf("🕊 %s %s нас...",
		helpers.RoleEmojiLink(*m),
		helpers.Gendered(u.Gender, "покинул", "покинула"),
	), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})
	return err
}

func (h *Handler) DemoteTgAdmin(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()
	if targetUser == nil {
		return cmd.ErrNoUser
	}

	if _, err := b.PromoteChatMember(ctx.TargetChatID(), targetUser.ID, nil); err != nil {
		return err
	}

	return ctx.Reply(b, "Пользователь разжалован", nil)
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

func getReason(firstArgument, secondArgument string, until *time.Time) string {
	if secondArgument != "" {
		return secondArgument
	}
	if until != nil {
		return ""
	}
	return firstArgument
}
