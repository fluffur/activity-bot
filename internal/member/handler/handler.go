package handler

import (
	service "activity-bot/internal/call"
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/user"
	"errors"
	"fmt"
	"html"
	"log"
	"log/slog"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	service     *member.Service
	chatService *chat.Service
	userService *user.Service
	callService *service.Service
}

func New(service *member.Service, chatService *chat.Service, userService *user.Service, callService *service.Service) *Handler {
	return &Handler{service, chatService, userService, callService}
}

func (h *Handler) UpdateMembersList(b *gotgbot.Bot, ctx *ext.Context, _ *cmd.Context) error {
	count, err := h.service.SyncChatMembers(ctx.EffectiveChat.Id)
	if err != nil {
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось обновить данные чата", nil)
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Чат обновлён. Найдено %d участников", count), nil)
	return err
}

func (h *Handler) ListRoles(b *gotgbot.Bot, ctx *ext.Context, _ *cmd.Context) error {
	if _, err := h.service.SyncChatMembers(ctx.EffectiveChat.Id); err != nil {
		slog.Warn("failed to update chat members", "chat_id", ctx.EffectiveChat.Id, "error", err)
	}
	members, err := h.service.GetMembersWithTitle(ctx.EffectiveChat.Id)
	if err != nil {
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось получить список ролей", nil)
		return err
	}

	if len(members) == 0 {
		_, err = ctx.EffectiveMessage.Reply(b, "В чате нет установленных ролей", nil)
		return err
	}

	var sb strings.Builder
	sb.WriteString("🎭 Роли всех участников:\n")
	for i, m := range members {
		sb.WriteString(fmt.Sprintf("\n%d. %s — %s", i+1, helpers.Link(m.User), html.EscapeString(m.CustomTitle)))
	}

	_, err = ctx.EffectiveMessage.Reply(b, sb.String(), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})

	return err
}

func (h *Handler) SetRole(b *gotgbot.Bot, ctx *ext.Context, cctx *cmd.Context) error {
	targetUser := cctx.FirstUser()
	role := cctx.FirstArgument()

	if targetUser == nil {
		return cmd.ErrNoUser
	}

	if role == "" {
		return nil
	}

	if len(role) > 32 {
		_, err := ctx.EffectiveMessage.Reply(b, "Слишком длинная роль (максимум 32 символа)", nil)

		return err
	}

	m, err := b.GetChatMember(ctx.EffectiveChat.Id, targetUser.ID, nil)
	if err != nil {
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось получить информацию о пользователе", nil)

		return err
	}

	if m.GetStatus() == "creator" {
		_, err := ctx.EffectiveMessage.Reply(b, "Я не могу изменить роль создателя чата", nil)
		return err
	}
	mergedMember := m.MergeChatMember()

	if m.GetStatus() == "administrator" {
		if !mergedMember.CanBeEdited {
			_, err := ctx.EffectiveMessage.Reply(b, "Я не могу изменить этого администратора (он назначен другим админом)", nil)
			return err
		}
		if _, err := b.SetChatAdministratorCustomTitle(ctx.EffectiveChat.Id, targetUser.ID, role, nil); err != nil {
			_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось изменить роль в Telegram", nil)

			return err
		}
	} else if m.GetStatus() == "member" {
		if ok, err := b.PromoteChatMember(ctx.EffectiveChat.Id, targetUser.ID, &gotgbot.PromoteChatMemberOpts{
			CanPinMessages:  true,
			CanPostMessages: true,
			CanEditMessages: true,
		}); err != nil || !ok {
			_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось назначить пользователя администратором. Проверьте права бота.", nil)
			return err
		}

		if _, err := b.SetChatAdministratorCustomTitle(ctx.EffectiveChat.Id, targetUser.ID, role, nil); err != nil {
			_, _ = ctx.EffectiveMessage.Reply(b, "Пользовтель назначен администратором, но не удалось изменить роль", nil)

			return err
		}

	} else {
		_, err := ctx.EffectiveMessage.Reply(b, "Пользователь не является полноправным участником чата", nil)

		return err
	}

	if err := h.service.SetMemberTitle(ctx.EffectiveChat.Id, targetUser.ID, &role); err != nil {
		_, _ = ctx.EffectiveMessage.Reply(b, "Роль в Telegram изменена, но не удалось сохранить у бота, можно попробовать !обновить чат", nil)

		return err
	}

	if err := h.service.SetMemberRole(ctx.EffectiveChat.Id, targetUser.ID, "administrator"); err != nil {
		slog.Warn("failed to set role in DB", "chat_id", ctx.EffectiveChat.Id, "user_id", targetUser.ID, "error", err)
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Роль пользователя обновлена на \"%s\"", html.EscapeString(role)), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})

	return err
}

