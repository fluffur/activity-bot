package chat

import (
	"activity-bot/internal/base"
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"context"
	"fmt"
	"html"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Handler struct {
	base.Handler
	service     *Service
	userService *user.Service
	dateParser  *DateParser
	setNormRe   *regexp.Regexp
	setExemptRe *regexp.Regexp
	setRoleRe   *regexp.Regexp
}

func NewHandler(service *Service, userService *user.Service, dateParser *DateParser, setNormRe, setExemptRe, setRoleRe *regexp.Regexp) *Handler {
	return &Handler{base.Handler{}, service, userService, dateParser, setNormRe, setExemptRe, setRoleRe}
}

func (h *Handler) ShowNorm(ctx context.Context, b *bot.Bot, update *models.Update) {
	norm, err := h.service.GetNorm(ctx, update.Message.Chat.ID)
	if err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось отправить норму чата")
		log.Println("Failed to show chat norm", err)
		return
	}

	h.AnswerMessage(ctx, b, update, fmt.Sprintf("Норма чата: %d сообщений", norm))
}

func (h *Handler) SetNorm(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !h.checkOwnerOrAdmin(ctx, b, update, update.Message.Chat.ID, update.Message.From.ID) {
		h.AnswerMessage(ctx, b, update, "Команда установки нормы доступна только создателю чата и администраторам бота")
		return
	}

	text := update.Message.Text
	matches := h.setNormRe.FindStringSubmatch(text)
	if len(matches) < 3 {
		h.AnswerMessage(ctx, b, update, "Неверный формат команды")
		return
	}
	norm, err := strconv.Atoi(matches[2])
	if err != nil {
		h.AnswerMessage(ctx, b, update, "Норма должна быть числом")
		return
	}

	if err := h.service.SetNorm(ctx, update.Message.Chat.ID, norm); err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось установить норму чата")
		log.Println("Failed to set chat norm", err)
		return
	}

	h.AnswerMessage(ctx, b, update, "Новая норма чата установлена")
}

func (h *Handler) ShowWeeklyReport(ctx context.Context, b *bot.Bot, update *models.Update) {
	if _, err := h.updateChatMembers(ctx, b, update.Message.Chat.ID); err != nil {
		log.Println("Auto-update chat members error", err)
	}

	report, err := h.service.GetMemberStats(ctx, update.Message.Chat.ID)
	if err != nil {
		log.Println("Get member stats error", err)
		h.AnswerMessage(ctx, b, update, "Не удалось получить отчёт")
		return
	}

	exemptMembers, err := h.service.GetExemptMembers(ctx, update.Message.Chat.ID)
	if err != nil {
		log.Println("Get exempt members error", err)
		h.AnswerMessage(ctx, b, update, "Не удалось получить отчёт")
		return
	}

	if len(report) == 0 && len(exemptMembers) == 0 {
		h.AnswerMessage(ctx, b, update, "Нет данных для отчёта на эту неделю")
		return
	}

	text := formatWeeklyReport(report, exemptMembers)
	h.AnswerMessage(ctx, b, update, text)
}

