package command

import (
	"fmt"
	"strings"

	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
	"github.com/celestix/gotgproto/ext"
	"github.com/go-faster/errors"
	"github.com/gotd/td/tg"
)

type HandlerFunc func(ctx *ext.Context, u *ext.Update) error

func (f HandlerFunc) CheckUpdate(ctx *ext.Context, u *ext.Update) error {
	return f(ctx, u)
}

func (c *Command) WrapCallback(filter filters.CallbackQueryFilter) HandlerFunc {
	return func(ctx *ext.Context, u *ext.Update) error {
		if u.CallbackQuery == nil {
			return nil
		}
		if filter != nil && !filter(u.CallbackQuery) {
			return nil
		}

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
		member, err := c.resolveMember(ctx, handlerCtx.chat, u.EffectiveUser())
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

		//_, _ = ctx.AnswerCallback(nil)

		if err != nil {
			return err
		}

		return nil
	}
}

func (c *Command) WrapEvent() HandlerFunc {
	return func(ctx *ext.Context, u *ext.Update) error {
		msg := u.EffectiveMessage
		if msg == nil {
			return nil
		}

		handlerCtx := &Context{Context: ctx}

		if c.scope == ScopeChat {
			chat, err := c.getChat(ctx, u)
			if err != nil {
				return fmt.Errorf("wrap event failed to get chat: %w", err)
			}

			handlerCtx.chat = &chat
		}

		senderMember, err := c.resolveMember(ctx, handlerCtx.chat, u.EffectiveUser())
		if err != nil {
			return err
		}
		handlerCtx.senderChatMember = senderMember

		if err = c.response(handlerCtx, u); err != nil {
			return err
		}

		return nil
	}
}
