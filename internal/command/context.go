package command

import (
	"activity-bot/internal/model"
	"activity-bot/internal/options"
	"context"
	"errors"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/tg"
)

var ErrNoValueDate = errors.New("no date found")
var ErrNoValueUser = errors.New("no user found")
var ErrNoValueChat = errors.New("no chat found")
var ErrNoValue = errors.New("no value found")

type Response func(ctx *Context, u *ext.Update) error

type Context struct {
	*ext.Context
	Command         *Command
	RawArgs         string
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

type UserStrategy int

const (
	UserFromArgs UserStrategy = iota
	UserFromReply
	UserFromSender
)

type UserFilter int

const (
	AllowBots UserFilter = iota
	OnlyHumans
)

func (c *Context) Chat() (model.Chat, error) {
	chat := c.chat
	if chat == nil {
		return model.Chat{}, ErrNoValueChat
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
	return c.ResolveUser(OnlyHumans, UserFromArgs, UserFromReply)
}

func (c *Context) AnyUser() (*model.ChatMember, error) {
	return c.ResolveUser(OnlyHumans, UserFromArgs, UserFromReply, UserFromSender)
}

func (c *Context) Sender() (*model.ChatMember, error) {
	return c.ResolveUser(OnlyHumans, UserFromSender)
}

func (c *Context) RequiredStatus() model.Status {
	return c.requiredStatus
}

func (c *Context) Reply(u *ext.Update, opts ...options.SendMessageOption) (*types.Message, error) {
	var msgID int

	switch {
	case u.EffectiveMessage != nil:
		msgID = u.EffectiveMessage.GetID()
	case u.CallbackQuery != nil:
		msgID = u.CallbackQuery.GetMsgID()
	default:
		return nil, errors.New("no effective message")
	}

	req := &tg.MessagesSendMessageRequest{}
	req.SetNoWebpage(true)
	req.ReplyTo = &tg.InputReplyToMessage{ReplyToMsgID: msgID}

	for _, opt := range opts {
		opt(req)
	}

	return c.SendMessage(u.EffectiveChat().GetID(), req)
}

func (c *Context) ResolveUser(filter UserFilter, strategies ...UserStrategy) (*model.ChatMember, error) {
	for _, s := range strategies {
		var u *model.ChatMember

		switch s {
		case UserFromArgs:
			if len(c.chatMembers) > 0 {
				u = &c.chatMembers[0]
			}
		case UserFromReply:
			u = c.replyChatMember
		case UserFromSender:
			u = c.senderChatMember
		}

		if u == nil {
			continue
		}

		if filter == OnlyHumans && u.User.IsBot {
			continue
		}

		return u, nil
	}

	return nil, ErrNoValueUser
}
func (c *Context) ReplyOnly(u *ext.Update, opts ...options.SendMessageOption) error {
	_, err := c.Reply(u, opts...)
	return err
}
