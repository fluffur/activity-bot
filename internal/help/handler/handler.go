package handler

import (
	"activity-bot/internal/command"
	"activity-bot/internal/options"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
)

type Handler struct {
	ownerUsername string
	commandsLink  string
}

func New(ownerUsername string, commandsLink string) *Handler {
	return &Handler{ownerUsername, commandsLink}
}

func (h *Handler) Start(ctx *command.Context, u *ext.Update) error {
	eu := u.EffectiveUser()
	eb := &entity.Builder{}
	eb.Plain("👋 Привет, ")
	eb.MentionName(eu.Username, eu.AsInput())
	eb.Plain("!\n")
	eb.Plain("Я чат-менеджер. Считаю сообщения и помогаю контролировать еженедельную активность\n\n")

	eb.Plain("Добавь меня в группу или ")
	eb.TextURL("открой список команд", h.commandsLink)

	return ctx.ReplyOnly(u, options.WithBuilder(eb), options.WithMarkup(getKb(ctx.Self.Username)))
}

func (h *Handler) Help(ctx *command.Context, u *ext.Update) error {
	eb := &entity.Builder{}
	eb.Plain("Помощь\n\n")

	eb.TextURL("* Команды бота\n", h.commandsLink)
	eb.TextURL("* Написать разработчику\n", "https://t.me/"+h.ownerUsername)

	return ctx.ReplyOnly(u, options.WithBuilder(eb), options.WithMarkup(getKb(ctx.Self.Username)))
}

func getKb(botUsername string) *tg.ReplyInlineMarkup {
	return &tg.ReplyInlineMarkup{
		Rows: []tg.KeyboardButtonRow{
			{Buttons: []tg.KeyboardButtonClass{
				&tg.KeyboardButtonURL{
					Text: "Добавить бота в чат",
					URL:  "https://t.me/" + botUsername + "?startgroup=true",
				},
			}},
		},
	}
}