func formatWeeklyReport(report []model.WeeklyMessageReportMember, exemptMembers []model.ExemptMember) string {
	now := time.Now()
	weekday := int(now.Weekday())
	daysSinceMonday := (weekday + 6) % 7
	monday := now.AddDate(0, 0, -daysSinceMonday)
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
	sunday := monday.AddDate(0, 0, 6)
	sunday = time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 0, sunday.Location())

	weekHeader := fmt.Sprintf("📊 Отчёт за неделю: %s — %s", monday.Format("02.01.2006"), sunday.Format("02.01.2006"))

	var passed, failed, rest []string

	for _, r := range report {
		name := html.EscapeString(r.FullName)
		log.Println(name)
		var line string
		if r.Username != nil {
			line = fmt.Sprintf(`<a href="t.me/%s">%s</a> (%d сообщений)`, *r.Username, name, r.MessagesCount)
		} else {
			line = fmt.Sprintf(`<a href="tg://openmessage?user_id=%d">%s</a> (%d сообщений)`, r.UserID, name, r.MessagesCount)
		}

		if r.NormDone {
			passed = append(passed, line)
		} else {
			failed = append(failed, line)
		}
	}

	for _, r := range exemptMembers {
		name := html.EscapeString(r.FullName)
		var untilText string
		if !r.ExemptUntil.IsZero() {
			untilText = r.ExemptUntil.Format("02.01.2006 15:04")
		} else {
			untilText = "неизвестно"
		}

		var line string
		if r.Username != nil {
			line = fmt.Sprintf(`<a href="t.me/%s">%s</a> до %s`, *r.Username, name, untilText)
		} else {
			line = fmt.Sprintf(`<a href="tg://openmessage?user_isd=%d">%s</a> до %s`, r.UserID, name, untilText)
		}
		rest = append(rest, line)
	}

	var sb strings.Builder
	sb.WriteString(weekHeader)

	if len(passed) > 0 {
		sb.WriteString("\n✅ Прошли норму\n")
		sb.WriteString(strings.Join(passed, "\n"))
	}
	if len(failed) > 0 {
		sb.WriteString("\n\n❎ Не прошли норму\n")
		sb.WriteString(strings.Join(failed, "\n"))
	}
	if len(rest) > 0 {
		sb.WriteString("\n\n💛 Рест\n")
		sb.WriteString(strings.Join(rest, "\n"))
	}

	return sb.String()
}
func (h *Handler) ExemptMember(ctx context.Context, b *bot.Bot, update *models.Update) {
	senderMember, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
		ChatID: update.Message.Chat.ID,
		UserID: update.Message.From.ID,
	})
	if err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось проверить статус пользователя")
		return
	}

	text := update.Message.Text
	matches := h.setExemptRe.FindStringSubmatch(text)
	if len(matches) < 2 {
		h.AnswerMessage(ctx, b, update, "Неверный формат команды")
		return
	}
	args := strings.TrimSpace(matches[2])

	targetUserID, restArg, found, err := h.extractTargetUser(ctx, update, args)
	if err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось найти пользователя")
		return
	}
	if !found {
		targetUserID = update.Message.From.ID
		restArg = args
	}

	date, ok := h.dateParser.Parse(restArg)
	if !ok {
		if restArg == "" {
			h.ShowMemberExempt(ctx, b, update)
			return
		}
		h.AnswerMessage(ctx, b, update,
			"Не понял формат. Примеры:\n+рест 12.01\n+рест 2 недели\n+рест месяц",
		)
		return
	}
	if date.Before(time.Now()) {
		h.AnswerMessage(ctx, b, update, "Нельзя указывать прошедшую дату")
		return
	}

	u, err := h.userService.GetUser(ctx, targetUserID)
	if err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось получить пользователя")
		return
	}

	isAdmin, err := h.service.IsAdmin(ctx, update.Message.Chat.ID, update.Message.From.ID)
	if err != nil {
		log.Println("Check admin error", err)
	}

	if senderMember.Owner == nil && !isAdmin {
		if err := h.createExemptRequest(ctx, b, update, targetUserID, u, date); err != nil {
			log.Println("Failed to create exempt request", err)
			h.AnswerMessage(ctx, b, update, "Не удалось создать заявку")
		}
		return
	}

	if err := h.service.ExemptUser(ctx, update.Message.Chat.ID, targetUserID, date); err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось создать рест")
		return
	}

	if targetUserID == update.Message.From.ID {
		h.AnswerMessage(ctx, b, update, fmt.Sprintf("Вы добавлены в рест до %s", date.Format("02.01.2006")))
	} else {
		h.AnswerMessage(ctx, b, update, fmt.Sprintf(`Пользователь <a href="tg://user?id=%d">%s</a> добавлен в рест до %s`, targetUserID, html.EscapeString(u.FirstName), date.Format("02.01.2006")))
	}
}

func (h *Handler) createExemptRequest(ctx context.Context, b *bot.Bot, update *models.Update, targetUserID int64, targetUser model.User, date time.Time) error {
	if err := h.service.CreateExemptRequest(ctx, update.Message.Chat.ID, targetUserID, int64(update.Message.ID), date); err != nil {
		return err
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "✅ Одобрить", CallbackData: fmt.Sprintf("approve:%d:%d", targetUserID, update.Message.ID)},
				{Text: "❌ Отклонить", CallbackData: fmt.Sprintf("reject:%d:%d", targetUserID, update.Message.ID)},
			},
		},
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: fmt.Sprintf(
			"Для пользователя <a href=\"tg://user?id=%d\">%s</a> запрошен рест до %s",
			targetUserID,
			html.EscapeString(targetUser.FirstName),
			date.Format("02.01.2006"),
		),
		ParseMode:   "HTML",
		ReplyMarkup: kb,
	})
	return err
}

