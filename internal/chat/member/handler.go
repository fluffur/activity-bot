package member

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/helpers"
	"activity-bot/internal/user"
	"context"
	"errors"
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
	sb.WriteString("🎭 Роли всех участников:\n")
	for i, m := range members {
		sb.WriteString(fmt.Sprintf("\n%d. %s — %s", i+1, helpers.FormatSilentMentionHTML(m.User), html.EscapeString(m.CustomTitle)))
	}

	helpers.AnswerMessage(ctx, b, update, sb.String())
}

func (h *Handler) SetRole(ctx context.Context, b *bot.Bot, update *models.Update) {
	args := h.setRoleRe.ReplaceAllString(update.Message.Text, "")
	args = strings.TrimSpace(args)

	targetUser, role, err := helpers.ExtractTargetUser(ctx, h.userService, update, args)
	if err != nil {
		if !errors.Is(err, helpers.ErrUserNotSpecified) {
			helpers.AnswerMessage(ctx, b, update, "Не удалось найти пользователя")
			return
		}
		if role == "" {
			targetUser, err = h.userService.GetUser(ctx, update.Message.From.ID)
			if err != nil {
				helpers.AnswerMessage(ctx, b, update, "Пользователь не найден, введите !обновить чат")
				return
			}
		} else {
			helpers.AnswerMessage(ctx, b, update, "Вы не указали кому хотите выдать роль. Укажите @mention или ответьте на сообщение.")
			return
		}

	}

	role = strings.TrimSpace(role)

	if role == "" {
		mTitle, err := h.service.GetMemberTitle(ctx, update.Message.Chat.ID, targetUser.ID)
		if err != nil {
			helpers.AnswerMessage(ctx, b, update, "Не удалось получить роль пользователя")
			return
		}

		if mTitle == "" {
			helpers.AnswerMessage(ctx, b, update, fmt.Sprintf("У пользователя %s нет роли", helpers.FormatSilentMentionHTML(targetUser)))
		} else {
			helpers.AnswerMessage(ctx, b, update, fmt.Sprintf("Роль пользователя %s — %s", helpers.FormatSilentMentionHTML(targetUser), html.EscapeString(mTitle)))
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
		UserID: targetUser.ID,
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
			UserID:      targetUser.ID,
			CustomTitle: role,
		}); err != nil {
			log.Println("Telegram set custom title error", err)
			helpers.AnswerMessage(ctx, b, update, "Не удалось изменить роль в Telegram")
			return
		}
	} else if member.Member != nil || member.Restricted != nil {
		if ok, err := b.PromoteChatMember(ctx, &bot.PromoteChatMemberParams{
			ChatID:          update.Message.Chat.ID,
			UserID:          targetUser.ID,
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
			UserID:      targetUser.ID,
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

	if err := h.service.SetMemberTitle(ctx, update.Message.Chat.ID, targetUser.ID, role); err != nil {
		log.Println("DB set custom title error", err)
		helpers.AnswerMessage(ctx, b, update, "Роль в Telegram изменена, но не удалось сохранить в базе данных")
		return
	}

	helpers.AnswerMessage(ctx, b, update, fmt.Sprintf("Роль пользователя обновлена на \"%s\"", html.EscapeString(role)))
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
		helpers.AnswerMessage(ctx, b, update, fmt.Sprintf("🕊 %s в роли \"%s\" покинул нас...", helpers.FormatSilentMentionHTML(helpers.MapFromUserToModel(leftMember)), html.EscapeString(title)))
	}
}
