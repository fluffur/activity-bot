package ai

import (
	"activity-bot/internal/action"
	"activity-bot/internal/cctx"
	"activity-bot/internal/chat/handler"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/rule"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cohesion-org/deepseek-go"

	"github.com/gotd/botapi"
)

type Handler struct {
	client *deepseek.Client
}

func NewHandler(client *deepseek.Client) *Handler {
	return &Handler{client: client}
}

func (h *Handler) Actions() []*command.Action {
	return []*command.Action{
		action.NewCommand(
			"ai",
			h.AI,
			i18n.Cmd.Chat.Ai.Desc,
			handler.CategoryChat,
			option.WithAliases("крис"),
			option.WithRules(rule.Text().Validate(func(s string) bool {
				return len([]rune(s)) <= 1024
			})),
		),
	}
}

func (h *Handler) AI(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	text, ok := cctx.MustArgs(c).Text()
	if !ok {
		return nil
	}
	request := &deepseek.ChatCompletionRequest{
		Model: deepseek.DeepSeekV4Flash,
		Messages: []deepseek.ChatCompletionMessage{
			{
				Role:    deepseek.ChatMessageRoleSystem,
				Content: "Отвечай кратко. Максимум 5 предложений\n" + ch.AISystemPrompt,
			},
			{
				Role:    deepseek.ChatMessageRoleUser,
				Content: text,
			},
		},
		Thinking: &deepseek.ThinkingConfig{
			Type: "disabled",
		},
		MaxTokens:   128,
		Temperature: 1.0,
	}

	ctx, cancel := context.WithTimeout(c.Background(), 15*time.Second)
	defer cancel()

	resp, err := h.client.CreateChatCompletion(ctx, request)

	if err != nil {
		return fmt.Errorf("bot: create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)

	if content == "" {
		return nil
	}

	_, err = c.Reply(content, botapi.WithParseMode(botapi.ParseModeMarkdownV2))

	return err
}