func (h *Handler) ShowMemberExempt(ctx context.Context, b *bot.Bot, update *models.Update) {
	targetUserID, _, found, err := h.extractTargetUser(ctx, update, "")
	if err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось найти пользователя")
		return
	}
	if !found {
		targetUserID = update.Message.From.ID
	}

	u, err := h.userService.GetUser(ctx, targetUserID)
	if err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось получить пользователя для информации о ресте")
		return
	}

	exempt, err := h.service.GetMemberExempt(ctx, update.Message.Chat.ID, targetUserID)
	if err != nil || exempt == nil {
		h.AnswerMessage(ctx, b, update, "Пользователь не находится в ресте")
		return
	}

	h.AnswerMessage(ctx, b, update,
		fmt.Sprintf(`Пользователь <a href="tg://user?id=%d">%s</a> находится в ресте до %s`,
			targetUserID,
			html.EscapeString(u.FirstName),
			exempt.Format("02.01.2006"),
		),
	)
}

func (h *Handler) EndMemberExempt(ctx context.Context, b *bot.Bot, update *models.Update) {
	targetUserID, _, found, err := h.extractTargetUser(ctx, update, "")
	if err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось найти пользователя")
		return
	}
	if !found {
		targetUserID = update.Message.From.ID
	}

	member, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
		ChatID: update.Message.Chat.ID,
		UserID: update.Message.From.ID,
	})
	if err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось проверить вашу роль в чате")
		return
	}

	if targetUserID != update.Message.From.ID && member.Owner == nil {
		h.AnswerMessage(ctx, b, update, "Вы можете удалить из реста только себя")
		return
	}

	u, err := h.userService.GetUser(ctx, targetUserID)
	if err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось получить пользователя")
		return
	}

	exempt, err := h.service.GetMemberExempt(ctx, update.Message.Chat.ID, targetUserID)

	if err != nil {
		log.Println("Get member exempt", err)
		h.AnswerMessage(ctx, b, update, "Ошибка, не удалось проверить рест пользователя")
		return
	}
	if exempt == nil {
		if targetUserID == update.Message.From.ID {
			h.AnswerMessage(ctx, b, update, "У вас нет реста")
		} else {
			h.AnswerMessage(ctx, b, update, "У пользователя нет реста")

		}
		return
	}

	if err := h.service.EndMemberExempt(ctx, update.Message.Chat.ID, targetUserID); err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось удалить пользователя из реста")
		return
	}

	if targetUserID == update.Message.From.ID {
		h.AnswerMessage(ctx, b, update, "Вы удалены из реста")
	} else {
		h.AnswerMessage(ctx, b, update,
			fmt.Sprintf(`Пользователь <a href="tg://user?id=%d">%s</a> удалён из реста`,
				targetUserID,
				html.EscapeString(u.FirstName),
			),
		)
	}
}

func (h *Handler) ApproveExemptRequest(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !h.checkOwnerOrAdmin(ctx, b, update, update.CallbackQuery.Message.Message.Chat.ID, update.CallbackQuery.From.ID) {
		h.AnswerCallback(ctx, b, update, "Кнопка доступна только создателю или пользователю с ролью администратор в боте")
		return
	}

	callbackData := update.CallbackQuery.Data
	parts := strings.SplitN(callbackData, ":", 3)
	if len(parts) != 3 {
		h.AnswerCallback(ctx, b, update, "Некорректная структура запроса")
		return
	}
	fromID, err := strconv.Atoi(parts[1])
	if err != nil {
		h.AnswerCallback(ctx, b, update, "Некорректный запрос, не найден айди пользователя")
		return
	}
	messageID, err := strconv.Atoi(parts[2])
	if err != nil {
		h.AnswerCallback(ctx, b, update, "Некорректный запрос, не найден айди сообщения")
		return
	}

	exemptRequest, err := h.service.GetExemptRequest(ctx, update.CallbackQuery.Message.Message.Chat.ID, fromID, messageID)
	if err != nil {
		log.Println("Exempt request not found", err)
		h.AnswerCallback(ctx, b, update, "Не найден запрос на рест")
		return
	}

	if err := h.service.ApproveExemptRequest(ctx, update.CallbackQuery.Message.Message.Chat.ID, int64(fromID), int64(messageID), exemptRequest.ExemptUntil); err != nil {
		log.Println("Failed to approve exempt request", err)
		h.AnswerCallback(ctx, b, update, "Не удалось одобрить запрос")
		return
	}
	u, err := h.userService.GetUser(ctx, int64(fromID))
	if err != nil {
		log.Println(err)
		h.AnswerCallback(ctx, b, update, "Не удалось найти пользователя")
		return
	}

	h.EditMessage(ctx, b, update,
		fmt.Sprintf(`Запрос одобрен. У <a href="tg://user?id=%d">%s</a> рест до %s`,
			fromID,
			html.EscapeString(u.FirstName),
			exemptRequest.ExemptUntil.Format("02.01.2006"),
		),
	)

}

