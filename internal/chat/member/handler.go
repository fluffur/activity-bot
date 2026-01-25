package member

//
//import (
//	"activity-bot/internal/admin"
//	"activity-bot/internal/helpers"
//	"activity-bot/internal/user"
//	"context"
//	"errors"
//	"fmt"
//	"html"
//	"log"
//	"regexp"
//	"strings"
//
//	"github.com/PaulSonOfLars/gotgbot/v2"
//	"github.com/PaulSonOfLars/gotgbot/v2/ext"
//	"github.com/go-telegram/bot"
//	"github.com/go-telegram/bot/models"
//)
//
//type Handler struct {
//	service      *Service
//	userService  *user.Service
//	adminService *admin.Service
//	setRoleRe    *regexp.Regexp
//}
//
//func NewHandler(service *Service, userService *user.Service, adminService *admin.Service, setRoleRe *regexp.Regexp) *Handler {
//	return &Handler{service, userService, adminService, setRoleRe}
//}
//
//func (h *Handler) UpdateMembersList(b *gotgbot.Bot, ctx *ext.Context, args []string) error {
//	count, err := helpers.UpdateChatMembers(b, h.service, ctx.EffectiveChat.Id)
//	if err != nil {
//		log.Println("Update chat members error", err)
//		_, err = ctx.EffectiveMessage.Reply(b, "Не удалось обновить данные чата", nil)
//		return err
//	}
//
//	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Чат обновлён. Найдено %d участников", count), nil)
//	return err
//}
//
//func (h *Handler) ListRoles(ctx context.Context, b *bot.Bot, update *models.Update) {
//	members, err := h.service.GetMembersWithTitle(ctx, update.Message.Chat.ID)
//	if err != nil {
//		helpers.SendMessage(ctx, b, update, "Не удалось получить список ролей")
//		return
//	}
//
//	if len(members) == 0 {
//		helpers.SendMessage(ctx, b, update, "В чате нет установленных ролей")
//		return
//	}
//
//	var sb strings.Builder
//	sb.WriteString("🎭 Роли всех участников:\n")
//	for i, m := range members {
//		sb.WriteString(fmt.Sprintf("\n%d. %s — %s", i+1, helpers.Link(m.User), html.EscapeString(m.CustomTitle)))
//	}
//
//	helpers.SendMessage(ctx, b, update, sb.String())
//}
//
//func (h *Handler) SetRole(ctx context.Context, b *bot.Bot, update *models.Update) {
//	args := h.setRoleRe.ReplaceAllString(update.Message.Text, "")
//	args = strings.TrimSpace(args)
//
//	targetUser, role, err := helpers.ExtractTargetUser(h.userService, update, args)
//	if err != nil {
//		if !errors.Is(err, helpers.ErrUserNotSpecified) {
//			helpers.SendMessage(ctx, b, update, "Не удалось найти пользователя")
//			return
//		}
//		if role == "" {
//			targetUser, err = h.userService.GetUser(update.Message.From.ID)
//			if err != nil {
//				helpers.SendMessage(ctx, b, update, "Пользователь не найден, введите !обновить чат")
//				return
//			}
//		} else {
//			helpers.SendMessage(ctx, b, update, "Вы не указали кому хотите выдать роль. Укажите @mention или ответьте на сообщение.")
//			return
//		}
//
//	}
//
//	role = strings.TrimSpace(role)
//
//	if role == "" {
//		mTitle, err := h.service.GetMemberTitle(ctx, update.Message.Chat.ID, targetUser.ID)
//		if err != nil {
//			helpers.SendMessage(ctx, b, update, "Не удалось получить роль пользователя")
//			return
//		}
//
//		if mTitle == "" {
//			helpers.SendMessage(ctx, b, update, fmt.Sprintf("У пользователя %s нет роли", helpers.Link(targetUser)))
//		} else {
//			helpers.SendMessage(ctx, b, update, fmt.Sprintf("Роль пользователя %s — %s", helpers.Link(targetUser), html.EscapeString(mTitle)))
//		}
//		return
//	}
//
//	if !helpers.CheckOwnerOrAdmin(ctx, b, h.adminService, update.Message.Chat.ID, update.Message.From.ID) {
//		helpers.SendMessage(ctx, b, update, "Команда изменения ролей доступна только создателю чата и администраторам бота")
//		return
//	}
//
//	if len(role) > 32 {
//		helpers.SendMessage(ctx, b, update, "Слишком длинная роль (максимум 32 символа)")
//		return
//	}
//
//	member, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
//		ChatID: update.Message.Chat.ID,
//		UserID: targetUser.ID,
//	})
//	if err != nil {
//		helpers.SendMessage(ctx, b, update, "Не удалось получить информацию о пользователе")
//		return
//	}
//
//	if member.Owner != nil {
//		helpers.SendMessage(ctx, b, update, "Нельзя изменить роль создателя чата")
//		return
//	}
//
//	if member.Administrator != nil {
//		if !member.Administrator.CanBeEdited {
//			helpers.SendMessage(ctx, b, update, "Я не могу изменить этого администратора (он назначен другим админом)")
//			return
//		}
//		if _, err := b.SetChatAdministratorCustomTitle(ctx, &bot.SetChatAdministratorCustomTitleParams{
//			ChatID:      update.Message.Chat.ID,
//			UserID:      targetUser.ID,
//			CustomTitle: role,
//		}); err != nil {
//			log.Println("Telegram set custom title error", err)
//			helpers.SendMessage(ctx, b, update, "Не удалось изменить роль в Telegram")
//			return
//		}
//	} else if member.Member != nil || member.Restricted != nil {
//		if ok, err := b.PromoteChatMember(ctx, &bot.PromoteChatMemberParams{
//			ChatID:          update.Message.Chat.ID,
//			UserID:          targetUser.ID,
//			CanPinMessages:  true,
//			CanPostMessages: true,
//			CanEditMessages: true,
//		}); err != nil || !ok {
//			log.Println("Telegram promote error", err)
//			helpers.SendMessage(ctx, b, update, "Не удалось назначить пользователя администратором. Проверьте права бота.")
//			return
//		}
//
//		if _, err := b.SetChatAdministratorCustomTitle(ctx, &bot.SetChatAdministratorCustomTitleParams{
//			ChatID:      update.Message.Chat.ID,
//			UserID:      targetUser.ID,
//			CustomTitle: role,
//		}); err != nil {
//			log.Println("Telegram set custom title after promote error", err)
//			helpers.SendMessage(ctx, b, update, "Пользователь назначен администратором, но не удалось установить роль")
//			return
//		}
//
//	} else {
//		helpers.SendMessage(ctx, b, update, "Пользователь не является участником чата")
//		return
//	}
//
//	if err := h.service.SetMemberTitle(ctx, update.Message.Chat.ID, targetUser.ID, role); err != nil {
//		log.Println("DB set custom title error", err)
//		helpers.SendMessage(ctx, b, update, "Роль в Telegram изменена, но не удалось сохранить в базе данных")
//		return
//	}
//
//	helpers.SendMessage(ctx, b, update, fmt.Sprintf("Роль пользователя обновлена на \"%s\"", html.EscapeString(role)))
//}
//
//func (h *Handler) OnLeftMember(ctx context.Context, b *bot.Bot, update *models.Update) {
//	if update.Message == nil || update.Message.LeftChatMember == nil {
//		return
//	}
//
//	leftMember := update.Message.LeftChatMember
//	title, err := h.service.ProcessLeftMember(ctx, update.Message.Chat.ID, leftMember.ID)
//	if err != nil {
//		log.Println("Process left member error", err)
//		return
//	}
//
//	if title != "" {
//		helpers.SendMessage(ctx, b, update, fmt.Sprintf("🕊 %s c ролью \"%s\" покинул нас...", helpers.Link(helpers.MapUser(leftMember)), html.EscapeString(title)))
//	}
//}
//
//func (h *Handler) OnBotPromote(b *gotgbot.Bot, ctx *ext.Context, args []string) {
//	count, err := helpers.UpdateChatMembers(b, h.service, ctx.EffectiveChat.Id)
//	if err != nil {
//		log.Println("Failed to update chat members on join:", err)
//		return
//	}
//	log.Printf("Updated chat %d members on bot join, total %d members\n", ctx.EffectiveChat.Id, count)
//}
