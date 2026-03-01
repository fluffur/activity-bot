package handler

import (
	"activity-bot/internal/cmd"
	"activity-bot/internal/help/view"
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Handler struct {
	ownerID   int64
	webappURL string
}

func New(ownerID int64, webappURL string) *Handler {
	return &Handler{ownerID, webappURL}
}

func (h *Handler) Start(b *gotgbot.Bot, ctx *cmd.Context) error {
	return ctx.Reply(b, view.FormatStartMessage(), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{
					{Text: "Добавить бота в группу", Url: fmt.Sprintf("https://t.me/%s?startgroup=true", b.User.Username), Style: "primary"},
				},
				{
					{Text: "Команды бота", Url: "https://telegra.ph/Komandy-bota-02-15-2"},
				},
				{
					{Text: "Приложение бота", WebApp: &gotgbot.WebAppInfo{Url: h.webappURL}},
				},
			},
		},
	})
}

func (h *Handler) Help(b *gotgbot.Bot, ctx *cmd.Context) error {

	return ctx.Reply(b, view.FormatHelpText(h.ownerID), &gotgbot.SendMessageOpts{
		ParseMode:   gotgbot.ParseModeHTML,
		ReplyMarkup: getKb(b),
	})
}

func getKb(b *gotgbot.Bot) gotgbot.InlineKeyboardMarkup {

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "Команды бота", Url: "https://telegra.ph/Komandy-bota-02-15-2", Style: "primary"},
			},
			{
				{Text: "Добавить бота в группу", Url: fmt.Sprintf("https://t.me/%s?startgroup=true", b.User.Username)},
			},
		},
	}
}