func (h *Handler) RejectExemptRequest(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !h.checkOwnerOrAdmin(ctx, b, update, update.CallbackQuery.Message.Message.Chat.ID, update.CallbackQuery.From.ID) {
		h.AnswerCallback(ctx, b, update, "Кнопка доступна только создателю или пользователю с ролью администратор в боте")
		return
	}

	if err := h.service.RejectExemptRequest(ctx, update.CallbackQuery.Message.Message.Chat.ID, update.CallbackQuery.From.ID, update.CallbackQuery.Message.Message.ID); err != nil {
		h.AnswerCallback(ctx, b, update, "Не удалось отклонить запрос")
		return
	}

	h.EditMessage(ctx, b, update, fmt.Sprintf("Запрос на рест для  <a href=\"tg://user?id=%d\">%s</a> отклонён", update.CallbackQuery.Message.Message.From.ID, update.CallbackQuery.Message.Message.From.FirstName))
}

func (h *Handler) extractTargetUser(ctx context.Context, update *models.Update, args string) (int64, string, bool, error) {
	var userID *int64

	if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From != nil {
		userID = &update.Message.ReplyToMessage.From.ID
	} else if update.Message.ExternalReply != nil && update.Message.ExternalReply.Origin.MessageOriginUser != nil {
		userID = &update.Message.ExternalReply.Origin.MessageOriginUser.SenderUser.ID
	}

	if userID != nil {
		return *userID, args, true, nil
	}

	textRunes := []rune(update.Message.Text)
	for _, e := range update.Message.Entities {
		if e.User != nil {
			name := string(textRunes[e.Offset : e.Offset+e.Length])
			restArg := strings.TrimSpace(strings.Replace(args, name, "", 1))
			return e.User.ID, restArg, true, nil
		} else if e.Type == "mention" {
			username := textRunes[e.Offset : e.Offset+e.Length]
			u, err := h.userService.GetUserByUsername(ctx, string(username[1:]))
			if err != nil {
				return 0, "", false, err
			}
			name := string(textRunes[e.Offset : e.Offset+e.Length])
			restArg := strings.TrimSpace(strings.Replace(args, name, "", 1))
			return u.ID, restArg, true, nil
		}
	}

	return 0, args, false, nil
}

func (h *Handler) checkOwnerOrAdmin(ctx context.Context, b *bot.Bot, update *models.Update, chatID, userID int64) bool {
	member, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		log.Printf("Failed to get chat member for owner check: %v", err)
		return false
	}
	if member.Owner != nil {
		return true
	}

	isAdmin, err := h.service.IsAdmin(ctx, chatID, userID)
	if err != nil {
		log.Printf("Failed to check db admin: %v", err)
		return false
	}
	return isAdmin
}

func (h *Handler) AddAdmin(ctx context.Context, b *bot.Bot, update *models.Update) {
	member, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
		ChatID: update.Message.Chat.ID,
		UserID: update.Message.From.ID,
	})
	if err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось проверить ваши права")
		return
	}
	if member.Owner == nil {
		h.AnswerMessage(ctx, b, update, "Только создатель чата может добавлять администраторов бота")
		return
	}

	targetUserID, _, found, err := h.extractTargetUser(ctx, update, "")
	if err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось найти пользователя")
		return
	}
	if !found {
		h.AnswerMessage(ctx, b, update, "Пользователь не найден")
		return
	}

	if err := h.service.AddAdmin(ctx, update.Message.Chat.ID, targetUserID); err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось добавить администратора")
		return
	}

	u, err := h.userService.GetUser(ctx, targetUserID)
	name := "пользователя"
	if err == nil {
		name = html.EscapeString(u.FirstName)
	}

	h.AnswerMessage(ctx, b, update, fmt.Sprintf("Пользователь <a href=\"tg://user?id=%d\">%s</a> назначен администратором бота", targetUserID, name))
}

