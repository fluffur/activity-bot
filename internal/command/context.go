package command

import (
	"activity-bot/internal/model"
	"context"
	"errors"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

var ErrNoValue = errors.New("no value found")

type Response func(b *gotgbot.Bot, ctx *Context) error

type Context struct {
	stdContext context.Context
	*ext.Context
	RawArgs     string
	RawArgsHTML string
	tokens      []string

	usedOffsets []Offset

	chat *model.Chat

	senderChatMember *model.ChatMember

	replyChatMember *model.ChatMember

	chatMembers []model.ChatMember

	dates          []time.Time
	numbers        []int
	texts          []string
	requiredStatus model.Status
}

func (c *Context) StdContext() context.Context {
	return c.stdContext
}

func (c *Context) Reply(b *gotgbot.Bot, text string, opts *gotgbot.SendMessageOpts) error {
	_, err := c.EffectiveMessage.Reply(b, text, opts)

	return err
}

func (c *Context) ReplyHTML(b *gotgbot.Bot, text string) error {
	_, err := c.EffectiveMessage.Reply(b, text, &gotgbot.SendMessageOpts{
		ParseMode: "HTML",
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})

	return err
}

func (c *Context) Chat() (model.Chat, error) {
	chat := c.chat
	if chat == nil {
		return model.Chat{}, ErrNoValue
	}
	return *chat, nil
}

func (c *Context) Number() (int, error) {
	if len(c.numbers) == 0 {
		return 0, ErrNoValue
	}
	return c.numbers[0], nil
}

func (c *Context) NumberOrDefault(def int) int {
	if len(c.numbers) == 0 {
		return def
	}
	return c.numbers[0]
}

func (c *Context) Date() (time.Time, error) {
	if len(c.dates) == 0 {
		return time.Time{}, ErrNoValue
	}
	return c.dates[0], nil
}

func (c *Context) Dates() []time.Time {
	return c.dates
}

func (c *Context) DateOrDefault(def time.Time) time.Time {
	if len(c.dates) == 0 {
		return def
	}
	return c.dates[0]
}

func (c *Context) Text() (string, error) {
	if len(c.texts) == 0 {
		return "", ErrNoValue
	}
	return c.texts[0], nil
}

func (c *Context) TextOrDefault(def string) string {
	if len(c.texts) == 0 {
		return def
	}
	return c.texts[0]
}

func (c *Context) User() (*model.ChatMember, error) {
	if len(c.chatMembers) == 0 {
		return nil, ErrNoValue
	}
	return &c.chatMembers[0], nil
}

func (c *Context) ReplyUser() (*model.ChatMember, error) {
	if c.replyChatMember == nil {
		return nil, ErrNoValue
	}
	return c.replyChatMember, nil
}

func (c *Context) AnyUser() (*model.ChatMember, error) {
	if len(c.chatMembers) > 0 {
		return &c.chatMembers[0], nil
	}
	if c.replyChatMember != nil {
		return c.replyChatMember, nil
	}
	if c.senderChatMember != nil {
		return c.senderChatMember, nil
	}
	return nil, ErrNoValue

}

func (c *Context) Sender() (*model.ChatMember, error) {
	if c.senderChatMember == nil {
		return c.senderChatMember, ErrNoValue
	}
	return c.senderChatMember, nil
}

func (c *Context) UserOrSender() (*model.ChatMember, error) {
	if len(c.chatMembers) > 0 {
		return &c.chatMembers[0], nil
	}
	if c.senderChatMember != nil {
		return c.senderChatMember, nil
	}
	return nil, ErrNoValue
}

func (c *Context) UserOrReply() (*model.ChatMember, error) {
	if len(c.chatMembers) > 0 {
		return &c.chatMembers[0], nil
	}
	if c.replyChatMember != nil {
		return c.replyChatMember, nil
	}
	return nil, ErrNoValue
}

func (c *Context) RequiredStatus() model.Status {
	return c.requiredStatus
}
