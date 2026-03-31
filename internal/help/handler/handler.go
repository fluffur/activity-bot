package handler

import (
	"activity-bot/internal/command"
	"activity-bot/internal/help/view"
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Handler struct {
	ownerUsername string
	commandsLink  string
}

func New(ownerUsername string, commandsLink string) *Handler {
	return &Handler{ownerUsername, commandsLink}
}

func (h *Handler) Start(b *gotgbot.Bot, ctx *command.Context) error {
	return ctx.Reply(b, view.FormatStartMessage(h.commandsLink), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{
					{Text: "Добавить бота в группу", Url: fmt.Sprintf("https://t.me/%s?startgroup=true", b.User.Username), Style: "primary"},
				},
			},
		},
	})
}

func (h *Handler) Help(b *gotgbot.Bot, ctx *command.Context) error {
	return ctx.Reply(b, view.FormatHelpText(h.ownerUsername, h.commandsLink), &gotgbot.SendMessageOpts{
		ParseMode:   gotgbot.ParseModeHTML,
		ReplyMarkup: getKb(b),
	})
}

func getKb(b *gotgbot.Bot) gotgbot.InlineKeyboardMarkup {

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "Добавить бота в группу", Url: fmt.Sprintf("https://t.me/%s?startgroup=true", b.User.Username), IconCustomEmojiId: "5289906211104247909"},
			},
		},
	}
}
