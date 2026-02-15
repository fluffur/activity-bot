package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/chat"
	"activity-bot/internal/chat/view"
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Handler struct {
	service      *chat.Service
	adminService *admin.Service
	dateParser   *helpers.DateParser
}

func New(service *chat.Service, adminService *admin.Service, dateParser *helpers.DateParser) *Handler {
	return &Handler{service, adminService, dateParser}
}

func (h *Handler) ShowNorm(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.service.GetChat(ctx.StdContext(), ctx.EffectiveChat.Id)
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
		if err := h.service.SetNorm(ctx.StdContext(), ctx.EffectiveChat.Id, norm); err != nil {
			return err
		}
	} else {
		if err := h.service.SetBanNorm(ctx.StdContext(), ctx.EffectiveChat.Id, norm); err != nil {
			return err
		}
	}

	return ctx.Reply(b, view.FormatNormSet(norm, action), nil)
}

func (h *Handler) ShowNewbieThreshold(b *gotgbot.Bot, ctx *cmd.Context) error {
	threshold, err := h.service.GetNewbieThreshold(ctx.StdContext(), ctx.EffectiveChat.Id)
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

	if err := h.service.SetNewbieThreshold(ctx.StdContext(), ctx.EffectiveChat.Id, int(days.Hours()/24)); err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatNewbieThresholdSet(int(days.Hours()/24)), nil)
}

func (h *Handler) SetPrompt(b *gotgbot.Bot, ctx *cmd.Context) error {
	if err := h.service.SetChatPrompt(ctx.StdContext(), ctx.EffectiveChat.Id, ctx.FirstArgument()); err != nil {
		return err
	}

	return ctx.Reply(b, "Промпт установлен успешно", nil)
}

func (h *Handler) ShowPrompt(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.service.GetChat(ctx.StdContext(), ctx.EffectiveChat.Id)
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
	if err := h.service.SetWeekStartDay(ctx.StdContext(), ctx.EffectiveChat.Id, weekStartDay); err != nil {
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
