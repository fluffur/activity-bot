package handler

import (
	"activity-bot/internal/adapter"
	"activity-bot/internal/admin"
	service "activity-bot/internal/call"
	"activity-bot/internal/call/view"
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/logger"
	"activity-bot/internal/member"
	memberview "activity-bot/internal/member/view"
	"activity-bot/internal/user"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/PaulSonOfLars/gotgbot/v2"
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

func (h *Handler) UpdateMembersList(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	count, err := h.service.SyncChatMembers(ctx.StdContext(), c.ID)
	if err != nil {
		_ = ctx.Reply(b, "Не удалось обновить данные чата", nil)
		return err
	}

	return ctx.Reply(b, memberview.FormatSyncResult(count), nil)
}

func (h *Handler) ListRoles(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	members, err := h.service.GetMembersWithTitle(ctx.StdContext(), c.ID)
	if err != nil {
		_ = ctx.Reply(b, "Не удалось получить список ролей", nil)
		return err
	}

	if len(members) == 0 {
		return ctx.Reply(b, "В чате нет установленных ролей", nil)
	}

	return ctx.ReplyHTML(b, memberview.FormatRolesList(members))
}
func (h *Handler) SetRole(b *gotgbot.Bot, ctx *command.Context) error {
	u, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	tag, err := ctx.Text()
	if err != nil {
		return err
	}

	if utf8.RuneCountInString(tag) > 16 {
		return ctx.Reply(b, "Слишком длинная роль (максимум 16 символа)", nil)
	}

	if err := h.service.SetMemberTitle(ctx.StdContext(), u.ChatID, u.User.ID, tag); err != nil {
		if errors.Is(err, adapter.ErrChatMemberNotFound) {
			return ctx.Reply(b, fmt.Sprintf("Участник не найден\n\nTelegram: %s", err.Error()), nil)
		} else if errors.Is(err, adapter.ErrChatMemberCantBeEdited) {
			return ctx.Reply(b, fmt.Sprintf("Я не могу изменить роль этого участника\n\nTelegram: %s", err.Error()), nil)
		} else if errors.Is(err, adapter.ErrChatMemberIsRestricted) {
			return ctx.Reply(b, fmt.Sprintf("Пользователь не является полноправным участником чата\n\nTelegram: %s", err.Error()), nil)
		} else if errors.Is(err, adapter.ErrChatMemberIsCreator) {
			return ctx.Reply(b, "Я не могу менять роль создателя чата", nil)
		}

		return fmt.Errorf("failed to set member title: %w", err)
	}

	return ctx.ReplyHTML(b, memberview.FormatRoleUpdated(*u, tag))
}

func (h *Handler) RestoreRoles(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	members, err := h.service.GetAnyMembersWithTitle(ctx.StdContext(), c.ID)
	if err != nil {
		_ = ctx.Reply(b, "Не удалось получить список ролей из базы", nil)
		return err
	}

	if len(members) == 0 {
		return ctx.Reply(b, "В базе данных нет сохраненных ролей для этого чата", nil)
	}

	var restoredCount int
	limiter := rate.NewLimiter(rate.Every(500*time.Millisecond), 1)

	for _, m := range members {
		if err := limiter.Wait(ctx.StdContext()); err != nil {
			return err
		}

		if ok, err := b.PromoteChatMember(c.ID, m.User.ID, &gotgbot.PromoteChatMemberOpts{
			CanManageChat:   true,
			CanPostMessages: true,
			CanEditMessages: true,
		}); err != nil || !ok {
			logger.L.Warn("failed to promote chat member", "chatID", c.ID, "userID", m.User.ID, "error", err)
			continue
		}

		restoredCount++
	}

	msgText := fmt.Sprintf("%s Восстановление завершено.\n\nВосстановлено: %d", helpers.SuccessEmoji(), restoredCount)

	return ctx.Reply(b, msgText, nil)
}

func (h *Handler) ShowRole(b *gotgbot.Bot, ctx *command.Context) error {
	u, err := ctx.User()
	if err != nil {
		return err
	}

	if u.Tag == "" {
		return ctx.ReplyHTML(b, fmt.Sprintf("У участника %s нет роли", helpers.UserLink(u.User)))
	}

	return ctx.ReplyHTML(b, memberview.FormatMemberRole(u.User, u.Tag))
}

func (h *Handler) OnJoinMember(b *gotgbot.Bot, ctx *command.Context) error {
	joinedMembers := ctx.EffectiveMessage.NewChatMembers
	for _, u := range joinedMembers {
		if u.Id == b.User.Id {
			return h.OnBotPromote(b, ctx)
		}
		if u.IsBot {
			continue
		}
		slog.Info("member joined", "chat_id", ctx.EffectiveChat.Id, "user_id", u.Id)
		if _, err := h.service.EnsureMemberExists(ctx.StdContext(), ctx.EffectiveChat.Id, u.Id, u.Username, u.FirstName, u.LastName, ctx.EffectiveMessage.SenderTag); err != nil {
			return err
		}
	}

	chatData, err := h.chatService.GetChat(ctx.StdContext(), ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	if chatData.CallOnJoin {
		members, err := h.callService.GetAllMembers(ctx.StdContext(), ctx.EffectiveChat.Id)
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

			chunkText := view.FormatCallChunk(message, members[i:end], chatData.MentionTypes)
			if _, sendErr := ctx.EffectiveMessage.Reply(b, chunkText, &gotgbot.SendMessageOpts{
				ParseMode: gotgbot.ParseModeHTML,
				LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
					IsDisabled: true,
				},
			}); sendErr != nil {
				return sendErr
			}
		}
	}

	return nil
}

