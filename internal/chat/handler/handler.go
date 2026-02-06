package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"fmt"
	"log"
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
	norm, err := h.service.GetNorm(ctx.StdContext(), ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Норма чата: %d сообщений", norm), nil)

	return err
}

func (h *Handler) SetNorm(b *gotgbot.Bot, ctx *cmd.Context) error {
	norm, err := strconv.Atoi(ctx.FirstArgument())
	if err != nil {
		return err
	}

	if err := h.service.SetNorm(ctx.StdContext(), ctx.EffectiveChat.Id, norm); err != nil {
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Установлена новая норма чата: %d", norm), nil)

	return err
}

func (h *Handler) ShowNewbieThreshold(b *gotgbot.Bot, ctx *cmd.Context) error {
	threshold, err := h.service.GetNewbieThreshold(ctx.StdContext(), ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Пользователи считаются новичками первые %d %s", threshold, helpers.PluralizeDays(threshold)), nil)
	return err
}

func (h *Handler) SetNewbieThreshold(b *gotgbot.Bot, ctx *cmd.Context) error {
	days, ok := h.dateParser.ParseDays(ctx.FirstArgument())
	if !ok {
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось распознать срок. Используйте формат: 3 дня, неделя, 14 дней или просто число.", nil)
		return err
	}

	if err := h.service.SetNewbieThreshold(ctx.StdContext(), ctx.EffectiveChat.Id, days); err != nil {
		return err
	}

	_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Теперь пользователи считаются новичками первые %d %s", days, helpers.PluralizeDays(days)), nil)
	return err
}

func (h *Handler) SetPrompt(b *gotgbot.Bot, ctx *cmd.Context) error {
	if err := h.service.SetChatPrompt(ctx.StdContext(), ctx.EffectiveChat.Id, ctx.FirstArgument()); err != nil {
		return err
	}

	_, err := ctx.EffectiveMessage.Reply(b, "Промпт установлен успешно", nil)
	return err
}

func (h *Handler) ShowPrompt(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.service.GetChat(ctx.StdContext(), ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Промпт: \"%s\"", c.AISystemPrompt), nil)
	return err
}

func (h *Handler) SetWeekStartDay(b *gotgbot.Bot, ctx *cmd.Context) error {
	log.Println(ctx.EffectiveMessage.GetText())
	arg := ctx.FirstArgument()
	if arg == "" {
		_, err := ctx.EffectiveMessage.Reply(
			b,
			"Укажите день начала недели (например: пн, ср, вс или число 1–7)",
			nil,
		)
		return err
	}
	weekStartDay, ok := parseWeekStartDay(arg)
	if !ok {
		_, err := ctx.EffectiveMessage.Reply(
			b,
			"Не удалось распознать день недели. Используйте пн/вт/ср/чт/пт/сб/вс или число 1–7.",
			nil,
		)
		return err
	}
	if err := h.service.SetWeekStartDay(ctx.StdContext(), ctx.EffectiveChat.Id, weekStartDay); err != nil {
		return err
	}

	_, err := ctx.EffectiveMessage.Reply(
		b,
		"📅 Начало недели в чате изменено",
		&gotgbot.SendMessageOpts{ParseMode: gotgbot.ParseModeHTML},
	)

	return err
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
