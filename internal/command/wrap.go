package command

import (
	"activity-bot/internal/logger"
	"fmt"
	"log"
	"strings"

	"github.com/celestix/gotgproto/dispatcher"
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
			return dispatcher.ContinueGroups
		}
		if filter != nil && !filter(u.CallbackQuery) {
			return dispatcher.ContinueGroups
		}

		cb := u.CallbackQuery
		if cb == nil {
			return dispatcher.ContinueGroups
		}

		handlerCtx := Context{
			Context: ctx,
		}
		handlerCtx.requiredStatus = c.requiredStatus

		if c.scope == ScopeChat {
			chat, err := c.getChat(ctx, u)
			if err != nil {
				return errors.Wrap(err, "callback: get chat failed")
			}
			handlerCtx.chat = &chat
			if s, err := c.chatProvider.GetCommandPermission(ctx.Context, chat.ID, c.name); err == nil {
				handlerCtx.requiredStatus = s
			}
		}
		member, err := c.resolveMember(ctx, handlerCtx.chat, u.EffectiveUser())
		if err != nil {
			return errors.Wrap(err, "callback: resolve sender failed")
		}
		handlerCtx.senderChatMember = member

		if !handlerCtx.senderChatMember.StatusGranted(handlerCtx.requiredStatus) {
			_, err = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
				Message: fmt.Sprintf("Требуются права: %s", handlerCtx.requiredStatus),
				Alert:   true,
				QueryID: cb.QueryID,
			})
			if err != nil {
				return errors.Wrap(err, "callback: answer failed")
			}
			return dispatcher.EndGroups
		}

		data := string(cb.Data)
		if idx := strings.Index(data, ":"); idx != -1 {
			handlerCtx.RawArgs = data[idx+1:]
		} else {
			handlerCtx.RawArgs = data
		}

		handlerCtx.texts = strings.Fields(handlerCtx.RawArgs)

		for _, middleware := range c.middlewares {
			if err := middleware.CheckUpdate(&handlerCtx, u); err != nil {
				if errors.Is(err, ErrStop) {
					return dispatcher.SkipCurrentGroup
				}
				logger.L.Error("middleware", "error", err)
				return dispatcher.SkipCurrentGroup
			}
		}

		if err = c.response(&handlerCtx, u); err != nil {
			return errors.Wrap(err, "callback: response failed")
		}

		return dispatcher.EndGroups
	}
}

func (c *Command) WrapEvent(filter filters.UpdateFilter) HandlerFunc {
	return func(ctx *ext.Context, u *ext.Update) error {
		if filter != nil && !filter(u) {
			return dispatcher.ContinueGroups
		}
		log.Println("event " + c.name)
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
			logger.L.Error("event: resolve sender failed", "error", err)
		}
		handlerCtx.senderChatMember = senderMember

		for _, middleware := range c.middlewares {
			if err := middleware.CheckUpdate(handlerCtx, u); err != nil {
				if errors.Is(err, ErrStop) {
					return dispatcher.SkipCurrentGroup
				}
				logger.L.Error("middleware", "error", err)
				return dispatcher.SkipCurrentGroup
			}
		}

		if err = c.response(handlerCtx, u); err != nil {
			return errors.Wrap(err, "event: response failed")
		}

		return nil
	}
}