func (h *Handler) OnLeftMember(b *gotgbot.Bot, ctx *command.Context) error {
	u := ctx.Message.LeftChatMember
	slog.Info("member left", "chat_id", ctx.EffectiveChat.Id, "user_id", u.Id)
	if u.IsBot {
		return nil
	}
	m, err := h.service.ProcessLeftMember(ctx.StdContext(), ctx.EffectiveChat.Id, u.Id)
	if err != nil {
		return err
	}

	admins, err := h.adminService.GetAdminsEnsured(ctx.StdContext(), ctx.EffectiveChat.Id, h.service.SyncChatMembers)
	if err != nil {
		return err
	}
	var sb strings.Builder
	for _, a := range admins {
		sb.WriteString(helpers.Mention(a.User.ID, "​"))
	}
	_, err = ctx.EffectiveChat.SendMessage(b, fmt.Sprintf("🕊 %s %s нас..."+sb.String(),
		helpers.RoleEmojiLink(m),
		helpers.Gendered(m.User.Gender, "покинул", "покинула"),
	), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})

	return err
}

func (h *Handler) OnBotPromote(_ *gotgbot.Bot, ctx *command.Context) error {
	count, err := h.service.SyncChatMembers(ctx.StdContext(), ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}
	if err := h.chatService.SetTitle(ctx.StdContext(), ctx.EffectiveChat.Id, ctx.EffectiveChat.Title); err != nil {
		return err
	}
	slog.Info("updated chat members on bot join", "chat_id", ctx.EffectiveChat.Id, "count", count)
	return nil
}

func (h *Handler) SetOnlyNewbies(b *gotgbot.Bot, ctx *cmd.Context) error {
	if len(ctx.Users()) == 0 {
		return ctx.Reply(b, "Укажите хотя бы одного участника", nil)
	}
	if err := h.service.SetOnlyNewbies(ctx.StdContext(), ctx.TargetChatID(), ctx.Users()); err != nil {
		_ = ctx.Reply(b, "Не удалось установить олдов", nil)
		return err
	}

	return ctx.Reply(b, "Олды установлены", nil)
}

func (h *Handler) SetNewbies(b *gotgbot.Bot, ctx *cmd.Context) error {
	if len(ctx.Users()) == 0 {
		return ctx.Reply(b, "Укажите хотя бы одного участника", nil)
	}
	if err := h.service.SetNewbies(ctx.StdContext(), ctx.TargetChatID(), ctx.Users()); err != nil {
		return err
	}

	return ctx.Reply(b, "Новички установлены", nil)
}

func (h *Handler) ShipRandom(b *gotgbot.Bot, ctx *command.Context) error {
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

	return ctx.ReplyHTML(b, text)
}

func (h *Handler) ShowEmoji(b *gotgbot.Bot, ctx *command.Context) error {
	m, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	if m.Emoji == "" {
		return ctx.ReplyHTML(b, fmt.Sprintf("У %s еще нет значка чата\n\nДобавить значок: <code>!значок @участник 😘</code>", helpers.RoleEmojiLink(*m)))

	}
	return ctx.ReplyHTML(b, fmt.Sprintf("Значок %s: %s", helpers.RoleLink(*m), m.Emoji))
}

func (h *Handler) SetEmoji(b *gotgbot.Bot, ctx *command.Context) error {
	m, err := ctx.AnyUser()
	if err != nil {
		return err
	}

	graphemes := helpers.ParseEmojis(ctx.RawArgsHTML)
	emojis := strings.Join(graphemes, "")
	if len(graphemes) > 3 {
		return ctx.Reply(b, "❌ Можно указать не более 3 значков на участника", nil)
	}
	if err := h.service.SetChatMemberEmoji(ctx.StdContext(), m.ChatID, m.User.ID, emojis); err != nil {
		return fmt.Errorf("failed to set chat member emoji: %w", err)
	}

	return ctx.ReplyHTML(b, fmt.Sprintf("Значок %s для %s успешно установлен", emojis, helpers.RoleLink(*m)))
}

func (h *Handler) RemoveEmoji(b *gotgbot.Bot, ctx *command.Context) error {
	m, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	if err := h.service.SetChatMemberEmoji(ctx.StdContext(), m.ChatID, m.User.ID, ""); err != nil {
		return fmt.Errorf("failed to set remove member emoji: %w", err)
	}

	return ctx.ReplyHTML(b, fmt.Sprintf("Значок %s для %s успешно удалён", "", helpers.RoleLink(*m)))
}
