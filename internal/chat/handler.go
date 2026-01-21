package chat

import (
	"activity-bot/internal/base"
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
}

func NewHandler(service *Service, userService *user.Service, dateParser *DateParser, setNormRe, setExemptRe *regexp.Regexp) *Handler {
	return &Handler{base.Handler{}, service, userService, dateParser, setNormRe, setExemptRe}
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
		name := r.DisplayName
		line := fmt.Sprintf(`<a href="tg://user?id=%d">%s</a> (%d сообщений)`, r.UserID, html.EscapeString(name), r.MessagesCount)

		if r.NormDone {
			passed = append(passed, line)
		} else {
			failed = append(failed, line)
		}
	}

	for _, r := range exemptMembers {
		name := r.DisplayName
		var untilText string
		if !r.ExemptUntil.IsZero() {
			untilText = r.ExemptUntil.Format("02.01.2006 15:04")
		} else {
			untilText = "неизвестно"
		}

		line := fmt.Sprintf(`<a href="tg://user?id=%d">%s</a> до %s`, r.UserID, html.EscapeString(name), untilText)
		rest = append(rest, line)
	}

	text := weekHeader

	if len(passed) > 0 {
		text += fmt.Sprintf("✅ Прошли норму\n%s", strings.Join(passed, "\n"))
	}
	if len(failed) > 0 {
		text += fmt.Sprintf("\n\n❎ Не прошли норму\n%s", strings.Join(failed, "\n"))
	}
	if len(rest) > 0 {
		text += fmt.Sprintf("\n\n💛 Рест\n%s", strings.Join(rest, "\n"))
	}

	h.AnswerMessage(ctx, b, update, text)
}
func (h *Handler) ExemptMember(ctx context.Context, b *bot.Bot, update *models.Update) {
	member, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
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

	var targetUserID int64
	var restArg string

	if update.Message.ReplyToMessage != nil {
		targetUserID = update.Message.ReplyToMessage.From.ID
		restArg = args

	} else {
		found := false
		for _, e := range update.Message.Entities {
			textRunes := []rune(update.Message.Text)
			if e.User != nil {
				targetUserID = e.User.ID
				name := string(textRunes[e.Offset : e.Offset+e.Length])
				restArg = strings.TrimSpace(strings.Replace(args, name, "", 1))
				found = true
				break
			} else if e.Type == "mention" {
				username := textRunes[e.Offset : e.Offset+e.Length]
				u, err := h.userService.GetUserByUsername(ctx, string(username[1:]))
				if err != nil {
					log.Println(username)
					h.AnswerMessage(ctx, b, update, "Не удалось найти пользователя по юзернейму")
					return
				}
				targetUserID = u.ID
				textRunes := []rune(update.Message.Text)
				name := string(textRunes[e.Offset : e.Offset+e.Length])
				restArg = strings.TrimSpace(strings.Replace(args, name, "", 1))
				found = true
				break
			}
		}

		if !found {
			targetUserID = update.Message.From.ID
			restArg = args
		}
	}
	log.Println(restArg)

	date, ok := h.dateParser.Parse(restArg)
	if !ok {
		h.AnswerMessage(ctx, b, update,
			"Не понял формат. Примеры:\n+рест 12.01\n+рест 2 недели\n+рест месяц",
		)
		return
	}
	if date.Before(time.Now()) {
		h.AnswerMessage(ctx, b, update, "Нельзя указывать прошедшую дату")
		return
	}

	if member.Owner == nil {

		if err := h.service.CreateExemptRequest(
			ctx,
			update.Message.Chat.ID,
			targetUserID,
			int64(update.Message.ID),
			date,
		); err != nil {
			log.Println("Failed to create exempt request", err)
			h.AnswerMessage(ctx, b, update, "Не удалось создать заявку")
			return
		}

		kb := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{
						Text:         "✅ Одобрить",
						CallbackData: fmt.Sprintf("approve:%d:%d", targetUserID, update.Message.ID),
					},
					{
						Text:         "❌ Отклонить",
						CallbackData: fmt.Sprintf("reject:%d:%d", targetUserID, update.Message.ID),
					},
				},
			},
		}

		u, err := h.userService.GetUser(ctx, targetUserID)
		if err != nil {
			log.Println("Failed to get user", err)
			h.AnswerMessage(ctx, b, update, "Не удалось получить пользователя которому направлен запрос на рест")
		}

		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text: fmt.Sprintf(
				"Для пользователя <a href=\"tg://user?id=%d\">%s</a> запрошен рест до %s",
				targetUserID,
				html.EscapeString(u.FirstName),
				date.Format("02.01.2006"),
			),
			ParseMode:   "HTML",
			ReplyMarkup: kb,
		}); err != nil {
			log.Println("Send Message Error", err)
		}
		return
	}

	u, err := h.userService.GetUser(ctx, targetUserID)
	if err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось получить пользователя которого нужно добавить в рест")
		return
	}

	if err := h.service.ExemptUser(
		ctx,
		update.Message.Chat.ID,
		targetUserID,
		date,
	); err != nil {
		h.AnswerMessage(ctx, b, update, "Не удалось создать рест")
		return
	}

	if targetUserID == update.Message.From.ID {
		h.AnswerMessage(ctx, b, update,
			fmt.Sprintf("Вы добавлены в рест до %s", date.Format("02.01.2006")),
		)
		return
	}

	h.AnswerMessage(ctx, b, update,
		fmt.Sprintf(
			`Пользователь <a href="tg://user?id=%d">%s</a> добавлен в рест до %s`,
			targetUserID,
			html.EscapeString(u.FirstName),
			date.Format("02.01.2006"),
		),
	)
}

