package fun

import (
	"activity-bot/internal/action"
	"activity-bot/internal/cctx"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/predicate"
	"activity-bot/internal/rule"
	"activity-bot/internal/utils/tghtml"
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/cohesion-org/deepseek-go"

	"github.com/gotd/botapi"
)

const CategoryFun command.Category = "fun"

type Handler struct {
	client            *deepseek.Client
	chatMemberService *chatmember.Service
}

func NewHandler(
	client *deepseek.Client,
	chatMemberService *chatmember.Service,
) *Handler {
	return &Handler{
		client:            client,
		chatMemberService: chatMemberService,
	}
}

func (h *Handler) Actions() []*command.Action {
	return []*command.Action{
		action.NewCommand(
			"ship",
			h.Ship,
			i18n.Cmd.Ship.Desc,
			CategoryFun,
			option.WithAliases("шипперим рандом", "шипперим"),
		),
		action.NewCommand(
			"fakeleave",
			h.FakeLeave,
			i18n.Cmd.FakeLeave.Desc,
			CategoryFun,
			option.WithAliases("фейклив"),
		),
		action.NewCommand(
			"ai",
			h.AI,
			i18n.Cmd.Ai.Desc,
			CategoryFun,
			option.WithAliases("крис"),
			option.WithRules(rule.Text().Validate(func(s string) bool {
				return len([]rune(s)) <= 1024
			})),
		),
		action.NewCommand("whois",
			h.WhoIsRandom,
			i18n.Cmd.WhoIs.Desc,
			CategoryFun,
			option.WithAliases("кто"),
			option.WithRules(rule.Text().Validate(func(s string) bool {
				return len([]rune(s)) <= 256
			})),
			option.WithPredicates(predicate.SensitiveCommand()),
		),
	}
}

func (h *Handler) Ship(c *botapi.Context) error {
	loc := cctx.MustLocalizer(c)
	ch := cctx.MustChat(c)

	members, err := h.chatMemberService.ListHumanPresentChatMembers(c, ch.ID)
	if err != nil {
		return fmt.Errorf("ship: list members: %w", err)
	}

	if len(members) < 2 {
		_, err = c.Reply(loc.T(i18n.Cmd.Ship.None, nil))
		return err
	}

	rand.Shuffle(len(members), func(i, j int) {
		members[i], members[j] = members[j], members[i]
	})

	first := members[0]
	second := members[1]

	firstMention := tghtml.MemberLink(loc, ch, first)
	secondMention := tghtml.MemberLink(loc, ch, second)

	var msg i18n.MessageID

	switch {
	case first.User.ID == second.User.ID:
		msg = randomShipMessage(
			i18n.Cmd.Ship.Self1,
			i18n.Cmd.Ship.Self2,
			i18n.Cmd.Ship.Self3,
			i18n.Cmd.Ship.Self4,
			i18n.Cmd.Ship.Self5,
			i18n.Cmd.Ship.Self6,
			i18n.Cmd.Ship.Self7,
			i18n.Cmd.Ship.Self8,
		)

	case first.User.IsBot && second.User.IsBot:
		msg = randomShipMessage(
			i18n.Cmd.Ship.BotBot1,
			i18n.Cmd.Ship.BotBot2,
			i18n.Cmd.Ship.BotBot3,
			i18n.Cmd.Ship.BotBot4,
			i18n.Cmd.Ship.BotBot5,
			i18n.Cmd.Ship.BotBot6,
			i18n.Cmd.Ship.BotBot7,
			i18n.Cmd.Ship.BotBot8,
		)

	case first.User.IsBot || second.User.IsBot:
		msg = randomShipMessage(
			i18n.Cmd.Ship.Bot1,
			i18n.Cmd.Ship.Bot2,
			i18n.Cmd.Ship.Bot3,
			i18n.Cmd.Ship.Bot4,
			i18n.Cmd.Ship.Bot5,
			i18n.Cmd.Ship.Bot6,
			i18n.Cmd.Ship.Bot7,
			i18n.Cmd.Ship.Bot8,
		)

	default:
		msg = randomShipMessage(
			i18n.Cmd.Ship.Normal1,
			i18n.Cmd.Ship.Normal2,
			i18n.Cmd.Ship.Normal3,
			i18n.Cmd.Ship.Normal4,
			i18n.Cmd.Ship.Normal5,
			i18n.Cmd.Ship.Normal6,
			i18n.Cmd.Ship.Normal7,
			i18n.Cmd.Ship.Normal8,
		)
	}

	_, err = c.Reply(
		loc.T(
			msg,
			i18n.CmdShipNormal1Data{
				First:  firstMention,
				Second: secondMention,
			},
		),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)

	return err
}

func randomShipMessage(ids ...i18n.MessageID) i18n.MessageID {
	return ids[rand.IntN(len(ids))]
}

func (h *Handler) FakeLeave(c *botapi.Context) error {
	msg := c.Message()
	if msg == nil {
		return nil
	}
	cm := cctx.MustChatMember(c)
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)
	chatID, _ := c.Chat()
	_ = c.Bot.DeleteMessage(c, chatID, msg.MessageID)
	text := loc.T(i18n.User.Left, i18n.UserLeftData{
		User: tghtml.MemberMention(loc, ch, cm),
	}, i18n.WithGender(cm.Gender()))

	if _, err := c.Bot.SendMessage(c, chatID, text, botapi.DisableWebPagePreview(), botapi.WithParseMode(botapi.ParseModeHTML)); err != nil {
		return err
	}
	return nil
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
		Temperature: 0.1,
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

func (h *Handler) WhoIsRandom(c *botapi.Context) error {
	loc := cctx.MustLocalizer(c)
	ch := cctx.MustChat(c)

	members, err := h.chatMemberService.ListHumanPresentChatMembers(c, ch.ID)
	if err != nil {
		return fmt.Errorf("who is random: %w", err)
	}
	if len(members) == 0 {
		return nil
	}

	m := randomMember(members)
	text, ok := cctx.MustArgs(c).Text()
	if !ok {
		return nil
	}

	keys := []i18n.MessageID{
		i18n.Cmd.WhoIs.Success1,
		i18n.Cmd.WhoIs.Success2,
		i18n.Cmd.WhoIs.Success3,
		i18n.Cmd.WhoIs.Success4,
		i18n.Cmd.WhoIs.Success5,
		i18n.Cmd.WhoIs.Success6,
		i18n.Cmd.WhoIs.Success7,
		i18n.Cmd.WhoIs.Success8,
	}
	key := keys[rand.IntN(len(keys))]

	_, err = c.Reply(loc.T(key, i18n.CmdWhoIsSuccess1Data{
		User: tghtml.MemberMention(loc, ch, m),
		Text: text,
	}), botapi.WithParseMode(botapi.ParseModeHTML))

	return err
}

func randomMember(m []chatmember.ChatMember) chatmember.ChatMember {
	return m[rand.IntN(len(m))]
}
