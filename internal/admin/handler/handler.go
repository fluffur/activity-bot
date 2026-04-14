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
	"activity-bot/internal/options"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
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

func (h *Handler) IsAdmin(ctx *command.Context, u *ext.Update) error {
	cm, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	return ctx.ReplyOnly(u, options.WithText(fmt.Sprintf("Ранг участника: %d", cm.Status)))
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
			return ctx.ReplyOnly(u, options.WithText("Участник уже является администратором"))
		}
		if errors.Is(err, admin.ErrUserIsCreator) {
			return ctx.ReplyOnly(u, options.WithText("Нельзя изменить статус владельца"))
		}
		if errors.Is(err, admin.ErrUserStatusInvalid) {
			return ctx.ReplyOnly(u, options.WithText("Нельзя установить участнику статус равный своему или выше"))
		}
		if errors.Is(err, admin.ErrUserIsNotPermitted) {
			return ctx.ReplyOnly(u, options.WithText("Недостаточно прав для изменения статуса"))
		}

		return ctx.ReplyOnly(u, options.WithText("Не удалось добавить администратора"))
	}

	eb := &entity.Builder{}
	view.WriteAdminAdded(eb, *m, status)
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
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
			return ctx.ReplyOnly(u, options.WithText("Пользователь не является администратором"))
		}

		if errors.Is(err, admin.ErrUserIsCreator) {
			return ctx.ReplyOnly(u, options.WithText("Нельзя удалить создателя из списка администраторов"))
		}

		return ctx.ReplyOnly(u, options.WithText("Не удалось удалить администратора"))
	}

	eb := &entity.Builder{}
	view.WriteAdminRemoved(eb, *target)
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) ListAdmins(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	admins, err := h.service.GetAdminsEnsured(ctx.StdContext(), c.ID, h.memberService.SyncChatMembers)
	if err != nil {
		return ctx.ReplyOnly(u, options.WithText("Не удалось получить список администраторов"))
	}

	if len(admins) == 0 {
		return ctx.ReplyOnly(u, options.WithText("Список администраторов пуст"))
	}

	eb := &entity.Builder{}
	view.WriteAdminsList(eb, admins)
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
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

	ebDM := &entity.Builder{}
	view.WriteDirectModerationAction(ebDM, *target, c.Title, "kick", time.Time{}, reason)
	caption, entities := ebDM.Complete()
	_, _ = ctx.Context.SendMessage(target.User.ID, &tg.MessagesSendMessageRequest{
		Message:  caption,
		Entities: entities,
	})

	if err := h.service.Kick(ctx.StdContext(), *target, *mod, reason); err != nil {
		if errors.Is(err, admin.ErrUserIsProtected) {
			return ctx.ReplyOnly(u, options.WithText("Нельзя кикнуть администратора или создателя"))
		}
		_, _ = h.memberService.ProcessLeftMember(ctx.StdContext(), target.ChatID, target.User.ID)
		return fmt.Errorf("failed to kick: %w", err)
	}

	if _, err := ctx.BanChatMember(c.ID, target.User.ID, 0); err != nil {
		return fmt.Errorf("failed to kick: %w", err)
	}

	_, _ = h.memberService.ProcessLeftMember(ctx.StdContext(), target.ChatID, target.User.ID)

	eb := &entity.Builder{}
	view.WriteModerationAction(eb, *target, "kick", time.Time{}, reason)
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
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
			return ctx.ReplyOnly(u, options.WithText("Нельзя забанить администратора или создателя"))
		}
		_, _ = h.memberService.ProcessLeftMember(ctx.StdContext(), c.ID, target.User.ID)
		return fmt.Errorf("failed to ban: %w", err)
	}
	if _, err := ctx.BanChatMember(c.ID, target.User.ID, 0); err != nil {
		return fmt.Errorf("failed to ban: %w", err)
	}

	ebDM := &entity.Builder{}
	view.WriteDirectModerationAction(ebDM, *target, c.Title, "ban", until, reason)
	caption, entities := ebDM.Complete()
	_, _ = ctx.Context.SendMessage(target.User.ID, &tg.MessagesSendMessageRequest{
		Message:  caption,
		Entities: entities,
	})

	_, _ = h.memberService.ProcessLeftMember(ctx.StdContext(), target.ChatID, target.User.ID)

	eb := &entity.Builder{}
	view.WriteModerationAction(eb, *target, "ban", until, reason)
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) Mute(ctx *command.Context, u *ext.Update) error {
	target, err := ctx.User()
	if err != nil {
		return err
	}
	c, err := ctx.Chat()
	if err != nil {
		return fmt.Errorf("chat %w", err)
	}
	until := ctx.DateOrDefault(time.Now().Add(time.Hour * 24 * 7 * 2))
	reason := ctx.TextOrDefault("")
	if reason == "навсегда" {
		reason = ""
		until = time.Time{}
	}
	mod, err := h.memberService.GetChatMember(ctx.StdContext(), target.ChatID, u.EffectiveUser().GetID())
	if err != nil {
		return err
	}

	if err := h.service.Mute(ctx.StdContext(), *target, mod, until, reason); err != nil {
		if errors.Is(err, admin.ErrUserIsProtected) {
			return ctx.ReplyOnly(u, options.WithText("Нельзя замутить администратора или создателя"))
		}
		if errors.Is(err, admin.ErrInvalidRange) {
			return ctx.ReplyOnly(u, options.WithText("Срок ограничения должен быть от 30 секунд до 366 дней"))
		}

		return err
	}

	chatPeer, err := ctx.ResolveInputPeerById(c.ID)
	if err != nil {
		return fmt.Errorf("failed to resolve chat peer: %w", err)
	}
	channelPeer, ok := chatPeer.(*tg.InputPeerChannel)
	if !ok {
		return fmt.Errorf("chat %d is not a channel/supergroup", c.ID)
	}

	participantPeer, err := ctx.ResolveInputPeerById(target.User.ID)
	if err != nil {
		return fmt.Errorf("failed to resolve participant peer: %w", err)
	}

	bannedRights := tg.ChatBannedRights{
		SendMessages:    true,
		SendMedia:       true,
		SendStickers:    true,
		SendGifs:        true,
		SendGames:       true,
		SendInline:      true,
		EmbedLinks:      true,
		SendPolls:       true,
		SendPhotos:      true,
		SendVideos:      true,
		SendRoundvideos: true,
		SendAudios:      true,
		SendVoices:      true,
		SendDocs:        true,
		SendPlain:       true,
	}
	if !until.IsZero() {
		bannedRights.UntilDate = int(until.Unix())
	}

	if _, err := ctx.Raw.ChannelsEditBanned(ctx, &tg.ChannelsEditBannedRequest{
		Channel: &tg.InputChannel{
			ChannelID:  channelPeer.ChannelID,
			AccessHash: channelPeer.AccessHash,
		},
		Participant:  participantPeer,
		BannedRights: bannedRights,
	}); err != nil {
		return fmt.Errorf("failed to mute: %w", err)
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

	eb := &entity.Builder{}
	view.WriteModerationAction(eb, *target, "mute", until, reason)
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
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
		eb := &entity.Builder{}
		helpers.WriteSuccessEmoji(eb)
		eb.Plain(" У ")
		helpers.WriteRoleEmojiLink(eb, *m)
		eb.Plain(" нет активных варнов")
		return ctx.ReplyOnly(u, options.WithBuilder(eb))
	}

	eb := &entity.Builder{}
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
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) WarnList(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	warns, err := h.service.GetWarnsByChat(ctx.StdContext(), c.ID)
	if err != nil {
		return ctx.ReplyOnly(u, options.WithText("Не удалось получить список предупреждений"))
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

	eb := &entity.Builder{}
	view.WriteWarnlist(eb, activeWarns, maxWarns)
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
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
			return ctx.ReplyOnly(u, options.WithText("Нельзя выдать предупреждение вышестоящему лицу"))
		}
		return fmt.Errorf("failed to give warn: %w", err)
	}
	if banned {
		if _, err := ctx.BanChatMember(target.ChatID, target.User.ID, 0); err != nil {
			return fmt.Errorf("failed to ban: %w", err)
		}

		_, _ = h.memberService.ProcessLeftMember(ctx.StdContext(), target.ChatID, target.User.ID)
	}
	maxWarns, _ := h.service.GetMaxWarns(ctx.StdContext(), target.ChatID)

	eb := &entity.Builder{}
	view.WriteWarnInfo(eb, *target, count, maxWarns, until, reason, banned)
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) ShowMaxWarns(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	return ctx.ReplyOnly(u, options.WithText(fmt.Sprintf("Текущий лимит предупреждений: %d", c.MaxWarns)))
}

