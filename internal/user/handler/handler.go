package handler

import (
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"activity-bot/internal/options"
	"activity-bot/internal/user"
	"fmt"
	"log"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/entity"
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

	eb := &entity.Builder{}
	eb.Plain("Пол ")
	helpers.WriteUserMention(eb, cm.User)
	eb.Plain(fmt.Sprintf(": %s", gender))
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) SetEmoji(ctx *command.Context, u *ext.Update) error {
	cm, err := ctx.AnyUser()
	if err != nil {
		return err
	}
	if cm.User.ID != u.EffectiveUser().GetID() {
		return ctx.ReplyOnly(u, options.WithText("Нельзя менять эмоджи других пользователей. Для эмоджи в рамках чата есть значки. Попробуйте через команду значок"))
	}
	emojis := helpers.ExtractEmoji(ctx.RawArgs, ctx.RawArgsEntities)
	if len(emojis) > 3 {
		return ctx.ReplyOnly(u, options.WithText("❌ Нужно отправить не более 3 эмоджи на пользователя"))
	}

	if err := h.service.SetEmoji(ctx.StdContext(), u.EffectiveUser().GetID(), emojis); err != nil {
		return fmt.Errorf("failed to set emoji: %w", err)
	}

	eb := &entity.Builder{}
	eb.Plain("Эмоджи ")
	helpers.DisplayEmoji(eb, emojis)
	eb.Plain(" установлено для ")
	helpers.WriteUserMention(eb, cm.User)

	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) RemoveEmoji(ctx *command.Context, u *ext.Update) error {
	member, err := ctx.Sender()
	if err != nil {
		return err
	}
	if len(member.User.Emojis) == 0 {
		return ctx.ReplyOnly(u, options.WithText("У вас не установлен эмоджи"))

	}

	if err := h.service.SetEmoji(ctx.StdContext(), u.EffectiveUser().GetID(), nil); err != nil {
		return fmt.Errorf("failed to set emoji: %w", err)
	}

	eb := &entity.Builder{}
	eb.Plain("Эмоджи ")
	helpers.DisplayEmoji(eb, member.User.Emojis)
	eb.Plain(" удалено")
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) ShowEmoji(ctx *command.Context, u *ext.Update) error {
	cm, err := ctx.AnyUser()
	if err != nil {
		return err
	}

	eb := &entity.Builder{}
	log.Println(cm.Emojis, cm.User.Emojis)
	if len(cm.User.Emojis) == 0 {
		eb.Plain("У ")
		helpers.WriteUserMention(eb, cm.User)
		eb.Plain(" ещё не установлен эмоджи\n\nКоманда: ")
		eb.Code("эмоджи 💤")
		return ctx.ReplyOnly(u, options.WithBuilder(eb))
	}

	eb.Plain("Эмоджи ")
	helpers.WriteUserMention(eb, cm.User)
	eb.Plain(": ")
	helpers.DisplayEmoji(eb, cm.User.Emojis)
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}
