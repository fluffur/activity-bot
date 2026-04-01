package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/chat"
	"activity-bot/internal/chat/view"
	"activity-bot/internal/cmd"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/logger"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"activity-bot/internal/session"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Handler struct {
	service        *chat.Service
	adminService   *admin.Service
	memberService  *member.Service
	sessionService *session.Service
	dateParser     *helpers.DateParser
}

type chatInfo struct {
	ID    int64
	Title string
}

func New(service *chat.Service, adminService *admin.Service, memberService *member.Service, sessionService *session.Service, dateParser *helpers.DateParser) *Handler {
	return &Handler{service, adminService, memberService, sessionService, dateParser}
}

func (h *Handler) ShowNorm(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	return ctx.ReplyHTML(b, view.FormatNorm(c.NormWarn, c.NormBan))
}

func (h *Handler) SetNorm(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	norm, err := ctx.Number()
	if err != nil {
		return err
	}
	action := ctx.TextOrDefault("варн")
	validActions := map[string]bool{
		"варн": true,
		"бан":  true,
	}
	if !validActions[action] {
		return ctx.Reply(b, fmt.Sprintf("❌ Неверное действие: '%s'. Допустимые: 'варн', 'бан'", action), nil)
	}

	if action == "варн" {
		if err := h.service.SetWarnNorm(ctx.StdContext(), c.ID, norm); err != nil {
			return err
		}
	} else {
		if err := h.service.SetBanNorm(ctx.StdContext(), c.ID, norm); err != nil {
			return err
		}
	}

	return ctx.Reply(b, view.FormatNormSet(norm, action), nil)
}

func (h *Handler) RemoveNorm(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	action := ctx.TextOrDefault("варн")
	var oldNorm int32
	var setFn func() error

	switch action {
	case "варн":
		oldNorm = c.NormWarn
		setFn = func() error {
			return h.service.SetWarnNorm(ctx.StdContext(), c.ID, 0)
		}
	case "бан":
		oldNorm = c.NormBan
		setFn = func() error {
			return h.service.SetBanNorm(ctx.StdContext(), c.ID, 0)
		}
	default:
		return ctx.Reply(b, "❌ Неверное действие. Допустимые: 'варн', 'бан'", nil)
	}

	if oldNorm == 0 {
		return ctx.Reply(b, fmt.Sprintf("❌ Норма на %s не была установлена", action), nil)
	}

	if err := setFn(); err != nil {
		return err
	}

	return ctx.Reply(b, fmt.Sprintf("Норма %d (%s) удалена", oldNorm, action), nil)
}

func (h *Handler) ShowNewbieThreshold(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatNewbieThreshold(c.NewbieThresholdDays), nil)
}

func (h *Handler) SetNewbieThreshold(b *gotgbot.Bot, ctx *command.Context) error {
	date, err := ctx.Date()
	if err != nil {
		return err
	}
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	duration := time.Until(date)
	days := int32(duration.Hours() / 24)
	if days < 0 {
		return ctx.Reply(b, "Дата не может быть в прошлом!", nil)
	}

	if err := h.service.SetNewbieThreshold(ctx.StdContext(), c.ID, days); err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatNewbieThresholdSet(days), nil)
}

func (h *Handler) SetPrompt(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	if err := h.service.SetChatPrompt(ctx.StdContext(), c.ID, ctx.RawArgsHTML); err != nil {
		return err
	}

	return ctx.Reply(b, "Промпт установлен успешно", nil)
}

func (h *Handler) ShowPrompt(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatPrompt(c.AISystemPrompt), nil)
}

func (h *Handler) ShowWeekStart(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	return ctx.ReplyHTML(b, view.FormatWeekStart(int(c.WeekStartDay), c.WeekStartTime))
}

func (h *Handler) SetWeekStart(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	date, err := ctx.Date()
	if err != nil {
		return err
	}
	weekday := date.Weekday()
	hour := date.Hour()
	minute := date.Minute()
	if err := h.service.SetWeekStartDay(ctx.StdContext(), c.ID, int(weekday)); err != nil {
		return err
	}
	newTime := fmt.Sprintf("%0.2d:%0.2d", hour, minute)
	if err := h.service.SetWeekStartTime(ctx.StdContext(), c.ID, newTime); err != nil {
		return err
	}

	return ctx.ReplyHTML(b, view.FormatWeekStartSet(int(weekday), newTime))
}

