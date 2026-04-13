package handler

import (
	"activity-bot/internal/adapter"
	"activity-bot/internal/admin"
	service "activity-bot/internal/call"
	"activity-bot/internal/call/view"
	"activity-bot/internal/chat"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/logger"
	"activity-bot/internal/member"
	memberview "activity-bot/internal/member/view"
	"activity-bot/internal/options"
	"activity-bot/internal/user"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"time"
	"unicode/utf8"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
	"golang.org/x/time/rate"
)

type Handler struct {
	service      *member.Service
	chatService  *chat.Service
	userService  *user.Service
	callService  *service.Service
	adminService *admin.Service
}

func New(service *member.Service, chatService *chat.Service, userService *user.Service, callService *service.Service, adminService *admin.Service) *Handler {
	return &Handler{service, chatService, userService, callService, adminService}
}

func (h *Handler) UpdateMembersList(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	count, err := h.service.SyncChatMembers(ctx.StdContext(), c.ID)
	if err != nil {
		_ = ctx.ReplyOnly(u, options.WithText("Не удалось обновить данные чата"))
		return err
	}

	return ctx.ReplyOnly(u, options.WithText(memberview.FormatSyncResult(count)))
}

func (h *Handler) ListRoles(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	members, err := h.service.GetMembersWithTitle(ctx.StdContext(), c.ID)
	if err != nil {
		_ = ctx.ReplyOnly(u, options.WithText("Не удалось получить список ролей"))
		return err
	}

	if len(members) == 0 {
		return ctx.ReplyOnly(u, options.WithText("В чате нет установленных ролей"))
	}

	return ctx.ReplyOnly(u, options.WithText(memberview.FormatRolesList(members)))
}
func (h *Handler) SetRole(ctx *command.Context, u *ext.Update) error {
	cm, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	tag, err := ctx.Text()
	if err != nil {
		return err
	}

	if utf8.RuneCountInString(tag) > 16 {
		return ctx.ReplyOnly(u, options.WithText("Слишком длинная роль (максимум 16 символа)"))
	}

	if err := h.service.SetMemberTitle(ctx.StdContext(), cm.ChatID, cm.User.ID, tag); err != nil {
		if errors.Is(err, adapter.ErrChatMemberNotFound) {
			return ctx.ReplyOnly(u, options.WithText(fmt.Sprintf("Участник не найден\n\nTelegram: %s", err.Error())))
		} else if errors.Is(err, adapter.ErrChatMemberCantBeEdited) {
			return ctx.ReplyOnly(u, options.WithText(fmt.Sprintf("Я не могу изменить роль этого участника\n\nTelegram: %s", err.Error())))
		} else if errors.Is(err, adapter.ErrChatMemberIsRestricted) {
			return ctx.ReplyOnly(u, options.WithText(fmt.Sprintf("Пользователь не является полноправным участником чата\n\nTelegram: %s", err.Error())))
		} else if errors.Is(err, adapter.ErrChatMemberIsCreator) {
			return ctx.ReplyOnly(u, options.WithText("Я не могу менять роль создателя чата"))
		}

		_ = ctx.ReplyOnly(u, options.WithText(fmt.Sprintf("Не удалось установить роль: %s", err.Error())))

		return fmt.Errorf("failed to set member title: %w", err)
	}

	return ctx.ReplyOnly(u, options.WithText(memberview.FormatRoleUpdated(*cm, tag)))
}

func (h *Handler) RestoreRoles(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	members, err := h.service.GetAnyMembersWithTitle(ctx.StdContext(), c.ID)
	if err != nil {
		_ = ctx.ReplyOnly(u, options.WithText("Не удалось получить список ролей из базы"))
		return err
	}

	if len(members) == 0 {
		return ctx.ReplyOnly(u, options.WithText("В базе данных нет сохраненных ролей для этого чата"))
	}

	var restoredCount int
	limiter := rate.NewLimiter(rate.Every(500*time.Millisecond), 1)

	for _, m := range members {
		if err := limiter.Wait(ctx.StdContext()); err != nil {
			return err
		}

		if _, err := ctx.Raw.ChannelsEditAdmin(ctx, &tg.ChannelsEditAdminRequest{
			Channel: &tg.InputChannel{ChannelID: c.ID},
			UserID:  m.User.AsInput(),
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
			Rank: m.Tag,
		}); err != nil {
			logger.L.Warn("failed to promote chat member", "chatID", c.ID, "userID", m.User.ID, "error", err)
			continue
		}

		restoredCount++
	}

	msgText := fmt.Sprintf("%s Восстановление завершено.\n\nВосстановлено: %d", helpers.SuccessEmoji(), restoredCount)

	return ctx.ReplyOnly(u, options.WithText(msgText))
}

