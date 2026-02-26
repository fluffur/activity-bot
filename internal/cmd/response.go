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
	ctx          context.Context
	args         []string
	html         string
	users        []*model.User
	targetChatID int64
}

func (c *Context) TargetChatID() int64 {
	return c.targetChatID
}

func (c *Context) FirstArgument() string {
	if len(c.args) > 0 {
		return c.args[0]
	}
	return ""
}

func (c *Context) SecondArgument() string {
	if len(c.args) > 1 {
		return c.args[1]
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

func (c *Context) HTML() string {
	return c.html
}

func (c *Context) StdContext() context.Context {
	return c.ctx
}

func (c *Context) Reply(b *gotgbot.Bot, text string, opts *gotgbot.SendMessageOpts) error {
	_, err := c.EffectiveMessage.Reply(b, text, opts)
	return err
}

func (c *Context) ReplyHTML(b *gotgbot.Bot, text string) error {
	_, err := c.EffectiveMessage.Reply(b, text, &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})
	return err
}
