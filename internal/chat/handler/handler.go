package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/chat"
	"activity-bot/internal/chat/view"
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/session"
	"context"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	service        *chat.Service
	adminService   *admin.Service
	sessionService *session.Service
	dateParser     *helpers.DateParser
}

func New(service *chat.Service, adminService *admin.Service, sessionService *session.Service, dateParser *helpers.DateParser) *Handler {
	return &Handler{service, adminService, sessionService, dateParser}
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

func (h *Handler) ShowPrompt(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.service.GetChat(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatPrompt(c.AISystemPrompt), nil)
}

func (h *Handler) SetWeekStartDay(b *gotgbot.Bot, ctx *cmd.Context) error {
	arg := ctx.FirstArgument()
	if arg == "" {
		return ctx.Reply(b, "Укажите день начала недели (например: пн, ср, вс или число 1–7)", nil)
	}
	weekStartDay, ok := parseWeekStartDay(arg)
	if !ok {
		return ctx.Reply(b, "Не удалось распознать день недели. Используйте пн/вт/ср/чт/пт/сб/вс или число 1–7.", nil)
	}
	if err := h.service.SetWeekStartDay(ctx.StdContext(), ctx.TargetChatID(), weekStartDay); err != nil {
		return err
	}

	return ctx.ReplyHTML(b, "📅 Начало недели в чате изменено")
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

	type chatInfo struct {
		ID    int64
		Title string
	}
	chats := make([]chatInfo, 0)
	for _, id := range chatIDs {
		c, err := b.GetChat(id, nil)
		if err != nil {
			continue
		}
		chats = append(chats, chatInfo{ID: id, Title: c.Title})
	}

	if len(chats) == 0 {
		return ctx.Reply(b, "❌ Не удалось получить информацию о ваших чатах", nil)
	}

	var buttons [][]gotgbot.InlineKeyboardButton
	for _, c := range chats {
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: c.Title, CallbackData: "manage:" + strconv.FormatInt(c.ID, 10)},
		})
	}

	return ctx.Reply(b, "Выберите чат для управления:", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: buttons},
	})
}

func (h *Handler) CallbackManage(b *gotgbot.Bot, ctx *ext.Context) error {
	data := ctx.CallbackQuery.Data
	chatIDStr := strings.TrimPrefix(data, "manage:")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return err
	}

	isAdmin, err := h.adminService.IsAdmin(context.Background(), chatID, ctx.EffectiveUser.Id)
	if err != nil || !isAdmin {
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "У вас больше нет прав администратора в этом чате",
			ShowAlert: true,
		})
		return err
	}

	if err := h.sessionService.SetActiveChat(context.Background(), ctx.EffectiveUser.Id, chatID); err != nil {
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
