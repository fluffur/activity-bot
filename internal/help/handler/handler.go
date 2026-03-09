package handler

import (
	"activity-bot/internal/cmd"
	"activity-bot/internal/help/view"
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Handler struct {
	ownerID int64
}

func New(ownerID int64) *Handler {
	return &Handler{ownerID}
}

func (h *Handler) Start(b *gotgbot.Bot, ctx *cmd.Context) error {
	return ctx.Reply(b, view.FormatStartMessage("https://telegra.ph/Komandy-bota-02-15-2"), &gotgbot.SendMessageOpts{
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

func (h *Handler) Help(b *gotgbot.Bot, ctx *cmd.Context) error {

	return ctx.Reply(b, view.FormatHelpText(h.ownerID, "https://telegra.ph/Komandy-bota-02-15-2"), &gotgbot.SendMessageOpts{
		ParseMode:   gotgbot.ParseModeHTML,
		ReplyMarkup: getKb(b),
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			ShowAboveText: true,
		},
	})
}

func getKb(b *gotgbot.Bot) gotgbot.InlineKeyboardMarkup {

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "Добавить бота в группу", Url: fmt.Sprintf("https://t.me/%s?startgroup=true", b.User.Username)},
			},
		},
	}
}
