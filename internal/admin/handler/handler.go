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
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
	"github.com/hibiken/asynq"
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

func (h *Handler) IsAdmin(ctx *command.Context, u *ext.Update) error {
	cm, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	_, err = ctx.Reply(u, ext.ReplyTextString(fmt.Sprintf("Ранг участника: %d", cm.Status)), nil)
	return err
}

func (h *Handler) SetStatus(ctx *command.Context, u *ext.Update) error {
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
			_, _ = ctx.Reply(u, ext.ReplyTextString("Участник уже является администратором"), nil)
			return nil
		}
		if errors.Is(err, admin.ErrUserIsCreator) {
			_, _ = ctx.Reply(u, ext.ReplyTextString("Нельзя изменить статус владельца"), nil)
			return nil
		}
		if errors.Is(err, admin.ErrUserStatusInvalid) {
			_, _ = ctx.Reply(u, ext.ReplyTextString("Нельзя установить участнику статус равный своему или выше"), nil)
			return nil
		}
		if errors.Is(err, admin.ErrUserIsNotPermitted) {
			_, _ = ctx.Reply(u, ext.ReplyTextString("Недостаточно прав для изменения статуса"), nil)
			return nil
		}

		_, _ = ctx.Reply(u, ext.ReplyTextString("Не удалось добавить администратора"), nil)
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		view.WriteAdminAdded(eb, *m, status)
		return nil
	})), nil)
	return err
}

func (h *Handler) RemoveAdmin(ctx *command.Context, u *ext.Update) error {
	target, err := ctx.User()
	if err != nil {
		return err
	}
	sender, err := ctx.Sender()
	if err != nil {
		return err
	}
	if err := h.service.SetStatus(ctx.StdContext(), *sender, *target, 0); err != nil {
		if errors.Is(err, admin.ErrUserIsNotAdmin) {
			_, _ = ctx.Reply(u, ext.ReplyTextString("Пользователь не является администратором"), nil)
			return nil
		}

		if errors.Is(err, admin.ErrUserIsCreator) {
			_, _ = ctx.Reply(u, ext.ReplyTextString("Нельзя удалить создателя из списка администраторов"), nil)
			return nil
		}

		_, _ = ctx.Reply(u, ext.ReplyTextString("Не удалось удалить администратора"), nil)
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		view.WriteAdminRemoved(eb, *target)
		return nil
	})), nil)
	return err
}

func (h *Handler) ListAdmins(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	admins, err := h.service.GetAdminsEnsured(ctx.StdContext(), c.ID, h.memberService.SyncChatMembers)
	if err != nil {
		_, _ = ctx.Reply(u, ext.ReplyTextString("Не удалось получить список администраторов"), nil)
		return err
	}

	if len(admins) == 0 {
		_, _ = ctx.Reply(u, ext.ReplyTextString("Список администраторов пуст"), nil)
		return nil
	}

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		eb.Plain(view.FormatAdminsList(admins))
		return nil
	})), nil)
	return err
}

func (h *Handler) Kick(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	target, err := ctx.User()
	if err != nil {
		return err
	}
	mod, err := ctx.Sender()
	if err != nil {
		return err
	}

	reason := ctx.TextOrDefault("")

	// DM Notification
	_, _ = ctx.Context.SendMessage(target.User.ID, &tg.MessagesSendMessageRequest{
		Message: styling.Custom(func(eb *entity.Builder) error {
			view.WriteDirectModerationAction(eb, *target, c.Title, "kick", time.Time{}, reason)
			return nil
		}),
	})

	if err := h.service.Kick(ctx.StdContext(), *target, *mod, reason); err != nil {
		if errors.Is(err, admin.ErrUserIsProtected) {
			_, _ = ctx.Reply(u, ext.ReplyTextString("Нельзя кикнуть администратора или создателя"), nil)
			return nil
		}
		_, _ = h.memberService.ProcessLeftMember(ctx.StdContext(), target.ChatID, target.User.ID)
		return fmt.Errorf("failed to kick: %w", err)
	}

	_, _ = h.memberService.ProcessLeftMember(ctx.StdContext(), target.ChatID, target.User.ID)

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		view.WriteModerationAction(eb, *target, "kick", time.Time{}, reason)
		return nil
	})), nil)
	return err
}

