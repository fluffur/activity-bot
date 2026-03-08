package handler

import (
	"activity-bot/internal/cmd"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

const (
	SetupStateNormWarn    = "setup_norm_warn"
	SetupStateNormBan     = "setup_norm_ban"
	SetupStateNewbie      = "setup_newbie"
	SetupStateWeekStart   = "setup_week_start"
	SetupStateMaxWarns    = "setup_max_warns"
	SetupStatePrefix      = "setup_prefix"
	SetupStatePrefixless  = "setup_prefixless"
	SetupStateMentions    = "setup_mentions"
	SetupStateWelcomeCall = "setup_welcome_call"
	SetupStateCallOnJoin  = "setup_call_on_join"
)

func setupCancelKeyboard() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "Пропустить", CallbackData: "setup_skip", Style: "primary"},
			},
			{
				{Text: "Отменить", CallbackData: "setup_cancel"},
			},
		},
	}
}

func (h *Handler) sendSetupPrompt(b *gotgbot.Bot, conversationChatID int64, text string) (*gotgbot.Message, error) {
	return b.SendMessage(conversationChatID, text, &gotgbot.SendMessageOpts{
		ReplyMarkup: setupCancelKeyboard(),
	})
}

func (h *Handler) SetupCancel(b *gotgbot.Bot, ctx *cmd.Context) error {
	if ctx.CallbackQuery != nil && ctx.CallbackQuery.Message != nil {
		_, _, _ = ctx.CallbackQuery.Message.EditText(b, "❌ Настройка отменена.", &gotgbot.EditMessageTextOpts{
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: [][]gotgbot.InlineKeyboardButton{}},
		})
		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Настройка отменена"})
	}
	return handlers.EndConversation()
}

func yesNoKeyboard() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "Да", CallbackData: "yes", Style: "success"},
				{Text: "Нет", CallbackData: "no", Style: "danger"},
			},
			{
				{Text: "Пропустить", CallbackData: "setup_skip", Style: "primary"},
			},
			{
				{Text: "Отменить", CallbackData: "setup_cancel"},
			},
		},
	}
}

func (h *Handler) sendYesNoPrompt(b *gotgbot.Bot, conversationChatID int64, text string) (*gotgbot.Message, error) {
	return b.SendMessage(conversationChatID, text, &gotgbot.SendMessageOpts{
		ReplyMarkup: yesNoKeyboard(),
	})
}

func (h *Handler) StartSetup(b *gotgbot.Bot, ctx *cmd.Context) error {
	stdCtx := context.Background()
	chatID := ctx.TargetChatID()

	isAdmin, err := h.adminService.IsAdmin(stdCtx, chatID, ctx.EffectiveSender.Id())
	if err != nil {
		return err
	}
	if !isAdmin {
		_, _ = ctx.EffectiveMessage.Reply(b, "Требуются права администратора.", nil)
		return handlers.EndConversation()
	}

	conversationChatID := ctx.EffectiveChat.Id

	msg, err := h.sendSetupPrompt(b, conversationChatID, "Начинаем опрос! Если вы пока не уверены в каком-либо вопросе, нажимайте пропустить. Начнём с нормы.\n\nУкажите норму на варн")
	if err != nil {
		return err
	}
	_ = msg
	return handlers.NextConversationState(SetupStateNormWarn)
}

func (h *Handler) HandleSetupNormWarn(b *gotgbot.Bot, ctx *cmd.Context) error {
	chatID := ctx.TargetChatID()
	text := strings.TrimSpace(ctx.EffectiveMessage.Text)
	n, err := strconv.Atoi(text)
	if err != nil || n < 0 {
		_, _ = ctx.EffectiveMessage.Reply(b, "❌ Введите неотрицательное число (норма сообщений в неделю).", nil)
		return handlers.NextConversationState(SetupStateNormWarn)
	}
	if err := h.service.SetNorm(context.Background(), chatID, n); err != nil {
		return err
	}
	msg, err := h.sendSetupPrompt(b, ctx.EffectiveChat.Id, "Укажите норму на бан")
	if err != nil {
		return err
	}
	_ = msg
	return handlers.NextConversationState(SetupStateNormBan)
}

func (h *Handler) HandleSetupNormBan(b *gotgbot.Bot, ctx *cmd.Context) error {
	chatID := ctx.TargetChatID()
	text := strings.TrimSpace(strings.ToLower(ctx.EffectiveMessage.Text))
	if text == "пропустить" || text == "пропуск" || text == "skip" || text == "-" {
		if err := h.service.SetBanNorm(context.Background(), chatID, 0); err != nil {
			return err
		}
	} else {
		n, err := strconv.Atoi(strings.TrimSpace(ctx.EffectiveMessage.Text))
		if err != nil || n < 0 {
			_, _ = ctx.EffectiveMessage.Reply(b, "❌ Введите число или «пропустить».", nil)
			return handlers.NextConversationState(SetupStateNormBan)
		}
		if err := h.service.SetBanNorm(context.Background(), chatID, n); err != nil {
			return err
		}
	}
	msg, err := h.sendSetupPrompt(b, ctx.EffectiveChat.Id, "Укажите количество дней, спустя которые участники являются новичками")
	if err != nil {
		return err
	}
	_ = msg
	return handlers.NextConversationState(SetupStateNewbie)
}

