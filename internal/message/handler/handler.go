package handler

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/command"
	"activity-bot/internal/member"
	"activity-bot/internal/message"
	"activity-bot/internal/options"
	"fmt"

	"github.com/cohesion-org/deepseek-go"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
)

type Handler struct {
	service        *message.Service
	memberService  *member.Service
	chatService    *chat.Service
	deepseekClient *deepseek.Client
}

func New(service *message.Service, memberService *member.Service, chatService *chat.Service, deepseekClient *deepseek.Client) *Handler {
	return &Handler{service, memberService, chatService, deepseekClient}
}

func (h *Handler) Bot(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return fmt.Errorf("bot: get chat: %w", err)
	}
	request := &deepseek.ChatCompletionRequest{
		Model: deepseek.DeepSeekChat,
		Messages: []deepseek.ChatCompletionMessage{
			{
				Role:    deepseek.ChatMessageRoleSystem,
				Content: "Отвечай коротко, 1-2 предложения. " + c.AISystemPrompt,
			},
			{
				Role:    deepseek.ChatMessageRoleUser,
				Content: ctx.TextOrDefault(""),
			},
		},
	}
	resp, err := h.deepseekClient.CreateChatCompletion(ctx.StdContext(), request)
	if err != nil {
		_ = ctx.ReplyOnly(u, options.WithText("Не удалось получить ответ от бота"))
		return fmt.Errorf("bot: create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		_ = ctx.ReplyOnly(u, options.WithText("Бот не вернул ответ"))
		return nil
	}

	return ctx.ReplyOnly(u, options.WithText(resp.Choices[0].Message.Content))
}

func (h *Handler) Message(ctx *command.Context, u *ext.Update) error {
	msg := u.EffectiveMessage
	if msg.Action != nil {
		switch msg.Action.(type) {
		case *tg.MessageActionChatJoinedByLink, *tg.MessageActionChatAddUser, *tg.MessageActionChatDeleteUser:
			return nil
		}
	}

	effectiveSender := u.EffectiveUser()
	effectiveChat := u.EffectiveChat()
	if effectiveSender == nil || effectiveChat == nil {
		return nil
	}
	if _, err := h.memberService.EnsureMemberExists(
		ctx.StdContext(),
		effectiveChat.GetID(),
		effectiveSender.GetID(),
		effectiveSender.Username,
		effectiveSender.FirstName,
		effectiveSender.LastName,
		"",
	); err != nil {
		return fmt.Errorf("message: ensure member exists: %w", err)
	}

	return h.service.Save(
		ctx.StdContext(),
		effectiveChat.GetID(),
		effectiveSender.GetID(),
		int64(msg.ID),
	)
}
