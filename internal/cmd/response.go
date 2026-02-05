package cmd

import (
	"activity-bot/internal/model"
	"context"
	"errors"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

var ErrNoUser = errors.New("failed to get user info from context")

type Response func(b *gotgbot.Bot, ctx *Context) error

type Context struct {
	*ext.Context
	ctx   context.Context
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

func (c *Context) StdContext() context.Context {
	return c.ctx
}