func parseTime(arg string) (int, int, bool) {
	var h, m int
	if _, err := fmt.Sscanf(arg, "%d:%d", &h, &m); err == nil {
		if h >= 0 && h <= 23 && m >= 0 && m <= 59 {
			return h, m, true
		}
	}
	if _, err := fmt.Sscanf(arg, "%d", &h); err == nil {
		if h >= 0 && h <= 23 {
			return h, 0, true
		}
	}
	return 0, 0, false
}

func parseWeekStartDay(arg string) (int, bool) {
	switch strings.ToLower(arg) {
	case "1", "пн", "понедельник":
		return 1, true
	case "2", "вт", "вторник":
		return 2, true
	case "3", "ср", "среда":
		return 3, true
	case "4", "чт", "четверг":
		return 4, true
	case "5", "пт", "пятница":
		return 5, true
	case "6", "сб", "суббота":
		return 6, true
	case "7", "вс", "воскресенье":
		return 7, true
	default:
		return 0, false
	}
}

func (h *Handler) ShowPrefix(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatPrefix(c.CommandPrefix), nil)
}

func (h *Handler) SetPrefix(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	prefix := ctx.TextOrDefault("")
	if prefix == "" {
		return ctx.Reply(b, "❌ Укажите префикс", nil)
	}

	if err := h.service.SetCommandPrefix(ctx.StdContext(), c.ID, prefix); err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatPrefixSet(prefix), nil)
}

func (h *Handler) EnablePrefixes(b *gotgbot.Bot, ctx *cmd.Context) error {
	if err := h.service.SetAllowPrefixless(ctx.StdContext(), ctx.TargetChatID(), true); err != nil {
		return err
	}
	return ctx.ReplyHTML(b, view.FormatPrefixlessToggle(true))
}

func (h *Handler) DisablePrefixes(b *gotgbot.Bot, ctx *cmd.Context) error {
	if err := h.service.SetAllowPrefixless(ctx.StdContext(), ctx.TargetChatID(), false); err != nil {
		return err
	}
	return ctx.ReplyHTML(b, view.FormatPrefixlessToggle(false))
}

func (h *Handler) ShowPrefixes(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.service.GetChat(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}
	return ctx.ReplyHTML(b, view.FormatPrefixlessStatus(c.AllowPrefixless))
}

func (h *Handler) Manage(b *gotgbot.Bot, ctx *command.Context) error {
	chatIDs, err := h.service.GetUserManagedChats(ctx.StdContext(), ctx.EffectiveUser.Id, h.adminService.OwnerID())
	if err != nil {
		return err
	}

	if len(chatIDs) == 0 {
		return h.SendDM(b, ctx, "❌ У вас нет доступных чатов для управления", nil)
	}

	chats := make([]chatInfo, 0)
	for _, c := range chatIDs {
		title := c.Title
		if title == "" {
			title = "Чат без названия"
		}

		chats = append(chats, chatInfo{ID: c.ID, Title: title})
	}

	if len(chats) == 0 {
		return h.SendDM(b, ctx, "❌ Не удалось получить информацию о ваших чатах", nil)
	}

	return h.SendDM(b, ctx, "Выберите чат для управления:", h.getManageKeyboard(chats, 1))
}

func (h *Handler) SendDM(b *gotgbot.Bot, ctx *command.Context, text string, replyMarkup gotgbot.ReplyMarkup) error {
	opts := &gotgbot.SendMessageOpts{
		MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
	}
	if replyMarkup != nil {
		opts.ReplyMarkup = replyMarkup
	}
	_, err := b.SendMessage(ctx.EffectiveSender.Id(), text, opts)
	if err != nil {
		return ctx.Reply(b, "Не удалось отправить сообщение в лс", nil)
	}
	if ctx.EffectiveChat.Type != gotgbot.ChatTypePrivate {
		return ctx.Reply(b, "Ответ отправлен в лс", nil)
	}
	return nil

}

