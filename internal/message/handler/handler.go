package handler

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/member"
	"activity-bot/internal/message"
	"context"
	"errors"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/cohesion-org/deepseek-go"
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

func (h *Handler) EnsureMemberCustomTitle(ctx context.Context, b *gotgbot.Bot, chatID, userID int64) (string, error) {
	m, err := h.memberService.GetMemberTitle(ctx, chatID, userID)
	if err != nil && !errors.Is(err, member.ErrInvalidCustomTitle) {
		return "", err
	}

	if m != "" {
		return m, nil
	}

	chatMember, err := b.GetChatMember(chatID, userID, nil)
	if err != nil {
		return "", err
	}

	return chatMember.MergeChatMember().CustomTitle, nil
}

func (h *Handler) Bot(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.chatService.GetChat(context.Background(), ctx.EffectiveChat.Id)
	if err != nil {
		return err
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
				Content: ctx.EffectiveSender.FirstName() + ": " + ctx.FirstArgument(),
			},
		},
	}
	resp, err := h.deepseekClient.CreateChatCompletion(context.Background(), request)
	if err != nil {
		_, _ = ctx.EffectiveMessage.Reply(b, "Не удалось получить ответ от бота", nil)
		return err
	}

	if len(resp.Choices) == 0 {
		_, _ = ctx.EffectiveMessage.Reply(b, "Бот не вернул ответ", nil)
		return nil
	}

	_, err = ctx.EffectiveMessage.Reply(b, resp.Choices[0].Message.Content, nil)
	return err
}

func (h *Handler) Message(b *gotgbot.Bot, ctx *ext.Context) error {
	u := ctx.EffectiveSender.User
	c := ctx.EffectiveChat
	if u == nil || c == nil || u.IsBot {
		return nil
	}

	m, err := h.memberService.EnsureMemberExists(context.Background(), c.Id, u.Id, u.Username, u.FirstName, u.LastName, "member")

	if err != nil {
		return err
	}

	if m.CustomTitle == "" {
		title, err := h.EnsureMemberCustomTitle(context.Background(), b, c.Id, u.Id)
		if err != nil {
			return err
		}
		if m.CustomTitle != title {
			if err := h.memberService.SetMemberTitle(context.Background(), c.Id, u.Id, &title); err != nil {
				return err
			}
		}
	}

	if err := h.service.Save(context.Background(), c.Id, u.Id); err != nil {
		return err
	}
	return nil
}
