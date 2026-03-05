package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/chat"
	"activity-bot/internal/chat/view"
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"activity-bot/internal/session"
	"context"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Handler struct {
	service       *chat.Service
	memberService interface {
		SetCommandLevel(ctx context.Context, chatID int64, commandID string, level int16) error
		GetCommandLevels(ctx context.Context, chatID int64) (map[string]int16, error)
	}
	adminService   *admin.Service
	sessionService *session.Service
	dateParser     *helpers.DateParser
	factory        *cmd.Factory
}

type chatInfo struct {
	ID    int64
	Title string
}

func New(service *chat.Service, memberService interface {
	SetCommandLevel(ctx context.Context, chatID int64, commandID string, level int16) error
	GetCommandLevels(ctx context.Context, chatID int64) (map[string]int16, error)
}, adminService *admin.Service, sessionService *session.Service, dateParser *helpers.DateParser, factory *cmd.Factory) *Handler {
	return &Handler{service, memberService, adminService, sessionService, dateParser, factory}
}

func (h *Handler) ShowNorm(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.service.GetChat(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatNorm(c.NormWarn, c.NormBan), nil)
}

func (h *Handler) SetNorm(b *gotgbot.Bot, ctx *cmd.Context) error {
	arg := ctx.FirstArgument()
	fields := strings.Fields(arg)
	if len(fields) == 0 {
		return ctx.Reply(b, "❌ Нужно указать норму (и опционально действие: 'варн' или 'бан')", nil)
	}

	norm, err := strconv.Atoi(fields[0])
	if err != nil {
		return ctx.Reply(b, "❌ Норма должна быть числом", nil)
	}

	action := "варн"
	if len(fields) > 1 {
		action = strings.ToLower(fields[1])
	}

	validActions := map[string]bool{
		"варн": true,
		"бан":  true,
	}
	if !validActions[action] {
		return ctx.Reply(b, "❌ Неверное действие. Допустимые: 'варн', 'бан'", nil)
	}

	if action == "варн" {
		if err := h.service.SetNorm(ctx.StdContext(), ctx.TargetChatID(), norm); err != nil {
			return err
		}
	} else {
		if err := h.service.SetBanNorm(ctx.StdContext(), ctx.TargetChatID(), norm); err != nil {
			return err
		}
	}

	return ctx.Reply(b, view.FormatNormSet(norm, action), nil)
}

func (h *Handler) ShowNewbieThreshold(b *gotgbot.Bot, ctx *cmd.Context) error {
	threshold, err := h.service.GetNewbieThreshold(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatNewbieThreshold(threshold), nil)
}

func (h *Handler) SetNewbieThreshold(b *gotgbot.Bot, ctx *cmd.Context) error {
	days, ok := h.dateParser.ParseDuration(ctx.FirstArgument())
	if !ok {
		return ctx.Reply(b, "Не удалось распознать срок. Используйте формат: 3 дня, неделя, 14 дней или просто число.", nil)
	}

	if err := h.service.SetNewbieThreshold(ctx.StdContext(), ctx.TargetChatID(), int(days.Hours()/24)); err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatNewbieThresholdSet(int(days.Hours()/24)), nil)
}

func (h *Handler) SetPrompt(b *gotgbot.Bot, ctx *cmd.Context) error {
	if err := h.service.SetChatPrompt(ctx.StdContext(), ctx.TargetChatID(), ctx.FirstArgument()); err != nil {
		return err
	}

	return ctx.Reply(b, "Промпт установлен успешно", nil)
}

