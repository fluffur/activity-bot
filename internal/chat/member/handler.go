package member

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/user"
	"fmt"
	"html"
	"log"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	service      *Service
	userService  *user.Service
	adminService *admin.Service
}

func NewHandler(service *Service, userService *user.Service, adminService *admin.Service) *Handler {
	return &Handler{service, userService, adminService}
}

func (h *Handler) UpdateMembersList(b *gotgbot.Bot, ctx *ext.Context, _ *command.Context) error {
	count, err := UpdateChatMembers(b, h.service, ctx.EffectiveChat.Id)
	if err != nil {
		log.Println("Update chat members error", err)
		_, err = ctx.EffectiveMessage.Reply(b, "Не удалось обновить данные чата", nil)
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Чат обновлён. Найдено %d участников", count), nil)
	return err
}

func (h *Handler) ListRoles(b *gotgbot.Bot, ctx *ext.Context, _ *command.Context) error {
	members, err := h.service.GetMembersWithTitle(ctx.EffectiveChat.Id)
	if err != nil {
		log.Println("Exists members error", err)
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
	role := cctx.Args[0]
	targetUser := cctx.Users[0]

	if len(role) > 32 {
		_, err := ctx.EffectiveMessage.Reply(b, "Слишком длинная роль (максимум 32 символа)", nil)

		return err
	}

	member, err := b.GetChatMember(ctx.EffectiveChat.Id, targetUser.ID, nil)
	if err != nil {
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось получить информацию о пользователе", nil)

		return err
	}

	if member.GetStatus() == "creator" {
		_, err := ctx.EffectiveMessage.Reply(b, "Нельзя изменить роль создателя чата", nil)
		return err
	}
	mergedMember := member.MergeChatMember()

	if member.GetStatus() == "administrator" {
		if !mergedMember.CanBeEdited {
			_, err := ctx.EffectiveMessage.Reply(b, "Я не могу изменить этого администратора (он назначен другим админом)", nil)
			return err
		}
		if _, err := b.SetChatAdministratorCustomTitle(ctx.EffectiveChat.Id, targetUser.ID, role, nil); err != nil {
			log.Println("Telegram set custom title error", err)
			_, err := ctx.EffectiveMessage.Reply(b, "Не удалось изменить роль в Telegram", nil)

			return err
		}
	} else if member.GetStatus() == "member" {
		if ok, err := b.PromoteChatMember(ctx.EffectiveChat.Id, targetUser.ID, &gotgbot.PromoteChatMemberOpts{
			CanPinMessages:  true,
			CanPostMessages: true,
			CanEditMessages: true,
		}); err != nil || !ok {
			log.Println("Telegram promote error", err)
			_, err := ctx.EffectiveMessage.Reply(b, "Не удалось назначить пользователя администратором. Проверьте права бота.", nil)
			return err
		}

		if _, err := b.SetChatAdministratorCustomTitle(ctx.EffectiveChat.Id, targetUser.ID, role, nil); err != nil {
			log.Println("Telegram set custom title error", err)
			_, err := ctx.EffectiveMessage.Reply(b, "Пользовтель назначен администратором, но не удалось изменить роль", nil)

			return err
		}

	} else {
		_, err := ctx.EffectiveMessage.Reply(b, "Пользователь не является полноправным участником чата", nil)

		return err
	}

	if err := h.service.SetMemberTitle(ctx.EffectiveChat.Id, targetUser.ID, role); err != nil {
		log.Println("DB set custom title error", err)
		_, err := ctx.EffectiveMessage.Reply(b, "Роль в Telegram изменена, но не удалось сохранить у бота, можно попробовать !обновить чат", nil)

		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Роль пользователя обновлена на \"%s\"", html.EscapeString(role)), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})

	return err
}

func (h *Handler) ShowRole(b *gotgbot.Bot, ctx *ext.Context, cctx *command.Context) error {
	targetUser := cctx.Users[0]
	mTitle, err := h.service.GetMemberTitle(ctx.EffectiveChat.Id, targetUser.ID)
	if err != nil {
		log.Println("DB get custom title error", err)
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
		log.Println("Joined member", u.Id)
		if _, err := h.service.EnsureMemberExists(ctx.EffectiveChat.Id, u.Id, u.Username, u.FirstName, u.LastName); err != nil {
			log.Println("Process left member error", err)
			return err
		}
	}

	return nil
}

func (h *Handler) OnLeftMember(b *gotgbot.Bot, ctx *ext.Context) error {
	leftMember := ctx.EffectiveSender
	u := leftMember.User
	if _, err := h.service.EnsureMemberExists(ctx.EffectiveChat.Id, u.Id, u.Username, u.FirstName, u.LastName); err != nil {
		log.Println("Process left member error", err)
		return err
	}
	title, err := h.service.ProcessLeftMember(ctx.EffectiveChat.Id, leftMember.Id())
	if err != nil {
		log.Println("Process left member error", err)
		return err
	}

	if title != "" {
		_, err = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("🕊 %s c ролью \"%s\" покинул нас...", helpers.Link(helpers.MapUser(leftMember.User)), html.EscapeString(title)), &gotgbot.SendMessageOpts{
			ParseMode: gotgbot.ParseModeHTML,
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
		})
	}

	return err
}

func (h *Handler) OnBotPromote(b *gotgbot.Bot, ctx *ext.Context) error {
	count, err := UpdateChatMembers(b, h.service, ctx.EffectiveChat.Id)
	if err != nil {
		log.Println("Failed to update chat members on join:", err)
		return err
	}
	log.Printf("Updated chat %d members on bot join, total %d members\n", ctx.EffectiveChat.Id, count)
	return nil
}
