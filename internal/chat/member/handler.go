package member

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/helpers"
	"activity-bot/internal/user"
	"context"
	"fmt"
	"html"
	"log"
	"regexp"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Handler struct {
	service      *Service
	userService  *user.Service
	adminService *admin.Service
	setRoleRe    *regexp.Regexp
}

func NewHandler(service *Service, userService *user.Service, adminService *admin.Service, setRoleRe *regexp.Regexp) *Handler {
	return &Handler{service, userService, adminService, setRoleRe}
}

func (h *Handler) UpdateMembersList(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !helpers.CheckOwnerOrAdmin(ctx, b, h.adminService, update.Message.Chat.ID, update.Message.From.ID) {
		helpers.AnswerMessage(ctx, b, update, "Команда доступна только создателю или пользователю с ролью администратор в боте")
		return
	}

	count, err := helpers.UpdateChatMembers(ctx, b, h.service, update.Message.Chat.ID)
	if err != nil {
		log.Println("Update chat members error", err)
		helpers.AnswerMessage(ctx, b, update, "Не удалось обновить данные чата")
		return
	}

	helpers.AnswerMessage(ctx, b, update, fmt.Sprintf("Чат обновлён. Найдено %d участников", count))
}

func (h *Handler) ListRoles(ctx context.Context, b *bot.Bot, update *models.Update) {
	members, err := h.service.GetMembersWithTitle(ctx, update.Message.Chat.ID)
	if err != nil {
		helpers.AnswerMessage(ctx, b, update, "Не удалось получить список ролей")
		return
	}

	if len(members) == 0 {
		helpers.AnswerMessage(ctx, b, update, "В чате нет установленных ролей")
		return
	}

	var sb strings.Builder
	sb.WriteString("🎭 Роли участников:\n")
	for _, m := range members {
		if m.Username != nil {
			sb.WriteString(fmt.Sprintf("\n<a href=\"https://t.me/%s\">%s</a>: <b>%s</b>", *m.Username, m.FirstName, html.EscapeString(m.CustomTitle)))
		} else {
			sb.WriteString(fmt.Sprintf("\n<a href=\"tg://openmessage?user_id=%d\">%s</a>: <b>%s</b>", m.UserID, m.FirstName, html.EscapeString(m.CustomTitle)))
		}
	}

	helpers.AnswerMessage(ctx, b, update, sb.String())
}

func (h *Handler) SetRole(ctx context.Context, b *bot.Bot, update *models.Update) {
	args := h.setRoleRe.ReplaceAllString(update.Message.Text, "")
	args = strings.TrimSpace(args)

	targetUserID, role, found, err := helpers.ExtractTargetUser(ctx, h.userService, update, args)
	if err != nil {
		helpers.AnswerMessage(ctx, b, update, "Не удалось найти пользователя")
		return
	}
	if !found {
		if role == "" {
			targetUserID = update.Message.From.ID
		} else {
			helpers.AnswerMessage(ctx, b, update, "Вы не указали кому хотите выдать роль. Укажите @mention или ответьте на сообщение.")
			return
		}
	}

	role = strings.TrimSpace(role)

	if role == "" {
		mTitle, err := h.service.GetMemberTitle(ctx, update.Message.Chat.ID, targetUserID)
		if err != nil {
			helpers.AnswerMessage(ctx, b, update, "Не удалось получить роль пользователя")
			return
		}

		u, err := h.userService.GetUser(ctx, targetUserID)
		name := "Пользователь"
		if err == nil {
			name = html.EscapeString(u.FirstName)
		}

		if mTitle == "" {
			helpers.AnswerMessage(ctx, b, update, fmt.Sprintf("У пользователя <a href=\"tg://user?id=%d\">%s</a> нет роли", targetUserID, name))
		} else {
			helpers.AnswerMessage(ctx, b, update, fmt.Sprintf("Роль пользователя <a href=\"tg://user?id=%d\">%s</a>: <b>%s</b>", targetUserID, name, html.EscapeString(mTitle)))
		}
		return
	}

	if !helpers.CheckOwnerOrAdmin(ctx, b, h.adminService, update.Message.Chat.ID, update.Message.From.ID) {
		helpers.AnswerMessage(ctx, b, update, "Команда изменения ролей доступна только создателю чата и администраторам бота")
		return
	}

	if len(role) > 32 {
		helpers.AnswerMessage(ctx, b, update, "Слишком длинная роль (максимум 32 символа)")
		return
	}

	member, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
		ChatID: update.Message.Chat.ID,
		UserID: targetUserID,
	})
	if err != nil {
		helpers.AnswerMessage(ctx, b, update, "Не удалось получить информацию о пользователе")
		return
	}

	if member.Owner != nil {
		helpers.AnswerMessage(ctx, b, update, "Нельзя изменить роль создателя чата")
		return
	}

	if member.Administrator != nil {
		if !member.Administrator.CanBeEdited {
			helpers.AnswerMessage(ctx, b, update, "Я не могу изменить этого администратора (он назначен другим админом)")
			return
		}
		if _, err := b.SetChatAdministratorCustomTitle(ctx, &bot.SetChatAdministratorCustomTitleParams{
			ChatID:      update.Message.Chat.ID,
			UserID:      targetUserID,
			CustomTitle: role,
		}); err != nil {
			log.Println("Telegram set custom title error", err)
			helpers.AnswerMessage(ctx, b, update, "Не удалось изменить роль в Telegram")
			return
		}
	} else if member.Member != nil || member.Restricted != nil {
		if ok, err := b.PromoteChatMember(ctx, &bot.PromoteChatMemberParams{
			ChatID:          update.Message.Chat.ID,
			UserID:          targetUserID,
			CanPinMessages:  true,
			CanPostMessages: true,
			CanEditMessages: true,
		}); err != nil || !ok {
			log.Println("Telegram promote error", err)
			helpers.AnswerMessage(ctx, b, update, "Не удалось назначить пользователя администратором. Проверьте права бота.")
			return
		}

		if _, err := b.SetChatAdministratorCustomTitle(ctx, &bot.SetChatAdministratorCustomTitleParams{
			ChatID:      update.Message.Chat.ID,
			UserID:      targetUserID,
			CustomTitle: role,
		}); err != nil {
			log.Println("Telegram set custom title after promote error", err)
			helpers.AnswerMessage(ctx, b, update, "Пользователь назначен администратором, но не удалось установить роль")
			return
		}

	} else {
		helpers.AnswerMessage(ctx, b, update, "Пользователь не является участником чата")
		return
	}

	if err := h.service.SetMemberTitle(ctx, update.Message.Chat.ID, targetUserID, role); err != nil {
		log.Println("DB set custom title error", err)
		helpers.AnswerMessage(ctx, b, update, "Роль в Telegram изменена, но не удалось сохранить в базе данных")
		return
	}

	helpers.AnswerMessage(ctx, b, update, fmt.Sprintf("Роль пользователя обновлена на: <b>%s</b>", html.EscapeString(role)))
}

func (h *Handler) OnLeftMember(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.LeftChatMember == nil {
		return
	}

	leftMember := update.Message.LeftChatMember
	title, err := h.service.ProcessLeftMember(ctx, update.Message.Chat.ID, leftMember.ID)
	if err != nil {
		log.Println("Process left member error", err)
		return
	}

	if title != "" {
		helpers.AnswerMessage(ctx, b, update, fmt.Sprintf("🕊 <a href=\"tg://user?id=%d\">%s</a> (<b>%s</b>) покинул нас...", leftMember.ID, html.EscapeString(leftMember.FirstName), html.EscapeString(title)))
	}
}
