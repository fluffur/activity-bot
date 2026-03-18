package handler

import (
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"fmt"
	"regexp"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/makeworld-the-better-one/go-isemoji"
	"github.com/rivo/uniseg"
)

var tgEmojiRegex = regexp.MustCompile(`<tg-emoji[^>]*>.*?</tg-emoji>`)

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

func (h *Handler) SetEmoji(b *gotgbot.Bot, ctx *cmd.Context) error {
	emojis := strings.TrimSpace(ctx.HTML())
	graphemes := parseEmojis(emojis)

	if len(graphemes) > 3 {
		return ctx.Reply(b, "❌ Нужно отправить не более 3 emojis", nil)
	}

	if err := h.service.SetEmoji(ctx.StdContext(), ctx.EffectiveSender.Id(), strings.Join(graphemes, "")); err != nil {
		return fmt.Errorf("failed to set emoji: %w", err)
	}
	return ctx.Reply(b, "Emoji установлено", nil)
}

func (h *Handler) RemoveEmoji(b *gotgbot.Bot, ctx *cmd.Context) error {
	if err := h.service.SetEmoji(ctx.StdContext(), ctx.EffectiveSender.Id(), ""); err != nil {
		return fmt.Errorf("failed to set emoji: %w", err)
	}
	return ctx.Reply(b, "Emoji удалено", nil)
}

func (h *Handler) ShowEmoji(b *gotgbot.Bot, ctx *cmd.Context) error {
	u := ctx.FirstUser()
	if u == nil {
		return cmd.ErrNoUser
	}

	if u.Emoji == "" {
		return ctx.ReplyHTML(b, "У пользователя еще нет emoji\n\nДобавить emoji: <code>!эмоджи 😘</code>")
	}

	return ctx.ReplyHTML(b, fmt.Sprintf("Emoji пользователя: %s", u.Emoji))

}

func parseEmojis(input string) []string {
	var result []string

	for len(input) > 0 {

		if strings.HasPrefix(input, "<tg-emoji") {
			loc := tgEmojiRegex.FindStringIndex(input)
			if loc != nil {
				result = append(result, input[loc[0]:loc[1]])
				input = input[loc[1]:]
				continue
			}
		}

		g := uniseg.NewGraphemes(input)
		if g.Next() {
			part := g.Str()

			if isemoji.IsEmoji(part) {
				result = append(result, part)
			}

			input = input[len(part):]
			continue
		}

		break
	}

	return result
}
