package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/admin/view"
	"activity-bot/internal/chat"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/logger"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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
	factory       *command.Factory
}

func New(service *admin.Service, memberService *member.Service, chatService *chat.Service, dateParser *helpers.DateParser, asyncClient *asynq.Client, factory *command.Factory) *Handler {
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

	s, err := ctx.Number()
	if err != nil {
		return err
	}
	status := model.Status(s)
	if err := h.service.SetStatus(ctx.StdContext(), *sender, *m, status); err != nil {
		if errors.Is(err, admin.ErrUserIsAlreadyAdmin) {
			return ctx.Reply(b, "Участник %s ", nil)
		}
		if errors.Is(err, admin.ErrUserIsCreator) {
			return ctx.Reply(b, "Нельзя изменить статус владельца", nil)
		}
		if errors.Is(err, admin.ErrUserStatusInvalid) {
			return ctx.Reply(b, "Нельзя установить участнику статус равный своему или выше", nil)
		}
		if errors.Is(err, admin.ErrUserIsNotPermitted) {
			return ctx.Reply(b, "Недостаточно прав для изменения статуса", nil)
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
	sender, err := ctx.Sender()
	if err != nil {
		return err
	}
	if err := h.service.SetStatus(ctx.StdContext(), *sender, *u, 0); err != nil {
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

func (h *Handler) Kick(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	u, err := ctx.User()
	if err != nil {
		return err
	}
	mod, err := ctx.Sender()
	if err != nil {
		return err
	}

	reason := ctx.TextOrDefault("")

	dmText := view.FormatDirectModerationAction(*u, c.Title, "kick", time.Time{}, reason)
	if _, err := b.SendMessage(u.User.ID, dmText, &gotgbot.SendMessageOpts{ParseMode: gotgbot.ParseModeHTML}); err != nil {
		slog.Warn("Failed to send kick DM notification", "user_id", u.User.ID, "error", err)
	}

	if err := h.service.Kick(ctx.StdContext(), *u, *mod, reason); err != nil {
		if errors.Is(err, admin.ErrUserIsProtected) {
			return ctx.Reply(b, "Нельзя кикнуть администратора или создателя", nil)
		}
		if _, err := h.memberService.ProcessLeftMember(ctx.StdContext(), u.ChatID, u.User.ID); err != nil {
			return err
		}
		return fmt.Errorf("failed to kick: %w", err)
	}

	if _, err := h.memberService.ProcessLeftMember(ctx.StdContext(), u.ChatID, u.User.ID); err != nil {
		return err
	}

	return ctx.ReplyHTML(b, view.FormatModerationAction(*u, "kick", time.Time{}, reason))
}

func (h *Handler) Ban(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	m, err := ctx.User()
	if err != nil {
		return err
	}
	mod, err := ctx.Sender()
	if err != nil {
		return err
	}

	until := ctx.DateOrDefault(time.Time{})
	reason := ctx.TextOrDefault("")

	title := c.Title

	if err := h.service.Ban(ctx.StdContext(), *m, *mod, until, reason); err != nil {
		if errors.Is(err, admin.ErrUserIsProtected) {
			return ctx.Reply(b, "Нельзя забанить администратора или создателя", nil)
		}
		if _, err := h.memberService.ProcessLeftMember(ctx.StdContext(), c.ID, m.User.ID); err != nil {
			return err
		}
		return fmt.Errorf("failed to ban: %w", err)
	}
	dmText := view.FormatDirectModerationAction(*m, title, "ban", until, reason)
	if _, err := b.SendMessage(m.User.ID, dmText, &gotgbot.SendMessageOpts{ParseMode: gotgbot.ParseModeHTML}); err != nil {
		slog.Warn("Failed to send ban DM notification", "user_id", m.User.ID, "error", err)
	}
	if _, err := h.memberService.ProcessLeftMember(ctx.StdContext(), m.ChatID, m.User.ID); err != nil {
		return err
	}

	return ctx.ReplyHTML(b, view.FormatModerationAction(*m, "ban", until, reason))
}

func (h *Handler) Mute(b *gotgbot.Bot, ctx *command.Context) error {
	m, err := ctx.User()
	if err != nil {
		return err
	}

	until := ctx.DateOrDefault(time.Now().Add(time.Hour * 24 * 7 * 2))
	reason := ctx.TextOrDefault("")

	mod, err := h.memberService.GetChatMember(ctx.StdContext(), m.ChatID, ctx.EffectiveSender.Id())
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

		return err
	}

	if !until.IsZero() {
		payload, _ := json.Marshal(model.RestoreRolePayload{
			ChatID: m.ChatID,
			UserID: m.User.ID,
		})
		task := asynq.NewTask("role:restore", payload)
		taskID := fmt.Sprintf("role:restore:%d:%d", m.ChatID, m.User.ID)
		_, err := h.asyncClient.Enqueue(task, asynq.ProcessAt(until), asynq.TaskID(taskID))
		if err != nil {
			slog.Error("Failed to enqueue restore task", "error", err)
		}
	}

	return ctx.ReplyHTML(b, view.FormatModerationAction(*m, "mute", until, reason))
}

func (h *Handler) ShowWarns(b *gotgbot.Bot, ctx *command.Context) error {
	m, err := ctx.AnyUser()
	if err != nil {
		return err
	}

	warns, err := h.service.GetWarns(ctx.StdContext(), m.ChatID, m.User.ID)
	if err != nil {
		return err
	}

	maxWarns, err := h.service.GetMaxWarns(ctx.StdContext(), m.ChatID)
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

func (h *Handler) WarnList(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	warns, err := h.service.GetWarnsByChat(ctx.StdContext(), c.ID)
	if err != nil {
		_ = ctx.Reply(b, "Не удалось получить список предупреждений", nil)
		return err
	}

	maxWarns, err := h.service.GetMaxWarns(ctx.StdContext(), c.ID)
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

func (h *Handler) Warn(b *gotgbot.Bot, ctx *command.Context) error {
	m, err := ctx.User()
	if err != nil {
		return err
	}
	mod, err := ctx.Sender()
	if err != nil {
		return err
	}
	until := ctx.DateOrDefault(time.Time{})
	reason := ctx.TextOrDefault("")

	count, banned, err := h.service.Warn(ctx.StdContext(), *m, *mod, reason, until)
	if err != nil {
		if errors.Is(err, admin.ErrUserIsProtected) {
			return ctx.Reply(b, "Нельзя выдать предупреждение администратору или создателю", nil)
		}
		if banned {
			if _, err := h.memberService.ProcessLeftMember(ctx.StdContext(), m.ChatID, m.User.ID); err != nil {
				return err
			}
		} else {
			_ = ctx.Reply(b, "Не удалось выдать предупреждение", nil)
		}
		return fmt.Errorf("failed to give warn: %w", err)
	}
	if banned {
		if _, err := h.memberService.ProcessLeftMember(ctx.StdContext(), m.ChatID, m.User.ID); err != nil {
			return err
		}
	}
	maxWarns, _ := h.service.GetMaxWarns(ctx.StdContext(), m.ChatID)

	return ctx.ReplyHTML(b, view.FormatWarnInfo(*m, count, maxWarns, until, reason, banned))
}

func (h *Handler) ShowMaxWarns(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	return ctx.Reply(b, fmt.Sprintf("Текущий лимит предупреждений: %d", c.MaxWarns), nil)
}

func (h *Handler) SetMaxWarns(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	maxWarns := ctx.NumberOrDefault(3)
	if maxWarns <= 0 {
		return ctx.Reply(b, "Лимит предупреждений должен быть положительным числом", nil)
	}

	if err := h.service.SetMaxWarns(ctx.StdContext(), c.ID, maxWarns); err != nil {
		_ = ctx.Reply(b, "Не удалось обновить лимит предупреждений", nil)
		return err
	}

	return ctx.Reply(b, fmt.Sprintf("Лимит предупреждений изменен на %d", maxWarns), nil)
}

func (h *Handler) Unban(b *gotgbot.Bot, ctx *command.Context) error {
	m, err := ctx.User()
	if err != nil {
		return err
	}
	if err := h.service.Unban(ctx.StdContext(), m.ChatID, m.User.ID); err != nil {
		_ = ctx.Reply(b, "Не удалось разбанить пользователя", nil)
		return err
	}

	return ctx.ReplyHTML(b, fmt.Sprintf("Пользователь %s %s",
		helpers.RoleEmojiLink(*m),
		helpers.Gendered(m.User.Gender, "разбанен", "разбанена"),
	))
}

func (h *Handler) Unmute(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	u, err := ctx.User()
	if err != nil {
		return err
	}

	if err := h.service.Unmute(ctx.StdContext(), u.ChatID, u.User.ID); err != nil {
		_ = ctx.Reply(b, "Не удалось размутить пользователя", nil)
		return err
	}

	if c.TagsEnabled {
		return ctx.ReplyHTML(b, view.FormatUnmuteInfo(*u))
	}

	title := u.Tag
	if title != "" {
		if ok, err := b.PromoteChatMember(c.ID, u.User.ID, &gotgbot.PromoteChatMemberOpts{
			CanManageChat:   true,
			CanPostMessages: true,
			CanEditMessages: true,
		}); err != nil || !ok {
			_ = ctx.Reply(b, "Пользователь размучен, но не удалось вернуть роль", nil)
			return err
		}
	}

	return ctx.ReplyHTML(b, view.FormatUnmuteInfo(*u))
}

func (h *Handler) Unwarn(b *gotgbot.Bot, ctx *command.Context) error {
	u, err := ctx.User()
	if u == nil {
		return err
	}

	count, err := h.service.Unwarn(ctx.StdContext(), u.ChatID, u.User.ID)
	if err != nil {
		_ = ctx.Reply(b, "Не удалось снять предупреждение", nil)
		return err
	}

	maxWarns, _ := h.service.GetMaxWarns(ctx.StdContext(), u.ChatID)

	return ctx.ReplyHTML(b, view.FormatUnwarnInfo(*u, count, maxWarns))
}

func (h *Handler) ToggleRights(b *gotgbot.Bot, ctx *command.Context) error {
	u, err := ctx.User()
	if err != nil {
		return err
	}
	sender, err := ctx.Sender()
	if err != nil {
		return err
	}
	s := ctx.NumberOrDefault(0)
	if s < 0 || s >= 5 {
		return errors.New("toggle rights invalid status")
	}
	status := model.Status(s)
	if err := h.service.SetStatus(ctx.StdContext(), *sender, *u, status); err != nil {
		return fmt.Errorf("failed to set dev status: %w", err)
	}

	return ctx.ReplyHTML(b,
		fmt.Sprintf("Права разработчика изменены на: %s", status.String()),
	)
}

func (h *Handler) UpdateChats(b *gotgbot.Bot, ctx *command.Context) error {
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
func (h *Handler) ClearWarns(b *gotgbot.Bot, ctx *command.Context) error {
	u, err := ctx.User()
	if err != nil {
		return err
	}

	if err := h.service.ClearWarns(ctx.StdContext(), u.ChatID, u.User.ID); err != nil {
		_ = ctx.Reply(b, "Не удалось очистить предупреждения", nil)
		return err
	}

	return ctx.ReplyHTML(b, view.FormatWarnsCleared(*u))
}

func (h *Handler) FakeLeave(b *gotgbot.Bot, ctx *command.Context) error {
	m, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	u := m.User
	_, err = b.SendMessage(m.ChatID, fmt.Sprintf("🕊 %s %s нас...",
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

func (h *Handler) DemoteTgAdmin(b *gotgbot.Bot, ctx *command.Context) error {
	u, err := ctx.User()
	if err != nil {
		return err
	}

	if _, err := b.PromoteChatMember(u.ChatID, u.User.ID, nil); err != nil {
		return err
	}

	return ctx.Reply(b, fmt.Sprintf("Участник %s %s", helpers.RoleEmojiLink(*u), helpers.Gendered(u.User.Gender, "разжалован", "разжалована")), nil)
}
