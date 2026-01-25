package command

import (
	"activity-bot/internal/model"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Context struct {
	Args  []string
	Users []*model.User
}

type Response func(b *gotgbot.Bot, ctx *ext.Context, cctx *Context) error