func (h *Handler) DeleteRole(b *gotgbot.Bot, ctx *ext.Context, cctx *cmd.Context) error {
	targetUser := cctx.FirstUser()

	if targetUser == nil {
		return cmd.ErrNoUser
	}

	if _, err := b.PromoteChatMember(ctx.EffectiveChat.Id, targetUser.ID, nil); err != nil {
		slog.Warn("Cannot demote chat member", "error", err)
	}

	if err := h.service.SetMemberTitle(ctx.EffectiveChat.Id, targetUser.ID, nil); err != nil {
		_, err := ctx.EffectiveMessage.Reply(b, "Администратор удалён, но роль в базе бота нет", nil)

		return err
	}
	_, err := ctx.EffectiveMessage.Reply(b, "Администратор удалён", nil)

	return err
}

func (h *Handler) ShowRole(b *gotgbot.Bot, ctx *ext.Context, cctx *cmd.Context) error {
	targetUser := cctx.FirstUser()

	if targetUser == nil {
		return cmd.ErrNoUser
	}

	mTitle, err := h.service.GetMemberTitle(ctx.EffectiveChat.Id, targetUser.ID)
	if err != nil && !errors.Is(err, member.ErrInvalidCustomTitle) {
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось получить роль пользователя", nil)
		return err
	}

	if mTitle == "" {
		_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("У пользователя %s нет роли", helpers.Link(*targetUser)), &gotgbot.SendMessageOpts{
			ParseMode: gotgbot.ParseModeHTML,
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
		})
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Роль пользователя %s — %s", helpers.Link(*targetUser), html.EscapeString(mTitle)), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})
	return err
}

func (h *Handler) OnJoinMember(b *gotgbot.Bot, ctx *ext.Context) error {
	joinedMembers := ctx.EffectiveMessage.NewChatMembers

	for _, u := range joinedMembers {
		if u.IsBot {
			continue
		}
		slog.Info("member joined", "chat_id", ctx.EffectiveChat.Id, "user_id", u.Id)
		if _, err := h.service.EnsureMemberExists(ctx.EffectiveChat.Id, u.Id, u.Username, u.FirstName, u.LastName, "member"); err != nil {
			return err
		}
	}

	chatData, err := h.chatService.GetChat(ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	if chatData.CallOnJoin {
		return h.callService.Call(b, ctx, chatData.WelcomeCallMessage)
	}

	return nil
}

func (h *Handler) OnLeftMember(b *gotgbot.Bot, ctx *ext.Context) error {
	u := ctx.Message.LeftChatMember
	slog.Info("member left", "chat_id", ctx.EffectiveChat.Id, "user_id", u.Id)
	if u.IsBot {
		return nil
	}
	if _, err := h.service.EnsureMemberExists(ctx.EffectiveChat.Id, u.Id, u.Username, u.FirstName, u.LastName, "member"); err != nil {
		return err
	}
	title, err := h.service.ProcessLeftMember(ctx.EffectiveChat.Id, u.Id)
	if err != nil {
		return err
	}

	if title != "" {
		_, err = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("🕊 %s c ролью \"%s\" покинул нас...", helpers.Link(helpers.MapUser(u)), html.EscapeString(title)), &gotgbot.SendMessageOpts{
			ParseMode: gotgbot.ParseModeHTML,
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
		})
	}

	return err
}

func (h *Handler) OnBotPromote(_ *gotgbot.Bot, ctx *ext.Context) error {
	count, err := h.service.SyncChatMembers(ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}
	slog.Info("updated chat members on bot join", "chat_id", ctx.EffectiveChat.Id, "count", count)
	return nil
}

func (h *Handler) SetOnlyNewbies(b *gotgbot.Bot, ctx *ext.Context, cctx *cmd.Context) error {
	if len(cctx.Users()) == 0 {
		_, err := ctx.EffectiveMessage.Reply(b, "Укажите хотя бы одного участника", nil)

		return err
	}
	if err := h.service.SetOnlyNewbies(ctx.EffectiveChat.Id, cctx.Users()); err != nil {
		log.Println("failed to set only-newbies", err)
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось установить олдов", nil)
		return err
	}

	_, err := ctx.EffectiveMessage.Reply(b, "Олды установлены", nil)

	return err
}

func (h *Handler) SetNewbies(b *gotgbot.Bot, ctx *ext.Context, cctx *cmd.Context) error {
	if len(cctx.Users()) == 0 {
		_, err := ctx.EffectiveMessage.Reply(b, "Укажите хотя бы одного участника", nil)

		return err
	}
	if err := h.service.SetNewbies(ctx.EffectiveChat.Id, cctx.Users()); err != nil {
		return err
	}

	_, err := ctx.EffectiveMessage.Reply(b, "Новички установлены", nil)

	return err
}
