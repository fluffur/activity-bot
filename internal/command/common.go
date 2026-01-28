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

func (c *Context) FirstArgument() string {
	if len(c.Args) > 0 {
		return c.Args[0]
	}
	return ""
}

func (c *Context) FirstUser() *model.User {
	if len(c.Users) > 0 && c.Users[0] != nil {
		return c.Users[0]
	}

	return nil
}

type Response func(b *gotgbot.Bot, ctx *ext.Context, cctx *Context) error