func (h *Handler) ShowRole(ctx *command.Context, u *ext.Update) error {
	cm, err := ctx.AnyUser()
	if err != nil {
		return err
	}

	if cm.Tag == "" {
		return ctx.ReplyOnly(u, options.WithText(fmt.Sprintf("У участника %s нет роли", helpers.UserLink(cm.User))))
	}

	return ctx.ReplyOnly(u, options.WithText(memberview.FormatMemberRole(cm.User, cm.Tag)))
}

func (h *Handler) OnJoinMember(ctx *command.Context, u *ext.Update) error {
	msg := u.EffectiveMessage
	if msg.Action == nil {
		return nil
	}

	effectiveChat := u.EffectiveChat()
	effectiveUser := u.EffectiveUser()

	var userIDs []int64

	switch action := msg.Action.(type) {

	case *tg.MessageActionChatAddUser:
		userIDs = action.Users

	case *tg.MessageActionChatJoinedByLink,
		*tg.MessageActionChatJoinedByRequest:
		userIDs = []int64{effectiveUser.ID}

	default:
		return nil
	}

	for _, id := range userIDs {
		if id == ctx.Self.ID {
			return h.OnBotPromote(ctx, u)
		}

		log.Println("member joined", id, effectiveChat, effectiveUser)

		if _, err := h.service.EnsureMemberExists(
			ctx.StdContext(),
			effectiveChat.GetID(),
			id,
			effectiveUser.Username,
			effectiveUser.FirstName,
			effectiveUser.LastName,
			"",
		); err != nil {
			return err
		}
	}

	switch msg.Action.(type) {
	case *tg.MessageActionChatJoinedByLink,
		*tg.MessageActionChatJoinedByRequest:

		chatData, err := ctx.Chat()
		if err != nil {
			return err
		}

		if chatData.CallOnJoin {
			members, err := h.callService.GetAllMembers(ctx.StdContext(), effectiveChat.GetID())
			if err != nil {
				return err
			}

			mentionsLimit := int(chatData.MentionsPerMessage)
			if mentionsLimit <= 0 {
				mentionsLimit = 5
			}

			message := chatData.WelcomeCallMessage
			if message != "" {
				message = view.ReplaceMentionsWithLinks(message)
			}

			for i := 0; i < len(members); i += mentionsLimit {
				end := i + mentionsLimit
				if end > len(members) {
					end = len(members)
				}

				eb := &entity.Builder{}
				view.FormatCallChunkBuilder(eb, message, members[i:end], chatData.MentionTypes)

				if err := ctx.ReplyOnly(u, options.WithBuilder(eb)); err != nil {
					logger.L.Error("failed to send on join call chuck", "error", err)
				}
			}
		}
	}

	return nil
}

func (h *Handler) OnLeftMember(ctx *command.Context, u *ext.Update) error {
	msg := u.EffectiveMessage
	if msg.Action == nil {
		return nil
	}
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	action, ok := msg.Action.(*tg.MessageActionChatDeleteUser)
	if !ok {
		return nil
	}

	userID := action.UserID
	slog.Info("member left", "chat_id", u.EffectiveChat().GetID(), "user_id", userID)

	m, err := h.service.ProcessLeftMember(ctx.StdContext(), u.EffectiveChat().GetID(), userID)
	if err != nil {
		return err
	}
	eb := &entity.Builder{}
	eb.Plain("🕊 ")
	helpers.WriteRoleEmojiLink(eb, m)
	eb.Plain(fmt.Sprintf(" %s нас", helpers.Gendered(m.User.Gender, "покинул", "покинула")))
	admins, err := h.adminService.GetAdminsEnsured(ctx.StdContext(), u.EffectiveChat().GetID(), h.service.SyncChatMembers)
	if err != nil {
		return err
	}
	eb.Plain("\n\n")
	for _, a := range admins {
		view.RenderMention(eb, a, c.MentionTypes)
	}

	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) OnBotPromote(ctx *command.Context, u *ext.Update) error {
	effectiveChat := u.GetChat()
	count, err := h.service.SyncChatMembers(ctx.StdContext(), effectiveChat.GetID())
	if err != nil {
		return err
	}
	if err := h.chatService.SetTitle(ctx.StdContext(), effectiveChat.GetID(), effectiveChat.GetTitle()); err != nil {
		return err
	}

	slog.Info("updated chat members on bot join", "chat_id", effectiveChat.GetID(), "count", count)
	return nil
}

