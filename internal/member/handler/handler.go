package handler

import (
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
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"unicode/utf8"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
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
		return fmt.Errorf("update members list: get chat: %w", err)
	}
	count, err := h.service.SyncChatMembers(ctx.StdContext(), c.ID)
	if err != nil {
		_ = ctx.ReplyOnly(u, options.WithText("Не удалось обновить данные чата"))
		return fmt.Errorf("update members list: sync chat members: %w", err)
	}

	return ctx.ReplyOnly(u, options.WithText(memberview.FormatSyncResult(count)))
}

func (h *Handler) ListRoles(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return fmt.Errorf("list roles: get chat: %w", err)
	}
	members, err := h.service.GetMembersWithTitle(ctx.StdContext(), c.ID)
	if err != nil {
		_ = ctx.ReplyOnly(u, options.WithText("Не удалось получить список ролей"))
		return fmt.Errorf("list roles: get members with title: %w", err)
	}

	if len(members) == 0 {
		return ctx.ReplyOnly(u, options.WithText("В чате нет установленных ролей"))
	}
	eb := &entity.Builder{}
	memberview.WriteRolesList(eb, members)
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}
func (h *Handler) SetRole(ctx *command.Context, u *ext.Update) error {
	cm, err := ctx.AnyUser()
	if err != nil {
		return fmt.Errorf("set role: resolve user: %w", err)
	}
	tag, err := ctx.Text()
	if err != nil {
		return fmt.Errorf("set role: parse tag: %w", err)
	}

	if utf8.RuneCountInString(tag) > 16 {
		return ctx.ReplyOnly(u, options.WithText("Слишком длинная роль (максимум 16 символа)"))
	}

	if err := h.service.SetMemberTitle(ctx.StdContext(), cm.ChatID, cm.User.ID, tag); err != nil {
		return fmt.Errorf("failed to set member title: %w", err)
	}

	c, err := ctx.Chat()
	if err != nil {
		return fmt.Errorf("set role: get chat: %w", err)
	}

	if c.TagsEnabled {
		if _, err := ctx.Raw.MessagesEditChatParticipantRank(ctx.StdContext(), &tg.MessagesEditChatParticipantRankRequest{
			Peer:        u.GetChannel().AsInputPeer(),
			Participant: &tg.InputPeerUser{UserID: cm.User.ID},
			Rank:        tag,
		}); err != nil {
			_ = ctx.ReplyOnly(u, options.WithText(
				fmt.Sprintf("Телеграм не позволил установить роль участнику.\nПроверьте, есть ли у бота право на изменение тегов участников\n\n%s", err.Error()),
			))
			return fmt.Errorf("failed to set rank: %w", err)
		}
	} else {
		if _, err := ctx.Raw.ChannelsEditAdmin(ctx, &tg.ChannelsEditAdminRequest{
			Channel: u.GetChannel().AsInput(),
			UserID:  &tg.InputUser{UserID: cm.User.ID},
			AdminRights: tg.ChatAdminRights{
				Other: true,
			},
			Rank: tag,
		}); err != nil {
			_ = ctx.ReplyOnly(u, options.WithText(
				fmt.Sprintf("Телеграм не позволил установить роль участнику.\nПроверьте, есть ли у бота право на добавление администраторов\n\n%s", err.Error()),
			))
			return fmt.Errorf("set role: channels edit admin: %w", err)
		}
	}

	eb := &entity.Builder{}
	memberview.WriteRoleUpdated(eb, *cm, tag)
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) ShowRole(ctx *command.Context, u *ext.Update) error {
	cm, err := ctx.AnyUser()
	if err != nil {
		return fmt.Errorf("show role: resolve user: %w", err)
	}

	if cm.Tag == "" {
		eb := &entity.Builder{}
		eb.Plain("У участника ")
		helpers.WriteRoleEmojiLink(eb, *cm)
		eb.Plain("еще не установлена роль\n\nПопробуйте установить командой: ")
		eb.Code("!роль @участник Название Роли")
		return ctx.ReplyOnly(u, options.WithBuilder(eb))
	}

	return ctx.ReplyOnly(u, options.WithText(memberview.FormatMemberRole(cm.Tag)))
}

