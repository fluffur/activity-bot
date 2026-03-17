package cmd

import (
	"activity-bot/internal/model"
	"context"
	"errors"
	"fmt"
	"strings"
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
	members       []*model.ChatMember
	targetChatID  int64
	parsedDates   []time.Time
	memberService *member.Service
	isCallback    bool
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

func (c *Context) FirstMember() *model.ChatMember {
	if len(c.members) > 0 && c.members[0] != nil {
		return c.members[0]
	}
	return nil
}

func (c *Context) FirstUser() *model.User {
	m := c.FirstMember()
	if m != nil {
		return &m.User
	}
	return nil
}

func (c *Context) Members() []*model.ChatMember {
	return c.members
}

func (c *Context) Users() []*model.User {
	users := make([]*model.User, len(c.members))
	for i, m := range c.members {
		users[i] = &m.User
	}
	return users
}

func (c *Context) Args() []string {
	return c.args
}

func (c *Context) ArgsString() string {
	return strings.Join(c.args, " ")
}

func (c *Context) ParsedDates() []time.Time {
	return c.parsedDates
}

func (c *Context) HTML() string {
	return c.html
}

func (c *Context) SetMembers(members []*model.ChatMember) {
	c.members = members
}

func (c *Context) SetUsers(users []*model.User) {
	members := make([]*model.ChatMember, len(users))
	for i, u := range users {
		members[i] = &model.ChatMember{User: *u}
	}
	c.members = members
}

func (c *Context) StdContext() context.Context {
	return c.ctx
}

func (c *Context) IsCallback() bool {
	return c.isCallback
}

func (c *Context) SetIsCallback(v bool) {
	c.isCallback = v
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
	if len(c.members) <= 1 {
		return false, nil
	}

	var buttons [][]gotgbot.InlineKeyboardButton
	for _, m := range c.members {
		u := m.User
		text := u.FirstName
		if m.Tag != "" {
			text = fmt.Sprintf("%s (%s)", u.FirstName, m.Tag)
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