func (h *Handler) RemoveAdmin(ctx context.Context, b *bot.Bot, update *models.Update) {
	member, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
		ChatID: update.Message.Chat.ID,
		UserID: update.Message.From.ID,
	})
	if err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось проверить ваши права")
		return
	}
	if member.Owner == nil {
		h.AnswerMessage(ctx, b, update, "Только создатель чата может удалять администраторов бота")
		return
	}

	targetUserID, _, found, err := h.extractTargetUser(ctx, update, "")
	if err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось найти пользователя")
		return
	}
	if !found {
		h.AnswerMessage(ctx, b, update, "Пользователь не найден")
		return
	}

	if err := h.service.RemoveAdmin(ctx, update.Message.Chat.ID, targetUserID); err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось удалить администратора")
		return
	}

	u, err := h.userService.GetUser(ctx, targetUserID)
	name := "пользователя"
	if err == nil {
		name = html.EscapeString(u.FirstName)
	}

	h.AnswerMessage(ctx, b, update, fmt.Sprintf("Пользователь <a href=\"tg://user?id=%d\">%s</a> удалён из администраторов бота", targetUserID, name))
}

func (h *Handler) ShowAdmins(ctx context.Context, b *bot.Bot, update *models.Update) {
	admins, err := h.service.GetAdmins(ctx, update.Message.Chat.ID)
	if err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось получить список администраторов")
		return
	}

	if len(admins) == 0 {
		h.AnswerMessage(ctx, b, update, "Список администраторов пуст")
		return
	}

	var sb strings.Builder
	sb.WriteString("👮 Администраторы бота:\n")
	for _, admin := range admins {
		sb.WriteString(fmt.Sprintf("\n<a href=\"tg://user?id=%d\">%s</a> (с %s)", admin.UserID, html.EscapeString(admin.DisplayName), admin.CreatedAt.Format("02.01.2006")))
	}
	h.AnswerMessage(ctx, b, update, sb.String())
}

func (h *Handler) UpdateChat(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !h.checkOwnerOrAdmin(ctx, b, update, update.Message.Chat.ID, update.Message.From.ID) {
		h.AnswerMessage(ctx, b, update, "Команда доступна только создателю или пользователю с ролью администратор в боте")
		return
	}

	count, err := h.updateChatMembers(ctx, b, update.Message.Chat.ID)
	if err != nil {
		log.Println("Update chat members error", err)
		h.AnswerMessage(ctx, b, update, "Не удалось обновить данные чата")
		return
	}

	h.AnswerMessage(ctx, b, update, fmt.Sprintf("Чат обновлён. Найдено %d участников", count))
}

func (h *Handler) updateChatMembers(ctx context.Context, b *bot.Bot, chatID int64) (int, error) {
	admins, err := b.GetChatAdministrators(ctx, &bot.GetChatAdministratorsParams{
		ChatID: chatID,
	})
	if err != nil {
		return 0, err
	}

	members := make([]ChatMemberUpdate, 0, len(admins))
	for _, admin := range admins {
		var chatUser *models.User
		var customTitle string

		if admin.Administrator != nil {
			chatUser = &admin.Administrator.User
			customTitle = admin.Administrator.CustomTitle
		} else if admin.Owner != nil {
			chatUser = admin.Owner.User
			customTitle = admin.Owner.CustomTitle
		} else {
			chatUser = nil
		}
		if chatUser == nil {
			continue
		}

		if chatUser.IsBot {
			continue
		}
		members = append(members, ChatMemberUpdate{
			User: model.User{
				ID:        chatUser.ID,
				FirstName: chatUser.FirstName,
				LastName:  chatUser.LastName,
				Username:  &chatUser.Username,
			},
			CustomTitle: customTitle,
		})
	}

	if err := h.service.UpdateChatMembers(ctx, chatID, members); err != nil {
		return 0, err
	}
	return len(members), nil
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
		h.AnswerMessage(ctx, b, update, fmt.Sprintf("🕊 <a href=\"tg://user?id=%d\">%s</a> (<b>%s</b>) покинул нас...", leftMember.ID, html.EscapeString(leftMember.FirstName), html.EscapeString(title)))
	}
}

func (h *Handler) ShowRoles(ctx context.Context, b *bot.Bot, update *models.Update) {
	members, err := h.service.GetRoles(ctx, update.Message.Chat.ID)
	if err != nil {
		log.Println()
		h.AnswerMessage(ctx, b, update, "Не удалось получить список ролей")
		return
	}

	if len(members) == 0 {
		h.AnswerMessage(ctx, b, update, "В чате нет установленных ролей")
		return
	}

	var sb strings.Builder
	sb.WriteString("🎭 Роли участников:\n")
	for _, m := range members {
		u, err := h.userService.GetUser(ctx, m.UserID)
		name := "Пользователь"
		if err == nil {
			name = html.EscapeString(u.FirstName)
		}
		sb.WriteString(fmt.Sprintf("\n<a href=\"tg://openmessage?user_id=%d\">%s</a>: <b>%s</b>", m.UserID, name, html.EscapeString(m.CustomTitle)))
	}

	h.AnswerMessage(ctx, b, update, sb.String())
}

