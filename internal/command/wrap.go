package command

import (
	"fmt"
	"strings"

	"github.com/celestix/gotgproto/ext"
	"github.com/go-faster/errors"
	"github.com/gotd/td/tg"
)

type HandlerFunc func(ctx *ext.Context, u *ext.Update) error

func (f HandlerFunc) CheckUpdate(ctx *ext.Context, u *ext.Update) error {
	return f(ctx, u)
}

func (c *Command) WrapCallback() HandlerFunc {
	return func(ctx *ext.Context, u *ext.Update) error {
		cb := u.CallbackQuery
		if cb == nil {
			return nil
		}

		handlerCtx := Context{
			Context: ctx,
		}

		if c.scope == ScopeChat {
			chat, err := c.getChat(ctx, u)
			if err != nil {
				return errors.Wrap(err, "callback: get chat failed")
			}
			handlerCtx.chat = &chat

			if s, err := c.chatProvider.GetCommandPermission(ctx.Context, chat.ID, c.name); err == nil {
				c.requiredStatus = s
			}
			handlerCtx.requiredStatus = c.requiredStatus
		}

		senderID := u.EffectiveUser().GetID()
		member, err := c.resolveMember(ctx.Context, handlerCtx.chat, senderID)
		if err != nil {
			return errors.Wrap(err, "callback: resolve sender failed")
		}
		handlerCtx.senderChatMember = member

		if !handlerCtx.senderChatMember.StatusGranted(c.requiredStatus) {
			_, err := ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
				Message: fmt.Sprintf("Требуются права: %s", c.requiredStatus),
				Alert:   true,
			})
			return err
		}

		data := string(cb.Data)
		if idx := strings.Index(data, ":"); idx != -1 {
			handlerCtx.RawArgs = data[idx+1:]
		} else {
			handlerCtx.RawArgs = data
		}

		handlerCtx.texts = strings.Fields(handlerCtx.RawArgs)

		err = c.response(&handlerCtx, u)

		_, _ = ctx.AnswerCallback(nil)

		return err
	}
}

//func (c *Command) WrapEvent() func(b *gotgbot.Bot, ctx *ext.Context) error {
//	return func(b *gotgbot.Bot, ctx *ext.Context) error {
//		stdCtx := context.Background()
//		msg := ctx.Message
//		if msg == nil {
//			return nil
//		}
//
//		handlerCtx := &Context{Context: ctx, stdContext: stdCtx}
//
//		if c.scope == ScopeChat {
//			chat, err := c.getChat(ctx, stdCtx)
//			if err != nil {
//				logger.L.Warn("WrapEvent: get chat failed", "error", err)
//			} else {
//				handlerCtx.chat = &chat
//			}
//		}
//
//		var senderMember *model.ChatMember
//		if ctx.Data != nil && ctx.Data["_cached_sender"] != nil {
//			if cached, ok := ctx.Data["_cached_sender"].(*model.ChatMember); ok {
//				senderMember = cached
//			}
//		}
//		if senderMember == nil && ctx.EffectiveUser != nil {
//			member, err := c.resolveMember(stdCtx, handlerCtx.chat, ctx.EffectiveUser.Id)
//			if err != nil {
//				logger.L.Warn("WrapEvent: resolve sender failed", "error", err)
//			} else {
//				senderMember = member
//				if ctx.Data == nil {
//					ctx.Data = make(map[string]interface{})
//				}
//				ctx.Data["_cached_sender"] = senderMember
//			}
//		}
//		handlerCtx.senderChatMember = senderMember
//
//		var eventUsers []gotgbot.User
//		switch {
//		case len(msg.NewChatMembers) > 0:
//			eventUsers = msg.NewChatMembers
//		case msg.LeftChatMember != nil:
//			eventUsers = []gotgbot.User{*msg.LeftChatMember}
//		}
//
//		for _, u := range eventUsers {
//			if u.IsBot {
//				continue
//			}
//			m, err := c.resolveMember(stdCtx, handlerCtx.chat, u.Id)
//			if err != nil {
//				m = &model.ChatMember{
//					User: model.User{
//						ID:        u.Id,
//						Username:  u.Username,
//						FirstName: u.FirstName,
//						LastName:  u.LastName,
//					},
//				}
//			}
//			handlerCtx.chatMembers = append(handlerCtx.chatMembers, *m)
//		}
//
//		return c.response(b, handlerCtx)
//	}
//}