func (h *Handler) SetMaxWarns(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	maxWarns := ctx.NumberOrDefault(3)
	if maxWarns <= 0 {
		return ctx.ReplyOnly(u, options.WithText("Лимит предупреждений должен быть положительным числом"))
	}

	if err := h.service.SetMaxWarns(ctx.StdContext(), c.ID, maxWarns); err != nil {
		_ = ctx.ReplyOnly(u, options.WithText("Не удалось обновить лимит предупреждений"))
		return err
	}

	return ctx.ReplyOnly(u, options.WithText(fmt.Sprintf("Лимит предупреждений изменен на %d", maxWarns)))
}

func (h *Handler) Unban(ctx *command.Context, u *ext.Update) error {
	target, err := ctx.User()
	if err != nil {
		return err
	}
	if err := h.service.Unban(ctx.StdContext(), target.ChatID, target.User.ID); err != nil {
		return ctx.ReplyOnly(u, options.WithText("Не удалось разбанить пользователя"))
	}

	if _, err := ctx.UnbanChatMember(target.ChatID, target.User.ID); err != nil {
		return fmt.Errorf("failed to unban: %w", err)
	}

	eb := &entity.Builder{}
	eb.Plain("Пользователь ")
	helpers.WriteRoleEmojiLink(eb, *target)
	eb.Plain(" ")
	eb.Plain(helpers.Gendered(target.User.Gender, "разбанен", "разбанена"))
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) Unmute(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	cm, err := ctx.User()
	if err != nil {
		return err
	}

	chatPeer, err := ctx.ResolveInputPeerById(c.ID)
	if err != nil {
		return fmt.Errorf("failed to resolve chat peer: %w", err)
	}
	channelPeer, ok := chatPeer.(*tg.InputPeerChannel)
	if !ok {
		return fmt.Errorf("chat %d is not a channel/supergroup", c.ID)
	}

	participantPeer, err := ctx.ResolveInputPeerById(cm.User.ID)
	if err != nil {
		return fmt.Errorf("failed to resolve participant peer: %w", err)
	}

	if _, err := ctx.Raw.ChannelsEditBanned(ctx, &tg.ChannelsEditBannedRequest{
		Channel: &tg.InputChannel{
			ChannelID:  channelPeer.ChannelID,
			AccessHash: channelPeer.AccessHash,
		},
		Participant:  participantPeer,
		BannedRights: tg.ChatBannedRights{},
	}); err != nil {
		return fmt.Errorf("failed to unmute: %w", err)
	}

	eb := &entity.Builder{}
	view.WriteUnmuteInfo(eb, *cm)

	if c.TagsEnabled {
		return ctx.ReplyOnly(u, options.WithBuilder(eb))
	}

	title := cm.Tag
	if title != "" {
		if _, err := ctx.Raw.ChannelsEditAdmin(ctx, &tg.ChannelsEditAdminRequest{
			Channel: c.AsInputChannel(),
			UserID:  cm.User.AsInput(),
			AdminRights: tg.ChatAdminRights{
				ChangeInfo:     true,
				PostMessages:   true,
				EditMessages:   true,
				DeleteMessages: true,
				BanUsers:       true,
				InviteUsers:    true,
				PinMessages:    true,
				AddAdmins:      true,
				Anonymous:      true,
				ManageCall:     true,
				Other:          true,
			},
			Rank: title,
		}); err != nil {
			_ = ctx.ReplyOnly(u, options.WithText("Пользователь размучен, но не удалось вернуть роль"))
			return err
		}
	}

	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) Unwarn(ctx *command.Context, u *ext.Update) error {
	cm, err := ctx.User()
	if cm == nil {
		return err
	}

	count, err := h.service.Unwarn(ctx.StdContext(), cm.ChatID, cm.User.ID)
	if err != nil {
		return ctx.ReplyOnly(u, options.WithText("Не удалось снять предупреждение"))
	}

	maxWarns, _ := h.service.GetMaxWarns(ctx.StdContext(), cm.ChatID)

	eb := &entity.Builder{}
	view.WriteUnwarnInfo(eb, *cm, count, maxWarns)
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) ToggleRights(ctx *command.Context, u *ext.Update) error {
	cm, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	s := ctx.NumberOrDefault(0)
	if s < 0 || s > 5 {
		return errors.New("toggle rights invalid status")
	}
	status := model.Status(s)
	if err := h.service.SetDevStatus(ctx.StdContext(), *cm, status); err != nil {
		return fmt.Errorf("failed to set dev status: %w", err)
	}

	return ctx.ReplyOnly(u,
		options.WithText(fmt.Sprintf("Права разработчика изменены на: %s", status.String())),
	)
}