func (h *Handler) getManageKeyboard(chats []chatInfo, page int) gotgbot.InlineKeyboardMarkup {
	const itemsPerPage = 8

	totalPages := (len(chats) + itemsPerPage - 1) / itemsPerPage
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	startIdx := (page - 1) * itemsPerPage
	endIdx := startIdx + itemsPerPage
	if endIdx > len(chats) {
		endIdx = len(chats)
	}

	var buttons [][]gotgbot.InlineKeyboardButton
	for _, c := range chats[startIdx:endIdx] {
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: c.Title, CallbackData: "manage:" + strconv.FormatInt(c.ID, 10)},
		})
	}

	var navButtons []gotgbot.InlineKeyboardButton
	if page > 1 {
		navButtons = append(navButtons, gotgbot.InlineKeyboardButton{
			Text:         "< Назад",
			CallbackData: "manage_page:" + strconv.Itoa(page-1),
			Style:        "primary",
		})
	}
	if page < totalPages {
		navButtons = append(navButtons, gotgbot.InlineKeyboardButton{
			Text:         "Вперед >",
			CallbackData: "manage_page:" + strconv.Itoa(page+1),
			Style:        "primary",
		})
	}

	if len(navButtons) > 0 {
		buttons = append(buttons, navButtons)
	}

	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: buttons}
}

func (h *Handler) CallbackManagePage(b *gotgbot.Bot, ctx *command.Context) error {
	data := ctx.CallbackQuery.Data
	pageStr := strings.TrimPrefix(data, "manage_page:")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return err
	}

	chatIDs, err := h.service.GetUserManagedChats(ctx.StdContext(), ctx.EffectiveUser.Id, h.adminService.OwnerID())
	if err != nil {
		return err
	}

	chats := make([]chatInfo, 0)
	for _, c := range chatIDs {

		title := c.Title
		if title == "" {
			title = "Чат без названия"
		}

		chats = append(chats, chatInfo{ID: c.ID, Title: title})
	}

	if len(chats) == 0 {
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "У вас нет доступных чатов для управления",
		})
		return err
	}

	_, _, err = ctx.EffectiveMessage.EditReplyMarkup(b, &gotgbot.EditMessageReplyMarkupOpts{
		ReplyMarkup: h.getManageKeyboard(chats, page),
	})
	if err != nil && !strings.Contains(err.Error(), "message is not modified") {
		return err
	}

	_, err = ctx.CallbackQuery.Answer(b, nil)
	return err
}

func (h *Handler) OnNewChatTitle(_ *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	newTitle := ctx.EffectiveMessage.NewChatTitle
	if newTitle == "" {
		return nil
	}

	_, err = h.service.EnsureChatExists(ctx.StdContext(), c.ID, newTitle)
	return err
}

func (h *Handler) CallbackManage(b *gotgbot.Bot, ctx *command.Context) error {
	data := ctx.CallbackQuery.Data
	chatIDStr := strings.TrimPrefix(data, "manage:")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return err
	}
	m, err := h.memberService.GetChatMember(ctx.StdContext(), chatID, ctx.EffectiveUser.Id)
	if err != nil {
		return err
	}
	if !m.StatusGranted(model.StatusModerator) && h.adminService.OwnerID() != m.User.ID {
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "У вас больше нет прав администратора в этом чате",
			ShowAlert: true,
		})
		return err
	}

	if err := h.sessionService.SetActiveChat(ctx.StdContext(), ctx.EffectiveUser.Id, chatID); err != nil {
		return err
	}

	cht, err := b.GetChat(chatID, nil)
	if err != nil {
		_, err = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Ошибка: " + err.Error(),
		})
		return err
	}
	title := cht.Title

	_, err = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
		Text: "Выбран чат: " + title,
	})
	if err != nil {
		return err
	}

	_, _, err = ctx.EffectiveMessage.EditText(b, "Теперь вы управляете чатом: **"+title+"**\nВсе команды настроек теперь будут применяться к этому чату.", &gotgbot.EditMessageTextOpts{
		ParseMode: gotgbot.ParseModeMarkdown,
	})
	return err
}

