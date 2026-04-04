package handler

import (
	"activity-bot/internal/command"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

type Handler struct {
	ownerUsername string
	commandsLink  string
}

func New(ownerUsername string, commandsLink string) *Handler {
	return &Handler{ownerUsername, commandsLink}
}

func (h *Handler) Start(ctx *command.Context, upd *ext.Update) error {
	upd.EffectiveMessage.GetReplies()
	u := upd.EffectiveUser()
	_, err := ctx.Reply(upd, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		eb.Plain("👋 Привет, ")
		eb.MentionName(u.Username, u.AsInput())
		eb.Plain("!\n")
		eb.Plain("Я чат-менеджер. Считаю сообщения и помогаю контролировать еженедельную активность\n\n")

		eb.Plain("Добавь меня в группу или ")
		eb.TextURL("открой список команд", h.commandsLink)

		return nil
	})), &ext.ReplyOpts{
		Markup: getKb(ctx.Self.Username),
	})
	return err
}

func (h *Handler) Help(ctx *command.Context, upd *ext.Update) error {
	_, err := ctx.Reply(upd, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		eb.Plain("Помощь\n\n")

		eb.TextURL("* Команды бота\n", h.commandsLink)
		eb.TextURL("* Написать разработчику\n", "https://t.me/"+h.ownerUsername)
		return nil
	})), &ext.ReplyOpts{
		Markup: getKb(ctx.Self.Username),
	})
	return err
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
