package handler

import (
	"activity-bot/internal/adapter"
	"activity-bot/internal/admin"
	service "activity-bot/internal/call"
	"activity-bot/internal/call/view"
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	memberview "activity-bot/internal/member/view"
	"activity-bot/internal/user"
	"errors"
	"fmt"
	"log"
	"log/slog"
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

func (h *Handler) UpdateMembersList(b *gotgbot.Bot, ctx *cmd.Context) error {
	count, err := h.service.SyncChatMembers(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		_ = ctx.Reply(b, "Не удалось обновить данные чата", nil)
		return err
	}

	return ctx.Reply(b, memberview.FormatSyncResult(count), nil)
}

func (h *Handler) ListRoles(b *gotgbot.Bot, ctx *cmd.Context) error {
	members, err := h.service.GetMembersWithTitle(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		_ = ctx.Reply(b, "Не удалось получить список ролей", nil)
		return err
	}

	if len(members) == 0 {
		return ctx.Reply(b, "В чате нет установленных ролей", nil)
	}

	return ctx.ReplyHTML(b, memberview.FormatRolesList(members))
}
func (h *Handler) SetRole(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()
	tag := ctx.FirstArgument()

	if targetUser == nil {
		return cmd.ErrNoUser
	}

	if tag == "" {
		return nil
	}

	if utf8.RuneCountInString(tag) > 16 {
		return ctx.Reply(b, "Слишком длинная роль (максимум 16 символа)", nil)
	}

	if err := h.service.SetMemberTitle(ctx.StdContext(), ctx.TargetChatID(), targetUser.ID, tag); err != nil {
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

	return ctx.ReplyHTML(b, memberview.FormatRoleUpdated(*targetUser, tag))
}

func (h *Handler) RestoreRoles(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetChatID := ctx.TargetChatID()
	members, err := h.service.GetAnyMembersWithTitle(ctx.StdContext(), targetChatID)
	if err != nil {
		_ = ctx.Reply(b, "Не удалось получить список ролей из базы", nil)
		return err
	}

	if len(members) == 0 {
		return ctx.Reply(b, "В базе данных нет сохраненных ролей для этого чата", nil)
	}

	var restoredCount int
	var errorsCount int
	limiter := rate.NewLimiter(rate.Every(1000*time.Millisecond), 1)

	for _, m := range members {
		if err := limiter.Wait(ctx.StdContext()); err != nil {
			return err
		}

		tgMember, err := b.GetChatMember(targetChatID, m.User.ID, nil)
		if err != nil {
			slog.Error("Failed to get member info", "error", err)
			_ = ctx.Reply(b, fmt.Sprintf("Не удалось получить пользователя %s", m.User.FirstName), nil)
			errorsCount++
			continue
		}

		if tgMember.GetStatus() == "creator" {
			continue
		}

		merged := tgMember.MergeChatMember()
		if tgMember.GetStatus() == "administrator" && merged.CustomTitle == m.CustomTitle {
			restoredCount++
			continue
		}

		var tgErr error
		if tgMember.GetStatus() != "administrator" {
			if ok, err := b.PromoteChatMember(targetChatID, m.User.ID, &gotgbot.PromoteChatMemberOpts{
				CanManageChat:   true,
				CanPostMessages: true,
				CanEditMessages: true,
			}); err != nil || !ok {
				tgErr = err
			} else {
				if _, err := b.SetChatAdministratorCustomTitle(targetChatID, m.User.ID, m.CustomTitle, nil); err != nil {
					tgErr = err
				}
			}
		} else if merged.CanBeEdited {
			if _, err := b.SetChatAdministratorCustomTitle(targetChatID, m.User.ID, m.CustomTitle, nil); err != nil {
				tgErr = err
			}
		} else {
			errorsCount++
			_ = ctx.Reply(b, fmt.Sprintf("Не могу изменить администратора %s (недостаточно прав)", m.User.FirstName), nil)
			continue
		}

		if tgErr != nil {
			errorsCount++
			_ = ctx.Reply(b, fmt.Sprintf("Ошибка при восстановлении пользователя %s: %s", m.User.FirstName, tgErr.Error()), nil)
			continue
		}

		restoredCount++
	}

	msgText := fmt.Sprintf("✅ Восстановление ролей завершено.\n\nВосстановлено: %d", restoredCount)
	if errorsCount > 0 {
		msgText += fmt.Sprintf("\nОшибок: %d", errorsCount)
	}

	return ctx.Reply(b, msgText, nil)
}

func (h *Handler) ShowRole(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()
	if targetUser == nil {
		return cmd.ErrNoUser
	}

	mTitle, err := h.service.GetMemberTitle(ctx.StdContext(), ctx.TargetChatID(), targetUser.ID)
	if err != nil && !errors.Is(err, member.ErrInvalidCustomTitle) {
		_ = ctx.Reply(b, "Не удалось получить роль пользователя", nil)
		return err
	}

	if mTitle == "" {
		return ctx.ReplyHTML(b, fmt.Sprintf("У пользователя %s нет роли", helpers.Link(*targetUser)))
	}

	return ctx.ReplyHTML(b, memberview.FormatMemberRole(*targetUser, mTitle))
}

func (h *Handler) OnJoinMember(b *gotgbot.Bot, ctx *cmd.Context) error {
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
		// handle call functionality inline
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

func (h *Handler) OnLeftMember(b *gotgbot.Bot, ctx *cmd.Context) error {
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
	title := m.CustomTitle
	if m.CustomTitle == "" {
		title = ctx.EffectiveSender.FirstName()
	}
	var sb strings.Builder
	for _, a := range admins {
		sb.WriteString(helpers.Mention(a.ID, "​"))
	}
	_, err = ctx.EffectiveChat.SendMessage(b, fmt.Sprintf("🕊 %s %s нас..."+sb.String(),
		helpers.LinkWithContent(m.User, title),
		helpers.Gendered(m.User.Gender, "покинул", "покинула", "покинул(а)"),
	), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})

	return err
}

func (h *Handler) OnBotPromote(_ *gotgbot.Bot, ctx *cmd.Context) error {
	log.Println("on bot promote")
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
