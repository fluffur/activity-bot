package cmd

import (
	"activity-bot/internal/model"
	"errors"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

var ErrNoUser = errors.New("failed to get user info from context")

type Response func(b *gotgbot.Bot, ctx *ext.Context, cctx *Context) error

type Context struct {
	args  []string
	users []*model.User
}

func (c *Context) FirstArgument() string {
	if len(c.args) > 0 {
		return c.args[0]
	}
	return ""
}

func (c *Context) FirstUser() *model.User {
	if len(c.users) > 0 && c.users[0] != nil {
		return c.users[0]
	}

	return nil
}

func (c *Context) Users() []*model.User {
	return c.users
}

func (c *Context) Args() []string {
	return c.args
}
