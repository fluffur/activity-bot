package command

import (
	"activity-bot/internal/logger"
	"fmt"
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

		logCommandHandled(c, u, handlerCtx.chat, cb.MsgID, string(cb.Data), handlerCtx.RawArgs, "callback", nil)
		return dispatcher.EndGroups
	}
}

func (c *Command) WrapEvent(filter filters.UpdateFilter) HandlerFunc {
	return func(ctx *ext.Context, u *ext.Update) error {
		if filter != nil && !filter(u) {
			return dispatcher.ContinueGroups
		}
		msg := u.EffectiveMessage
		if msg == nil {
			return nil
		}
		if msg.EditDate != 0 {
			return dispatcher.ContinueGroups
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

		for _, rule := range c.argRules {
			switch rule.Type {
			case ArgTypeOnlyUserSender:
				if _, ok := getReplyToMessageID(msg); ok {
					return dispatcher.ContinueGroups
				}
				members, _, err := c.extractMembersFromEntities(ctx, handlerCtx.chat, msg.Text, msg.Entities)
				if err != nil {
					return errors.Wrap(err, "failed to extract users")
				}
				if len(members) > 0 {
					return dispatcher.ContinueGroups
				}

			case ArgTypeAnyUser, ArgTypeMentionedUser:
				if err := c.resolveUsers(ctx, handlerCtx, msg, msg.Text, msg.Entities); err != nil {
					return fmt.Errorf("failed to resolve users event: %w", err)
				}
				if c.scope == ScopeChat && handlerCtx.chat != nil && len(handlerCtx.chatMembers) == 0 && handlerCtx.replyChatMember == nil {
					toks := freeTokens(handlerCtx.RawArgs, handlerCtx.usedOffsets)
					matched := false
					for i := 0; i < len(toks) && !matched; {
						for width := 3; width >= 1; width-- {
							if i+width > len(toks) {
								continue
							}
							words := make([]string, width)
							for k := 0; k < width; k++ {
								words[k] = toks[i+k].text
							}
							tag := strings.Join(words, " ")
							if len([]rune(tag)) <= 16 {
								members, err := c.chatMemberProvider.FindChatMembersByTag(ctx.Context, handlerCtx.chat.ID, tag)
								if err == nil && len(members) > 0 {
									handlerCtx.chatMembers = append(handlerCtx.chatMembers, members...)
									for k := 0; k < width; k++ {
										handlerCtx.usedOffsets = append(handlerCtx.usedOffsets, Offset{toks[i+k].start, toks[i+k].end})
									}
									matched = true
									break
								}
							}
						}
						if !matched {
							i++
						}
					}
				}

				if rule.Type == ArgTypeMentionedUser {
					totalUsers := handlerCtx.chatMembers
					if replyUser := handlerCtx.replyChatMember; replyUser != nil {
						totalUsers = append(totalUsers, *replyUser)
					}
					if len(totalUsers) < rule.Min {
						return dispatcher.ContinueGroups
					}
				}
			default:
			}

		}

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