func (h *Handler) SetRole(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Remove command from text to get args
	args := h.setRoleRe.ReplaceAllString(update.Message.Text, "")
	args = strings.TrimSpace(args)

	targetUserID, role, found, err := h.extractTargetUser(ctx, update, args)
	if err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось найти пользователя")
		return
	}
	if !found {
		// Default to self if no role title provided
		if role == "" {
			targetUserID = update.Message.From.ID
		} else {
			h.AnswerMessage(ctx, b, update, "Вы не указали кому хотите выдать роль. Укажите @mention или ответьте на сообщение.")
			return
		}
	}

	role = strings.TrimSpace(role)

	// Case 1: Just show the role (no title provided)
	if role == "" {
		mTitle, err := h.service.GetMemberRole(ctx, update.Message.Chat.ID, targetUserID)
		if err != nil {
			h.AnswerMessage(ctx, b, update, "Не удалось получить роль пользователя")
			return
		}

		u, err := h.userService.GetUser(ctx, targetUserID)
		name := "Пользователь"
		if err == nil {
			name = html.EscapeString(u.FirstName)
		}

		if mTitle == "" {
			h.AnswerMessage(ctx, b, update, fmt.Sprintf("У пользователя <a href=\"tg://user?id=%d\">%s</a> нет роли", targetUserID, name))
		} else {
			h.AnswerMessage(ctx, b, update, fmt.Sprintf("Роль пользователя <a href=\"tg://user?id=%d\">%s</a>: <b>%s</b>", targetUserID, name, html.EscapeString(mTitle)))
		}
		return
	}

	// Case 2: Set the role (title provided) - Check Admin permissions
	if !h.checkOwnerOrAdmin(ctx, b, update, update.Message.Chat.ID, update.Message.From.ID) {
		h.AnswerMessage(ctx, b, update, "Команда изменения ролей доступна только создателю чата и администраторам бота")
		return
	}

	if len(role) > 32 {
		h.AnswerMessage(ctx, b, update, "Слишком длинная роль (максимум 32 символа)")
		return
	}

	member, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
		ChatID: update.Message.Chat.ID,
		UserID: targetUserID,
	})
	if err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось получить информацию о пользователе")
		return
	}

	if member.Owner != nil {
		h.AnswerMessage(ctx, b, update, "Нельзя изменить роль создателя чата")
		return
	}

	if member.Administrator != nil {
		if !member.Administrator.CanBeEdited {
			h.AnswerMessage(ctx, b, update, "Я не могу изменить этого администратора (он назначен другим админом)")
			return
		}
		if _, err := b.SetChatAdministratorCustomTitle(ctx, &bot.SetChatAdministratorCustomTitleParams{
			ChatID:      update.Message.Chat.ID,
			UserID:      targetUserID,
			CustomTitle: role,
		}); err != nil {
			log.Println("Telegram set custom title error", err)
			h.AnswerMessage(ctx, b, update, "Не удалось изменить роль в Telegram")
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
			h.AnswerMessage(ctx, b, update, "Не удалось назначить пользователя администратором. Проверьте права бота.")
			return
		}

		if _, err := b.SetChatAdministratorCustomTitle(ctx, &bot.SetChatAdministratorCustomTitleParams{
			ChatID:      update.Message.Chat.ID,
			UserID:      targetUserID,
			CustomTitle: role,
		}); err != nil {
			log.Println("Telegram set custom title after promote error", err)
			h.AnswerMessage(ctx, b, update, "Пользователь назначен администратором, но не удалось установить роль")
			return
		}

	} else {
		h.AnswerMessage(ctx, b, update, "Пользователь не является участником чата")
		return
	}

	if err := h.service.SetMemberTitle(ctx, update.Message.Chat.ID, targetUserID, role); err != nil {
		log.Println("DB set custom title error", err)
		h.AnswerMessage(ctx, b, update, "Роль в Telegram изменена, но не удалось сохранить в базе данных")
		return
	}

	h.AnswerMessage(ctx, b, update, fmt.Sprintf("Роль пользователя обновлена на: <b>%s</b>", html.EscapeString(role)))
}
