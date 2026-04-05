package command

import (
	"activity-bot/internal/model"
	"context"
	"errors"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
)

var ErrNoValueDate = errors.New("no date found")
var ErrNoValueUser = errors.New("no user found")
var ErrNoValueChat = errors.New("no chat found")
var ErrNoValue = errors.New("no value found")

type Response func(ctx *Context, u *ext.Update) error

type Context struct {
	*ext.Context
	RawArgs         string
	RawArgsHTML     string
	RawArgsEntities []tg.MessageEntityClass
	tokens          []string

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
	return c.Context.Context
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
		return time.Time{}, ErrNoValueDate
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
		return nil, ErrNoValueUser
	}
	return &c.chatMembers[0], nil
}

func (c *Context) ReplyUser() (*model.ChatMember, error) {
	if c.replyChatMember == nil {
		return nil, ErrNoValueUser
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
		return c.senderChatMember, ErrNoValueUser
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
	return nil, ErrNoValueUser
}

func (c *Context) UserOrReply() (*model.ChatMember, error) {
	if len(c.chatMembers) > 0 {
		return &c.chatMembers[0], nil
	}
	if c.replyChatMember != nil {
		return c.replyChatMember, nil
	}
	return nil, ErrNoValueUser
}

func (c *Context) RequiredStatus() model.Status {
	return c.requiredStatus
}