func (h *Handler) OnJoinMember(ctx *command.Context, u *ext.Update) error {
	log.Println("join member")
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
			false,
		); err != nil {
			return fmt.Errorf("join member: ensure member exists: %w", err)
		}
	}

	switch msg.Action.(type) {
	case *tg.MessageActionChatJoinedByLink,
		*tg.MessageActionChatJoinedByRequest:

		chatData, err := ctx.Chat()
		if err != nil {
			return fmt.Errorf("join member: get chat: %w", err)
		}

		if chatData.CallOnJoin {
			members, err := h.callService.GetAllMembers(ctx.StdContext(), effectiveChat.GetID())
			if err != nil {
				return fmt.Errorf("join member: get call members: %w", err)
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
		return fmt.Errorf("left member: get chat: %w", err)
	}
	action, ok := msg.Action.(*tg.MessageActionChatDeleteUser)
	if !ok {
		return nil
	}

	userID := action.UserID
	slog.Info("member left", "chat_id", u.EffectiveChat().GetID(), "user_id", userID)

	m, err := h.service.ProcessLeftMember(ctx.StdContext(), u.EffectiveChat().GetID(), userID)
	if err != nil {
		return fmt.Errorf("left member: process leave: %w", err)
	}
	eb := &entity.Builder{}
	eb.Plain("🕊 ")
	helpers.WriteRoleEmojiLink(eb, m)
	eb.Plain(fmt.Sprintf(" %s нас", helpers.Gendered(m.User.Gender, "покинул", "покинула")))
	admins, err := h.adminService.GetAdminsEnsured(ctx.StdContext(), u.EffectiveChat().GetID(), h.service.SyncChatMembers)
	if err != nil {
		return fmt.Errorf("left member: get admins ensured: %w", err)
	}
	eb.Plain("\n\n")
	for _, a := range admins {
		view.RenderMention(eb, a, c.MentionTypes)
	}

	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) OnBotPromote(ctx *command.Context, u *ext.Update) error {
	log.Println("bot promote")
	effectiveChat := u.EffectiveChat()
	count, err := h.service.SyncChatMembers(ctx.StdContext(), effectiveChat.GetID())
	if err != nil {
		return fmt.Errorf("bot promote: sync chat members: %w", err)
	}
	ch := u.GetChat()
	channel := u.GetChannel()
	var title string
	if ch != nil {
		title = ch.Title
	} else if channel != nil {
		title = channel.Title
	}
	if err := h.chatService.SetTitle(ctx.StdContext(), effectiveChat.GetID(), title); err != nil {
		return fmt.Errorf("bot promote: set chat title: %w", err)
	}

	slog.Info("updated chat members on bot join", "chat_id", effectiveChat.GetID(), "count", count)
	return nil
}

func (h *Handler) ShipRandom(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return fmt.Errorf("ship random: get chat: %w", err)
	}

	members, err := h.service.GetChatMembersIncludingBots(ctx.StdContext(), c.ID)
	if err != nil {
		return fmt.Errorf("ship random: get chat members: %w", err)
	}
	if len(members) == 0 {
		return ctx.ReplyOnly(u, options.WithText("Нет участников для шипа"))
	}

	first := members[rand.Intn(len(members))]
	second := members[rand.Intn(len(members))]

	phrases := []string{
		"Любите друг друга и берегите",
		"Кажется, это судьба",
		"Это выглядит подозрительно идеально",
	}

	selfPhrases := []string{
		"Самодостаточность — новый тренд",
		"Любовь начинается с себя ❤️",
		"Ну тут без вариантов",
	}

	botPhrases := []string{
		"Любовь с машиной? Будущее наступило",
		"Главное — не забудь обновить прошивку отношений",
	}

	botBotPhrases := []string{
		"💞 Союз машин заключён. Человечеству стоит насторожиться...",
		"Два бота нашли друг друга. Скайнет близко",
		"01001100 01001111 01010110 01000101",
	}

	eb := &entity.Builder{}
	helpers.WriteCustomEmoji(eb, "5258276353949575281", "❤️")
	eb.Bold(" Шипперим рандом: ")

	helpers.WriteRoleEmojiMention(eb, first)
	eb.Plain(" + ")
	helpers.WriteRoleEmojiMention(eb, second)
	eb.Plain("\n\n")

	if first.User.ID == second.User.ID {
		phrase := selfPhrases[rand.Intn(len(selfPhrases))]
		eb.Plain(phrase)
		return ctx.ReplyOnly(u, options.WithBuilder(eb))
	}

	isBot1 := first.User.IsBot
	isBot2 := second.User.IsBot

	if isBot1 && isBot2 {
		phrase := botBotPhrases[rand.Intn(len(botBotPhrases))]
		eb.Plain(phrase)
		return ctx.ReplyOnly(u, options.WithBuilder(eb))
	}

	if isBot1 || isBot2 {
		phrase := botPhrases[rand.Intn(len(botPhrases))]
		eb.Plain(phrase)
		return ctx.ReplyOnly(u, options.WithBuilder(eb))
	}

	phrase := phrases[rand.Intn(len(phrases))]
	eb.Plain(phrase)

	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) ShowEmoji(ctx *command.Context, u *ext.Update) error {
	cm, err := ctx.AnyUser()
	if err != nil {
		return fmt.Errorf("show member emoji: resolve user: %w", err)
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
		return fmt.Errorf("set member emoji: resolve user: %w", err)
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
		return fmt.Errorf("remove member emoji: resolve user: %w", err)
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
