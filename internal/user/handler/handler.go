package handler

import (
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/logger"
	"activity-bot/internal/model"
	"activity-bot/internal/options"
	"activity-bot/internal/user"
	"fmt"
	"strings"

	"github.com/celestix/gotgproto/ext"
)

type Handler struct {
	service *user.Service
}

func New(service *user.Service) *Handler {
	return &Handler{service}
}

func (h *Handler) SetGender(ctx *command.Context, u *ext.Update) error {
	gender := ctx.TextOrDefault("м")

	switch gender {
	case "м", "муж", "мужской", "male", "m":
		gender = model.GenderMale
	case "ж", "жен", "женский", "female", "f":
		gender = model.GenderFemale
	default:
		return ctx.ReplyOnly(u, options.WithText("Неизвестный пол. Используйте: м или ж"))
	}

	if err := h.service.SetGender(ctx.StdContext(), u.EffectiveUser().GetID(), gender); err != nil {
		_ = ctx.ReplyOnly(u, options.WithText("Не удалось установить пол"))
		return err
	}

	genderName := "мужской"
	if gender == model.GenderFemale {
		genderName = "женский"
	}

	return ctx.ReplyOnly(u, options.WithText(fmt.Sprintf("Пол установлен: %s", genderName)))
}

func (h *Handler) ShowGender(ctx *command.Context, u *ext.Update) error {
	cm, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	gender := "неизвестен"
	if cm.User.Gender == model.GenderFemale {
		gender = "женский"
	} else if cm.User.Gender == model.GenderMale {
		gender = "мужской"
	}

	return ctx.ReplyOnly(u, options.WithText(fmt.Sprintf("Пол %s: %s", helpers.LinkWithContent(cm.User, cm.User.FirstName), gender)))
}

func (h *Handler) SetEmoji(ctx *command.Context, u *ext.Update) error {
	cm, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	if cm.User.ID != u.EffectiveUser().GetID() {
		return ctx.ReplyOnly(u, options.WithText("Нельзя менять эмоджи других пользователей. Для эмоджи в рамках чата есть значки. Попробуйте через команду значок"))
	}
	emojis := strings.TrimSpace(ctx.RawArgsHTML)
	graphemes := helpers.ParseEmojis(emojis)

	if len(graphemes) > 3 {
		return ctx.ReplyOnly(u, options.WithText("❌ Нужно отправить не более 3 эмоджи на пользователя"))
	}

	if err := h.service.SetEmoji(ctx.StdContext(), u.EffectiveUser().GetID(), strings.Join(graphemes, "")); err != nil {
		return fmt.Errorf("failed to set emoji: %w", err)
	}
	logger.L.Info("set emoji", "emoji", emojis)
	return ctx.ReplyOnly(u, options.WithText(fmt.Sprintf("Эмоджи %s установлено для %s", emojis, helpers.UserLink(cm.User))))
}

func (h *Handler) RemoveEmoji(ctx *command.Context, u *ext.Update) error {
	if err := h.service.SetEmoji(ctx.StdContext(), u.EffectiveUser().GetID(), ""); err != nil {
		return fmt.Errorf("failed to set emoji: %w", err)
	}
	return ctx.ReplyOnly(u, options.WithText("Emoji удалено"))
}

func (h *Handler) ShowEmoji(ctx *command.Context, u *ext.Update) error {
	cm, err := ctx.AnyUser()
	if err != nil {
		return err
	}

	if cm.Emoji == "" {
		return ctx.ReplyOnly(u, options.WithText(fmt.Sprintf("У пользователя %s еще нет эмоджи\n\nДобавить emoji: !эмоджи 😘", helpers.UserLink(cm.User))))
	}

	return ctx.ReplyOnly(u, options.WithText(fmt.Sprintf("Эмоджи пользователя %s: %s", helpers.UserLink(cm.User), cm.Emoji)))
}