func (h *Handler) HandleSetupNewbie(b *gotgbot.Bot, ctx *cmd.Context) error {
	chatID := ctx.TargetChatID()
	text := strings.TrimSpace(ctx.EffectiveMessage.Text)
	dur, ok := h.dateParser.ParseDuration(text)
	if !ok {
		if n, err := strconv.Atoi(text); err == nil && n > 0 && n <= 365 {
			dur = time.Duration(n) * 24 * time.Hour
			ok = true
		}
	}
	if !ok || dur <= 0 {
		_, _ = ctx.EffectiveMessage.Reply(b, "❌ Не удалось распознать срок. Введите число дней или, например: 3 дня, неделя.", nil)
		return handlers.NextConversationState(SetupStateNewbie)
	}
	days := int(dur.Hours() / 24)
	if days < 1 {
		days = 1
	}
	if err := h.service.SetNewbieThreshold(context.Background(), chatID, days); err != nil {
		return err
	}
	msg, err := h.sendSetupPrompt(b, ctx.EffectiveChat.Id, "Начало недели (день и время, например: пн 00:00 или вт 12:00):")
	if err != nil {
		return err
	}
	_ = msg
	return handlers.NextConversationState(SetupStateWeekStart)
}

func (h *Handler) HandleSetupWeekStart(b *gotgbot.Bot, ctx *cmd.Context) error {
	chatID := ctx.TargetChatID()

	// Получаем текущие значения недели и времени
	c, err := h.service.GetChat(ctx.StdContext(), chatID)
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

	args := strings.Fields(strings.TrimSpace(ctx.EffectiveMessage.Text))
	for _, arg := range args {
		if day, ok := parseWeekStartDay(arg); ok && !daySet {
			newDay = day
			daySet = true
		}

		if hour, minute, ok := parseTime(arg); ok && !timeSet {
			newTime = fmt.Sprintf("%02d:%02d", hour, minute)
			timeSet = true
		}
	}

	if !daySet && !timeSet {
		_, _ = ctx.EffectiveMessage.Reply(b, "❌ Не удалось распознать день недели или время. Используйте пн/вт/... или ЧЧ:ММ.", nil)
		return handlers.NextConversationState(SetupStateWeekStart)
	}

	if err := h.service.SetWeekStartDay(ctx.StdContext(), chatID, newDay); err != nil {
		return err
	}
	if err := h.service.SetWeekStartTime(ctx.StdContext(), chatID, newTime); err != nil {
		return err
	}

	msg, err := h.sendSetupPrompt(b, chatID, "Укажите максимум предупреждений перед наказанием (по умолчанию 3):")
	if err != nil {
		return err
	}
	_ = msg
	return handlers.NextConversationState(SetupStateMaxWarns)
}
func (h *Handler) HandleSetupMaxWarns(b *gotgbot.Bot, ctx *cmd.Context) error {
	chatID := ctx.TargetChatID()
	n, err := strconv.Atoi(strings.TrimSpace(ctx.EffectiveMessage.Text))
	if err != nil || n < 0 {
		_, _ = ctx.EffectiveMessage.Reply(b, "❌ Введите неотрицательное число.", nil)
		return handlers.NextConversationState(SetupStateMaxWarns)
	}
	if err := h.adminService.SetMaxWarns(context.Background(), chatID, n); err != nil {
		return err
	}
	msg, err := h.sendSetupPrompt(b, ctx.EffectiveChat.Id, "Укажите кастомный префикс бота (например <code>ботик</code>)")
	if err != nil {
		return err
	}
	_ = msg
	return handlers.NextConversationState(SetupStatePrefix)
}

func (h *Handler) HandleSetupPrefix(b *gotgbot.Bot, ctx *cmd.Context) error {
	chatID := ctx.TargetChatID()
	text := strings.TrimSpace(ctx.EffectiveMessage.Text)
	if err := h.service.SetCommandPrefix(context.Background(), chatID, text); err != nil {
		return err
	}
	msg, err := h.sendSetupPrompt(b, ctx.EffectiveChat.Id, "Разрешить команды без префикса?")
	if err != nil {
		return err
	}
	_ = msg
	return handlers.NextConversationState(SetupStatePrefixless)
}