func (h *Handler) Ban(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	target, err := ctx.User()
	if err != nil {
		return err
	}
	mod, err := ctx.Sender()
	if err != nil {
		return err
	}

	until := ctx.DateOrDefault(time.Time{})
	reason := ctx.TextOrDefault("")

	if err := h.service.Ban(ctx.StdContext(), *target, *mod, until, reason); err != nil {
		if errors.Is(err, admin.ErrUserIsProtected) {
			_, _ = ctx.Reply(u, ext.ReplyTextString("Нельзя забанить администратора или создателя"), nil)
			return nil
		}
		_, _ = h.memberService.ProcessLeftMember(ctx.StdContext(), c.ID, target.User.ID)
		return fmt.Errorf("failed to ban: %w", err)
	}

	// DM Notification
	_, _ = ctx.Context.SendMessage(target.User.ID, &tg.MessagesSendMessageRequest{
		Message: styling.Custom(func(eb *entity.Builder) error {
			view.WriteDirectModerationAction(eb, *target, c.Title, "ban", until, reason)
			return nil
		}),
	})

	_, _ = h.memberService.ProcessLeftMember(ctx.StdContext(), target.ChatID, target.User.ID)

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		view.WriteModerationAction(eb, *target, "ban", until, reason)
		return nil
	})), nil)
	return err
}

func (h *Handler) Mute(ctx *command.Context, u *ext.Update) error {
	target, err := ctx.User()
	if err != nil {
		return err
	}

	until := ctx.DateOrDefault(time.Now().Add(time.Hour * 24 * 7 * 2))
	reason := ctx.TextOrDefault("")

	mod, err := h.memberService.GetChatMember(ctx.StdContext(), target.ChatID, ctx.EffectiveSender.Id())
	if err != nil {
		return err
	}

	if err := h.service.Mute(ctx.StdContext(), *target, mod, until, reason); err != nil {
		if errors.Is(err, admin.ErrUserIsProtected) {
			_, _ = ctx.Reply(u, ext.ReplyTextString("Нельзя замутить администратора или создателя"), nil)
			return nil
		}
		if errors.Is(err, admin.ErrInvalidRange) {
			_, _ = ctx.Reply(u, ext.ReplyTextString("Срок ограничения должен быть от 30 секунд до 366 дней"), nil)
			return nil
		}

		return err
	}

	if !until.IsZero() {
		payload, _ := json.Marshal(model.RestoreRolePayload{
			ChatID: target.ChatID,
			UserID: target.User.ID,
		})
		task := asynq.NewTask("role:restore", payload)
		taskID := fmt.Sprintf("role:restore:%d:%d", target.ChatID, target.User.ID)
		_, err := h.asyncClient.Enqueue(task, asynq.ProcessAt(until), asynq.TaskID(taskID))
		if err != nil {
			slog.Error("Failed to enqueue restore task", "error", err)
		}
	}

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		view.WriteModerationAction(eb, *target, "mute", until, reason)
		return nil
	})), nil)
	return err
}

