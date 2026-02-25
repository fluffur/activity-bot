package handler

import (
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"fmt"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Handler struct {
	service *user.Service
}

func New(service *user.Service) *Handler {
	return &Handler{service}
}

func (h *Handler) SetGender(b *gotgbot.Bot, ctx *cmd.Context) error {
	arg := strings.ToLower(ctx.FirstArgument())
	if arg == "" {
		return ctx.Reply(b, "Укажите пол: м (мужской) или ж (женский)", nil)
	}

	var gender string
	switch arg {
	case "м", "муж", "мужской", "male", "m":
		gender = model.GenderMale
	case "ж", "жен", "женский", "female", "f":
		gender = model.GenderFemale
	default:
		return ctx.Reply(b, "Неизвестный пол. Используйте: м или ж", nil)
	}

	if err := h.service.SetGender(ctx.StdContext(), ctx.EffectiveSender.Id(), gender); err != nil {
		_ = ctx.Reply(b, "Не удалось установить пол", nil)
		return err
	}

	genderName := "мужской"
	if gender == model.GenderFemale {
		genderName = "женский"
	}

	return ctx.Reply(b, fmt.Sprintf("Пол установлен: %s", genderName), nil)
}

func (h *Handler) ShowGender(b *gotgbot.Bot, ctx *cmd.Context) error {
	u := ctx.FirstUser()
	if u == nil {
		return cmd.ErrNoUser
	}
	gender := "неизвестен"
	if u.Gender == model.GenderFemale {
		gender = "женский"
	} else if u.Gender == model.GenderMale {
		gender = "мужской"
	}

	return ctx.ReplyHTML(b, fmt.Sprintf("Пол %s: %s", helpers.LinkWithContent(*u, u.FirstName), gender))
}