func (h *Handler) ShipRandom(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	members, err := h.service.GetChatMembers(ctx.StdContext(), c.ID)
	if err != nil {
		return err
	}

	rand.Shuffle(len(members), func(i, j int) {
		members[i], members[j] = members[j], members[i]
	})

	phrases := []string{
		"Любите друг друга и берегите",
		"Кажется, это судьба",
	}

	phrase := phrases[rand.Intn(len(phrases))]

	first := members[0]
	second := members[1]

	text := fmt.Sprintf("%s <b>Шипперим рандом</b>: %s + %s\n%s", helpers.CustomEmoji("5258276353949575281", "❤️"), helpers.RoleMentionEmoji(first), helpers.RoleMentionEmoji(second), phrase)

	return ctx.ReplyOnly(u, options.WithText(text))
}

func (h *Handler) ShowEmoji(ctx *command.Context, u *ext.Update) error {
	cm, err := ctx.AnyUser()
	if err != nil {
		return err
	}

	eb := &entity.Builder{}

	if len(cm.Emojis) == 0 {
		eb.Plain("У ")
		helpers.WriteUserMention(eb, cm.User)
		eb.Plain(" еще нет значка чата\n\nДобавить: ")
		eb.Code("!значок @участник 💤")

		return ctx.ReplyOnly(u, options.WithBuilder(eb))
	}

	eb.Plain("Значок ")
	helpers.WriteUserMention(eb, cm.User)
	eb.Plain(": ")
	helpers.DisplayEmoji(eb, cm.Emojis)

	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) SetEmoji(ctx *command.Context, u *ext.Update) error {
	cm, err := ctx.AnyUser()
	if err != nil {
		return err
	}

	emojis := helpers.ExtractEmoji(ctx.RawArgs, ctx.RawArgsEntities)

	if len(emojis) == 0 {
		return ctx.ReplyOnly(u, options.WithText("Отправьте emoji"))
	}

	if len(emojis) > 3 {
		return ctx.ReplyOnly(u, options.WithText("❌ Можно указать не более 3 значков"))
	}

	if err := h.service.SetChatMemberEmoji(
		ctx.StdContext(),
		cm.ChatID,
		cm.User.ID,
		emojis,
	); err != nil {
		return fmt.Errorf("failed to set chat member emoji: %w", err)
	}

	eb := &entity.Builder{}
	eb.Plain("Значок ")
	helpers.DisplayEmoji(eb, emojis)
	eb.Plain(" установлен для ")
	helpers.WriteUserMention(eb, cm.User)

	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) RemoveEmoji(ctx *command.Context, u *ext.Update) error {
	cm, err := ctx.AnyUser()
	if err != nil {
		return err
	}

	if len(cm.Emojis) == 0 {
		return ctx.ReplyOnly(u, options.WithText("У пользователя нет значка"))
	}

	if err := h.service.SetChatMemberEmoji(
		ctx.StdContext(),
		cm.ChatID,
		cm.User.ID,
		nil,
	); err != nil {
		return fmt.Errorf("failed to remove member emoji: %w", err)
	}

	eb := &entity.Builder{}
	eb.Plain("Значок ")
	helpers.DisplayEmoji(eb, cm.Emojis)
	eb.Plain(" удалён у ")
	helpers.WriteUserMention(eb, cm.User)

	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}