func (h *Handler) SetCommandLevel(b *gotgbot.Bot, ctx *cmd.Context) error {
	fields := strings.Fields(ctx.FirstArgument())

	overrides, _ := h.memberService.GetCommandLevels(ctx.StdContext(), ctx.TargetChatID())
	commands := h.factory.RegisteredCommands()
	if len(fields) == 0 {
		if len(fields) == 0 {
			categorySet := make(map[string]struct{})
			for _, c := range commands {
				if c.Category() != "" {
					categorySet[c.Category()] = struct{}{}
				}
			}

			if len(categorySet) == 0 {
				return ctx.Reply(b, "❌ Нет доступных категорий команд", nil)
			}

			var sb strings.Builder
			sb.WriteString("📂 <b>Доступные категории:</b>\n\n")

			for category := range categorySet {
				sb.WriteString(fmt.Sprintf("• <code>%s</code>\n", html.EscapeString(category)))
			}

			sb.WriteString("\nИспользование:\n")
			sb.WriteString("дк команда уровень\n")
			sb.WriteString("дк команда\n")
			sb.WriteString("дк категория")

			return ctx.Reply(b, sb.String(), &gotgbot.SendMessageOpts{
				ParseMode: gotgbot.ParseModeHTML,
			})
		}
	}

	arg1 := fields[0]

	// 1. Check if it's a SET operation: дк <id_команды> <уровень>
	if len(fields) >= 2 {
		levelStr := fields[1]
		level, err := strconv.Atoi(levelStr)
		if err == nil && level >= 0 && level <= 5 {
			var targetCmd *cmd.Command
			for _, c := range commands {
				if strings.EqualFold(c.ID(), arg1) {
					targetCmd = c
					break
				}
				for _, alias := range c.Aliases() {
					if strings.EqualFold(alias, arg1) {
						targetCmd = c
						break
					}
				}
				if targetCmd != nil {
					break
				}
			}

			if targetCmd == nil {
				return ctx.Reply(b, fmt.Sprintf("❌ Команда <code>%s</code> не найдена", html.EscapeString(arg1)), &gotgbot.SendMessageOpts{
					ParseMode: gotgbot.ParseModeHTML,
				})
			}

			if err := h.memberService.SetCommandLevel(ctx.StdContext(), ctx.TargetChatID(), targetCmd.ID(), int16(level)); err != nil {
				return err
			}
			return ctx.Reply(b, fmt.Sprintf("✅ Уровень доступа для команды <code>%s</code> установлен на %d", html.EscapeString(targetCmd.PrimaryAlias()), level), &gotgbot.SendMessageOpts{
				ParseMode: gotgbot.ParseModeHTML,
			})
		}
	}

	// 2. VIEW operation: дк <id_команды> or дк <категория>

	// 2.1 Check if arg1 is a command ID or alias
	var targetCmd *cmd.Command
	for _, c := range commands {
		if strings.EqualFold(c.ID(), arg1) {
			targetCmd = c
			break
		}
		for _, alias := range c.Aliases() {
			if strings.EqualFold(alias, arg1) {
				targetCmd = c
				break
			}
		}
		if targetCmd != nil {
			break
		}
	}

	if targetCmd != nil {
		lvl := targetCmd.Level()
		if overrideLvl, overridden := overrides[targetCmd.ID()]; overridden {
			lvl = overrideLvl
		}
		return ctx.Reply(b, fmt.Sprintf("📊 Уровень доступа команды <code>%s</code>: <b>%d</b>", html.EscapeString(targetCmd.PrimaryAlias()), lvl), &gotgbot.SendMessageOpts{
			ParseMode: gotgbot.ParseModeHTML,
		})
	}

	// 2.2 Check if arg1 is a category
	var matches []*cmd.Command
	for _, c := range commands {
		if strings.EqualFold(c.Category(), arg1) {
			matches = append(matches, c)
		}
	}

	if len(matches) > 0 {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("📂 <b>Уровни доступа в категории [%s]:</b>\n\n", html.EscapeString(arg1)))

		for _, c := range matches {
			lvl := c.Level()
			if overrideLvl, overridden := overrides[c.ID()]; overridden {
				lvl = overrideLvl
			}
			sb.WriteString(fmt.Sprintf("• <code>%s</code> — <b>%d</b>\n", html.EscapeString(c.PrimaryAlias()), lvl))
		}

		return ctx.Reply(b, sb.String(), &gotgbot.SendMessageOpts{
			ParseMode: gotgbot.ParseModeHTML,
		})
	}

	return ctx.Reply(b, fmt.Sprintf("❌ Команда или категория <code>%s</code> не найдена", html.EscapeString(arg1)), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
	})
}

func (h *Handler) ShowPrompt(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.service.GetChat(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatPrompt(c.AISystemPrompt), nil)
}

