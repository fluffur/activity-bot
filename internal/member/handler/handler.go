package handler

import (
	"activity-bot/internal/admin"
	service "activity-bot/internal/call"
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/member/view"
	"activity-bot/internal/user"
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
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

	return ctx.Reply(b, view.FormatSyncResult(count), nil)
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

	return ctx.ReplyHTML(b, view.FormatRolesList(members))
}
func (h *Handler) SetRole(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()
	role := ctx.FirstArgument()

	if targetUser == nil {
		return cmd.ErrNoUser
	}

	if role == "" {
		return nil
	}

	if len(role) > 32 {
		return ctx.Reply(b, "Слишком длинная роль (максимум 32 символа)", nil)
	}

	m, err := b.GetChatMember(ctx.TargetChatID(), targetUser.ID, nil)
	if err != nil {
		return ctx.Reply(b, "Не удалось получить информацию о пользователе", nil)
	}

	if m.GetStatus() == "creator" {
		return ctx.Reply(b, "Я не могу изменить роль создателя чата", nil)
	}

	mergedMember := m.MergeChatMember()
	var tgErr error

	if m.GetStatus() == "administrator" {
		if !mergedMember.CanBeEdited {
			return ctx.Reply(b, "Я не могу изменить этого администратора", nil)
		}

		_, tgErr = b.SetChatAdministratorCustomTitle(
			ctx.TargetChatID(),
			targetUser.ID,
			role,
			nil,
		)

	} else if m.GetStatus() == "member" {
		if ok, err := b.PromoteChatMember(ctx.TargetChatID(), targetUser.ID, &gotgbot.PromoteChatMemberOpts{
			CanManageChat:   true,
			CanPostMessages: true,
			CanEditMessages: true,
		}); err != nil || !ok {
			tgErr = err
		} else {
			_, tgErr = b.SetChatAdministratorCustomTitle(
				ctx.TargetChatID(),
				targetUser.ID,
				role,
				nil,
			)
		}

	} else {
		return ctx.Reply(b, "Пользователь не является участником чата", nil)
	}

	serviceErr := h.service.SetMemberTitle(
		ctx.StdContext(),
		ctx.TargetChatID(),
		targetUser.ID,
		&role,
	)

	if tgErr != nil && serviceErr != nil {
		return ctx.Reply(b, "Ошибка в Telegram и при сохранении роли у бота", nil)
	}

	if tgErr != nil {
		return ctx.Reply(b, "Роль сохранена у бота, но не удалось установить в Telegram\n"+tgErr.Error(), nil)
	}

	if serviceErr != nil {
		return ctx.Reply(b, "Роль изменена в Telegram, но не удалось сохранить у бота", nil)
	}

	return ctx.ReplyHTML(b, view.FormatRoleUpdated(*targetUser, role))
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
	limiter := rate.NewLimiter(rate.Every(100*time.Millisecond), 1)

	for _, m := range members {
		if err := limiter.Wait(ctx.StdContext()); err != nil {
			return err
		}

		tgMember, err := b.GetChatMember(targetChatID, m.User.ID, nil)
		if err != nil {
			log.Printf("Ошибка при получении информации о пользователе %s: %v\n", m.User.FirstName, err)
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

func (h *Handler) DeleteRole(b *gotgbot.Bot, ctx *cmd.Context) error {
	targetUser := ctx.FirstUser()
	targetChatID := ctx.TargetChatID()

	if targetUser == nil {
		return cmd.ErrNoUser
	}

	if _, err := b.PromoteChatMember(targetChatID, targetUser.ID, nil); err != nil {
		slog.Warn("Cannot demote chat member", "error", err)
	}

	if err := h.service.SetMemberTitle(ctx.StdContext(), targetChatID, targetUser.ID, nil); err != nil {
		_ = ctx.Reply(b, "Администратор удалён, но роль в базе бота нет", nil)

		return err
	}
	return ctx.Reply(b, "Администратор удалён", nil)
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

	return ctx.ReplyHTML(b, view.FormatMemberRole(*targetUser, mTitle))
}

func (h *Handler) OnJoinMember(b *gotgbot.Bot, ctx *ext.Context) error {
	joinedMembers := ctx.EffectiveMessage.NewChatMembers
	cctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, u := range joinedMembers {
		if u.IsBot {
			continue
		}
		slog.Info("member joined", "chat_id", ctx.EffectiveChat.Id, "user_id", u.Id)
		if _, err := h.service.EnsureMemberExists(cctx, ctx.EffectiveChat.Id, u.Id, u.Username, u.FirstName, u.LastName, "member"); err != nil {
			return err
		}
	}

	chatData, err := h.chatService.GetChat(cctx, ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	if chatData.CallOnJoin {
		return h.callService.Call(cctx, b, ctx, chatData.WelcomeCallMessage)
	}

	return nil
}

func (h *Handler) OnLeftMember(b *gotgbot.Bot, ctx *ext.Context) error {
	u := ctx.Message.LeftChatMember
	cctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	slog.Info("member left", "chat_id", ctx.EffectiveChat.Id, "user_id", u.Id)
	if u.IsBot {
		return nil
	}
	if _, err := h.service.EnsureMemberExists(cctx, ctx.EffectiveChat.Id, u.Id, u.Username, u.FirstName, u.LastName, "member"); err != nil {
		return err
	}
	m, err := h.service.ProcessLeftMember(cctx, ctx.EffectiveChat.Id, u.Id)
	if err != nil {
		return err
	}

	admins, err := h.adminService.GetAdminsEnsured(context.Background(), ctx.EffectiveChat.Id, h.service.SyncChatMembers)
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

func (h *Handler) OnBotPromote(_ *gotgbot.Bot, ctx *ext.Context) error {
	cctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	count, err := h.service.SyncChatMembers(cctx, ctx.EffectiveChat.Id)
	if err != nil {
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
		log.Println("failed to set only-newbies", err)
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