func (h *Handler) ShowMemberExempt(ctx context.Context, b *bot.Bot, update *models.Update) {
	var targetUserID int64

	if update.Message.ReplyToMessage != nil {
		targetUserID = update.Message.ReplyToMessage.From.ID
	} else {
		found := false
		for _, e := range update.Message.Entities {
			if e.User != nil {
				targetUserID = e.User.ID
				found = true
				break
			} else if e.Type == "mention" {
				textRunes := []rune(update.Message.Text)
				username := textRunes[e.Offset : e.Offset+e.Length]
				u, err := h.userService.GetUserByUsername(ctx, string(username[1:]))
				if err == nil {
					targetUserID = u.ID
					found = true
					break
				}
			}
		}
		if !found {
			targetUserID = update.Message.From.ID
		}
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
	var targetUserID int64

	if update.Message.ReplyToMessage != nil {
		targetUserID = update.Message.ReplyToMessage.From.ID
	} else {
		found := false
		for _, e := range update.Message.Entities {
			if e.User != nil {
				targetUserID = e.User.ID
				found = true
				break
			} else if e.Type == "mention" {
				textRunes := []rune(update.Message.Text)
				username := textRunes[e.Offset : e.Offset+e.Length]
				u, err := h.userService.GetUserByUsername(ctx, string(username[1:]))
				if err == nil {
					targetUserID = u.ID
					found = true
					break
				}
			}
		}
		if !found {
			targetUserID = update.Message.From.ID
		}
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
		h.AnswerMessage(ctx, b, update, "Не удалось получить пользователя, которого нужно удалить из реста")
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
	member, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		UserID: update.CallbackQuery.From.ID,
	})
	if err != nil {
		h.AnswerCallback(ctx, b, update, "Не удалось проверить вашу роль")
		return
	}
	if member.Owner == nil {
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
	member, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		UserID: update.CallbackQuery.From.ID,
	})
	if err != nil {
		h.AnswerCallback(ctx, b, update, "Не удалось проверить вашу роль")
		return
	}
	if member.Owner == nil {
		h.AnswerCallback(ctx, b, update, "Кнопка доступна только создателю или пользователю с ролью администратор в боте")
		return
	}

	if err := h.service.RejectExemptRequest(ctx, update.CallbackQuery.Message.Message.Chat.ID, update.CallbackQuery.From.ID, update.CallbackQuery.Message.Message.ID); err != nil {
		h.AnswerCallback(ctx, b, update, "Не удалось отклонить запрос")
		return
	}

	h.EditMessage(ctx, b, update, fmt.Sprintf("Запрос на рест для  <a href=\"tg://user?id=%d\">%s</a> отклонён", update.CallbackQuery.Message.Message.From.ID, update.CallbackQuery.Message.Message.From.FirstName))

}