func (h *Handler) UserChats(b *gotgbot.Bot, ctx *command.Context) error {
	chats, err := h.service.ListChatsWithoutNorm(
		ctx.StdContext(),
		ctx.EffectiveSender.Id(),
	)
	if err != nil {
		return err
	}
	logger.L.Info("chats", "chats", chats)
	if len(chats) == 0 {
		_, err := ctx.EffectiveChat.SendMessage(
			b,
			fmt.Sprintf("%s Все недельные нормы выполнены.", helpers.SuccessEmoji()),
			nil,
		)
		return err
	}

	var warnChats []model.ChatWithoutNorm
	var banChats []model.ChatWithoutNorm

	for _, c := range chats {
		if c.NormBan > 0 && c.WeekCount < int64(c.NormBan) {
			banChats = append(banChats, c)
			continue
		}

		if c.NormWarn > 0 && c.WeekCount < int64(c.NormWarn) {
			warnChats = append(warnChats, c)
		}
	}

	var text strings.Builder
	text.WriteString("📉 <b>Невыполненные нормы</b>\n\n")

	if len(banChats) > 0 {
		text.WriteString("🚫 <b>Бан</b>\n")
		text.WriteString("<blockquote expandable>")

		for i, c := range banChats {
			text.WriteString(fmt.Sprintf(
				"<a href=\"%s\">%s</a>\n",
				chatLink(c.ID),
				html.EscapeString(c.Title),
			))

			writeNormInfo(&text, c)

			text.WriteString(fmt.Sprintf(
				"Сообщений: %d\n",
				c.WeekCount,
			))
			if i < len(banChats)-1 {
				text.WriteString("\n")
			}
		}

		text.WriteString("</blockquote>")
	}

	if len(warnChats) > 0 {
		text.WriteString("⚠ <b>Варн</b>\n")
		text.WriteString("<blockquote expandable>")

		for i, c := range warnChats {
			text.WriteString(fmt.Sprintf(
				"<a href=\"%s\">%s</a>\n",
				chatLink(c.ID),
				html.EscapeString(c.Title),
			))

			writeNormInfo(&text, c)

			text.WriteString(fmt.Sprintf(
				"Сообщений: %d\n",
				c.WeekCount,
			))
			if i < len(warnChats)-1 {
				text.WriteString("\n")
			}

		}

		text.WriteString("</blockquote>\n")
	}

	text.WriteString(fmt.Sprintf(
		"Всего проблемных чатов: <b>%d</b>",
		len(banChats)+len(warnChats),
	))

	if _, err = b.SendMessage(
		ctx.EffectiveSender.Id(),
		text.String(),
		&gotgbot.SendMessageOpts{
			ParseMode: "HTML",
		},
	); err != nil {
		if ctx.EffectiveChat.Type != "private" {
			return ctx.Reply(
				b,
				"Не удалось отправить список норм в личные сообщения, убедитесь в том, что бот имеет доступ к личным сообщениям",
				nil,
			)
		}

		return err
	}

	if ctx.EffectiveChat.Type != "private" {
		return ctx.ReplyHTML(
			b,
			fmt.Sprintf("Список невыполненных норм отправлен вам в личные сообщения %s", helpers.SuccessEmoji()),
		)
	}

	return nil
}

func (h *Handler) EnableTags(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	if err := h.service.EnableTags(ctx.StdContext(), c.ID, true); err != nil {
		return err
	}

	return ctx.Reply(b, "Поддержка тегов в чате включена. Теперь при установке роли админка не выдается", nil)
}

func (h *Handler) DisableTags(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	if err := h.service.EnableTags(ctx.StdContext(), c.ID, false); err != nil {
		return err
	}

	return ctx.Reply(b, "Поддержка тегов в чате выключена. Теперь при установке роли выдается админка с минимальными правами", nil)
}

func (h *Handler) ShowTags(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	if c.TagsEnabled {
		return ctx.ReplyHTML(b, fmt.Sprintf("%s В чате поддерживаются теги. Это значит, что при установке роли пользователю устанавливается телеграм-тег, а не минимальные права администратора с подписью", helpers.SuccessEmoji()))
	}

	return ctx.Reply(b, "❌ В чате не поддерживаются теги. Это значит, что при установке роли пользователю выдаются минимальные права администратора с подписью, а не телеграм-тег", nil)
}

func writeNormInfo(text *strings.Builder, c model.ChatWithoutNorm) {
	var normParts []string

	if c.NormWarn > 0 {
		normParts = append(normParts,
			fmt.Sprintf("&lt;%d – варн", c.NormWarn),
		)
	}

	if c.NormBan > 0 {
		normParts = append(normParts,
			fmt.Sprintf("&lt;%d – бан", c.NormBan),
		)
	}

	if len(normParts) > 0 {
		text.WriteString(fmt.Sprintf("Норма: %d (", c.NormWarn))
		text.WriteString(strings.Join(normParts, ", "))
		text.WriteString(")\n")
	}
}

func chatLink(id int64) string {
	s := strconv.FormatInt(id, 10)

	if strings.HasPrefix(s, "-100") {
		internal := strings.TrimPrefix(s, "-100")
		return fmt.Sprintf("https://t.me/c/%s/1", internal)
	}

	return ""
}