func (h *Handler) HandleSetupPrefixless(b *gotgbot.Bot, ctx *cmd.Context) error {
	chatID := ctx.TargetChatID()
	t := strings.TrimSpace(strings.ToLower(ctx.EffectiveMessage.Text))
	var allow bool
	switch t {
	case "да", "yes", "1", "true", "включить":
		allow = true
	case "нет", "no", "0", "false", "выключить":
		allow = false
	default:
		_, _ = ctx.EffectiveMessage.Reply(b, "❌ Введите «да» или «нет».", nil)
		return handlers.NextConversationState(SetupStatePrefixless)
	}
	if err := h.service.SetAllowPrefixless(context.Background(), chatID, allow); err != nil {
		return err
	}
	msg, err := h.sendSetupPrompt(b, ctx.EffectiveChat.Id, "Лимит упоминаний в одном сообщении созыва (1–100):")
	if err != nil {
		return err
	}
	_ = msg
	return handlers.NextConversationState(SetupStateMentions)
}

func (h *Handler) HandleSetupMentions(b *gotgbot.Bot, ctx *cmd.Context) error {
	chatID := ctx.TargetChatID()
	n, err := strconv.Atoi(strings.TrimSpace(ctx.EffectiveMessage.Text))
	if err != nil || n < 1 || n > 100 {
		_, _ = ctx.EffectiveMessage.Reply(b, "❌ Введите число от 1 до 100.", nil)
		return handlers.NextConversationState(SetupStateMentions)
	}
	if err := h.service.SetMentionsPerMessage(context.Background(), chatID, int32(n)); err != nil {
		return err
	}
	msg, err := h.sendSetupPrompt(b, ctx.EffectiveChat.Id, "Сообщение созыва по умолчанию (текст или «пропустить»):")
	if err != nil {
		return err
	}
	_ = msg
	return handlers.NextConversationState(SetupStateWelcomeCall)
}

func (h *Handler) HandleSetupWelcomeCall(b *gotgbot.Bot, ctx *cmd.Context) error {
	chatID := ctx.TargetChatID()
	text := strings.TrimSpace(ctx.EffectiveMessage.Text)
	if strings.ToLower(text) == "пропустить" || strings.ToLower(text) == "пропуск" || text == "skip" || text == "-" {
		text = ""
	}
	if err := h.service.SetWelcomeCallMessage(context.Background(), chatID, text); err != nil {
		return err
	}
	msg, err := h.sendSetupPrompt(b, ctx.EffectiveChat.Id, "Включить созыв при входе нового участника? (да/нет):")
	if err != nil {
		return err
	}
	_ = msg
	return handlers.NextConversationState(SetupStateCallOnJoin)
}

func (h *Handler) HandleSetupCallOnJoin(b *gotgbot.Bot, ctx *cmd.Context) error {
	chatID := ctx.TargetChatID()
	t := strings.TrimSpace(strings.ToLower(ctx.EffectiveMessage.Text))
	var enable bool
	switch t {
	case "да", "yes", "1", "true", "включить":
		enable = true
	case "нет", "no", "0", "false", "выключить":
		enable = false
	default:
		_, _ = ctx.EffectiveMessage.Reply(b, "❌ Введите «да» или «нет».", nil)
		return handlers.NextConversationState(SetupStateCallOnJoin)
	}
	if err := h.service.UpdateCallOnJoin(context.Background(), chatID, enable); err != nil {
		return err
	}
	_, _ = ctx.EffectiveMessage.Reply(b, "✅ Настройка чата завершена.", nil)
	return handlers.EndConversation()
}

func (h *Handler) SkipSetupNormWarnCallback(b *gotgbot.Bot, ctx *cmd.Context) error {
	_, _, err := ctx.EffectiveMessage.EditText(b, "Укажите норму на бан", &gotgbot.EditMessageTextOpts{
		ReplyMarkup: setupCancelKeyboard(),
	})
	if err != nil {
		return err
	}
	return handlers.NextConversationState(SetupStateNormBan)
}

func (h *Handler) SkipSetupNormBanCallback(b *gotgbot.Bot, ctx *cmd.Context) error {
	_, _, err := ctx.EffectiveMessage.EditText(b, "Укажите количество дней, спустя которые участники являются новичками", &gotgbot.EditMessageTextOpts{
		ReplyMarkup: setupCancelKeyboard(),
	})
	if err != nil {
		return err
	}
	return handlers.NextConversationState()
}

func (h *Handler) HandleSkip(state string) func(b *gotgbot.Bot, ctx *cmd.Context) error {
	return func(b *gotgbot.Bot, ctx *cmd.Context) error {
		if _, err := ctx.CallbackQuery.Answer(b, nil); err != nil {
			return err
		}
		return handlers.NextConversationState(state)
	}
}
