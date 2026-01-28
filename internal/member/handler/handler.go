package handler

import (
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/user"
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
	userService *user.Service
}

func New(service *member.Service, userService *user.Service) *Handler {
	return &Handler{service, userService}
}

func (h *Handler) UpdateMembersList(b *gotgbot.Bot, ctx *ext.Context, _ *command.Context) error {
	count, err := h.service.SyncChatMembers(ctx.EffectiveChat.Id)
	if err != nil {
		slog.Error("failed to update chat members", "chat_id", ctx.EffectiveChat.Id, "error", err)
		_, err = ctx.EffectiveMessage.Reply(b, "Не удалось обновить данные чата", nil)
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Чат обновлён. Найдено %d участников", count), nil)
	return err
}

func (h *Handler) ListRoles(b *gotgbot.Bot, ctx *ext.Context, _ *command.Context) error {
	members, err := h.service.GetMembersWithTitle(ctx.EffectiveChat.Id)
	if err != nil {
		slog.Error("failed to get members with titles", "chat_id", ctx.EffectiveChat.Id, "error", err)
		_, err = ctx.EffectiveMessage.Reply(b, "Не удалось получить список ролей", nil)
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

func (h *Handler) SetRole(b *gotgbot.Bot, ctx *ext.Context, cctx *command.Context) error {
	if len(cctx.Users) == 0 {
		_, err := ctx.EffectiveMessage.Reply(b, "Пользователь не найден в базе данных бота. Попробуйте упомянуть его через ответ на сообщение или дождитесь, пока он напишет что-нибудь.", nil)
		return err
	}

	role := cctx.Args[0]
	targetUser := cctx.Users[0]
	log.Println(role, targetUser)
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
			slog.Error("failed to set custom title in Telegram", "chat_id", ctx.EffectiveChat.Id, "user_id", targetUser.ID, "error", err)
			_, err := ctx.EffectiveMessage.Reply(b, "Не удалось изменить роль в Telegram", nil)

			return err
		}
	} else if m.GetStatus() == "member" {
		if ok, err := b.PromoteChatMember(ctx.EffectiveChat.Id, targetUser.ID, &gotgbot.PromoteChatMemberOpts{
			CanPinMessages:  true,
			CanPostMessages: true,
			CanEditMessages: true,
		}); err != nil || !ok {
			slog.Error("failed to promote member in Telegram", "chat_id", ctx.EffectiveChat.Id, "user_id", targetUser.ID, "error", err)
			_, err := ctx.EffectiveMessage.Reply(b, "Не удалось назначить пользователя администратором. Проверьте права бота.", nil)
			return err
		}

		if _, err := b.SetChatAdministratorCustomTitle(ctx.EffectiveChat.Id, targetUser.ID, role, nil); err != nil {
			slog.Error("failed to set custom title for new administrator in Telegram", "chat_id", ctx.EffectiveChat.Id, "user_id", targetUser.ID, "error", err)
			_, err := ctx.EffectiveMessage.Reply(b, "Пользовтель назначен администратором, но не удалось изменить роль", nil)

			return err
		}

	} else {
		_, err := ctx.EffectiveMessage.Reply(b, "Пользователь не является полноправным участником чата", nil)

		return err
	}

	if err := h.service.SetMemberTitle(ctx.EffectiveChat.Id, targetUser.ID, &role); err != nil {
		slog.Error("failed to set custom title in DB", "chat_id", ctx.EffectiveChat.Id, "user_id", targetUser.ID, "error", err)
		_, err := ctx.EffectiveMessage.Reply(b, "Роль в Telegram изменена, но не удалось сохранить у бота, можно попробовать !обновить чат", nil)

		return err
	}

	if err := h.service.SetMemberRole(ctx.EffectiveChat.Id, targetUser.ID, "administrator"); err != nil {
		slog.Error("failed to set role in DB", "chat_id", ctx.EffectiveChat.Id, "user_id", targetUser.ID, "error", err)
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Роль пользователя обновлена на \"%s\"", html.EscapeString(role)), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})

	return err
}

func (h *Handler) DeleteRole(b *gotgbot.Bot, ctx *ext.Context, cctx *command.Context) error {
	if len(cctx.Users) == 0 {
		slog.Error("No user in DeleteRole")
		return nil
	}

	targetUser := cctx.Users[0]

	if _, err := b.PromoteChatMember(ctx.EffectiveChat.Id, targetUser.ID, nil); err != nil {
		slog.Error("Cannot demote chat member", "error", err)
	}

	if err := h.service.SetMemberTitle(ctx.EffectiveChat.Id, targetUser.ID, nil); err != nil {
		_, err := ctx.EffectiveMessage.Reply(b, "Администратор удалён, но роль в базе бота нет", nil)

		return err
	}
	_, err := ctx.EffectiveMessage.Reply(b, "Администратор удалён", nil)

	return err
}

func (h *Handler) ShowRole(b *gotgbot.Bot, ctx *ext.Context, cctx *command.Context) error {
	if len(cctx.Users) == 0 {
		slog.Error("No user in ShowRole")
		return nil
	}

	targetUser := cctx.Users[0]
	mTitle, err := h.service.GetMemberTitle(ctx.EffectiveChat.Id, targetUser.ID)
	if err != nil {
		slog.Error("failed to get custom title from DB", "chat_id", ctx.EffectiveChat.Id, "user_id", targetUser.ID, "error", err)
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось получить роль пользователя", nil)
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

func (h *Handler) OnJoinMember(_ *gotgbot.Bot, ctx *ext.Context) error {
	joinedMembers := ctx.EffectiveMessage.NewChatMembers
	for _, u := range joinedMembers {
		if u.IsBot {
			continue
		}
		slog.Info("member joined", "chat_id", ctx.EffectiveChat.Id, "user_id", u.Id)
		if _, err := h.service.EnsureMemberExists(ctx.EffectiveChat.Id, u.Id, u.Username, u.FirstName, u.LastName, "member"); err != nil {
			slog.Error("failed to ensure joined member exists", "chat_id", ctx.EffectiveChat.Id, "user_id", u.Id, "error", err)
			return err
		}
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
		slog.Error("failed to ensure left member exists", "chat_id", ctx.EffectiveChat.Id, "user_id", u.Id, "error", err)
		return err
	}
	title, err := h.service.ProcessLeftMember(ctx.EffectiveChat.Id, u.Id)
	if err != nil {
		slog.Error("failed to process left member", "chat_id", ctx.EffectiveChat.Id, "user_id", u.Id, "error", err)
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
		slog.Error("failed to update chat members on bot promote", "chat_id", ctx.EffectiveChat.Id, "error", err)
		return err
	}
	slog.Info("updated chat members on bot join", "chat_id", ctx.EffectiveChat.Id, "count", count)
	return nil
}
