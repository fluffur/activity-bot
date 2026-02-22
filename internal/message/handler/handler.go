package handler

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/member"
	"activity-bot/internal/message"
	"context"
	"errors"
	"time"

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
	c, err := h.chatService.GetChat(ctx.StdContext(), ctx.EffectiveChat.Id)
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
				Content: ctx.FirstArgument(),
			},
		},
	}
	resp, err := h.deepseekClient.CreateChatCompletion(ctx.StdContext(), request)
	if err != nil {
		_ = ctx.Reply(b, "Не удалось получить ответ от бота", nil)
		return err
	}

	if len(resp.Choices) == 0 {
		_ = ctx.Reply(b, "Бот не вернул ответ", nil)
		return nil
	}

	return ctx.Reply(b, resp.Choices[0].Message.Content, &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeMarkdown,
	})
}

func (h *Handler) Message(b *gotgbot.Bot, ctx *ext.Context) error {
	cctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	u := ctx.EffectiveSender.User
	c := ctx.EffectiveChat
	if u == nil || c == nil || u.IsBot {
		return nil
	}

	m, err := h.memberService.EnsureMemberExists(cctx, c.Id, u.Id, u.Username, u.FirstName, u.LastName, "member")

	if err != nil {
		return err
	}

	if m.CustomTitle == "" {
		title, err := h.EnsureMemberCustomTitle(cctx, b, c.Id, u.Id)
		if err != nil {
			return err
		}
		if m.CustomTitle != title {
			if err := h.memberService.SetMemberTitle(cctx, c.Id, u.Id, &title); err != nil {
				return err
			}
		}
	}

	return h.service.Save(cctx, c.Id, u.Id, ctx.EffectiveMessage.MessageId)
}
