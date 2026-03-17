package cmd

import (
	"activity-bot/internal/model"
	"context"
	"errors"
	"fmt"
	"time"

	"activity-bot/internal/member"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

var ErrNoUser = errors.New("failed to get user info from context")

type Response func(b *gotgbot.Bot, ctx *Context) error

type Context struct {
	*ext.Context
	ctx           context.Context
	args          []string
	html          string
	users         []*model.User
	targetChatID  int64
	parsedDates   []time.Time
	memberService *member.Service
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

func (c *Context) ParsedDates() []time.Time {
	return c.parsedDates
}

func (c *Context) HTML() string {
	return c.html
}

func (c *Context) SetUsers(users []*model.User) {
	c.users = users
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

func (c *Context) ResolveUserAmbiguity(b *gotgbot.Bot, callbackPrefix string, extraData string) (bool, error) {
	if len(c.users) <= 1 {
		return false, nil
	}

	var buttons [][]gotgbot.InlineKeyboardButton
	for _, u := range c.users {
		text := u.FirstName
		if c.memberService != nil && c.EffectiveChat.Type != "private" {
			m, err := c.memberService.GetChatMember(c.ctx, c.EffectiveChat.Id, u.ID)
			if err == nil && m.CustomTitle != "" {
				text = fmt.Sprintf("%s (%s)", u.FirstName, m.CustomTitle)
			}
		}

		// If extraData is needed by the caller, it's appended to the callback data.
		data := fmt.Sprintf("%s:%d", callbackPrefix, u.ID)
		if extraData != "" {
			data += ":" + extraData
		}

		btn := gotgbot.InlineKeyboardButton{
			Text:         text,
			CallbackData: data,
		}
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{btn})
	}

	kb := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}

	_, err := c.EffectiveMessage.Reply(b, "<b>Обнаружено несколько участников.</b>\nВыберите нужного:", &gotgbot.SendMessageOpts{
		ParseMode:   gotgbot.ParseModeHTML,
		ReplyMarkup: kb,
	})
	return true, err
}