func (h *Handler) ShowWeekStart(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.service.GetChat(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	return ctx.ReplyHTML(b, view.FormatWeekStart(int(c.WeekStartDay), c.WeekStartTime))
}

func (h *Handler) ManageWeekStart(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.service.GetChat(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	newDay := int(c.WeekStartDay)
	newTime := c.WeekStartTime
	daySet := false
	timeSet := false

	if len(ctx.ParsedDates()) > 0 {
		d := ctx.ParsedDates()[0]
		wd := d.Weekday()
		if wd == time.Sunday {
			newDay = 7
		} else {
			newDay = int(wd)
		}
		newTime = d.Format("15:04")
		daySet = true
		timeSet = true
	}

	args := strings.Fields(ctx.FirstArgument())

	for _, arg := range args {
		if day, ok := parseWeekStartDay(arg); ok && !daySet {
			newDay = day
			daySet = true
			continue
		}

		if hour, minute, ok := parseTime(arg); ok && !timeSet {
			newTime = fmt.Sprintf("%02d:%02d", hour, minute)
			timeSet = true
			continue
		}
	}

	if !daySet && !timeSet {
		return ctx.Reply(b, "❌ Не удалось распознать день недели или время. Используйте пн/вт/... или ЧЧ:ММ.", nil)
	}

	if err := h.service.SetWeekStartDay(ctx.StdContext(), ctx.TargetChatID(), newDay); err != nil {
		return err
	}
	if err := h.service.SetWeekStartTime(ctx.StdContext(), ctx.TargetChatID(), newTime); err != nil {
		return err
	}

	return ctx.ReplyHTML(b, view.FormatWeekStartSet(newDay, newTime))
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

func (h *Handler) ShowPrefix(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.service.GetChat(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatPrefix(c.CommandPrefix), nil)
}

func (h *Handler) SetPrefix(b *gotgbot.Bot, ctx *cmd.Context) error {
	prefix := strings.TrimSpace(ctx.FirstArgument())
	if prefix == "" {
		return ctx.Reply(b, "❌ Укажите префикс", nil)
	}

	if err := h.service.SetCommandPrefix(ctx.StdContext(), ctx.TargetChatID(), prefix); err != nil {
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

func (h *Handler) Manage(b *gotgbot.Bot, ctx *cmd.Context) error {
	if ctx.EffectiveChat.Type != "private" {
		return ctx.Reply(b, "❌ Команда доступна только в ЛС бота", nil)
	}

	chatIDs, err := h.adminService.GetUserManagedChats(ctx.StdContext(), ctx.EffectiveUser.Id)
	if err != nil {
		return err
	}

	if len(chatIDs) == 0 {
		return ctx.Reply(b, "❌ У вас нет доступных чатов для управления", nil)
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
		return ctx.Reply(b, "❌ Не удалось получить информацию о ваших чатах", nil)
	}

	return ctx.Reply(b, "Выберите чат для управления:", &gotgbot.SendMessageOpts{
		ReplyMarkup: h.getManageKeyboard(chats, 1),
	})
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

func (h *Handler) CallbackManagePage(b *gotgbot.Bot, ctx *cmd.Context) error {
	data := ctx.CallbackQuery.Data
	pageStr := strings.TrimPrefix(data, "manage_page:")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return err
	}

	chatIDs, err := h.adminService.GetUserManagedChats(ctx.StdContext(), ctx.EffectiveUser.Id)
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

func (h *Handler) OnNewChatTitle(_ *gotgbot.Bot, ctx *cmd.Context) error {
	newTitle := ctx.EffectiveMessage.NewChatTitle
	if newTitle == "" {
		return nil
	}

	_, err := h.service.EnsureChatExists(ctx.StdContext(), ctx.TargetChatID(), newTitle)
	return err
}

func (h *Handler) CallbackManage(b *gotgbot.Bot, ctx *cmd.Context) error {
	data := ctx.CallbackQuery.Data
	chatIDStr := strings.TrimPrefix(data, "manage:")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return err
	}

	isAdmin, err := h.adminService.IsAdmin(ctx.StdContext(), chatID, ctx.EffectiveUser.Id)
	if err != nil || !isAdmin {
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
	title := "чатом"
	if err == nil {
		title = cht.Title
	}

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

func (h *Handler) UserChats(b *gotgbot.Bot, ctx *cmd.Context) error {
	chats, err := h.service.ListChatsWithoutNorm(
		ctx.StdContext(),
		ctx.EffectiveSender.Id(),
	)
	if err != nil {
		return err
	}

	if len(chats) == 0 {
		_, err := ctx.EffectiveChat.SendMessage(
			b,
			"✅ Все недельные нормы выполнены.",
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
		return ctx.Reply(
			b,
			"Список невыполненных норм отправлен вам в личные сообщения ✅",
			nil,
		)
	}

	return nil
}

func (h *Handler) EnableTags(b *gotgbot.Bot, ctx *cmd.Context) error {
	if err := h.service.EnableTags(ctx.StdContext(), ctx.TargetChatID(), true); err != nil {
		return err
	}

	return ctx.Reply(b, "✅Поддержка тегов в чате включена. Теперь при установке роли админка не выдается", nil)
}

func (h *Handler) DisableTags(b *gotgbot.Bot, ctx *cmd.Context) error {
	if err := h.service.EnableTags(ctx.StdContext(), ctx.TargetChatID(), false); err != nil {
		return err
	}

	return ctx.Reply(b, "❌ Поддержка тегов в чате выключена. Теперь при установке роли выдается админка с минимальными правами", nil)
}

func (h *Handler) ShowTags(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.service.GetChat(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	if c.TagsEnabled {
		return ctx.Reply(b, "✅ В чате поддерживаются теги. Это значит, что при установке роли  пользователю устанавливается телеграм-тег, а не минимальные права администратора с подписью", nil)
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