func (h *Handler) UpdateChats(ctx *command.Context, u *ext.Update) error {
	chats, err := h.chatService.GetChatsWithoutTitle(ctx.StdContext())
	if err != nil {
		return err
	}

	limiter := rate.NewLimiter(rate.Every(1000*time.Millisecond), 2)

	for _, c := range chats {
		if err := limiter.Wait(ctx.StdContext()); err != nil {
			return err
		}

		ch, err := ctx.Raw.MessagesGetFullChat(ctx, c.ID)
		if err != nil {
			slog.Error("failed to get chat", "chat", c, "err", err)
			continue
		}
		title := ""
		for _, chat := range ch.Chats {
			if chat.GetID() == c.ID {
				if chObj, ok := chat.(*tg.Chat); ok {
					title = chObj.Title
				}
				break
			}
		}

		logger.L.Info("found chat title", "title", title, "id", c.ID)
		if err := h.chatService.SetTitle(ctx.StdContext(), c.ID, title); err != nil {
			return err
		}
	}
	return ctx.ReplyOnly(u, options.WithText("Чаты обновлены"))
}
func (h *Handler) ClearWarns(ctx *command.Context, u *ext.Update) error {
	target, err := ctx.User()
	if err != nil {
		return err
	}

	if err := h.service.ClearWarns(ctx.StdContext(), target.ChatID, target.User.ID); err != nil {
		return ctx.ReplyOnly(u, options.WithText("Не удалось очистить предупреждения"))
	}

	eb := &entity.Builder{}
	view.WriteWarnsCleared(eb, *target)
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) FakeLeave(ctx *command.Context, u *ext.Update) error {
	m, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	eu := m.User

	eb := &entity.Builder{}
	eb.Plain("🕊 ")
	helpers.WriteRoleEmojiLink(eb, *m)
	eb.Plain(" ")
	eb.Plain(helpers.Gendered(eu.Gender, "покинул", "покинула"))
	eb.Plain(" нас...")

	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) DemoteTgAdmin(ctx *command.Context, u *ext.Update) error {
	cm, err := ctx.User()
	if err != nil {
		return err
	}

	if _, err := ctx.Raw.ChannelsEditAdmin(ctx, &tg.ChannelsEditAdminRequest{
		Channel:     cm.AsInputChannel(),
		UserID:      cm.User.AsInput(),
		AdminRights: tg.ChatAdminRights{},
	}); err != nil {
		return err
	}

	eb := &entity.Builder{}
	eb.Plain("Участник ")
	helpers.WriteRoleEmojiLink(eb, *cm)
	eb.Plain(" " + helpers.Gendered(cm.User.Gender, "разжалован", "разжалована"))
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}
