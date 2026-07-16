package handler

import (
	"activity-bot/internal/action"
	"activity-bot/internal/cctx"
	"activity-bot/internal/command"
	"activity-bot/internal/emoji"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/rp"
	"activity-bot/internal/rule"
	"activity-bot/internal/user"
	"activity-bot/internal/utils/tghtml"
	"errors"
	"fmt"
	"html"
	regexp "regexp"
	"strings"

	"github.com/gotd/botapi"
)

const CategoryRP = "rp"

type Handler struct {
	repo rp.Repository
}

func NewHandler(repo rp.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Actions() []*command.Action {
	return []*command.Action{
		action.NewCommand(
			"addrp",
			h.AddRPCommand,
			i18n.Cmd.AddRp.Desc,
			CategoryRP,
			option.WithAliases("рп", "+рп"),
			option.WithRules(rule.Text()),
			option.WithExamples(i18n.Cmd.AddRp.Example1, i18n.Cmd.AddRp.Example2),
		),
		action.NewRPCommand(
			"rp",
			h.HandleRPCommand,
			CategoryRP,
			option.WithRules(rule.User(), rule.Text().Optional()),
		),
		action.NewCommand(
			"listrp",
			h.ListRPCommand,
			i18n.Cmd.ListRp.Desc,
			CategoryRP,
			option.WithAliases("рп"),
		),
		action.NewCommand(
			"delrp",
			h.DeleteRPCommand,
			i18n.Cmd.DelRp.Desc,
			CategoryRP,
			option.WithRules(rule.Text()),
			option.WithAliases("-рп"),
		),
	}
}

var ErrInvalidFormat = errors.New("invalid format")

func ParseRP(html string) (rp.Definition, error) {
	parts := strings.SplitN(strings.TrimSpace(html), "\n", 2)
	if len(parts) != 2 {
		return rp.Definition{}, ErrInvalidFormat
	}

	trigger := strings.TrimSpace(parts[0])

	actionHTML := strings.TrimSpace(parts[1])

	emojis := emoji.Extract(actionHTML)

	act := actionHTML
	for _, e := range emojis {
		act = strings.ReplaceAll(act, e, "")
	}

	act = strings.TrimSpace(act)

	return rp.Definition{
		Trigger: trigger,
		Action:  act,
		Emojis:  strings.Join(emojis, ""),
	}, nil
}

func (h *Handler) AddRPCommand(c *botapi.Context) error {
	message := cctx.MustArgsMessage(c)

	cmd, err := ParseRP(message.OriginalTextHTML())
	if err != nil {
		return fmt.Errorf("add rp parse: %w", err)
	}

	ch := cctx.MustChat(c)
	cm := cctx.MustChatMember(c)
	cmd.ChatID = ch.ID
	cmd.CreatedBy = cm.ID()

	if err := h.repo.Upsert(c, cmd); err != nil {
		return fmt.Errorf("add rp upsert: %w", err)
	}

	_, err = c.Reply(
		fmt.Sprintf("Команда успешно добавлена.\nПопробуйте написать <code>%s @%s</code>", cmd.Trigger, c.Bot.Self().Username),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)
	return err
}

var genderRegex = regexp.MustCompile(`([^\s|]+)\|([^\s|]+)`)

func Genderize(s string, gender user.Gender) string {
	return genderRegex.ReplaceAllStringFunc(s, func(match string) string {
		parts := strings.SplitN(match, "|", 2)
		if len(parts) != 2 {
			return match
		}

		if gender == user.GenderFemale {
			return parts[1]
		}
		return parts[0]
	})
}

func (h *Handler) HandleRPCommand(c *botapi.Context) error {
	rpCmd := cctx.MustRPCommand(c)
	parsed := cctx.MustArgs(c)
	var extra, speech string
	if len(parsed.Texts) > 0 {
		text := parsed.Texts[0]

		parts := strings.SplitN(text, "\n", 2)

		extra = strings.TrimSpace(parts[0])

		if len(parts) == 2 {
			speech = strings.TrimSpace(parts[1])
		}
	}

	sender := cctx.MustChatMember(c)
	target, ok := cctx.MustArgs(c).AnyUser()
	if !ok {
		return nil
	}

	act := Genderize(rpCmd.Action, sender.Gender())
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	var b strings.Builder

	if rpCmd.Emojis != "" {
		b.WriteString(rpCmd.Emojis)
		b.WriteString(" | ")
	}
	b.WriteString(tghtml.MemberMention(loc, ch, sender))
	b.WriteByte(' ')
	b.WriteString(act)
	b.WriteByte(' ')
	b.WriteString(tghtml.MemberMention(loc, ch, target))

	if extra != "" {
		b.WriteByte(' ')
		b.WriteString(extra)
	}

	if speech != "" {
		b.WriteByte('\n')
		b.WriteString(loc.T(i18n.Cmd.Rp.Speech, i18n.CmdRpSpeechData{
			Text: html.EscapeString(speech),
		}))
	}

	_, err := c.Reply(
		b.String(),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}

func (h *Handler) ListRPCommand(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)
	list, err := h.repo.List(c, ch.ID)
	if err != nil {
		return fmt.Errorf("list rp: %w", err)
	}

	if len(list) == 0 {
		_, err := c.Reply(loc.T(i18n.Cmd.ListRp.Empty, nil))
		return err
	}

	var b strings.Builder
	b.WriteString(loc.T(i18n.Cmd.ListRp.Title, nil))
	b.WriteString("\n\n")
	for _, item := range list {
		b.WriteString(fmt.Sprintf("• <code>%s</code>\n", item.Trigger))
	}

	_, err = c.Reply(b.String(), botapi.WithParseMode(botapi.ParseModeHTML))
	return err
}

func (h *Handler) DeleteRPCommand(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	parsed := cctx.MustArgs(c)

	trigger, ok := parsed.Text()
	if !ok {
		return nil
	}

	if err := h.repo.Delete(c, ch.ID, trigger); err != nil {
		return fmt.Errorf("delete rp: %w", err)
	}
	loc := cctx.MustLocalizer(c)
	_, err := c.Reply(loc.T(i18n.Cmd.DelRp.Success, i18n.CmdDelRpSuccessData{
		Trigger: trigger,
	}), botapi.WithParseMode(botapi.ParseModeHTML))
	return err
}
