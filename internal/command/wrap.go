package command

import (
	"activity-bot/internal/logger"
	"activity-bot/internal/model"
	"context"
	"fmt"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func (c *Command) WrapCallback() func(b *gotgbot.Bot, ctx *ext.Context) error {
	return func(b *gotgbot.Bot, ctx *ext.Context) error {
		stdCtx := context.Background()

		cq := ctx.CallbackQuery
		if cq == nil {
			return nil
		}

		handlerCtx := &Context{Context: ctx, stdContext: stdCtx}

		if c.scope == ScopeChat {
			chat, err := c.getChat(ctx, stdCtx)
			if err != nil {
				logger.L.Warn("WrapCallback: get chat failed", "error", err)
			} else {
				handlerCtx.chat = &chat
			}
		}

		var senderMember *model.ChatMember
		if ctx.Data != nil && ctx.Data["_cached_sender"] != nil {
			if cached, ok := ctx.Data["_cached_sender"].(*model.ChatMember); ok {
				senderMember = cached
			}
		}
		if senderMember == nil && ctx.EffectiveUser != nil {
			member, err := c.resolveMember(stdCtx, handlerCtx.chat, ctx.EffectiveUser)
			if err != nil {
				logger.L.Warn("WrapCallback: resolve sender failed", "error", err)
			} else {
				senderMember = member
				if ctx.Data == nil {
					ctx.Data = make(map[string]interface{})
				}
				ctx.Data["_cached_sender"] = senderMember
			}
		}
		handlerCtx.senderChatMember = senderMember

		if c.requiredStatus > 0 && handlerCtx.senderChatMember != nil && !handlerCtx.senderChatMember.StatusGranted(c.requiredStatus) {
			_, err := cq.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
				Text: fmt.Sprintf("[%d/%d] Недостаточно прав для выполнения команды", handlerCtx.senderChatMember.Status, c.requiredStatus),
			})
			return err
		}

		data := cq.Data
		if idx := strings.Index(data, ":"); idx != -1 {
			handlerCtx.RawArgs = data[idx+1:]
		}
		handlerCtx.texts = strings.Fields(handlerCtx.RawArgs)

		if err := c.response(b, handlerCtx); err != nil {
			return err
		}

		_, _ = cq.Answer(b, nil)
		return nil
	}
}

func (c *Command) WrapEvent() func(b *gotgbot.Bot, ctx *ext.Context) error {
	return func(b *gotgbot.Bot, ctx *ext.Context) error {
		stdCtx := context.Background()
		msg := ctx.Message
		if msg == nil {
			return nil
		}

		handlerCtx := &Context{Context: ctx, stdContext: stdCtx}

		if c.scope == ScopeChat {
			chat, err := c.getChat(ctx, stdCtx)
			if err != nil {
				logger.L.Warn("WrapEvent: get chat failed", "error", err)
			} else {
				handlerCtx.chat = &chat
			}
		}

		var senderMember *model.ChatMember
		if ctx.Data != nil && ctx.Data["_cached_sender"] != nil {
			if cached, ok := ctx.Data["_cached_sender"].(*model.ChatMember); ok {
				senderMember = cached
			}
		}
		if senderMember == nil && ctx.EffectiveUser != nil {
			member, err := c.resolveMember(stdCtx, handlerCtx.chat, ctx.EffectiveUser)
			if err != nil {
				logger.L.Warn("WrapEvent: resolve sender failed", "error", err)
			} else {
				senderMember = member
				if ctx.Data == nil {
					ctx.Data = make(map[string]interface{})
				}
				ctx.Data["_cached_sender"] = senderMember
			}
		}
		handlerCtx.senderChatMember = senderMember

		var eventUsers []gotgbot.User
		switch {
		case len(msg.NewChatMembers) > 0:
			eventUsers = msg.NewChatMembers
		case msg.LeftChatMember != nil:
			eventUsers = []gotgbot.User{*msg.LeftChatMember}
		}

		for _, u := range eventUsers {
			if u.IsBot {
				continue
			}
			m, err := c.resolveMember(stdCtx, handlerCtx.chat, &u)
			if err != nil {
				m = &model.ChatMember{
					User: model.User{
						ID:        u.Id,
						Username:  u.Username,
						FirstName: u.FirstName,
						LastName:  u.LastName,
					},
				}
			}
			handlerCtx.chatMembers = append(handlerCtx.chatMembers, *m)
		}

		return c.response(b, handlerCtx)
	}
}
