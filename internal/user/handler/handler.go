package handler

import (
	"activity-bot/internal/action"
	"activity-bot/internal/cctx"
	"activity-bot/internal/command"
	"activity-bot/internal/emoji"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/rule"
	"activity-bot/internal/user"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"strings"

	"github.com/gotd/botapi"
)

const CategoryUser command.Category = "user"

type Handler struct {
	repo user.Repository
}

func NewHandler(repo user.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Actions() []*command.Action {
	return []*command.Action{
		action.NewCommand(
			"setemoji",
			h.SetEmoji,
			i18n.Cmd.User.SetEmoji.Desc,
			CategoryUser,
			option.WithAliases("мой эмоджи", "эмоджи", "эмодзи"),
			option.WithScope(command.ScopeAny),
			option.WithRules(rule.Text()),
		),
		action.NewCommand(
			"emoji",
			h.ShowEmoji,
			i18n.Cmd.User.ShowEmoji.Desc,
			CategoryUser,
			option.WithAliases("мой эмоджи", "эмоджи", "эмодзи"),
			option.WithScope(command.ScopeAny),
		),
	}
}

func (h *Handler) SetEmoji(c *botapi.Context) error {
	msg := cctx.MustArgsMessage(c)

	emojis := emoji.Extract(msg.OriginalTextHTML())

	if len(emojis) == 0 || len(emojis) > 3 {
		return nil
	}

	emojisString := strings.Join(emojis, "")

	cm := cctx.MustChatMember(c)

	if err := h.repo.SetEmoji(c, cm.ID(), emojisString); err != nil {
		return fmt.Errorf("set emoji: %w", err)
	}

	loc := cctx.MustLocalizer(c)

	_, err := c.Reply(loc.T(i18n.Cmd.User.SetEmoji.Success, i18n.CmdUserSetEmojiSuccessData{
		Emoji: emojisString,
		User:  tghtml.MemberMentionCustom(loc, cm, false),
	}), botapi.WithParseMode(botapi.ParseModeHTML))

	return err
}

func (h *Handler) ShowEmoji(c *botapi.Context) error {
	cm := cctx.MustChatMember(c)
	emojis := emoji.Extract(cm.User.Emojis)
	loc := cctx.MustLocalizer(c)

	if len(emojis) == 0 {
		_, err := c.Reply(loc.T(i18n.Cmd.User.ShowEmoji.NoEmoji, i18n.CmdUserShowEmojiNoEmojiData{
			User: tghtml.MemberMentionCustom(loc, cm, false),
		}), botapi.WithParseMode(botapi.ParseModeHTML))

		return err
	}

	_, err := c.Reply(loc.T(i18n.Cmd.User.ShowEmoji.Success, i18n.CmdUserShowEmojiSuccessData{
		User:  tghtml.MemberMentionCustom(loc, cm, false),
		Emoji: cm.User.Emojis,
	}), botapi.WithParseMode(botapi.ParseModeHTML))

	return err
}