func (h *Handler) ShowWarns(ctx *command.Context, u *ext.Update) error {
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
		_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
			helpers.WriteSuccessEmoji(eb)
			eb.Plain(" У ")
			helpers.WriteRoleEmojiLink(eb, *m)
			eb.Plain(" нет активных варнов")
			return nil
		})), nil)
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		eb.Plain("⚠️ Варны пользователя ")
		helpers.WriteRoleEmojiLink(eb, *m)
		eb.Plain(fmt.Sprintf(" (активные: %d/%d):\n\n", len(activeWarns), maxWarns))

		for i, w := range activeWarns {
			eb.Plain(fmt.Sprintf("%d. Выдан ", i+1))
			helpers.FormattedDate(eb, w.CreatedAt)
			eb.Plain(" модератором ")
			helpers.WriteRoleEmojiLink(eb, w.Moderator)
			if !w.ExpiresAt.IsZero() {
				eb.Plain(", истекает ")
				helpers.FormattedDate(eb, w.ExpiresAt)
			}

			if w.Reason != "" {
				eb.Plain(fmt.Sprintf(", причина: %s", w.Reason))
			}
			eb.Plain("\n")
		}
		return nil
	})), nil)
	return err
}

func (h *Handler) WarnList(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	warns, err := h.service.GetWarnsByChat(ctx.StdContext(), c.ID)
	if err != nil {
		_, _ = ctx.Reply(u, ext.ReplyTextString("Не удалось получить список предупреждений"), nil)
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

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		view.WriteWarnlist(activeWarns, maxWarns)
		return nil
	})), nil)
	return err
}

func (h *Handler) Warn(ctx *command.Context, u *ext.Update) error {
	target, err := ctx.User()
	if err != nil {
		return err
	}
	mod, err := ctx.Sender()
	if err != nil {
		return err
	}
	until := ctx.DateOrDefault(time.Time{})
	reason := ctx.TextOrDefault("")

	count, banned, err := h.service.Warn(ctx.StdContext(), *target, *mod, reason, until)
	if err != nil {
		if errors.Is(err, admin.ErrUserIsProtected) {
			_, _ = ctx.Reply(u, ext.ReplyTextString("Нельзя выдать предупреждение администратору или создателю"), nil)
			return nil
		}
		if banned {
			_, _ = h.memberService.ProcessLeftMember(ctx.StdContext(), target.ChatID, target.User.ID)
		} else {
			_, _ = ctx.Reply(u, ext.ReplyTextString("Не удалось выдать предупреждение"), nil)
		}
		return fmt.Errorf("failed to give warn: %w", err)
	}
	if banned {
		_, _ = h.memberService.ProcessLeftMember(ctx.StdContext(), target.ChatID, target.User.ID)
	}
	maxWarns, _ := h.service.GetMaxWarns(ctx.StdContext(), target.ChatID)

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		view.WriteWarnInfo(eb, *target, count, maxWarns, until, reason, banned)
		return nil
	})), nil)
	return err
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

func (h *Handler) Unban(ctx *command.Context, u *ext.Update) error {
	target, err := ctx.User()
	if err != nil {
		return err
	}
	if err := h.service.Unban(ctx.StdContext(), target.ChatID, target.User.ID); err != nil {
		_, _ = ctx.Reply(u, ext.ReplyTextString("Не удалось разбанить пользователя"), nil)
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		eb.Plain("Пользователь ")
		helpers.WriteRoleEmojiLink(eb, *target)
		eb.Plain(" ")
		eb.Plain(helpers.Gendered(target.User.Gender, "разбанен", "разбанена"))
		return nil
	})), nil)
	return err
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
	u, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	s := ctx.NumberOrDefault(0)
	if s < 0 || s > 5 {
		return errors.New("toggle rights invalid status")
	}
	status := model.Status(s)
	if err := h.service.SetDevStatus(ctx.StdContext(), *u, status); err != nil {
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
func (h *Handler) ClearWarns(ctx *command.Context, u *ext.Update) error {
	target, err := ctx.User()
	if err != nil {
		return err
	}

	if err := h.service.ClearWarns(ctx.StdContext(), target.ChatID, target.User.ID); err != nil {
		_, _ = ctx.Reply(u, ext.ReplyTextString("Не удалось очистить предупреждения"), nil)
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		eb.Plain(view.FormatWarnsCleared(*target))
		return nil
	})), nil)
	return err
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
