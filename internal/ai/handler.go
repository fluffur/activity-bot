package ai

import (
	"activity-bot/internal/action"
	"activity-bot/internal/cctx"
	"activity-bot/internal/chat/handler"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/rule"
	"fmt"
	"strings"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/gotd/log"

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
		Model: deepseek.DeepSeekChat,
		Messages: []deepseek.ChatCompletionMessage{
			{
				Role:    deepseek.ChatMessageRoleSystem,
				Content: ch.AISystemPrompt,
			},
			{
				Role:    deepseek.ChatMessageRoleUser,
				Content: text,
			},
		},
		MaxTokens:   128,
		Temperature: 0.5,
	}
	start := time.Now()
	resp, err := h.client.CreateChatCompletion(c, request)
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

	log.For(c.Bot.Logger()).Info(c, "ai request",
		log.Duration("elapsed", time.Since(start)),
		log.Int("prompt_tokens", resp.Usage.PromptTokens),
		log.Int("completion_tokens", resp.Usage.CompletionTokens),
	)

	_, err = c.Reply(content, botapi.WithParseMode(botapi.ParseModeMarkdownV2))

	return err
}
